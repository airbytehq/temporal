package workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	enumspb "go.temporal.io/api/enums/v1"
	persistencespb "go.temporal.io/server/api/persistence/v1"
	"go.temporal.io/server/common"
	"go.temporal.io/server/common/definition"
	"go.temporal.io/server/common/primitives/timestamp"
	historyi "go.temporal.io/server/service/history/interfaces"
	"go.temporal.io/server/service/history/tasks"
	"go.temporal.io/server/service/history/tests"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	timerSequenceSuite struct {
		suite.Suite
		*require.Assertions

		controller       *gomock.Controller
		mockMutableState *historyi.MockMutableState

		workflowKey   definition.WorkflowKey
		timerSequence *timerSequenceImpl
	}
)

func TestTimerSequenceSuite(t *testing.T) {
	s := new(timerSequenceSuite)
	suite.Run(t, s)
}

func (s *timerSequenceSuite) SetupSuite() {

}

func (s *timerSequenceSuite) TearDownSuite() {

}

func (s *timerSequenceSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockMutableState = historyi.NewMockMutableState(s.controller)

	s.workflowKey = definition.NewWorkflowKey(
		tests.NamespaceID.String(),
		tests.WorkflowID,
		tests.RunID,
	)
	s.mockMutableState.EXPECT().GetWorkflowKey().Return(s.workflowKey).AnyTimes()
	s.timerSequence = NewTimerSequence(s.mockMutableState)
}

func (s *timerSequenceSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *timerSequenceSuite) TestCreateNextUserTimer_AlreadyCreated_AfterWorkflowExpiry() {
	now := time.Now().UTC()
	timerExpiry := timestamppb.New(now.Add(100))
	timerInfo := &persistencespb.TimerInfo{
		Version:        123,
		TimerId:        "some random timer ID",
		StartedEventId: 456,
		ExpiryTime:     timerExpiry,
		TaskStatus:     TimerTaskStatusCreated,
	}
	timerInfos := map[string]*persistencespb.TimerInfo{timerInfo.TimerId: timerInfo}
	s.mockMutableState.EXPECT().GetPendingTimerInfos().Return(timerInfos)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: timestamppb.New(timerExpiry.AsTime().Add(-1 * time.Second)),
	})

	modified, err := s.timerSequence.CreateNextUserTimer()
	s.NoError(err)
	s.False(modified)
}

func (s *timerSequenceSuite) TestCreateNextUserTimer_AlreadyCreated_BeforeWorkflowExpiry() {
	now := time.Now().UTC()
	timerExpiry := timestamppb.New(now.Add(100))
	timerInfo := &persistencespb.TimerInfo{
		Version:        123,
		TimerId:        "some random timer ID",
		StartedEventId: 456,
		ExpiryTime:     timerExpiry,
		TaskStatus:     TimerTaskStatusCreated,
	}
	timerInfos := map[string]*persistencespb.TimerInfo{timerInfo.TimerId: timerInfo}
	s.mockMutableState.EXPECT().GetPendingTimerInfos().Return(timerInfos)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: timestamppb.New(timerExpiry.AsTime().Add(1 * time.Second)),
	})

	modified, err := s.timerSequence.CreateNextUserTimer()
	s.NoError(err)
	s.False(modified)
}

func (s *timerSequenceSuite) TestCreateNextUserTimer_AlreadyCreated_NoWorkflowExpiry() {
	now := time.Now().UTC()
	timer1Expiry := timestamppb.New(now.Add(100))
	timerInfo := &persistencespb.TimerInfo{
		Version:        123,
		TimerId:        "some random timer ID",
		StartedEventId: 456,
		ExpiryTime:     timer1Expiry,
		TaskStatus:     TimerTaskStatusCreated,
	}
	timerInfos := map[string]*persistencespb.TimerInfo{timerInfo.TimerId: timerInfo}
	s.mockMutableState.EXPECT().GetPendingTimerInfos().Return(timerInfos)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: nil,
	})

	modified, err := s.timerSequence.CreateNextUserTimer()
	s.NoError(err)
	s.False(modified)
}

func (s *timerSequenceSuite) TestCreateNextUserTimer_NotCreated_AfterWorkflowExpiry() {
	now := time.Now().UTC()
	timerExpiry := timestamppb.New(now.Add(100))
	timerInfo := &persistencespb.TimerInfo{
		Version:        123,
		TimerId:        "some random timer ID",
		StartedEventId: 456,
		ExpiryTime:     timerExpiry,
		TaskStatus:     TimerTaskStatusNone,
	}
	timerInfos := map[string]*persistencespb.TimerInfo{timerInfo.TimerId: timerInfo}
	s.mockMutableState.EXPECT().GetPendingTimerInfos().Return(timerInfos)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: timestamppb.New(timerExpiry.AsTime().Add(-1 * time.Second)),
	})

	modified, err := s.timerSequence.CreateNextUserTimer()
	s.NoError(err)
	s.False(modified)
}

