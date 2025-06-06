//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination activity_state_replicator_mock.go

package ndc

import (
	"context"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/serviceerror"
	enumsspb "go.temporal.io/server/api/enums/v1"
	historyspb "go.temporal.io/server/api/history/v1"
	"go.temporal.io/server/api/historyservice/v1"
	persistencespb "go.temporal.io/server/api/persistence/v1"
	"go.temporal.io/server/common"
	"go.temporal.io/server/common/definition"
	"go.temporal.io/server/common/locks"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/namespace"
	"go.temporal.io/server/common/persistence"
	"go.temporal.io/server/common/persistence/versionhistory"
	"go.temporal.io/server/common/primitives/timestamp"
	serviceerrors "go.temporal.io/server/common/serviceerror"
	"go.temporal.io/server/service/history/consts"
	historyi "go.temporal.io/server/service/history/interfaces"
	"go.temporal.io/server/service/history/tasks"
	"go.temporal.io/server/service/history/workflow"
	wcache "go.temporal.io/server/service/history/workflow/cache"
)

const (
	resendMissingEventMessage  = "Resend missed sync activity events"
	resendHigherVersionMessage = "Resend sync activity events due to a higher version received"
)

type (
	ActivityStateReplicator interface {
		SyncActivityState(
			ctx context.Context,
			request *historyservice.SyncActivityRequest,
		) error
		SyncActivitiesState(
			ctx context.Context,
			request *historyservice.SyncActivitiesRequest,
		) error
	}

	ActivityStateReplicatorImpl struct {
		shardContext  historyi.ShardContext
		workflowCache wcache.Cache
		logger        log.Logger
	}
)

func NewActivityStateReplicator(
	shardContext historyi.ShardContext,
	workflowCache wcache.Cache,
	logger log.Logger,
) *ActivityStateReplicatorImpl {

	return &ActivityStateReplicatorImpl{
		shardContext:  shardContext,
		workflowCache: workflowCache,
		logger:        log.With(logger, tag.ComponentActivityStateReplicator),
	}
}

func (r *ActivityStateReplicatorImpl) SyncActivityState(
	ctx context.Context,
	request *historyservice.SyncActivityRequest,
) (retError error) {

	// sync activity info will only be sent from active side, when
	// 1. activity retry
	// 2. activity start
	// 3. activity heart beat
	// no sync activity task will be sent when active side fail / timeout activity,
	namespaceID := namespace.ID(request.GetNamespaceId())
	execution := commonpb.WorkflowExecution{
		WorkflowId: request.WorkflowId,
		RunId:      request.RunId,
	}

	executionContext, release, err := r.workflowCache.GetOrCreateWorkflowExecution(
		ctx,
		r.shardContext,
		namespaceID,
		&execution,
		locks.PriorityHigh,
	)
	if err != nil {
		// for get workflow execution context, with valid run id
		// err will not be of type EntityNotExistsError
		return err
	}
	defer func() { release(retError) }()

	mutableState, err := executionContext.LoadMutableState(ctx, r.shardContext)
	if err != nil {
		if _, isNotFound := err.(*serviceerror.NotFound); isNotFound {
			// this can happen if the workflow start event and this sync activity task are out of order
			// or the target workflow is long gone
			// the safe solution to this is to throw away the sync activity task
			// or otherwise, worker attempt will exceed limit and put this message to DLQ
			return nil
		}
		return err
	}
	applied, err := r.syncSingleActivityState(
		&definition.WorkflowKey{
			NamespaceID: request.NamespaceId,
			WorkflowID:  request.WorkflowId,
			RunID:       request.RunId,
		},
		mutableState,
		&historyservice.ActivitySyncInfo{
			Version:                    request.Version,
			ScheduledEventId:           request.ScheduledEventId,
			ScheduledTime:              request.ScheduledTime,
			StartedEventId:             request.StartedEventId,
			StartedTime:                request.StartedTime,
			LastHeartbeatTime:          request.LastHeartbeatTime,
			Details:                    request.Details,
			Attempt:                    request.Attempt,
			LastFailure:                request.LastFailure,
			LastWorkerIdentity:         request.LastWorkerIdentity,
			LastStartedBuildId:         request.LastStartedBuildId,
			LastStartedRedirectCounter: request.LastStartedRedirectCounter,
			VersionHistory:             request.VersionHistory,
			FirstScheduledTime:         request.FirstScheduledTime,
			LastAttemptCompleteTime:    request.LastAttemptCompleteTime,
			Stamp:                      request.Stamp,
			Paused:                     request.Paused,
			RetryInitialInterval:       request.RetryInitialInterval,
			RetryMaximumInterval:       request.RetryMaximumInterval,
			RetryMaximumAttempts:       request.RetryMaximumAttempts,
			RetryBackoffCoefficient:    request.RetryBackoffCoefficient,
		},
	)
	if err != nil {
		return err
	}
	if !applied {
		return consts.ErrDuplicate
	}

	// passive logic need to explicitly call create timer
	if _, err := workflow.NewTimerSequence(
		mutableState,
	).CreateNextActivityTimer(); err != nil {
		return err
	}

	if r.shardContext.GetConfig().EnableUpdateWorkflowModeIgnoreCurrent() {
		return executionContext.UpdateWorkflowExecutionAsPassive(ctx, r.shardContext)
	}

	// TODO: remove following code once EnableUpdateWorkflowModeIgnoreCurrent config is deprecated.
	updateMode := persistence.UpdateWorkflowModeUpdateCurrent
	if state, _ := mutableState.GetWorkflowStateStatus(); state == enumsspb.WORKFLOW_EXECUTION_STATE_ZOMBIE {
		updateMode = persistence.UpdateWorkflowModeBypassCurrent
	}

	return executionContext.UpdateWorkflowExecutionWithNew(
		ctx,
		r.shardContext,
		updateMode,
		nil, // no new workflow
		nil, // no new workflow
		historyi.TransactionPolicyPassive,
		nil,
	)
}

