package updateactivityoptions

import (
	"context"

	activitypb "go.temporal.io/api/activity/v1"
	commandpb "go.temporal.io/api/command/v1"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/api/historyservice/v1"
	persistencespb "go.temporal.io/server/api/persistence/v1"
	"go.temporal.io/server/common/definition"
	"go.temporal.io/server/common/namespace"
	"go.temporal.io/server/common/util"
	"go.temporal.io/server/service/history/api"
	"go.temporal.io/server/service/history/consts"
	historyi "go.temporal.io/server/service/history/interfaces"
	"go.temporal.io/server/service/history/workflow"
)

func Invoke(
	ctx context.Context,
	request *historyservice.UpdateActivityOptionsRequest,
	shardContext historyi.ShardContext,
	workflowConsistencyChecker api.WorkflowConsistencyChecker,
) (resp *historyservice.UpdateActivityOptionsResponse, retError error) {
	validator := api.NewCommandAttrValidator(
		shardContext.GetNamespaceRegistry(),
		shardContext.GetConfig(),
		nil,
	)

	var response *historyservice.UpdateActivityOptionsResponse

	err := api.GetAndUpdateWorkflowWithNew(
		ctx,
		nil,
		definition.NewWorkflowKey(
			request.NamespaceId,
			request.GetUpdateRequest().GetExecution().GetWorkflowId(),
			request.GetUpdateRequest().GetExecution().GetRunId(),
		),
		func(workflowLease api.WorkflowLease) (*api.UpdateWorkflowAction, error) {
			mutableState := workflowLease.GetMutableState()
			var err error
			response, err = processActivityOptionsRequest(validator, mutableState, request)
			if err != nil {
				return nil, err
			}
			return &api.UpdateWorkflowAction{
				Noop:               false,
				CreateWorkflowTask: false,
			}, nil
		},
		nil,
		shardContext,
		workflowConsistencyChecker,
	)

	if err != nil {
		return nil, err
	}

	return response, err
}

func processActivityOptionsRequest(
	validator *api.CommandAttrValidator,
	mutableState historyi.MutableState,
	request *historyservice.UpdateActivityOptionsRequest,
) (*historyservice.UpdateActivityOptionsResponse, error) {
	if !mutableState.IsWorkflowExecutionRunning() {
		return nil, consts.ErrWorkflowCompleted
	}
	updateRequest := request.GetUpdateRequest()
	mergeFrom := updateRequest.GetActivityOptions()
	if mergeFrom == nil {
		return nil, serviceerror.NewInvalidArgument("ActivityOptions are not provided")
	}

	var activityIDs []string
	switch a := updateRequest.GetActivity().(type) {
	case *workflowservice.UpdateActivityOptionsRequest_Id:
		activityIDs = append(activityIDs, a.Id)
	case *workflowservice.UpdateActivityOptionsRequest_Type:
		activityType := a.Type
		for _, ai := range mutableState.GetPendingActivityInfos() {
			if ai.ActivityType.Name == activityType {
				activityIDs = append(activityIDs, ai.ActivityId)
			}
		}
	}

	if len(activityIDs) == 0 {
		return nil, consts.ErrActivityNotFound
	}

	mask := updateRequest.GetUpdateMask()
	if mask == nil {
		return nil, serviceerror.NewInvalidArgument("UpdateMask is not provided")
	}

	updateFields := util.ParseFieldMask(mask)

	var adjustedOptions *activitypb.ActivityOptions
	var err error
	for _, activityId := range activityIDs {
		ai, activityFound := mutableState.GetActivityByActivityID(activityId)

		if !activityFound {
			return nil, consts.ErrActivityNotFound
		}

		if adjustedOptions, err = updateActivityOptions(validator, mutableState, request, ai, mergeFrom, updateFields); err != nil {
			return nil, err
		}
	}

	// fill the response
	response := &historyservice.UpdateActivityOptionsResponse{
		ActivityOptions: adjustedOptions,
	}
	return response, nil
}

