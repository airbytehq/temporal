package replication

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	enumsspb "go.temporal.io/server/api/enums/v1"
	"go.temporal.io/server/api/historyservicemock/v1"
	persistencespb "go.temporal.io/server/api/persistence/v1"
	"go.temporal.io/server/client"
	"go.temporal.io/server/common"
	"go.temporal.io/server/common/cluster"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/metrics"
	"go.temporal.io/server/common/namespace"
	"go.temporal.io/server/common/persistence"
	"go.temporal.io/server/common/persistence/serialization"
	"go.temporal.io/server/service/history/configs"
	historyi "go.temporal.io/server/service/history/interfaces"
	"go.temporal.io/server/service/history/shard"
	"go.temporal.io/server/service/history/tasks"
	"go.temporal.io/server/service/history/tests"
	"go.uber.org/mock/gomock"
)

type (
	taskProcessorManagerSuite struct {
		suite.Suite
		*require.Assertions

		controller                        *gomock.Controller
		mockShard                         *historyi.MockShardContext
		mockEngine                        *historyi.MockEngine
		mockClientBean                    *client.MockBean
		mockClusterMetadata               *cluster.MockMetadata
		mockHistoryClient                 *historyservicemock.MockHistoryServiceClient
		mockReplicationTaskExecutor       *MockTaskExecutor
		mockReplicationTaskFetcherFactory *MockTaskFetcherFactory

		mockExecutionManager *persistence.MockExecutionManager

		shardID     int32
		shardOwner  string
		config      *configs.Config
		requestChan chan *replicationTaskRequest

		taskProcessorManager *taskProcessorManagerImpl
	}
)

func TestTaskProcessorManagerSuite(t *testing.T) {
	s := new(taskProcessorManagerSuite)
	suite.Run(t, s)
}

func (s *taskProcessorManagerSuite) SetupSuite() {
}

func (s *taskProcessorManagerSuite) TearDownSuite() {
}

func (s *taskProcessorManagerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.controller = gomock.NewController(s.T())

	s.config = tests.NewDynamicConfig()
	s.requestChan = make(chan *replicationTaskRequest, 10)

	s.shardID = rand.Int31()
	s.shardOwner = "test-shard-owner"
	s.mockShard = historyi.NewMockShardContext(s.controller)
	s.mockEngine = historyi.NewMockEngine(s.controller)
	s.mockClientBean = client.NewMockBean(s.controller)

	s.mockReplicationTaskExecutor = NewMockTaskExecutor(s.controller)
	s.mockHistoryClient = historyservicemock.NewMockHistoryServiceClient(s.controller)
	s.mockReplicationTaskFetcherFactory = NewMockTaskFetcherFactory(s.controller)
	serializer := serialization.NewSerializer()
	s.mockClusterMetadata = cluster.NewMockMetadata(s.controller)
	s.mockClientBean.EXPECT().GetHistoryClient().Return(s.mockHistoryClient).AnyTimes()
	s.mockShard.EXPECT().GetClusterMetadata().Return(s.mockClusterMetadata).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockShard.EXPECT().GetHistoryClient().Return(nil).AnyTimes()
	s.mockShard.EXPECT().GetNamespaceRegistry().Return(namespace.NewMockRegistry(s.controller)).AnyTimes()
	s.mockShard.EXPECT().GetConfig().Return(s.config).AnyTimes()
	s.mockShard.EXPECT().GetLogger().Return(log.NewNoopLogger()).AnyTimes()
	s.mockShard.EXPECT().GetMetricsHandler().Return(metrics.NoopMetricsHandler).AnyTimes()
	s.mockShard.EXPECT().GetPayloadSerializer().Return(serializer).AnyTimes()
	s.mockExecutionManager = persistence.NewMockExecutionManager(s.controller)
	s.mockShard.EXPECT().GetExecutionManager().Return(s.mockExecutionManager).AnyTimes()
	s.mockShard.EXPECT().GetShardID().Return(s.shardID).AnyTimes()
	s.mockShard.EXPECT().GetOwner().Return(s.shardOwner).AnyTimes()
	s.taskProcessorManager = NewTaskProcessorManager(
		s.config,
		s.mockShard,
		s.mockEngine,
		nil,
		nil,
		s.mockClientBean,
		serializer,
		s.mockReplicationTaskFetcherFactory,
		func(params TaskExecutorParams) TaskExecutor {
			return s.mockReplicationTaskExecutor
		},
		NewExecutionManagerDLQWriter(s.mockExecutionManager),
	)
}

