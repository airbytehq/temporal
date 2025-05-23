package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/server/common"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/metrics"
	"go.temporal.io/server/common/namespace"
	"go.temporal.io/server/common/persistence"
	"go.uber.org/mock/gomock"
)

type (
	eventsCacheSuite struct {
		suite.Suite
		*require.Assertions

		controller           *gomock.Controller
		mockExecutionManager *persistence.MockExecutionManager

		logger log.Logger

		cache *CacheImpl
	}
)

func TestEventsCacheSuite(t *testing.T) {
	s := new(eventsCacheSuite)
	suite.Run(t, s)
}

func (s *eventsCacheSuite) SetupSuite() {

}

func (s *eventsCacheSuite) TearDownSuite() {

}

func (s *eventsCacheSuite) SetupTest() {
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockExecutionManager = persistence.NewMockExecutionManager(s.controller)

	s.logger = log.NewTestLogger()
	s.cache = s.newTestEventsCache()
}

func (s *eventsCacheSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *eventsCacheSuite) newTestEventsCache() *CacheImpl {
	return newEventsCache(s.mockExecutionManager,
		metrics.NoopMetricsHandler,
		s.logger,
		32,
		time.Minute,
		false)
}

func (s *eventsCacheSuite) TestEventsCacheHitSuccess() {
	namespaceID := namespace.ID("events-cache-hit-success-namespace")
	workflowID := "events-cache-hit-success-workflow-id"
	runID := "events-cache-hit-success-run-id"
	eventID := int64(23)
	shardID := int32(10)
	event := &historypb.HistoryEvent{
		EventId:    eventID,
		EventType:  enumspb.EVENT_TYPE_ACTIVITY_TASK_STARTED,
		Attributes: &historypb.HistoryEvent_ActivityTaskStartedEventAttributes{ActivityTaskStartedEventAttributes: &historypb.ActivityTaskStartedEventAttributes{}},
	}

	s.cache.PutEvent(
		EventKey{namespaceID, workflowID, runID, eventID, common.EmptyVersion},
		event)
	actualEvent, err := s.cache.GetEvent(
		context.Background(),
		shardID,
		EventKey{namespaceID, workflowID, runID, eventID, common.EmptyVersion},
		eventID, nil)
	s.Nil(err)
	s.Equal(event, actualEvent)
}

func (s *eventsCacheSuite) TestEventsCacheMissMultiEventsBatchV2Success() {
	namespaceID := namespace.ID("events-cache-miss-multi-events-batch-v2-success-namespace")
	workflowID := "events-cache-miss-multi-events-batch-v2-success-workflow-id"
	runID := "events-cache-miss-multi-events-batch-v2-success-run-id"
	event1 := &historypb.HistoryEvent{
		EventId:    11,
		EventType:  enumspb.EVENT_TYPE_WORKFLOW_TASK_COMPLETED,
		Attributes: &historypb.HistoryEvent_WorkflowTaskCompletedEventAttributes{WorkflowTaskCompletedEventAttributes: &historypb.WorkflowTaskCompletedEventAttributes{}},
	}
	event2 := &historypb.HistoryEvent{
		EventId:    12,
		EventType:  enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
		Attributes: &historypb.HistoryEvent_ActivityTaskScheduledEventAttributes{ActivityTaskScheduledEventAttributes: &historypb.ActivityTaskScheduledEventAttributes{}},
	}
	event3 := &historypb.HistoryEvent{
		EventId:    13,
		EventType:  enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
		Attributes: &historypb.HistoryEvent_ActivityTaskScheduledEventAttributes{ActivityTaskScheduledEventAttributes: &historypb.ActivityTaskScheduledEventAttributes{}},
	}
	event4 := &historypb.HistoryEvent{
		EventId:    14,
		EventType:  enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
		Attributes: &historypb.HistoryEvent_ActivityTaskScheduledEventAttributes{ActivityTaskScheduledEventAttributes: &historypb.ActivityTaskScheduledEventAttributes{}},
	}
	event5 := &historypb.HistoryEvent{
		EventId:    15,
		EventType:  enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
		Attributes: &historypb.HistoryEvent_ActivityTaskScheduledEventAttributes{ActivityTaskScheduledEventAttributes: &historypb.ActivityTaskScheduledEventAttributes{}},
	}
	event6 := &historypb.HistoryEvent{
		EventId:    16,
		EventType:  enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
		Attributes: &historypb.HistoryEvent_ActivityTaskScheduledEventAttributes{ActivityTaskScheduledEventAttributes: &historypb.ActivityTaskScheduledEventAttributes{}},
	}

	shardID := int32(10)
	s.mockExecutionManager.EXPECT().ReadHistoryBranch(gomock.Any(), &persistence.ReadHistoryBranchRequest{
		BranchToken:   []byte("store_token"),
		MinEventID:    event1.GetEventId(),
		MaxEventID:    event6.GetEventId() + 1,
		PageSize:      1,
		NextPageToken: nil,
		ShardID:       shardID,
	}).Return(&persistence.ReadHistoryBranchResponse{
		HistoryEvents: []*historypb.HistoryEvent{event1, event2, event3, event4, event5, event6},
		NextPageToken: nil,
	}, nil)

	s.cache.PutEvent(
		EventKey{namespaceID, workflowID, runID, event2.GetEventId(), common.EmptyVersion},
		event2)
	actualEvent, err := s.cache.GetEvent(
		context.Background(),
		shardID,
		EventKey{namespaceID, workflowID, runID, event6.GetEventId(), common.EmptyVersion},
		event1.GetEventId(), []byte("store_token"))
	s.Nil(err)
	s.Equal(event6, actualEvent)
}