func (r *ActivityStateReplicatorImpl) SyncActivitiesState(
	ctx context.Context,
	request *historyservice.SyncActivitiesRequest,
) (retError error) {
	// sync activity info will only be sent from active side, when
	// 1. activity retry
	// 2. activity start
	// 3. activity heart beat
	// no sync activity task will be sent when active side fail / timeout activity,
	namespaceID := namespace.ID(request.GetNamespaceId())
	execution := &commonpb.WorkflowExecution{
		WorkflowId: request.WorkflowId,
		RunId:      request.RunId,
	}

	executionContext, release, err := r.workflowCache.GetOrCreateWorkflowExecution(
		ctx,
		r.shardContext,
		namespaceID,
		execution,
		locks.PriorityHigh,
	)
	if err != nil {
		// for get workflow execution context, with valid run id
		// err will not be of type EntityNotExistsError
		return err
	}
	defer func() { release(retError) }()

	mutableState, err := executionContext.LoadMutableState(ctx, r.shardContext)
	if err != nil {
		if _, isNotFound := err.(*serviceerror.NotFound); isNotFound {
			// this can happen if the workflow start event and this sync activity task are out of order
			// or the target workflow is long gone
			// the safe solution to this is to throw away the sync activity task
			// or otherwise, worker attempt will exceed limit and put this message to DLQ

			// TODO: this should return serviceerrors.NewRetryReplication to trigger a resend
			// resend logic will handle not found case and drop the task.
			return nil
		}
		return err
	}
	anyEventApplied := false
	for _, syncActivityInfo := range request.ActivitiesInfo {
		applied, err := r.syncSingleActivityState(
			&definition.WorkflowKey{
				NamespaceID: request.NamespaceId,
				WorkflowID:  request.WorkflowId,
				RunID:       request.RunId,
			},
			mutableState,
			syncActivityInfo,
		)
		if err != nil {
			return err
		}
		anyEventApplied = anyEventApplied || applied
	}
	if !anyEventApplied {
		return consts.ErrDuplicate
	}

	// passive logic need to explicitly call create timer
	if _, err := workflow.NewTimerSequence(
		mutableState,
	).CreateNextActivityTimer(); err != nil {
		return err
	}

	if r.shardContext.GetConfig().EnableUpdateWorkflowModeIgnoreCurrent() {
		return executionContext.UpdateWorkflowExecutionAsPassive(ctx, r.shardContext)
	}

	// TODO: remove following code once EnableUpdateWorkflowModeIgnoreCurrent config is deprecated.
	updateMode := persistence.UpdateWorkflowModeUpdateCurrent
	if state, _ := mutableState.GetWorkflowStateStatus(); state == enumsspb.WORKFLOW_EXECUTION_STATE_ZOMBIE {
		updateMode = persistence.UpdateWorkflowModeBypassCurrent
	}

	return executionContext.UpdateWorkflowExecutionWithNew(
		ctx,
		r.shardContext,
		updateMode,
		nil, // no new workflow
		nil, // no new workflow
		historyi.TransactionPolicyPassive,
		nil,
	)
}

func (r *ActivityStateReplicatorImpl) syncSingleActivityState(
	workflowKey *definition.WorkflowKey,
	mutableState historyi.MutableState,
	activitySyncInfo *historyservice.ActivitySyncInfo,
) (applied bool, retError error) {
	scheduledEventID := activitySyncInfo.GetScheduledEventId()
	shouldApply, err := r.compareVersionHistory(
		namespace.ID(workflowKey.NamespaceID),
		workflowKey.WorkflowID,
		workflowKey.RunID,
		scheduledEventID,
		mutableState,
		activitySyncInfo.GetVersionHistory(),
	)
	if err != nil || !shouldApply {
		return false, err
	}

	activityInfo, ok := mutableState.GetActivityInfo(scheduledEventID)
	if !ok {
		// this should not retry, can be caused by out of order delivery
		// since the activity is already finished
		return false, nil
	}
	if shouldApply := r.compareActivity(
		activitySyncInfo.GetVersion(),
		activitySyncInfo.GetAttempt(),
		activitySyncInfo.GetStamp(),
		timestamp.TimeValue(activitySyncInfo.GetLastHeartbeatTime()),
		activityInfo,
	); !shouldApply {
		return false, nil
	}

	// sync activity with empty started ID means activity retry
	eventTime := timestamp.TimeValue(activitySyncInfo.GetScheduledTime())
	if activitySyncInfo.StartedEventId == common.EmptyEventID && activitySyncInfo.Attempt > activityInfo.GetAttempt() {
		mutableState.AddTasks(&tasks.ActivityRetryTimerTask{
			WorkflowKey:         *workflowKey,
			VisibilityTimestamp: eventTime,
			EventID:             activitySyncInfo.GetScheduledEventId(),
			Version:             activitySyncInfo.GetVersion(),
			Attempt:             activitySyncInfo.GetAttempt(),
			Stamp:               activitySyncInfo.GetStamp(),
		})
	}

	if err := mutableState.UpdateActivityInfo(
		activitySyncInfo,
		mutableState.ShouldResetActivityTimerTaskMask(
			activityInfo,
			&persistencespb.ActivityInfo{
				Version: activitySyncInfo.GetVersion(),
				Attempt: activitySyncInfo.GetAttempt(),
			},
		),
	); err != nil {
		return false, err
	}

	return true, nil
}

