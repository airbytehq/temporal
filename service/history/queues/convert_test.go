package queues

import (
	"math/rand"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/api/temporalproto"
	enumsspb "go.temporal.io/server/api/enums/v1"
	"go.temporal.io/server/common/predicates"
	"go.temporal.io/server/service/history/tasks"
)

type (
	convertSuite struct {
		suite.Suite
		*require.Assertions
	}
)

func TestConvertSuite(t *testing.T) {
	s := new(convertSuite)
	suite.Run(t, s)
}

func (s *convertSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *convertSuite) TestConvertPredicate_All() {
	predicate := predicates.Universal[tasks.Task]()
	s.Equal(predicate, FromPersistencePredicate(ToPersistencePredicate(predicate)))
}

func (s *convertSuite) TestConvertPredicate_Empty() {
	predicate := predicates.Empty[tasks.Task]()
	s.Equal(predicate, FromPersistencePredicate(ToPersistencePredicate(predicate)))
}

func (s *convertSuite) TestConvertPredicate_And() {
	testCases := []tasks.Predicate{
		predicates.And(
			predicates.Universal[tasks.Task](),
			predicates.Empty[tasks.Task](),
		),
		predicates.And(
			predicates.Or[tasks.Task](
				tasks.NewNamespacePredicate([]string{uuid.New()}),
				tasks.NewNamespacePredicate([]string{uuid.New()}),
			),
			predicates.Or[tasks.Task](
				tasks.NewTypePredicate([]enumsspb.TaskType{
					enumsspb.TASK_TYPE_ACTIVITY_RETRY_TIMER,
				}),
				tasks.NewTypePredicate([]enumsspb.TaskType{
					enumsspb.TASK_TYPE_DELETE_HISTORY_EVENT,
				}),
			),
		),
		predicates.And(
			predicates.Not(predicates.Empty[tasks.Task]()),
			predicates.And[tasks.Task](
				tasks.NewNamespacePredicate([]string{uuid.New()}),
				tasks.NewNamespacePredicate([]string{uuid.New()}),
			),
		),
		predicates.And(
			predicates.Not(predicates.Empty[tasks.Task]()),
			predicates.And[tasks.Task](
				tasks.NewNamespacePredicate([]string{uuid.New()}),
				tasks.NewTypePredicate([]enumsspb.TaskType{
					enumsspb.TASK_TYPE_DELETE_HISTORY_EVENT,
				}),
			),
		),
	}

	for _, predicate := range testCases {
		s.Equal(predicate, FromPersistencePredicate(ToPersistencePredicate(predicate)))
	}
}

func (s *convertSuite) TestConvertPredicate_Or() {
	testCases := []tasks.Predicate{
		predicates.Or(
			predicates.Universal[tasks.Task](),
			predicates.Empty[tasks.Task](),
		),
		predicates.Or(
			predicates.And[tasks.Task](
				tasks.NewNamespacePredicate([]string{uuid.New()}),
				tasks.NewNamespacePredicate([]string{uuid.New()}),
			),
			predicates.And[tasks.Task](
				tasks.NewTypePredicate([]enumsspb.TaskType{
					enumsspb.TASK_TYPE_ACTIVITY_RETRY_TIMER,
				}),
				tasks.NewTypePredicate([]enumsspb.TaskType{
					enumsspb.TASK_TYPE_DELETE_HISTORY_EVENT,
				}),
			),
		),
		predicates.Or(
			predicates.Not(predicates.Empty[tasks.Task]()),
			predicates.And[tasks.Task](
				tasks.NewNamespacePredicate([]string{uuid.New()}),
				tasks.NewNamespacePredicate([]string{uuid.New()}),
			),
		),
		predicates.Or(
			predicates.Not(predicates.Empty[tasks.Task]()),
			predicates.And[tasks.Task](
				tasks.NewNamespacePredicate([]string{uuid.New()}),
				tasks.NewTypePredicate([]enumsspb.TaskType{
					enumsspb.TASK_TYPE_DELETE_HISTORY_EVENT,
				}),
			),
		),
	}

	for _, predicate := range testCases {
		s.Equal(predicate, FromPersistencePredicate(ToPersistencePredicate(predicate)))
	}
}