func updateActivityOptions(
	validator *api.CommandAttrValidator,
	mutableState historyi.MutableState,
	request *historyservice.UpdateActivityOptionsRequest,
	ai *persistencespb.ActivityInfo,
	mergeFrom *activitypb.ActivityOptions,
	updateFields map[string]struct{},
) (*activitypb.ActivityOptions, error) {

	mergeInto := &activitypb.ActivityOptions{
		TaskQueue: &taskqueuepb.TaskQueue{
			Name: ai.TaskQueue,
		},
		ScheduleToCloseTimeout: ai.ScheduleToCloseTimeout,
		ScheduleToStartTimeout: ai.ScheduleToStartTimeout,
		StartToCloseTimeout:    ai.StartToCloseTimeout,
		HeartbeatTimeout:       ai.HeartbeatTimeout,
		RetryPolicy: &commonpb.RetryPolicy{
			BackoffCoefficient: ai.RetryBackoffCoefficient,
			InitialInterval:    ai.RetryInitialInterval,
			MaximumInterval:    ai.RetryMaximumInterval,
			MaximumAttempts:    ai.RetryMaximumAttempts,
		},
	}

	// update activity options
	if err := applyActivityOptions(mergeInto, mergeFrom, updateFields); err != nil {
		return nil, err
	}

	// validate the updated options
	adjustedOptions, err := adjustActivityOptions(validator, request.NamespaceId, ai.ActivityId, ai.ActivityType, mergeInto)
	if err != nil {
		return nil, err
	}

	if err = mutableState.UpdateActivity(ai.ScheduledEventId, func(activityInfo *persistencespb.ActivityInfo, _ historyi.MutableState) error {
		// update activity info with new options
		activityInfo.TaskQueue = adjustedOptions.TaskQueue.Name
		activityInfo.ScheduleToCloseTimeout = adjustedOptions.ScheduleToCloseTimeout
		activityInfo.ScheduleToStartTimeout = adjustedOptions.ScheduleToStartTimeout
		activityInfo.StartToCloseTimeout = adjustedOptions.StartToCloseTimeout
		activityInfo.HeartbeatTimeout = adjustedOptions.HeartbeatTimeout
		activityInfo.RetryMaximumInterval = adjustedOptions.RetryPolicy.MaximumInterval
		activityInfo.RetryBackoffCoefficient = adjustedOptions.RetryPolicy.BackoffCoefficient
		activityInfo.RetryInitialInterval = adjustedOptions.RetryPolicy.InitialInterval
		activityInfo.RetryMaximumAttempts = adjustedOptions.RetryPolicy.MaximumAttempts

		// move forward activity version
		activityInfo.Stamp++

		// invalidate timers
		activityInfo.TimerTaskStatus = workflow.TimerTaskStatusNone
		return nil
	}); err != nil {
		return nil, err
	}

	if workflow.GetActivityState(ai) == enumspb.PENDING_ACTIVITY_STATE_SCHEDULED {
		// in this case we always want to generate a new retry task

		// two options - activity can be in backoff, or scheduled (waiting to be started)
		// if activity in backoff
		// 		in this case there is already old retry task
		// 		it will be ignored because of stamp mismatch
		// if activity is scheduled and waiting to be started
		// 		eventually matching service will call history service (recordActivityTaskStarted)
		// 		history service will return error based on stamp. Task will be dropped

		nextScheduledTime := workflow.GetNextScheduledTime(ai)
		err = mutableState.RegenerateActivityRetryTask(ai, nextScheduledTime)
		if err != nil {
			return nil, err
		}
	}

	return adjustedOptions, nil
}