func (r *ActivityStateReplicatorImpl) compareActivity(
	version int64,
	attempt int32,
	stamp int32,
	lastHeartbeatTime time.Time,
	activityInfo *persistencespb.ActivityInfo,
) bool {

	if activityInfo.Version > version {
		// this should not retry, can be caused by failover or reset
		return false
	}

	if activityInfo.Version < version {
		// incoming version larger then local version, should update activity
		return true
	}

	if activityInfo.Stamp < stamp {
		// stamp changed, should update activity
		return true
	}

	if activityInfo.Stamp > stamp {
		// stamp is older than we have, should not update activity
		return false
	}

	// activityInfo.Version == version
	if activityInfo.Attempt > attempt {
		// this should not retry, can be caused by failover or reset
		return false
	}

	// activityInfo.Version == version
	if activityInfo.Attempt < attempt {
		// version equal & attempt larger then existing, should update activity
		return true
	}

	// activityInfo.Version == version & activityInfo.Attempt == attempt

	// last heartbeat after existing heartbeat & should update activity
	if activityInfo.LastHeartbeatUpdateTime != nil &&
		!activityInfo.LastHeartbeatUpdateTime.AsTime().IsZero() &&
		activityInfo.LastHeartbeatUpdateTime.AsTime().After(lastHeartbeatTime) {
		// this should not retry, can be caused by out of order delivery
		return false
	}
	return true
}

func (r *ActivityStateReplicatorImpl) compareVersionHistory(
	namespaceID namespace.ID,
	workflowID string,
	runID string,
	scheduledEventID int64,
	mutableState historyi.MutableState,
	incomingVersionHistory *historyspb.VersionHistory,
) (bool, error) {

	currentVersionHistory, err := versionhistory.GetCurrentVersionHistory(
		mutableState.GetExecutionInfo().GetVersionHistories(),
	)
	if err != nil {
		return false, err
	}

	lastLocalItem, err := versionhistory.GetLastVersionHistoryItem(currentVersionHistory)
	if err != nil {
		return false, err
	}

	lastIncomingItem, err := versionhistory.GetLastVersionHistoryItem(incomingVersionHistory)
	if err != nil {
		return false, err
	}

	lcaItem, err := versionhistory.FindLCAVersionHistoryItem(currentVersionHistory, incomingVersionHistory)
	if err != nil {
		return false, err
	}

	// case 1: local version history is superset of incoming version history
	//  or incoming version history is superset of local version history
	//  resend the missing event if local version history doesn't have the schedule event

	// case 2: local version history and incoming version history diverged
	//  case 2-1: local version history has the higher version and discard the incoming event
	//  case 2-2: incoming version history has the higher version and resend the missing incoming events
	if versionhistory.IsLCAVersionHistoryItemAppendable(currentVersionHistory, lcaItem) ||
		versionhistory.IsLCAVersionHistoryItemAppendable(incomingVersionHistory, lcaItem) {
		// case 1
		if scheduledEventID > lcaItem.GetEventId() {
			return false, serviceerrors.NewRetryReplication(
				resendMissingEventMessage,
				namespaceID.String(),
				workflowID,
				runID,
				lcaItem.GetEventId(),
				lcaItem.GetVersion(),
				common.EmptyEventID,
				common.EmptyVersion,
			)
		}
	} else {
		// case 2
		if lastIncomingItem.GetVersion() < lastLocalItem.GetVersion() {
			// case 2-1
			return false, nil
		} else if lastIncomingItem.GetVersion() > lastLocalItem.GetVersion() {
			// case 2-2
			return false, serviceerrors.NewRetryReplication(
				resendHigherVersionMessage,
				namespaceID.String(),
				workflowID,
				runID,
				lcaItem.GetEventId(),
				lcaItem.GetVersion(),
				common.EmptyEventID,
				common.EmptyVersion,
			)
		}
	}

	state, _ := mutableState.GetWorkflowStateStatus()
	return state != enumsspb.WORKFLOW_EXECUTION_STATE_COMPLETED, nil
}