func (s *convertSuite) TestConvertPredicate_Not() {
	testCases := []tasks.Predicate{
		predicates.Not(predicates.Universal[tasks.Task]()),
		predicates.Not(predicates.Empty[tasks.Task]()),
		predicates.Not(predicates.And[tasks.Task](
			tasks.NewNamespacePredicate([]string{uuid.New()}),
			tasks.NewTypePredicate([]enumsspb.TaskType{}),
		)),
		predicates.Not(predicates.Or[tasks.Task](
			tasks.NewNamespacePredicate([]string{uuid.New()}),
			tasks.NewTypePredicate([]enumsspb.TaskType{}),
		)),
		predicates.Not(predicates.Not(predicates.Empty[tasks.Task]())),
		predicates.Not[tasks.Task](tasks.NewNamespacePredicate([]string{uuid.New()})),
		predicates.Not[tasks.Task](tasks.NewTypePredicate([]enumsspb.TaskType{
			enumsspb.TASK_TYPE_ACTIVITY_RETRY_TIMER,
		})),
	}

	for _, predicate := range testCases {
		s.Equal(predicate, FromPersistencePredicate(ToPersistencePredicate(predicate)))
	}
}

func (s *convertSuite) TestConvertPredicate_NamespaceID() {
	testCases := []tasks.Predicate{
		tasks.NewNamespacePredicate(nil),
		tasks.NewNamespacePredicate([]string{}),
		tasks.NewNamespacePredicate([]string{uuid.New(), uuid.New(), uuid.New()}),
	}

	for _, predicate := range testCases {
		s.Equal(predicate, FromPersistencePredicate(ToPersistencePredicate(predicate)))
	}
}

func (s *convertSuite) TestConvertPredicate_TaskType() {
	testCases := []tasks.Predicate{
		tasks.NewTypePredicate(nil),
		tasks.NewTypePredicate([]enumsspb.TaskType{}),
		tasks.NewTypePredicate([]enumsspb.TaskType{
			enumsspb.TASK_TYPE_ACTIVITY_RETRY_TIMER,
			enumsspb.TASK_TYPE_ACTIVITY_TIMEOUT,
			enumsspb.TASK_TYPE_DELETE_HISTORY_EVENT,
		}),
	}

	for _, predicate := range testCases {
		s.Equal(predicate, FromPersistencePredicate(ToPersistencePredicate(predicate)))
	}
}

func (s *convertSuite) TestConvertTaskKey() {
	key := NewRandomKey()
	s.Equal(key, FromPersistenceTaskKey(
		ToPersistenceTaskKey(key),
	))
}

func (s *convertSuite) TestConvertTaskRange() {
	r := NewRandomRange()
	s.Equal(r, FromPersistenceRange(
		ToPersistenceRange(r),
	))
}

func (s *convertSuite) TestConvertScope() {
	scope := NewScope(
		NewRandomRange(),
		tasks.NewNamespacePredicate([]string{uuid.New(), uuid.New()}),
	)

	s.True(temporalproto.DeepEqual(scope, FromPersistenceScope(
		ToPersistenceScope(scope),
	)))
}

func (s *convertSuite) TestConvertQueueState() {
	readerScopes := map[int64][]Scope{
		0: {},
		1: {
			NewScope(
				NewRandomRange(),
				tasks.NewNamespacePredicate([]string{uuid.New(), uuid.New()}),
			),
		},
		123: {
			NewScope(
				NewRandomRange(),
				tasks.NewNamespacePredicate([]string{uuid.New(), uuid.New()}),
			),
			NewScope(
				NewRandomRange(),
				tasks.NewTypePredicate([]enumsspb.TaskType{
					enumsspb.TASK_TYPE_ACTIVITY_TIMEOUT,
					enumsspb.TASK_TYPE_ACTIVITY_RETRY_TIMER,
				}),
			),
		},
	}

	queueState := &queueState{
		readerScopes:                 readerScopes,
		exclusiveReaderHighWatermark: tasks.NewKey(time.Unix(0, rand.Int63()).UTC(), 0),
	}

	s.True(temporalproto.DeepEqual(queueState, FromPersistenceQueueState(
		ToPersistenceQueueState(queueState),
	)))
}