func applyActivityOptions(
	mergeInto *activitypb.ActivityOptions,
	mergeFrom *activitypb.ActivityOptions,
	updateFields map[string]struct{},
) error {

	if _, ok := updateFields["taskQueue.name"]; ok {
		if mergeFrom.TaskQueue == nil {
			return serviceerror.NewInvalidArgument("TaskQueue is not provided")
		}
		if mergeInto.TaskQueue == nil {
			mergeInto.TaskQueue = mergeFrom.TaskQueue
		}
		mergeInto.TaskQueue.Name = mergeFrom.TaskQueue.Name
	}

	if _, ok := updateFields["scheduleToCloseTimeout"]; ok {
		mergeInto.ScheduleToCloseTimeout = mergeFrom.ScheduleToCloseTimeout
	}

	if _, ok := updateFields["scheduleToStartTimeout"]; ok {
		mergeInto.ScheduleToStartTimeout = mergeFrom.ScheduleToStartTimeout
	}

	if _, ok := updateFields["startToCloseTimeout"]; ok {
		mergeInto.StartToCloseTimeout = mergeFrom.StartToCloseTimeout
	}

	if _, ok := updateFields["heartbeatTimeout"]; ok {
		mergeInto.HeartbeatTimeout = mergeFrom.HeartbeatTimeout
	}

	if mergeInto.RetryPolicy == nil {
		mergeInto.RetryPolicy = &commonpb.RetryPolicy{}
	}

	if _, ok := updateFields["retryPolicy.initialInterval"]; ok {
		if mergeFrom.RetryPolicy == nil {
			return serviceerror.NewInvalidArgument("RetryPolicy is not provided")
		}
		mergeInto.RetryPolicy.InitialInterval = mergeFrom.RetryPolicy.InitialInterval
	}

	if _, ok := updateFields["retryPolicy.backoffCoefficient"]; ok {
		if mergeFrom.RetryPolicy == nil {
			return serviceerror.NewInvalidArgument("RetryPolicy is not provided")
		}
		mergeInto.RetryPolicy.BackoffCoefficient = mergeFrom.RetryPolicy.BackoffCoefficient
	}

	if _, ok := updateFields["retryPolicy.maximumInterval"]; ok {
		if mergeFrom.RetryPolicy == nil {
			return serviceerror.NewInvalidArgument("RetryPolicy is not provided")
		}
		mergeInto.RetryPolicy.MaximumInterval = mergeFrom.RetryPolicy.MaximumInterval
	}
	if _, ok := updateFields["retryPolicy.maximumAttempts"]; ok {
		if mergeFrom.RetryPolicy == nil {
			return serviceerror.NewInvalidArgument("RetryPolicy is not provided")
		}
		mergeInto.RetryPolicy.MaximumAttempts = mergeFrom.RetryPolicy.MaximumAttempts
	}

	return nil
}

func adjustActivityOptions(
	validator *api.CommandAttrValidator,
	namespaceID string,
	activityID string,
	activityType *commonpb.ActivityType,
	ao *activitypb.ActivityOptions,
) (*activitypb.ActivityOptions, error) {
	attributes := &commandpb.ScheduleActivityTaskCommandAttributes{
		TaskQueue:              ao.TaskQueue,
		ScheduleToCloseTimeout: ao.ScheduleToCloseTimeout,
		ScheduleToStartTimeout: ao.ScheduleToStartTimeout,
		StartToCloseTimeout:    ao.StartToCloseTimeout,
		HeartbeatTimeout:       ao.HeartbeatTimeout,
		ActivityId:             activityID,
		ActivityType:           activityType,
	}

	_, err := validator.ValidateActivityScheduleAttributes(namespace.ID(namespaceID), attributes, nil)
	if err != nil {
		return nil, err
	}

	ao.ScheduleToCloseTimeout = attributes.ScheduleToCloseTimeout
	ao.ScheduleToStartTimeout = attributes.ScheduleToStartTimeout
	ao.StartToCloseTimeout = attributes.StartToCloseTimeout
	ao.HeartbeatTimeout = attributes.HeartbeatTimeout

	return ao, nil
}