func (s *eventsCacheSuite) TestEventsCacheMissV2Failure() {
	namespaceID := namespace.ID("events-cache-miss-failure-namespace")
	workflowID := "events-cache-miss-failure-workflow-id"
	runID := "events-cache-miss-failure-run-id"

	shardID := int32(10)
	expectedErr := errors.New("persistence call failed")
	s.mockExecutionManager.EXPECT().ReadHistoryBranch(gomock.Any(), &persistence.ReadHistoryBranchRequest{
		BranchToken:   []byte("store_token"),
		MinEventID:    int64(11),
		MaxEventID:    int64(15),
		PageSize:      1,
		NextPageToken: nil,
		ShardID:       shardID,
	}).Return(nil, expectedErr)

	actualEvent, err := s.cache.GetEvent(
		context.Background(),
		shardID,
		EventKey{namespaceID, workflowID, runID, int64(14), common.EmptyVersion},
		int64(11), []byte("store_token"))
	s.Nil(actualEvent)
	s.Equal(expectedErr, err)
}

func (s *eventsCacheSuite) TestEventsCacheDisableSuccess() {
	namespaceID := namespace.ID("events-cache-disable-success-namespace")
	workflowID := "events-cache-disable-success-workflow-id"
	runID := "events-cache-disable-success-run-id"
	event1 := &historypb.HistoryEvent{
		EventId:    23,
		EventType:  enumspb.EVENT_TYPE_ACTIVITY_TASK_STARTED,
		Attributes: &historypb.HistoryEvent_ActivityTaskStartedEventAttributes{ActivityTaskStartedEventAttributes: &historypb.ActivityTaskStartedEventAttributes{}},
	}
	event2 := &historypb.HistoryEvent{
		EventId:    32,
		EventType:  enumspb.EVENT_TYPE_ACTIVITY_TASK_STARTED,
		Attributes: &historypb.HistoryEvent_ActivityTaskStartedEventAttributes{ActivityTaskStartedEventAttributes: &historypb.ActivityTaskStartedEventAttributes{}},
	}

	shardID := int32(10)
	s.mockExecutionManager.EXPECT().ReadHistoryBranch(gomock.Any(), &persistence.ReadHistoryBranchRequest{
		BranchToken:   []byte("store_token"),
		MinEventID:    event2.GetEventId(),
		MaxEventID:    event2.GetEventId() + 1,
		PageSize:      1,
		NextPageToken: nil,
		ShardID:       shardID,
	}).Return(&persistence.ReadHistoryBranchResponse{
		HistoryEvents: []*historypb.HistoryEvent{event2},
		NextPageToken: nil,
	}, nil)

	s.cache.PutEvent(
		EventKey{namespaceID, workflowID, runID, event1.GetEventId(), common.EmptyVersion},
		event1)
	s.cache.PutEvent(
		EventKey{namespaceID, workflowID, runID, event2.GetEventId(), common.EmptyVersion},
		event2)
	s.cache.disabled = true
	actualEvent, err := s.cache.GetEvent(
		context.Background(),
		shardID,
		EventKey{namespaceID, workflowID, runID, event2.GetEventId(), common.EmptyVersion},
		event2.GetEventId(), []byte("store_token"))
	s.Nil(err)
	s.Equal(event2, actualEvent)
}