func (s *timerSequenceSuite) TestCreateNextUserTimer_NotCreated_BeforeWorkflowExpiry() {
	now := time.Now().UTC()
	timerExpiry := timestamppb.New(now.Add(100))
	timerInfo := &persistencespb.TimerInfo{
		Version:        123,
		TimerId:        "some random timer ID",
		StartedEventId: 456,
		ExpiryTime:     timerExpiry,
		TaskStatus:     TimerTaskStatusNone,
	}
	timerInfos := map[string]*persistencespb.TimerInfo{timerInfo.TimerId: timerInfo}
	s.mockMutableState.EXPECT().GetPendingTimerInfos().Return(timerInfos)
	s.mockMutableState.EXPECT().GetUserTimerInfoByEventID(timerInfo.StartedEventId).Return(timerInfo, true)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: timestamppb.New(timerExpiry.AsTime().Add(1 * time.Second)),
	})

	var timerInfoUpdated = common.CloneProto(timerInfo) // make a copy
	timerInfoUpdated.TaskStatus = TimerTaskStatusCreated
	s.mockMutableState.EXPECT().UpdateUserTimerTaskStatus(timerInfo.TimerId, timerInfoUpdated.TaskStatus).Return(nil)
	s.mockMutableState.EXPECT().AddTasks(&tasks.UserTimerTask{
		// TaskID is set by shard
		WorkflowKey:         s.workflowKey,
		VisibilityTimestamp: timerExpiry.AsTime(),
		EventID:             timerInfo.GetStartedEventId(),
	})

	modified, err := s.timerSequence.CreateNextUserTimer()
	s.NoError(err)
	s.True(modified)
}

func (s *timerSequenceSuite) TestCreateNextUserTimer_NotCreated_NoWorkflowExpiry() {
	now := time.Now().UTC()
	timerExpiry := timestamppb.New(now.Add(100))
	timerInfo := &persistencespb.TimerInfo{
		Version:        123,
		TimerId:        "some random timer ID",
		StartedEventId: 456,
		ExpiryTime:     timerExpiry,
		TaskStatus:     TimerTaskStatusNone,
	}
	timerInfos := map[string]*persistencespb.TimerInfo{timerInfo.TimerId: timerInfo}
	s.mockMutableState.EXPECT().GetPendingTimerInfos().Return(timerInfos)
	s.mockMutableState.EXPECT().GetUserTimerInfoByEventID(timerInfo.StartedEventId).Return(timerInfo, true)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: nil,
	})

	var timerInfoUpdated = common.CloneProto(timerInfo) // make a copy
	timerInfoUpdated.TaskStatus = TimerTaskStatusCreated
	s.mockMutableState.EXPECT().UpdateUserTimerTaskStatus(timerInfoUpdated.TimerId, timerInfoUpdated.TaskStatus).Return(nil)
	s.mockMutableState.EXPECT().AddTasks(&tasks.UserTimerTask{
		// TaskID is set by shard
		WorkflowKey:         s.workflowKey,
		VisibilityTimestamp: timerExpiry.AsTime(),
		EventID:             timerInfo.GetStartedEventId(),
	})

	modified, err := s.timerSequence.CreateNextUserTimer()
	s.NoError(err)
	s.True(modified)
}

func (s *timerSequenceSuite) TestCreateNextActivityTimer_AlreadyCreated_AfterWorkflowExpiry() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToClose | TimerTaskStatusCreatedScheduleToStart,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: timestamppb.New(now.Add(-2000 * time.Second)),
	})

	modified, err := s.timerSequence.CreateNextActivityTimer()
	s.NoError(err)
	s.False(modified)
}

func (s *timerSequenceSuite) TestCreateNextActivityTimer_AlreadyCreated_BeforeWorkflowExpiry() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToClose | TimerTaskStatusCreatedScheduleToStart,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: timestamppb.New(now.Add(2000 * time.Second)),
	})

	modified, err := s.timerSequence.CreateNextActivityTimer()
	s.NoError(err)
	s.False(modified)
}

func (s *timerSequenceSuite) TestCreateNextActivityTimer_AlreadyCreated_NoWorkflowExpiry() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToClose | TimerTaskStatusCreatedScheduleToStart,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: nil,
	})

	modified, err := s.timerSequence.CreateNextActivityTimer()
	s.NoError(err)
	s.False(modified)
}

func (s *timerSequenceSuite) TestCreateNextActivityTimer_NotCreated_AfterWorkflowExpiry() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: timestamppb.New(now.Add(-2000 * time.Second)),
	})

	modified, err := s.timerSequence.CreateNextActivityTimer()
	s.NoError(err)
	s.False(modified)
}