func (s *taskProcessorManagerSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *taskProcessorManagerSuite) TestCleanupReplicationTask_Noop() {
	ackedTaskID := int64(12345)
	s.mockShard.EXPECT().GetQueueExclusiveHighReadWatermark(tasks.CategoryReplication).Return(tasks.NewImmediateKey(ackedTaskID + 2)).AnyTimes()
	s.mockShard.EXPECT().GetQueueState(tasks.CategoryReplication).Return(&persistencespb.QueueState{
		ExclusiveReaderHighWatermark: nil,
		ReaderStates: map[int64]*persistencespb.QueueReaderState{
			shard.ReplicationReaderIDFromClusterShardID(cluster.TestAlternativeClusterInitialFailoverVersion, s.shardID): {
				Scopes: []*persistencespb.QueueSliceScope{{
					Range: &persistencespb.QueueSliceRange{
						InclusiveMin: shard.ConvertToPersistenceTaskKey(
							tasks.NewImmediateKey(ackedTaskID + 1),
						),
						ExclusiveMax: shard.ConvertToPersistenceTaskKey(
							tasks.NewImmediateKey(math.MaxInt64),
						),
					},
					Predicate: &persistencespb.Predicate{
						PredicateType: enumsspb.PREDICATE_TYPE_UNIVERSAL,
						Attributes:    &persistencespb.Predicate_UniversalPredicateAttributes{},
					},
				}},
			},
		},
	}, true)

	s.taskProcessorManager.minTxAckedTaskID = ackedTaskID
	err := s.taskProcessorManager.cleanupReplicationTasks()
	s.NoError(err)
}

func (s *taskProcessorManagerSuite) TestCleanupReplicationTask_Cleanup() {
	ackedTaskID := int64(12345)
	s.mockShard.EXPECT().GetQueueExclusiveHighReadWatermark(tasks.CategoryReplication).Return(tasks.NewImmediateKey(ackedTaskID + 2)).AnyTimes()
	s.mockShard.EXPECT().GetQueueState(tasks.CategoryReplication).Return(&persistencespb.QueueState{
		ExclusiveReaderHighWatermark: nil,
		ReaderStates: map[int64]*persistencespb.QueueReaderState{
			shard.ReplicationReaderIDFromClusterShardID(cluster.TestAlternativeClusterInitialFailoverVersion, common.MapShardID(
				cluster.TestAllClusterInfo[cluster.TestCurrentClusterName].ShardCount,
				cluster.TestAllClusterInfo[cluster.TestAlternativeClusterName].ShardCount,
				s.shardID,
			)[0]): {
				Scopes: []*persistencespb.QueueSliceScope{{
					Range: &persistencespb.QueueSliceRange{
						InclusiveMin: shard.ConvertToPersistenceTaskKey(
							tasks.NewImmediateKey(ackedTaskID + 1),
						),
						ExclusiveMax: shard.ConvertToPersistenceTaskKey(
							tasks.NewImmediateKey(math.MaxInt64),
						),
					},
					Predicate: &persistencespb.Predicate{
						PredicateType: enumsspb.PREDICATE_TYPE_UNIVERSAL,
						Attributes:    &persistencespb.Predicate_UniversalPredicateAttributes{},
					},
				}},
			},
		},
	}, true)
	s.taskProcessorManager.minTxAckedTaskID = ackedTaskID - 1
	s.mockExecutionManager.EXPECT().RangeCompleteHistoryTasks(
		gomock.Any(),
		&persistence.RangeCompleteHistoryTasksRequest{
			ShardID:             s.shardID,
			TaskCategory:        tasks.CategoryReplication,
			ExclusiveMaxTaskKey: tasks.NewImmediateKey(ackedTaskID + 1),
		},
	).Return(nil).Times(1)
	err := s.taskProcessorManager.cleanupReplicationTasks()
	s.NoError(err)
}