func (s *eventsCacheSuite) TestEventsCacheGetCachesResult() {
	namespaceID := namespace.ID("events-cache-get-caches-namespace")
	workflowID := "events-cache-get-caches-workflow-id"
	runID := "events-cache-get-caches-run-id"
	branchToken := []byte("store_token")

	shardID := int32(10)
	event1 := &historypb.HistoryEvent{
		EventId:   14,
		EventType: enumspb.EVENT_TYPE_ACTIVITY_TASK_STARTED,
	}
	s.mockExecutionManager.EXPECT().ReadHistoryBranch(gomock.Any(), &persistence.ReadHistoryBranchRequest{
		BranchToken:   branchToken,
		MinEventID:    int64(11),
		MaxEventID:    int64(15),
		PageSize:      1,
		NextPageToken: nil,
		ShardID:       shardID,
	}).Return(&persistence.ReadHistoryBranchResponse{
		HistoryEvents: []*historypb.HistoryEvent{event1},
		NextPageToken: nil,
	}, nil).Times(1) // will only be called once with two calls to GetEvent

	gotEvent1, _ := s.cache.GetEvent(
		context.Background(),
		shardID,
		EventKey{namespaceID, workflowID, runID, int64(14), common.EmptyVersion},
		int64(11), branchToken)
	s.Equal(gotEvent1, event1)
	gotEvent2, _ := s.cache.GetEvent(
		context.Background(),
		shardID,
		EventKey{namespaceID, workflowID, runID, int64(14), common.EmptyVersion},
		int64(11), branchToken)
	s.Equal(gotEvent2, event1)
}

func (s *eventsCacheSuite) TestEventsCacheInvalidKey() {
	namespaceID := namespace.ID("events-cache-invalid-key-namespace")
	workflowID := "events-cache-invalid-key-workflow-id"
	runID := "" // <-- this is invalid
	branchToken := []byte("store_token")

	shardID := int32(10)
	event1 := &historypb.HistoryEvent{
		EventId:   14,
		EventType: enumspb.EVENT_TYPE_ACTIVITY_TASK_STARTED,
	}
	s.mockExecutionManager.EXPECT().ReadHistoryBranch(gomock.Any(), &persistence.ReadHistoryBranchRequest{
		BranchToken:   branchToken,
		MinEventID:    int64(11),
		MaxEventID:    int64(15),
		PageSize:      1,
		NextPageToken: nil,
		ShardID:       shardID,
	}).Return(&persistence.ReadHistoryBranchResponse{
		HistoryEvents: []*historypb.HistoryEvent{event1},
		NextPageToken: nil,
	}, nil).Times(2) // will be called twice since the key is invalid

	s.cache.PutEvent(
		EventKey{namespaceID, workflowID, runID, event1.EventId, common.EmptyVersion},
		event1)

	gotEvent1, _ := s.cache.GetEvent(
		context.Background(),
		shardID,
		EventKey{namespaceID, workflowID, runID, int64(14), common.EmptyVersion},
		int64(11), branchToken)
	s.Equal(gotEvent1, event1)
	gotEvent2, _ := s.cache.GetEvent(
		context.Background(),
		shardID,
		EventKey{namespaceID, workflowID, runID, int64(14), common.EmptyVersion},
		int64(11), branchToken)
	s.Equal(gotEvent2, event1)
}