func (s *timerSequenceSuite) TestCreateNextActivityTimer_NotCreated_BeforeWorkflowExpiry() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)
	s.mockMutableState.EXPECT().GetActivityInfo(activityInfo.ScheduledEventId).Return(activityInfo, true)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: timestamppb.New(now.Add(2000 * time.Second)),
	})

	var activityInfoUpdated = common.CloneProto(activityInfo) // make a copy
	activityInfoUpdated.TimerTaskStatus = TimerTaskStatusCreatedScheduleToStart
	s.mockMutableState.EXPECT().UpdateActivityTaskStatusWithTimerHeartbeat(activityInfoUpdated.ScheduledEventId, activityInfoUpdated.TimerTaskStatus, nil).Return(nil)
	s.mockMutableState.EXPECT().AddTasks(&tasks.ActivityTimeoutTask{
		// TaskID is set by shard
		WorkflowKey:         s.workflowKey,
		VisibilityTimestamp: activityInfo.ScheduledTime.AsTime().Add(activityInfo.ScheduleToStartTimeout.AsDuration()),
		TimeoutType:         enumspb.TIMEOUT_TYPE_SCHEDULE_TO_START,
		EventID:             activityInfo.ScheduledEventId,
		Attempt:             activityInfo.Attempt,
	})

	modified, err := s.timerSequence.CreateNextActivityTimer()
	s.NoError(err)
	s.True(modified)
}

func (s *timerSequenceSuite) TestCreateNextActivityTimer_NotCreated_NoWorkflowExpiry() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)
	s.mockMutableState.EXPECT().GetActivityInfo(activityInfo.ScheduledEventId).Return(activityInfo, true)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: nil,
	})

	var activityInfoUpdated = common.CloneProto(activityInfo) // make a copy
	activityInfoUpdated.TimerTaskStatus = TimerTaskStatusCreatedScheduleToStart
	s.mockMutableState.EXPECT().UpdateActivityTaskStatusWithTimerHeartbeat(activityInfoUpdated.ScheduledEventId, activityInfoUpdated.TimerTaskStatus, nil).Return(nil)
	s.mockMutableState.EXPECT().AddTasks(&tasks.ActivityTimeoutTask{
		// TaskID is set by shard
		WorkflowKey:         s.workflowKey,
		VisibilityTimestamp: activityInfo.ScheduledTime.AsTime().Add(activityInfo.ScheduleToStartTimeout.AsDuration()),
		TimeoutType:         enumspb.TIMEOUT_TYPE_SCHEDULE_TO_START,
		EventID:             activityInfo.ScheduledEventId,
		Attempt:             activityInfo.Attempt,
	})

	modified, err := s.timerSequence.CreateNextActivityTimer()
	s.NoError(err)
	s.True(modified)
}

func (s *timerSequenceSuite) TestCreateNextActivityTimer_HeartbeatTimer_AfterWorkflowExpiry() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: timestamppb.New(now.Add(-2000 * time.Second)),
	})

	modified, err := s.timerSequence.CreateNextActivityTimer()
	s.NoError(err)
	s.False(modified)
}

func (s *timerSequenceSuite) TestCreateNextActivityTimer_HeartbeatTimer_BeforeWorkflowExpiry() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)
	s.mockMutableState.EXPECT().GetActivityInfo(activityInfo.ScheduledEventId).Return(activityInfo, true)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: timestamppb.New(now.Add(2000 * time.Second)),
	})

	taskVisibilityTimestamp := activityInfo.StartedTime.AsTime().Add(activityInfo.HeartbeatTimeout.AsDuration())

	var activityInfoUpdated = common.CloneProto(activityInfo) // make a copy
	activityInfoUpdated.TimerTaskStatus = TimerTaskStatusCreatedHeartbeat
	s.mockMutableState.EXPECT().UpdateActivityTaskStatusWithTimerHeartbeat(activityInfo.ScheduledEventId, activityInfoUpdated.TimerTaskStatus, &taskVisibilityTimestamp).Return(nil)
	s.mockMutableState.EXPECT().AddTasks(&tasks.ActivityTimeoutTask{
		// TaskID is set by shard
		WorkflowKey:         s.workflowKey,
		VisibilityTimestamp: taskVisibilityTimestamp,
		TimeoutType:         enumspb.TIMEOUT_TYPE_HEARTBEAT,
		EventID:             activityInfo.ScheduledEventId,
		Attempt:             activityInfo.Attempt,
	})

	modified, err := s.timerSequence.CreateNextActivityTimer()
	s.NoError(err)
	s.True(modified)
}

func (s *timerSequenceSuite) TestCreateNextActivityTimer_HeartbeatTimer_NoWorkflowExpiry() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)
	s.mockMutableState.EXPECT().GetActivityInfo(activityInfo.ScheduledEventId).Return(activityInfo, true)
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistencespb.WorkflowExecutionInfo{
		WorkflowRunExpirationTime: nil,
	})

	taskVisibilityTimestamp := activityInfo.StartedTime.AsTime().Add(activityInfo.HeartbeatTimeout.AsDuration())

	var activityInfoUpdated = common.CloneProto(activityInfo) // make a copy
	activityInfoUpdated.TimerTaskStatus = TimerTaskStatusCreatedHeartbeat
	s.mockMutableState.EXPECT().UpdateActivityTaskStatusWithTimerHeartbeat(activityInfo.ScheduledEventId, activityInfoUpdated.TimerTaskStatus, &taskVisibilityTimestamp).Return(nil)
	s.mockMutableState.EXPECT().AddTasks(&tasks.ActivityTimeoutTask{
		// TaskID is set by shard
		WorkflowKey:         s.workflowKey,
		VisibilityTimestamp: taskVisibilityTimestamp,
		TimeoutType:         enumspb.TIMEOUT_TYPE_HEARTBEAT,
		EventID:             activityInfo.ScheduledEventId,
		Attempt:             activityInfo.Attempt,
	})

	modified, err := s.timerSequence.CreateNextActivityTimer()
	s.NoError(err)
	s.True(modified)
}

func (s *timerSequenceSuite) TestLoadAndSortUserTimers_None() {
	timerInfos := map[string]*persistencespb.TimerInfo{}
	s.mockMutableState.EXPECT().GetPendingTimerInfos().Return(timerInfos)

	timerSequenceIDs := s.timerSequence.LoadAndSortUserTimers()
	s.Empty(timerSequenceIDs)
}

func (s *timerSequenceSuite) TestLoadAndSortUserTimers_One() {
	now := time.Now().UTC()
	timer1Expiry := timestamppb.New(now.Add(100))
	timerInfo := &persistencespb.TimerInfo{
		Version:        123,
		TimerId:        "some random timer ID",
		StartedEventId: 456,
		ExpiryTime:     timer1Expiry,
		TaskStatus:     TimerTaskStatusCreated,
	}
	timerInfos := map[string]*persistencespb.TimerInfo{timerInfo.TimerId: timerInfo}
	s.mockMutableState.EXPECT().GetPendingTimerInfos().Return(timerInfos)

	timerSequenceIDs := s.timerSequence.LoadAndSortUserTimers()
	s.Equal([]TimerSequenceID{{
		EventID:      timerInfo.GetStartedEventId(),
		Timestamp:    timer1Expiry.AsTime(),
		TimerType:    enumspb.TIMEOUT_TYPE_START_TO_CLOSE,
		TimerCreated: true,
		Attempt:      1,
	}}, timerSequenceIDs)
}

func (s *timerSequenceSuite) TestLoadAndSortUserTimers_Multiple() {
	now := time.Now().UTC()
	timer1Expiry := timestamppb.New(now.Add(100))
	timer2Expiry := timestamppb.New(now.Add(200))
	timerInfo1 := &persistencespb.TimerInfo{
		Version:        123,
		TimerId:        "some random timer ID",
		StartedEventId: 456,
		ExpiryTime:     timer1Expiry,
		TaskStatus:     TimerTaskStatusCreated,
	}
	timerInfo2 := &persistencespb.TimerInfo{
		Version:        1234,
		TimerId:        "other random timer ID",
		StartedEventId: 4567,
		ExpiryTime:     timestamppb.New(now.Add(200)),
		TaskStatus:     TimerTaskStatusNone,
	}
	timerInfos := map[string]*persistencespb.TimerInfo{
		timerInfo1.TimerId: timerInfo1,
		timerInfo2.TimerId: timerInfo2,
	}
	s.mockMutableState.EXPECT().GetPendingTimerInfos().Return(timerInfos)

	timerSequenceIDs := s.timerSequence.LoadAndSortUserTimers()
	s.Equal([]TimerSequenceID{
		{
			EventID:      timerInfo1.GetStartedEventId(),
			Timestamp:    timer1Expiry.AsTime(),
			TimerType:    enumspb.TIMEOUT_TYPE_START_TO_CLOSE,
			TimerCreated: true,
			Attempt:      1,
		},
		{
			EventID:      timerInfo2.GetStartedEventId(),
			Timestamp:    timer2Expiry.AsTime(),
			TimerType:    enumspb.TIMEOUT_TYPE_START_TO_CLOSE,
			TimerCreated: false,
			Attempt:      1,
		},
	}, timerSequenceIDs)
}

func (s *timerSequenceSuite) TestLoadAndSortActivityTimers_None() {
	activityInfos := map[int64]*persistencespb.ActivityInfo{}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)

	timerSequenceIDs := s.timerSequence.LoadAndSortActivityTimers()
	s.Empty(timerSequenceIDs)
}

func (s *timerSequenceSuite) TestLoadAndSortActivityTimers_One_NotScheduled() {
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        common.EmptyEventID,
		ScheduledTime:           nil,
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusNone,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)

	timerSequenceIDs := s.timerSequence.LoadAndSortActivityTimers()
	s.Empty(timerSequenceIDs)
}

func (s *timerSequenceSuite) TestLoadAndSortActivityTimers_One_Scheduled_NotStarted() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToClose | TimerTaskStatusCreatedScheduleToStart,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)

	timerSequenceIDs := s.timerSequence.LoadAndSortActivityTimers()
	s.Equal([]TimerSequenceID{
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.ScheduledTime.AsTime().Add(activityInfo.ScheduleToStartTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_START,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.ScheduledTime.AsTime().Add(activityInfo.ScheduleToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
	}, timerSequenceIDs)
}

func (s *timerSequenceSuite) TestLoadAndSortActivityTimers_One_Scheduled_Started_WithHeartbeatTimeout() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToClose | TimerTaskStatusCreatedStartToClose | TimerTaskStatusCreatedHeartbeat,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)

	timerSequenceIDs := s.timerSequence.LoadAndSortActivityTimers()
	s.Equal([]TimerSequenceID{
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.StartedTime.AsTime().Add(activityInfo.HeartbeatTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_HEARTBEAT,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.StartedTime.AsTime().Add(activityInfo.StartToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_START_TO_CLOSE,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.ScheduledTime.AsTime().Add(activityInfo.ScheduleToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
	}, timerSequenceIDs)
}

func (s *timerSequenceSuite) TestLoadAndSortActivityTimers_One_Scheduled_Started_WithoutHeartbeatTimeout() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToClose | TimerTaskStatusCreatedStartToClose,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)

	timerSequenceIDs := s.timerSequence.LoadAndSortActivityTimers()
	s.Equal([]TimerSequenceID{
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.StartedTime.AsTime().Add(activityInfo.StartToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_START_TO_CLOSE,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.ScheduledTime.AsTime().Add(activityInfo.ScheduleToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
	}, timerSequenceIDs)
}

func (s *timerSequenceSuite) TestLoadAndSortActivityTimers_One_Scheduled_Started_Heartbeated_WithHeartbeatTimeout() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToClose | TimerTaskStatusCreatedStartToClose | TimerTaskStatusCreatedHeartbeat,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)

	timerSequenceIDs := s.timerSequence.LoadAndSortActivityTimers()
	s.Equal([]TimerSequenceID{
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.LastHeartbeatUpdateTime.AsTime().Add(activityInfo.HeartbeatTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_HEARTBEAT,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.StartedTime.AsTime().Add(activityInfo.StartToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_START_TO_CLOSE,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.ScheduledTime.AsTime().Add(activityInfo.ScheduleToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
	}, timerSequenceIDs)
}

func (s *timerSequenceSuite) TestLoadAndSortActivityTimers_One_Scheduled_Started_Heartbeated_WithoutHeartbeatTimeout() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToClose | TimerTaskStatusCreatedStartToClose,
		Attempt:                 12,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)

	timerSequenceIDs := s.timerSequence.LoadAndSortActivityTimers()
	s.EqualValues([]TimerSequenceID{
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.StartedTime.AsTime().Add(activityInfo.StartToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_START_TO_CLOSE,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.ScheduledTime.AsTime().Add(activityInfo.ScheduleToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
	}, timerSequenceIDs)
}

func (s *timerSequenceSuite) TestLoadAndSortActivityTimers_Multiple() {
	now := time.Now().UTC()
	activityInfo1 := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}
	activityInfo2 := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        2345,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "other random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(11),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1001),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(101),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(6),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(800 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 21,
	}
	activityInfos := map[int64]*persistencespb.ActivityInfo{
		activityInfo1.ScheduledEventId: activityInfo1,
		activityInfo2.ScheduledEventId: activityInfo2,
	}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)

	timerSequenceIDs := s.timerSequence.LoadAndSortActivityTimers()
	s.Equal([]TimerSequenceID{
		{
			EventID:      activityInfo2.ScheduledEventId,
			Timestamp:    activityInfo2.ScheduledTime.AsTime().Add(activityInfo2.ScheduleToStartTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_START,
			TimerCreated: false,
			Attempt:      activityInfo2.Attempt,
		},
		{
			EventID:      activityInfo1.ScheduledEventId,
			Timestamp:    activityInfo1.StartedTime.AsTime().Add(activityInfo1.StartToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_START_TO_CLOSE,
			TimerCreated: false,
			Attempt:      activityInfo1.Attempt,
		},
		{
			EventID:      activityInfo1.ScheduledEventId,
			Timestamp:    activityInfo1.ScheduledTime.AsTime().Add(activityInfo1.ScheduleToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
			TimerCreated: false,
			Attempt:      activityInfo1.Attempt,
		},
		{
			EventID:      activityInfo2.ScheduledEventId,
			Timestamp:    activityInfo2.ScheduledTime.AsTime().Add(activityInfo2.ScheduleToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
			TimerCreated: false,
			Attempt:      activityInfo2.Attempt,
		},
	}, timerSequenceIDs)
}

func (s *timerSequenceSuite) TestGetUserTimerTimeout() {
	now := time.Now().UTC()
	timerExpiry := timestamppb.New(now.Add(100))
	timerInfo := &persistencespb.TimerInfo{
		Version:        123,
		TimerId:        "some random timer ID",
		StartedEventId: 456,
		ExpiryTime:     timerExpiry,
		TaskStatus:     TimerTaskStatusCreated,
	}

	expectedTimerSequence := &TimerSequenceID{
		EventID:      timerInfo.StartedEventId,
		Timestamp:    timerExpiry.AsTime(),
		TimerType:    enumspb.TIMEOUT_TYPE_START_TO_CLOSE,
		TimerCreated: true,
		Attempt:      1,
	}

	timerSequence := s.timerSequence.getUserTimerTimeout(timerInfo)
	s.Equal(expectedTimerSequence, timerSequence)

	timerInfo.TaskStatus = TimerTaskStatusNone
	expectedTimerSequence.TimerCreated = false
	timerSequence = s.timerSequence.getUserTimerTimeout(timerInfo)
	s.Equal(expectedTimerSequence, timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityScheduleToStartTimeout_WithTimeout_NotScheduled() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        common.EmptyEventID,
		ScheduledTime:           nil,
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityScheduleToStartTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityScheduleToStartTimeout_WithTimeout_Scheduled_NotStarted() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToStart,
		Attempt:                 12,
	}

	expectedTimerSequence := &TimerSequenceID{
		EventID:      activityInfo.ScheduledEventId,
		Timestamp:    activityInfo.ScheduledTime.AsTime().Add(activityInfo.ScheduleToStartTimeout.AsDuration()),
		TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_START,
		TimerCreated: true,
		Attempt:      12,
	}

	timerSequence := s.timerSequence.getActivityScheduleToStartTimeout(activityInfo)
	s.Equal(expectedTimerSequence, timerSequence)

	activityInfo.TimerTaskStatus = TimerTaskStatusNone
	expectedTimerSequence.TimerCreated = false
	timerSequence = s.timerSequence.getActivityScheduleToStartTimeout(activityInfo)
	s.Equal(expectedTimerSequence, timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityScheduleToStartTimeout_WithTimeout_Scheduled_Started() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Second)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToStart,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityScheduleToStartTimeout(activityInfo)
	s.Empty(timerSequence)

	activityInfo.TimerTaskStatus = TimerTaskStatusNone
	timerSequence = s.timerSequence.getActivityScheduleToStartTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityScheduleToStartTimeout_WithoutTimeout_NotScheduled() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        common.EmptyEventID,
		ScheduledTime:           nil,
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(0),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityScheduleToStartTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityScheduleToStartTimeout_WithoutTimeout_Scheduled_NotStarted() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(0),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToStart,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityScheduleToStartTimeout(activityInfo)
	s.Empty(timerSequence)

	activityInfo.TimerTaskStatus = TimerTaskStatusNone
	timerSequence = s.timerSequence.getActivityScheduleToStartTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityScheduleToStartTimeout_WithoutTimeout_Scheduled_Started() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Second)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(0),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToStart,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityScheduleToStartTimeout(activityInfo)
	s.Empty(timerSequence)

	activityInfo.TimerTaskStatus = TimerTaskStatusNone
	timerSequence = s.timerSequence.getActivityScheduleToStartTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityScheduleToCloseTimeout_WithTimeout_NotScheduled() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        common.EmptyEventID,
		ScheduledTime:           nil,
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityScheduleToCloseTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityScheduleToCloseTimeout_WithTimeout_Scheduled() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		FirstScheduledTime:      timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToClose,
		Attempt:                 12,
	}

	expectedTimerSequence := &TimerSequenceID{
		EventID:      activityInfo.ScheduledEventId,
		Timestamp:    activityInfo.ScheduledTime.AsTime().Add(activityInfo.ScheduleToCloseTimeout.AsDuration()),
		TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
		TimerCreated: true,
		Attempt:      12,
	}

	timerSequence := s.timerSequence.getActivityScheduleToCloseTimeout(activityInfo)
	s.Equal(expectedTimerSequence, timerSequence)

	activityInfo.TimerTaskStatus = TimerTaskStatusNone
	expectedTimerSequence.TimerCreated = false
	timerSequence = s.timerSequence.getActivityScheduleToCloseTimeout(activityInfo)
	s.Equal(expectedTimerSequence, timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityScheduleToCloseTimeout_WithoutTimeout_NotScheduled() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        common.EmptyEventID,
		ScheduledTime:           nil,
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(0),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityScheduleToCloseTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityScheduleToCloseTimeout_WithoutTimeout_Scheduled() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(0),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusCreatedScheduleToClose,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityScheduleToCloseTimeout(activityInfo)
	s.Empty(timerSequence)

	activityInfo.TimerTaskStatus = TimerTaskStatusNone
	timerSequence = s.timerSequence.getActivityScheduleToCloseTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityStartToCloseTimeout_WithTimeout_NotStarted() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityStartToCloseTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityStartToCloseTimeout_WithTimeout_Started() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusCreatedStartToClose,
		Attempt:                 12,
	}

	expectedTimerSequence := &TimerSequenceID{
		EventID:      activityInfo.ScheduledEventId,
		Timestamp:    activityInfo.StartedTime.AsTime().Add(activityInfo.StartToCloseTimeout.AsDuration()),
		TimerType:    enumspb.TIMEOUT_TYPE_START_TO_CLOSE,
		TimerCreated: true,
		Attempt:      12,
	}

	timerSequence := s.timerSequence.getActivityStartToCloseTimeout(activityInfo)
	s.Equal(expectedTimerSequence, timerSequence)

	activityInfo.TimerTaskStatus = TimerTaskStatusNone
	expectedTimerSequence.TimerCreated = false
	timerSequence = s.timerSequence.getActivityStartToCloseTimeout(activityInfo)
	s.Equal(expectedTimerSequence, timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityStartToCloseTimeout_WithoutTimeout_NotStarted() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(0),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityStartToCloseTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityStartToCloseTimeout_WithoutTimeout_Started() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(0),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusCreatedStartToClose,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityStartToCloseTimeout(activityInfo)
	s.Empty(timerSequence)

	activityInfo.TimerTaskStatus = TimerTaskStatusNone
	timerSequence = s.timerSequence.getActivityStartToCloseTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityHeartbeatTimeout_WithHeartbeat_NotStarted() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityHeartbeatTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityHeartbeatTimeout_WithHeartbeat_Started_NoHeartbeat() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusCreatedHeartbeat,
		Attempt:                 12,
	}

	expectedTimerSequence := &TimerSequenceID{
		EventID:      activityInfo.ScheduledEventId,
		Timestamp:    activityInfo.StartedTime.AsTime().Add(activityInfo.HeartbeatTimeout.AsDuration()),
		TimerType:    enumspb.TIMEOUT_TYPE_HEARTBEAT,
		TimerCreated: true,
		Attempt:      12,
	}

	timerSequence := s.timerSequence.getActivityHeartbeatTimeout(activityInfo)
	s.Equal(expectedTimerSequence, timerSequence)

	activityInfo.TimerTaskStatus = TimerTaskStatusNone
	expectedTimerSequence.TimerCreated = false
	timerSequence = s.timerSequence.getActivityHeartbeatTimeout(activityInfo)
	s.Equal(expectedTimerSequence, timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityHeartbeatTimeout_WithHeartbeat_Started_Heartbeated() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(1),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusCreatedHeartbeat,
		Attempt:                 12,
	}

	expectedTimerSequence := &TimerSequenceID{
		EventID:      activityInfo.ScheduledEventId,
		Timestamp:    activityInfo.LastHeartbeatUpdateTime.AsTime().Add(activityInfo.HeartbeatTimeout.AsDuration()),
		TimerType:    enumspb.TIMEOUT_TYPE_HEARTBEAT,
		TimerCreated: true,
		Attempt:      12,
	}

	timerSequence := s.timerSequence.getActivityHeartbeatTimeout(activityInfo)
	s.Equal(expectedTimerSequence, timerSequence)

	activityInfo.TimerTaskStatus = TimerTaskStatusNone
	expectedTimerSequence.TimerCreated = false
	timerSequence = s.timerSequence.getActivityHeartbeatTimeout(activityInfo)
	s.Equal(expectedTimerSequence, timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityHeartbeatTimeout_WithoutHeartbeat_NotStarted() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          common.EmptyEventID,
		StartedTime:             nil,
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusNone,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityHeartbeatTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityHeartbeatTimeout_WithoutHeartbeat_Started_NoHeartbeat() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: nil,
		TimerTaskStatus:         TimerTaskStatusCreatedHeartbeat,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityHeartbeatTimeout(activityInfo)
	s.Empty(timerSequence)

	activityInfo.TimerTaskStatus = TimerTaskStatusNone
	timerSequence = s.timerSequence.getActivityHeartbeatTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestGetActivityHeartbeatTimeout_WithoutHeartbeat_Started_Heartbeated() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		Version:                 123,
		ScheduledEventId:        234,
		ScheduledTime:           timestamppb.New(now),
		StartedEventId:          345,
		StartedTime:             timestamppb.New(now.Add(200 * time.Millisecond)),
		ActivityId:              "some random activity ID",
		ScheduleToStartTimeout:  timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout:  timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:     timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:        timestamp.DurationFromSeconds(0),
		LastHeartbeatUpdateTime: timestamppb.New(now.Add(400 * time.Millisecond)),
		TimerTaskStatus:         TimerTaskStatusCreatedHeartbeat,
		Attempt:                 12,
	}

	timerSequence := s.timerSequence.getActivityHeartbeatTimeout(activityInfo)
	s.Empty(timerSequence)

	activityInfo.TimerTaskStatus = TimerTaskStatusNone
	timerSequence = s.timerSequence.getActivityHeartbeatTimeout(activityInfo)
	s.Empty(timerSequence)
}

func (s *timerSequenceSuite) TestConversion() {
	s.Equal(int32(TimerTaskStatusCreatedStartToClose), timerTypeToTimerMask(enumspb.TIMEOUT_TYPE_START_TO_CLOSE))
	s.Equal(int32(TimerTaskStatusCreatedScheduleToStart), timerTypeToTimerMask(enumspb.TIMEOUT_TYPE_SCHEDULE_TO_START))
	s.Equal(int32(TimerTaskStatusCreatedScheduleToClose), timerTypeToTimerMask(enumspb.TIMEOUT_TYPE_SCHEDULE_TO_CLOSE))
	s.Equal(int32(TimerTaskStatusCreatedHeartbeat), timerTypeToTimerMask(enumspb.TIMEOUT_TYPE_HEARTBEAT))

	s.Equal(TimerTaskStatusNone, 0)
	s.Equal(TimerTaskStatusCreated, 1)
	s.Equal(TimerTaskStatusCreatedStartToClose, 1)
	s.Equal(TimerTaskStatusCreatedScheduleToStart, 2)
	s.Equal(TimerTaskStatusCreatedScheduleToClose, 4)
	s.Equal(TimerTaskStatusCreatedHeartbeat, 8)
}

func (s *timerSequenceSuite) TestLess_CompareTime() {
	now := time.Now().UTC()
	timerSequenceID1 := TimerSequenceID{
		EventID:      123,
		Timestamp:    now,
		TimerType:    enumspb.TIMEOUT_TYPE_HEARTBEAT,
		TimerCreated: true,
		Attempt:      12,
	}

	timerSequenceID2 := TimerSequenceID{
		EventID:      123,
		Timestamp:    now.Add(time.Second),
		TimerType:    enumspb.TIMEOUT_TYPE_HEARTBEAT,
		TimerCreated: true,
		Attempt:      12,
	}

	timerSequenceIDs := TimerSequenceIDs([]TimerSequenceID{timerSequenceID1, timerSequenceID2})
	s.True(timerSequenceIDs.Less(0, 1))
	s.False(timerSequenceIDs.Less(1, 0))
}

func (s *timerSequenceSuite) TestLess_CompareEventID() {
	now := time.Now().UTC()
	timerSequenceID1 := TimerSequenceID{
		EventID:      122,
		Timestamp:    now,
		TimerType:    enumspb.TIMEOUT_TYPE_HEARTBEAT,
		TimerCreated: true,
		Attempt:      12,
	}

	timerSequenceID2 := TimerSequenceID{
		EventID:      123,
		Timestamp:    now,
		TimerType:    enumspb.TIMEOUT_TYPE_HEARTBEAT,
		TimerCreated: true,
		Attempt:      12,
	}

	timerSequenceIDs := TimerSequenceIDs([]TimerSequenceID{timerSequenceID1, timerSequenceID2})
	s.True(timerSequenceIDs.Less(0, 1))
	s.False(timerSequenceIDs.Less(1, 0))
}

func (s *timerSequenceSuite) TestLess_CompareType() {
	now := time.Now().UTC()
	timerSequenceID1 := TimerSequenceID{
		EventID:      123,
		Timestamp:    now,
		TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
		TimerCreated: true,
		Attempt:      12,
	}

	timerSequenceID2 := TimerSequenceID{
		EventID:      123,
		Timestamp:    now,
		TimerType:    enumspb.TIMEOUT_TYPE_HEARTBEAT,
		TimerCreated: true,
		Attempt:      12,
	}

	timerSequenceIDs := TimerSequenceIDs([]TimerSequenceID{timerSequenceID1, timerSequenceID2})
	s.True(timerSequenceIDs.Less(0, 1))
	s.False(timerSequenceIDs.Less(1, 0))
}

func (s *timerSequenceSuite) TestLoadAndSortActivityTimers_FirstScheduledTime() {
	now := time.Now().UTC()
	activityInfo := &persistencespb.ActivityInfo{
		ScheduledEventId:       234,
		ScheduledTime:          timestamppb.New(now),
		ScheduleToStartTimeout: timestamp.DurationFromSeconds(10),
		ScheduleToCloseTimeout: timestamp.DurationFromSeconds(1000),
		StartToCloseTimeout:    timestamp.DurationFromSeconds(100),
		HeartbeatTimeout:       timestamp.DurationFromSeconds(1),
		TimerTaskStatus:        TimerTaskStatusCreatedScheduleToClose | TimerTaskStatusCreatedScheduleToStart,
		Attempt:                12,
	}
	activityInfo.FirstScheduledTime = timestamppb.New(now.Add(1 * time.Second))
	activityInfos := map[int64]*persistencespb.ActivityInfo{activityInfo.ScheduledEventId: activityInfo}
	s.mockMutableState.EXPECT().GetPendingActivityInfos().Return(activityInfos)

	timerSequenceIDs := s.timerSequence.LoadAndSortActivityTimers()
	s.Equal([]TimerSequenceID{
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.ScheduledTime.AsTime().Add(activityInfo.ScheduleToStartTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_START,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
		{
			EventID:      activityInfo.ScheduledEventId,
			Timestamp:    activityInfo.FirstScheduledTime.AsTime().Add(activityInfo.ScheduleToCloseTimeout.AsDuration()),
			TimerType:    enumspb.TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
			TimerCreated: true,
			Attempt:      activityInfo.Attempt,
		},
	}, timerSequenceIDs)
}
