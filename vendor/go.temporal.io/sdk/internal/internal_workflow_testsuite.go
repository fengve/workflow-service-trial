// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package internal

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/facebookgo/clock"
	"github.com/golang/mock/gomock"
	"github.com/robfig/cron"
	"github.com/stretchr/testify/mock"
	commandpb "go.temporal.io/api/command/v1"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/api/workflowservicemock/v1"
	"google.golang.org/grpc"

	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/internal/common"
	"go.temporal.io/sdk/internal/common/metrics"
	ilog "go.temporal.io/sdk/internal/log"
	"go.temporal.io/sdk/log"
)

const (
	defaultTestNamespace        = "default-test-namespace"
	defaultTestTaskQueue        = "default-test-taskqueue"
	defaultTestWorkflowID       = "default-test-workflow-id"
	defaultTestRunID            = "default-test-run-id"
	defaultTestWorkflowTypeName = "default-test-workflow-type-name"
	workflowTypeNotSpecified    = "workflow-type-not-specified"

	// These are copied from service implementation
	reservedTaskQueuePrefix = "/__temporal_sys/"
	maxIDLengthLimit        = 1000
	maxWorkflowTimeout      = 24 * time.Hour * 365 * 10

	defaultMaximumAttemptsForUnitTest = 10
)

type (
	testTimerHandle struct {
		env            *testWorkflowEnvironmentImpl
		callback       ResultHandler
		timer          *clock.Timer
		wallTimer      *clock.Timer
		duration       time.Duration
		mockTimeToFire time.Time
		wallTimeToFire time.Time
		timerID        int64
	}

	testActivityHandle struct {
		callback         ResultHandler
		activityType     string
		heartbeatDetails *commonpb.Payloads
	}

	testWorkflowHandle struct {
		env      *testWorkflowEnvironmentImpl
		callback ResultHandler
		handled  bool
		params   *ExecuteWorkflowParams
		err      error
	}

	testCallbackHandle struct {
		callback          func()
		startWorkflowTask bool // start a new workflow task after callback() is handled.
		env               *testWorkflowEnvironmentImpl
	}

	activityExecutorWrapper struct {
		*activityExecutor
		env *testWorkflowEnvironmentImpl
	}

	workflowExecutorWrapper struct {
		*workflowExecutor
		env *testWorkflowEnvironmentImpl
	}

	mockWrapper struct {
		env           *testWorkflowEnvironmentImpl
		name          string
		fn            interface{}
		isWorkflow    bool
		dataConverter converter.DataConverter
	}

	taskQueueSpecificActivity struct {
		fn         interface{}
		taskQueues map[string]struct{}
	}

	// testWorkflowEnvironmentShared is the shared data between parent workflow and child workflow test environments
	testWorkflowEnvironmentShared struct {
		locker    sync.Mutex
		testSuite *WorkflowTestSuite

		taskQueueSpecificActivities map[string]*taskQueueSpecificActivity

		mock                      *mock.Mock
		service                   workflowservice.WorkflowServiceClient
		logger                    log.Logger
		metricsHandler            metrics.Handler
		contextPropagators        []ContextPropagator
		identity                  string
		detachedChildWaitDisabled bool

		mockClock *clock.Mock
		wallClock clock.Clock

		callbackChannel chan testCallbackHandle
		testTimeout     time.Duration
		header          *commonpb.Header

		counterID        int64
		activities       map[string]*testActivityHandle
		localActivities  map[string]*localActivityTask
		timers           map[string]*testTimerHandle
		runningWorkflows map[string]*testWorkflowHandle

		runningCount int

		expectedMockCalls map[string]struct{}

		onActivityStartedListener        func(activityInfo *ActivityInfo, ctx context.Context, args converter.EncodedValues)
		onActivityCompletedListener      func(activityInfo *ActivityInfo, result converter.EncodedValue, err error)
		onActivityCanceledListener       func(activityInfo *ActivityInfo)
		onLocalActivityStartedListener   func(activityInfo *ActivityInfo, ctx context.Context, args []interface{})
		onLocalActivityCompletedListener func(activityInfo *ActivityInfo, result converter.EncodedValue, err error)
		onLocalActivityCanceledListener  func(activityInfo *ActivityInfo)
		onActivityHeartbeatListener      func(activityInfo *ActivityInfo, details converter.EncodedValues)
		onChildWorkflowStartedListener   func(workflowInfo *WorkflowInfo, ctx Context, args converter.EncodedValues)
		onChildWorkflowCompletedListener func(workflowInfo *WorkflowInfo, result converter.EncodedValue, err error)
		onChildWorkflowCanceledListener  func(workflowInfo *WorkflowInfo)
		onTimerScheduledListener         func(timerID string, duration time.Duration)
		onTimerFiredListener             func(timerID string)
		onTimerCanceledListener          func(timerID string)
	}

	// testWorkflowEnvironmentImpl is the environment that runs the workflow/activity unit tests.
	testWorkflowEnvironmentImpl struct {
		*testWorkflowEnvironmentShared
		parentEnv *testWorkflowEnvironmentImpl
		registry  *registry

		workflowInfo   *WorkflowInfo
		workflowDef    WorkflowDefinition
		changeVersions map[string]Version
		openSessions   map[string]*SessionInfo

		workflowCancelHandler func()
		signalHandler         func(name string, input *commonpb.Payloads, header *commonpb.Header) error
		queryHandler          func(string, *commonpb.Payloads, *commonpb.Header) (*commonpb.Payloads, error)
		startedHandler        func(r WorkflowExecution, e error)

		isWorkflowCompleted bool
		testResult          converter.EncodedValue
		testError           error
		doneChannel         chan struct{}
		workerOptions       WorkerOptions
		dataConverter       converter.DataConverter
		runTimeout          time.Duration

		heartbeatDetails *commonpb.Payloads

		workerStopChannel  chan struct{}
		sessionEnvironment *testSessionEnvironmentImpl

		// True if this was created only for testing activities not workflows.
		activityEnvOnly bool
	}

	testSessionEnvironmentImpl struct {
		*sessionEnvironmentImpl
		testWorkflowEnvironment *testWorkflowEnvironmentImpl
	}
)

func newTestWorkflowEnvironmentImpl(s *WorkflowTestSuite, parentRegistry *registry) *testWorkflowEnvironmentImpl {
	var r *registry
	if parentRegistry == nil {
		r = newRegistry()
	} else {
		r = parentRegistry
	}

	env := &testWorkflowEnvironmentImpl{
		testWorkflowEnvironmentShared: &testWorkflowEnvironmentShared{
			testSuite:                   s,
			taskQueueSpecificActivities: make(map[string]*taskQueueSpecificActivity),

			logger:            s.logger,
			metricsHandler:    s.metricsHandler,
			mockClock:         clock.NewMock(),
			wallClock:         clock.New(),
			timers:            make(map[string]*testTimerHandle),
			activities:        make(map[string]*testActivityHandle),
			localActivities:   make(map[string]*localActivityTask),
			runningWorkflows:  make(map[string]*testWorkflowHandle),
			callbackChannel:   make(chan testCallbackHandle, 1000),
			testTimeout:       3 * time.Second,
			expectedMockCalls: make(map[string]struct{}),
		},

		workflowInfo: &WorkflowInfo{
			Namespace: defaultTestNamespace,
			WorkflowExecution: WorkflowExecution{
				ID:    defaultTestWorkflowID,
				RunID: defaultTestRunID,
			},
			WorkflowType:  WorkflowType{Name: workflowTypeNotSpecified},
			TaskQueueName: defaultTestTaskQueue,

			WorkflowExecutionTimeout: maxWorkflowTimeout,
			WorkflowTaskTimeout:      1 * time.Second,
			Attempt:                  1,
		},
		registry: r,

		changeVersions: make(map[string]Version),
		openSessions:   make(map[string]*SessionInfo),

		doneChannel:       make(chan struct{}),
		workerStopChannel: make(chan struct{}),
		dataConverter:     converter.GetDefaultDataConverter(),
		runTimeout:        maxWorkflowTimeout,
	}

	if debugMode {
		env.testTimeout = time.Hour * 24
		env.workerOptions.DeadlockDetectionTimeout = unlimitedDeadlockDetectionTimeout
	}

	// move forward the mock clock to start time.
	env.setStartTime(time.Now())

	// put current workflow as a running workflow so child can send signal to parent
	env.runningWorkflows[env.workflowInfo.WorkflowExecution.ID] = &testWorkflowHandle{env: env, callback: func(result *commonpb.Payloads, err error) {}}

	if env.logger == nil {
		env.logger = ilog.NewDefaultLogger()
	}
	if env.metricsHandler == nil {
		env.metricsHandler = metrics.NopHandler
	}
	env.contextPropagators = s.contextPropagators
	env.header = s.header

	// setup mock service
	mockCtrl := gomock.NewController(ilog.NewTestReporter(env.logger))
	mockService := workflowservicemock.NewMockWorkflowServiceClient(mockCtrl)

	mockHeartbeatFn := func(c context.Context, r *workflowservice.RecordActivityTaskHeartbeatRequest, opts ...grpc.CallOption) error {
		activityID := ActivityID{id: string(r.TaskToken)}
		env.locker.Lock() // need lock as this is running in activity worker's goroutinue
		activityHandle, ok := env.getActivityHandle(activityID.id, GetActivityInfo(c).WorkflowExecution.RunID)
		env.locker.Unlock()
		if !ok {
			env.logger.Debug("RecordActivityTaskHeartbeat: ActivityID not found, could be already completed or canceled.",
				tagActivityID, activityID)
			return serviceerror.NewNotFound("")
		}
		activityHandle.heartbeatDetails = r.Details
		activityInfo := env.getActivityInfo(activityID, activityHandle.activityType)
		if env.onActivityHeartbeatListener != nil {
			// If we're only in an activity environment, posted callbacks are not
			// invoked
			if env.activityEnvOnly {
				env.onActivityHeartbeatListener(activityInfo, newEncodedValues(r.Details, env.GetDataConverter()))
			} else {
				env.postCallback(func() {
					env.onActivityHeartbeatListener(activityInfo, newEncodedValues(r.Details, env.GetDataConverter()))
				}, false)
			}
		}

		env.logger.Debug("RecordActivityTaskHeartbeat", tagActivityID, activityID)
		return nil
	}

	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(
		ctx context.Context,
		r *workflowservice.RecordActivityTaskHeartbeatRequest,
		opts ...grpc.CallOption,
	) (*workflowservice.RecordActivityTaskHeartbeatResponse, error) {
		if err := mockHeartbeatFn(ctx, r, opts...); err != nil {
			return nil, err
		}
		return &workflowservice.RecordActivityTaskHeartbeatResponse{CancelRequested: false}, nil
	}).AnyTimes()

	env.service = mockService

	return env
}

func (env *testWorkflowEnvironmentImpl) setStartTime(startTime time.Time) {
	// move forward the mock clock to start time.
	if startTime.IsZero() {
		// if start time not set, use current clock time
		startTime = env.wallClock.Now()
	}
	env.mockClock.Add(startTime.Sub(env.mockClock.Now()))
	env.workflowInfo.WorkflowStartTime = env.mockClock.Now()
}

func (env *testWorkflowEnvironmentImpl) newTestWorkflowEnvironmentForChild(params *ExecuteWorkflowParams, callback ResultHandler, startedHandler func(r WorkflowExecution, e error)) (*testWorkflowEnvironmentImpl, error) {
	// create a new test env
	childEnv := newTestWorkflowEnvironmentImpl(env.testSuite, env.registry)
	childEnv.parentEnv = env
	childEnv.startedHandler = startedHandler
	childEnv.testWorkflowEnvironmentShared = env.testWorkflowEnvironmentShared
	childEnv.workerOptions = env.workerOptions
	childEnv.dataConverter = params.DataConverter
	childEnv.registry = env.registry
	childEnv.detachedChildWaitDisabled = env.detachedChildWaitDisabled

	if params.TaskQueueName == "" {
		return nil, serviceerror.NewWorkflowExecutionAlreadyStarted("Empty task queue name", "", "")
	}

	if params.WorkflowID == "" {
		params.WorkflowID = env.workflowInfo.WorkflowExecution.RunID + "_" + getStringID(env.nextID())
	}
	var cronSchedule string
	if len(params.CronSchedule) > 0 {
		cronSchedule = params.CronSchedule
	}
	// set workflow info data for child workflow
	childEnv.header = params.Header
	childEnv.workflowInfo.Attempt = params.attempt
	childEnv.workflowInfo.WorkflowExecution.ID = params.WorkflowID
	childEnv.workflowInfo.WorkflowExecution.RunID = params.WorkflowID + "_RunID"
	childEnv.workflowInfo.Namespace = params.Namespace
	childEnv.workflowInfo.TaskQueueName = params.TaskQueueName
	childEnv.workflowInfo.WorkflowExecutionTimeout = params.WorkflowExecutionTimeout
	childEnv.workflowInfo.WorkflowRunTimeout = params.WorkflowRunTimeout
	childEnv.workflowInfo.WorkflowTaskTimeout = params.WorkflowTaskTimeout
	childEnv.workflowInfo.lastCompletionResult = params.lastCompletionResult
	childEnv.workflowInfo.CronSchedule = cronSchedule
	childEnv.workflowInfo.ParentWorkflowNamespace = env.workflowInfo.Namespace
	childEnv.workflowInfo.ParentWorkflowExecution = &env.workflowInfo.WorkflowExecution
	childEnv.runTimeout = params.WorkflowRunTimeout
	if workflowHandler, ok := env.runningWorkflows[params.WorkflowID]; ok {
		// duplicate workflow ID
		if !workflowHandler.handled {
			return nil, serviceerror.NewWorkflowExecutionAlreadyStarted("Workflow execution already started", "", "")
		}
		if params.WorkflowIDReusePolicy == enumspb.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE {
			return nil, serviceerror.NewWorkflowExecutionAlreadyStarted("Workflow execution already started", "", "")
		}
		if workflowHandler.err == nil && params.WorkflowIDReusePolicy == enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY {
			return nil, serviceerror.NewWorkflowExecutionAlreadyStarted("Workflow execution already started", "", "")
		}
	}

	env.runningWorkflows[params.WorkflowID] = &testWorkflowHandle{env: childEnv, callback: callback, params: params}

	return childEnv, nil
}

func (env *testWorkflowEnvironmentImpl) setWorkerOptions(options WorkerOptions) {
	env.workerOptions = options
	env.registry.interceptors = options.Interceptors
	if env.workerOptions.EnableSessionWorker && env.sessionEnvironment == nil {
		env.registry.RegisterActivityWithOptions(sessionCreationActivity, RegisterActivityOptions{
			Name:                          sessionCreationActivityName,
			DisableAlreadyRegisteredCheck: true,
		})
		env.registry.RegisterActivityWithOptions(sessionCompletionActivity, RegisterActivityOptions{
			Name:                          sessionCompletionActivityName,
			DisableAlreadyRegisteredCheck: true,
		})
	}
}

func (env *testWorkflowEnvironmentImpl) setIdentity(identity string) {
	env.identity = identity
}

func (env *testWorkflowEnvironmentImpl) setDataConverter(dataConverter converter.DataConverter) {
	env.dataConverter = dataConverter
}

func (env *testWorkflowEnvironmentImpl) setContextPropagators(contextPropagators []ContextPropagator) {
	env.contextPropagators = contextPropagators
}

func (env *testWorkflowEnvironmentImpl) setWorkerStopChannel(c chan struct{}) {
	env.workerStopChannel = c
}

func (env *testWorkflowEnvironmentImpl) setDetachedChildWaitDisabled(detachedChildWaitDisabled bool) {
	env.detachedChildWaitDisabled = detachedChildWaitDisabled
}

func (env *testWorkflowEnvironmentImpl) setActivityTaskQueue(taskqueue string, activityFns ...interface{}) {
	for _, activityFn := range activityFns {
		fnName := getActivityFunctionName(env.registry, activityFn)
		taskQueueActivity, ok := env.taskQueueSpecificActivities[fnName]
		if !ok {
			taskQueueActivity = &taskQueueSpecificActivity{fn: activityFn, taskQueues: make(map[string]struct{})}
			env.taskQueueSpecificActivities[fnName] = taskQueueActivity
		}
		taskQueueActivity.taskQueues[taskqueue] = struct{}{}
	}
}

func (env *testWorkflowEnvironmentImpl) executeWorkflow(workflowFn interface{}, args ...interface{}) {
	fType := reflect.TypeOf(workflowFn)
	if getKind(fType) == reflect.Func {
		env.RegisterWorkflowWithOptions(workflowFn, RegisterWorkflowOptions{DisableAlreadyRegisteredCheck: true})
	}
	workflowType, input, err := getValidatedWorkflowFunction(workflowFn, args, env.GetDataConverter(), env.GetRegistry())
	if err != nil {
		panic(err)
	}
	env.executeWorkflowInternal(0, workflowType.Name, input)
}

func (env *testWorkflowEnvironmentImpl) executeWorkflowInternal(delayStart time.Duration, workflowType string, input *commonpb.Payloads) {
	env.locker.Lock()
	wInfo := env.workflowInfo
	if wInfo.WorkflowType.Name != workflowTypeNotSpecified {
		// Current TestWorkflowEnvironment only support to run one workflow.
		// Created task to support testing multiple workflows with one env instance
		// https://github.com/temporalio/go-sdk/issues/50
		panic(fmt.Sprintf("Current TestWorkflowEnvironment is used to execute %v. Please create a new TestWorkflowEnvironment for %v.", wInfo.WorkflowType.Name, workflowType))
	}
	wInfo.WorkflowType.Name = workflowType
	if wInfo.WorkflowRunTimeout == 0 {
		wInfo.WorkflowRunTimeout = env.runTimeout
	}
	if wInfo.WorkflowExecutionTimeout == 0 {
		wInfo.WorkflowExecutionTimeout = maxWorkflowTimeout
	}
	if wInfo.WorkflowTaskTimeout == 0 {
		wInfo.WorkflowTaskTimeout = 1 * time.Second
	}
	env.locker.Unlock()

	workflowDefinition, err := env.getWorkflowDefinition(wInfo.WorkflowType)
	if err != nil {
		panic(err)
	}
	env.workflowDef = workflowDefinition

	// env.workflowDef.Execute() method will execute dispatcher. We want the dispatcher to only run in main loop.
	// In case of child workflow, this executeWorkflowInternal() is run in separate goroutinue, so use postCallback
	// to make sure workflowDef.Execute() is run in main loop.
	env.postCallback(func() {
		env.workflowDef.Execute(env, env.header, input)
		// kick off first workflow task to start the workflow
		if delayStart == 0 {
			env.startWorkflowTask()
		} else {
			// we need to delayStart start workflow, decrease runningCount so mockClock could auto forward
			env.runningCount--
			env.registerDelayedCallback(func() {
				env.runningCount++
				env.startWorkflowTask()
			}, delayStart)
		}
	}, false)

	if env.runTimeout > 0 {
		timeoutDuration := env.runTimeout + delayStart
		env.registerDelayedCallback(func() {
			if !env.isWorkflowCompleted {
				env.Complete(nil, ErrDeadlineExceeded)
			}
		}, timeoutDuration)
	}
	env.startMainLoop()
}

func (env *testWorkflowEnvironmentImpl) getWorkflowDefinition(wt WorkflowType) (WorkflowDefinition, error) {
	wf, ok := env.registry.getWorkflowFn(wt.Name)
	if !ok {
		supported := strings.Join(env.registry.getRegisteredWorkflowTypes(), ", ")
		return nil, fmt.Errorf("unable to find workflow type: %v. Supported types: [%v]", wt.Name, supported)
	}
	wd := &workflowExecutorWrapper{
		workflowExecutor: &workflowExecutor{workflowType: wt.Name, fn: wf, interceptors: env.registry.interceptors},
		env:              env,
	}
	return newSyncWorkflowDefinition(wd), nil
}

func (env *testWorkflowEnvironmentImpl) executeActivity(
	activityFn interface{},
	args ...interface{},
) (converter.EncodedValue, error) {
	activityType, err := getValidatedActivityFunction(activityFn, args, env.registry)
	if err != nil {
		panic(err)
	}

	input, err := encodeArgs(env.GetDataConverter(), args)
	if err != nil {
		panic(err)
	}

	parameters := ExecuteActivityParams{
		ExecuteActivityOptions: ExecuteActivityOptions{
			ScheduleToCloseTimeout: 600 * time.Second,
			StartToCloseTimeout:    600 * time.Second,
		},
		ActivityType: *activityType,
		Input:        input,
		Header:       env.header,
	}

	scheduleTaskAttr := &commandpb.ScheduleActivityTaskCommandAttributes{}
	if parameters.ActivityID == "" {
		scheduleTaskAttr.ActivityId = getStringID(env.nextID())
	} else {
		scheduleTaskAttr.ActivityId = parameters.ActivityID
	}
	scheduleTaskAttr.ActivityType = &commonpb.ActivityType{Name: parameters.ActivityType.Name}
	scheduleTaskAttr.TaskQueue = &taskqueuepb.TaskQueue{Name: parameters.TaskQueueName, Kind: enumspb.TASK_QUEUE_KIND_NORMAL}
	scheduleTaskAttr.Input = parameters.Input
	scheduleTaskAttr.ScheduleToCloseTimeout = &parameters.ScheduleToCloseTimeout
	scheduleTaskAttr.StartToCloseTimeout = &parameters.StartToCloseTimeout
	scheduleTaskAttr.ScheduleToStartTimeout = &parameters.ScheduleToStartTimeout
	scheduleTaskAttr.HeartbeatTimeout = &parameters.HeartbeatTimeout
	scheduleTaskAttr.RetryPolicy = parameters.RetryPolicy
	scheduleTaskAttr.Header = parameters.Header

	workflowType := env.workflowInfo.WorkflowType.Name
	if workflowType == workflowTypeNotSpecified {
		workflowType = "0"
	}
	task := newTestActivityTask(
		env.workflowInfo.WorkflowExecution.ID,
		env.workflowInfo.WorkflowExecution.RunID,
		workflowType,
		env.workflowInfo.Namespace,
		scheduleTaskAttr,
	)

	task.HeartbeatDetails = env.heartbeatDetails

	// ensure activityFn is registered to defaultTestTaskQueue
	taskHandler := env.newTestActivityTaskHandler(defaultTestTaskQueue, env.GetDataConverter())
	activityHandle := &testActivityHandle{callback: func(result *commonpb.Payloads, err error) {}, activityType: parameters.ActivityType.Name}
	activityID := ActivityID{id: scheduleTaskAttr.GetActivityId()}
	env.setActivityHandle(activityID.id, env.workflowInfo.WorkflowExecution.RunID, activityHandle)

	result, err := taskHandler.Execute(defaultTestTaskQueue, task)
	if err != nil {
		if err == context.DeadlineExceeded {
			env.logger.Debug(fmt.Sprintf("Activity %v timed out", task.ActivityType.Name))
			return nil, env.wrapActivityError(activityID, scheduleTaskAttr.ActivityType.Name, enumspb.RETRY_STATE_TIMEOUT, NewTimeoutError("Activity timeout", enumspb.TIMEOUT_TYPE_START_TO_CLOSE, err))
		}
		topLine := fmt.Sprintf("activity for %s [panic]:", defaultTestTaskQueue)
		st := getStackTraceRaw(topLine, 7, 0)
		return nil, env.wrapActivityError(activityID, scheduleTaskAttr.ActivityType.Name, enumspb.RETRY_STATE_UNSPECIFIED, newPanicError(err.Error(), st))
	}

	if result == ErrActivityResultPending {
		return nil, ErrActivityResultPending
	}

	switch request := result.(type) {
	case *workflowservice.RespondActivityTaskCanceledRequest:
		details := newEncodedValues(request.Details, env.GetDataConverter())
		return nil, env.wrapActivityError(activityID, scheduleTaskAttr.ActivityType.Name, enumspb.RETRY_STATE_NON_RETRYABLE_FAILURE, NewCanceledError(details))
	case *workflowservice.RespondActivityTaskFailedRequest:
		return nil, env.wrapActivityError(activityID, scheduleTaskAttr.ActivityType.Name, enumspb.RETRY_STATE_UNSPECIFIED, ConvertFailureToError(request.GetFailure(), env.GetDataConverter()))
	case *workflowservice.RespondActivityTaskCompletedRequest:
		return newEncodedValue(request.Result, env.GetDataConverter()), nil
	default:
		// will never happen
		return nil, fmt.Errorf("unsupported respond type %T", result)
	}
}

func (env *testWorkflowEnvironmentImpl) executeLocalActivity(
	activityFn interface{},
	args ...interface{},
) (val converter.EncodedValue, err error) {
	params := ExecuteLocalActivityParams{
		ExecuteLocalActivityOptions: ExecuteLocalActivityOptions{
			ScheduleToCloseTimeout: env.testTimeout,
		},
		ActivityFn:   activityFn,
		InputArgs:    args,
		WorkflowInfo: env.workflowInfo,
	}
	task := &localActivityTask{
		activityID: "test-local-activity",
		params:     &params,
		callback: func(lar *LocalActivityResultWrapper) {
		},
		attempt: 1,
	}
	taskHandler := localActivityTaskHandler{
		userContext:    env.workerOptions.BackgroundActivityContext,
		metricsHandler: env.metricsHandler,
		logger:         env.logger,
		interceptors:   env.registry.interceptors,
	}

	result := taskHandler.executeLocalActivityTask(task)
	if result.err != nil {
		activityType, _ := getValidatedActivityFunction(activityFn, args, env.registry)
		return nil, env.wrapActivityError(ActivityID{id: task.activityID}, activityType.Name, enumspb.RETRY_STATE_UNSPECIFIED, result.err)
	}
	return newEncodedValue(result.result, env.GetDataConverter()), nil
}

func (env *testWorkflowEnvironmentImpl) startWorkflowTask() {
	if !env.isWorkflowCompleted {
		env.workflowDef.OnWorkflowTaskStarted(env.workerOptions.DeadlockDetectionTimeout)
	}
}

func (env *testWorkflowEnvironmentImpl) isChildWorkflow() bool {
	return env.parentEnv != nil
}

func (env *testWorkflowEnvironmentImpl) startMainLoop() {
	if env.isChildWorkflow() {
		// child workflow rely on parent workflow's main loop to process events
		<-env.doneChannel // wait until workflow is complete
		return
	}

	// notify all child workflows to exit their main loop
	defer close(env.doneChannel)

	for !env.shouldStopEventLoop() {
		// use non-blocking-select to check if there is anything pending in the main thread.
		select {
		case c := <-env.callbackChannel:
			// this will drain the callbackChannel
			c.processCallback()
		default:
			// nothing to process, main thread is blocked at this moment, now check if we should auto fire next timer
			if !env.autoFireNextTimer() {
				if env.shouldStopEventLoop() {
					return
				}

				// no timer to fire, wait for things to do or timeout.
				select {
				case c := <-env.callbackChannel:
					c.processCallback()
				case <-time.After(env.testTimeout):
					// not able to complete workflow within test timeout, workflow likely stuck somewhere,
					// check workflow stack for more details.
					panicMsg := fmt.Sprintf("test timeout: %v, workflow stack: %v",
						env.testTimeout, env.workflowDef.StackTrace())
					panic(panicMsg)
				}
			}
		}
	}
}

func (env *testWorkflowEnvironmentImpl) shouldStopEventLoop() bool {
	// Check if any detached children are still running if not disabled.
	if !env.detachedChildWaitDisabled {
		for _, handle := range env.runningWorkflows {
			if env.workflowInfo.WorkflowExecution.ID == handle.env.workflowInfo.WorkflowExecution.ID {
				// ignore root workflow
				continue
			}

			if !handle.handled && (handle.params.ParentClosePolicy == enumspb.PARENT_CLOSE_POLICY_ABANDON ||
				handle.params.ParentClosePolicy == enumspb.PARENT_CLOSE_POLICY_REQUEST_CANCEL) {
				return false
			}
		}
	}

	return env.isWorkflowCompleted
}

func (env *testWorkflowEnvironmentImpl) registerDelayedCallback(f func(), delayDuration time.Duration) {
	timerCallback := func(result *commonpb.Payloads, err error) {
		f()
	}
	if delayDuration == 0 {
		env.postCallback(f, false)
		return
	}
	mainLoopCallback := func() {
		env.newTimer(delayDuration, timerCallback, false)
	}
	env.postCallback(mainLoopCallback, false)
}

func (c *testCallbackHandle) processCallback() {
	c.env.locker.Lock()
	defer c.env.locker.Unlock()
	c.callback()
	if c.startWorkflowTask {
		c.env.startWorkflowTask()
	}
}

func (env *testWorkflowEnvironmentImpl) autoFireNextTimer() bool {
	if len(env.timers) == 0 {
		return false
	}

	// find next timer
	var nextTimer *testTimerHandle
	for _, t := range env.timers {
		if nextTimer == nil {
			nextTimer = t
		} else if t.mockTimeToFire.Before(nextTimer.mockTimeToFire) ||
			(t.mockTimeToFire.Equal(nextTimer.mockTimeToFire) && t.timerID < nextTimer.timerID) {
			nextTimer = t
		}
	}

	if nextTimer == nil {
		return false
	}

	// function to fire timer
	fireTimer := func(th *testTimerHandle) {
		skipDuration := th.mockTimeToFire.Sub(env.mockClock.Now())
		env.logger.Debug("Auto fire timer",
			tagTimerID, th.timerID,
			"TimerDuration", th.duration,
			"TimeSkipped", skipDuration)

		// Move mockClock forward, this will fire the timer, and the timer callback will remove timer from timers.
		env.mockClock.Add(skipDuration)
	}

	// fire timer if there is no running activity
	if env.runningCount == 0 {
		if nextTimer.wallTimer != nil {
			nextTimer.wallTimer.Stop()
			nextTimer.wallTimer = nil
		}
		fireTimer(nextTimer)
		return true
	}

	durationToFire := nextTimer.mockTimeToFire.Sub(env.mockClock.Now())
	wallTimeToFire := env.wallClock.Now().Add(durationToFire)

	if nextTimer.wallTimer != nil && nextTimer.wallTimeToFire.Before(wallTimeToFire) {
		// nextTimer already set, meaning we already have a wall clock timer for the nextTimer setup earlier. And the
		// previously scheduled wall time to fire is before the wallTimeToFire calculated this time. This could happen
		// if workflow was blocked while there was activity running, and when that activity completed, there are some
		// other activities still running while the nextTimer is still that same nextTimer. In that case, we should not
		// reset the wall time to fire for the nextTimer.
		return false
	}
	if nextTimer.wallTimer != nil {
		// wallTimer was scheduled, but the wall time to fire should be earlier based on current calculation.
		nextTimer.wallTimer.Stop()
	}

	// there is running activities, we would fire next timer only if wall time passed by nextTimer duration.
	nextTimer.wallTimeToFire, nextTimer.wallTimer = wallTimeToFire, env.wallClock.AfterFunc(durationToFire, func() {
		// make sure it is running in the main loop
		nextTimer.env.postCallback(func() {
			if timerHandle, ok := env.timers[getStringID(nextTimer.timerID)]; ok {
				fireTimer(timerHandle)
			}
		}, true)
	})

	return false
}

func (env *testWorkflowEnvironmentImpl) postCallback(cb func(), startWorkflowTask bool) {
	env.callbackChannel <- testCallbackHandle{callback: cb, startWorkflowTask: startWorkflowTask, env: env}
}

func (env *testWorkflowEnvironmentImpl) RequestCancelActivity(activityID ActivityID) {
	handle, ok := env.getActivityHandle(activityID.id, env.workflowInfo.WorkflowExecution.RunID)
	if !ok {
		env.logger.Debug("RequestCancelActivity failed, Activity not exists or already completed.", tagActivityID, activityID)
		return
	}
	activityInfo := env.getActivityInfo(activityID, handle.activityType)
	env.logger.Debug("RequestCancelActivity", tagActivityID, activityID)
	env.deleteHandle(activityID.id, env.workflowInfo.WorkflowExecution.RunID)
	env.postCallback(func() {
		handle.callback(nil, NewCanceledError())
		if env.onActivityCanceledListener != nil {
			env.onActivityCanceledListener(activityInfo)
		}
	}, true)
}

// RequestCancelTimer request to cancel timer on this testWorkflowEnvironmentImpl.
func (env *testWorkflowEnvironmentImpl) RequestCancelTimer(timerID TimerID) {
	env.logger.Debug("RequestCancelTimer", tagTimerID, timerID)
	timerHandle, ok := env.timers[timerID.id]
	if !ok {
		env.logger.Debug("RequestCancelTimer failed, TimerID not exists.", tagTimerID, timerID)
		return
	}

	delete(env.timers, timerID.id)
	timerHandle.timer.Stop()
	timerHandle.env.postCallback(func() {
		timerHandle.callback(nil, NewCanceledError())
		if timerHandle.env.onTimerCanceledListener != nil {
			timerHandle.env.onTimerCanceledListener(timerID.id)
		}
	}, true)
}

func (env *testWorkflowEnvironmentImpl) Complete(result *commonpb.Payloads, err error) {
	if env.isWorkflowCompleted {
		env.logger.Debug("Workflow already completed.")
		return
	}
	env.workflowDef.Close()
	var canceledErr *CanceledError
	if errors.As(err, &canceledErr) && env.workflowCancelHandler != nil {
		env.workflowCancelHandler()
	}

	dc := env.GetDataConverter()
	env.isWorkflowCompleted = true

	if err != nil {
		var continueAsNewErr *ContinueAsNewError
		var timeoutErr *TimeoutError
		var workflowPanicErr *workflowPanicError
		var workflowExecutionAlreadyStartedErr *serviceerror.WorkflowExecutionAlreadyStarted
		if errors.As(err, &canceledErr) || errors.As(err, &continueAsNewErr) || errors.As(err, &timeoutErr) || errors.As(err, &workflowExecutionAlreadyStartedErr) {
			env.testError = err
		} else if errors.As(err, &workflowPanicErr) {
			env.testError = newPanicError(workflowPanicErr.value, workflowPanicErr.stackTrace)
		} else {
			failure := ConvertErrorToFailure(err, dc)
			env.testError = ConvertFailureToError(failure, dc)
		}

		if !env.isChildWorkflow() {
			env.testError = NewWorkflowExecutionError(
				env.WorkflowInfo().WorkflowExecution.ID,
				env.WorkflowInfo().WorkflowExecution.RunID,
				env.WorkflowInfo().WorkflowType.Name,
				env.testError,
			)
		}
	} else {
		env.testResult = newEncodedValue(result, dc)
	}

	if env.isChildWorkflow() {
		// this is completion of child workflow
		childWorkflowID := env.workflowInfo.WorkflowExecution.ID
		if childWorkflowHandle, ok := env.runningWorkflows[childWorkflowID]; ok && !childWorkflowHandle.handled {
			// It is possible that child workflow could complete after cancellation. In that case, childWorkflowHandle
			// would have already been removed from the runningWorkflows map by RequestCancelWorkflow().
			childWorkflowHandle.handled = true
			// check if a retry is needed
			if childWorkflowHandle.rerunAsChild() {
				// rerun requested, so we don't want to post the error to parent workflow, return here.
				return
			}

			// no rerun, child workflow is done.
			env.parentEnv.postCallback(func() {
				// deliver result
				if env.testError != nil {
					childWorkflowHandle.err = NewChildWorkflowExecutionError(
						defaultTestNamespace,
						env.WorkflowInfo().WorkflowExecution.ID,
						env.WorkflowInfo().WorkflowExecution.RunID,
						env.WorkflowInfo().WorkflowType.Name,
						0,
						0,
						enumspb.RETRY_STATE_UNSPECIFIED,
						env.testError,
					)
				}
				childWorkflowHandle.callback(result, childWorkflowHandle.err)
				if env.onChildWorkflowCompletedListener != nil {
					env.onChildWorkflowCompletedListener(env.workflowInfo, env.testResult, childWorkflowHandle.err)
				}
			}, true /* true to trigger parent workflow to resume to handle child workflow's result */)
		}
	}

	// properly handle child workflows based on their ParentClosePolicy
	env.handleParentClosePolicy()
}

func (env *testWorkflowEnvironmentImpl) handleParentClosePolicy() {
	for _, handle := range env.runningWorkflows {
		if handle.env.parentEnv != nil &&
			env.workflowInfo.WorkflowExecution.ID == handle.env.parentEnv.workflowInfo.WorkflowExecution.ID {

			switch handle.params.ParentClosePolicy {
			case enumspb.PARENT_CLOSE_POLICY_ABANDON:
				// noop
			case enumspb.PARENT_CLOSE_POLICY_TERMINATE:
				handle.env.Complete(nil, newTerminatedError())
			case enumspb.PARENT_CLOSE_POLICY_REQUEST_CANCEL:
				handle.env.cancelWorkflow(func(result *commonpb.Payloads, err error) {})
			}
		}
	}
}

func (h *testWorkflowHandle) rerunAsChild() bool {
	env := h.env
	if !env.isChildWorkflow() {
		return false
	}
	params := h.params

	// pass down the last completion result
	var result *commonpb.Payloads
	// TODO (shtin): convert env.testResult to *commonpb.Payloads
	if ev, ok := env.testResult.(*EncodedValue); ev != nil && ok {
		result = ev.value
	}
	if result == nil {
		// not successful run this time, carry over from whatever previous run pass to this run.
		result = env.workflowInfo.lastCompletionResult
	}
	params.lastCompletionResult = result

	if params.RetryPolicy != nil && env.testError != nil {
		var expireTime time.Time
		if params.WorkflowOptions.WorkflowExecutionTimeout > 0 {
			expireTime = params.scheduledTime.Add(params.WorkflowOptions.WorkflowExecutionTimeout)
		}
		backoff := getRetryBackoffFromProtoRetryPolicy(params.RetryPolicy, env.workflowInfo.Attempt, env.testError, env.Now(), expireTime)
		if backoff > 0 {
			// remove the current child workflow from the pending child workflow map because
			// the childWorkflowID will be the same for retry run.
			delete(env.runningWorkflows, env.workflowInfo.WorkflowExecution.ID)
			params.attempt++
			env.parentEnv.executeChildWorkflowWithDelay(backoff, *params, h.callback, nil /* child workflow already started */)

			return true
		}
	}

	if len(params.CronSchedule) > 0 {
		schedule, err := cron.ParseStandard(params.CronSchedule)
		if err != nil {
			panic(fmt.Errorf("invalid cron schedule %v, err: %v", params.CronSchedule, err))
		}

		workflowNow := env.Now().In(time.UTC)
		backoff := schedule.Next(workflowNow).Sub(workflowNow)
		if backoff > 0 {
			delete(env.runningWorkflows, env.workflowInfo.WorkflowExecution.ID)
			params.attempt = 1
			params.scheduledTime = env.Now()
			env.parentEnv.executeChildWorkflowWithDelay(backoff, *params, h.callback, nil /* child workflow already started */)
			return true
		}
	}

	return false
}

func (env *testWorkflowEnvironmentImpl) CompleteActivity(taskToken []byte, result interface{}, err error) error {
	if taskToken == nil {
		return errors.New("nil task token provided")
	}
	var data *commonpb.Payloads
	if result != nil {
		var encodeErr error
		data, encodeErr = encodeArg(env.GetDataConverter(), result)
		if encodeErr != nil {
			return encodeErr
		}
	}

	activityID := ActivityID{id: string(taskToken)}
	env.postCallback(func() {
		activityHandle, ok := env.getActivityHandle(activityID.id, env.workflowInfo.WorkflowExecution.RunID)
		if !ok {
			env.logger.Debug("CompleteActivity: ActivityID not found, could be already completed or canceled.",
				tagActivityID, activityID)
			return
		}
		// We do allow canceled error to be passed here
		cancelAllowed := true
		request := convertActivityResultToRespondRequest("test-identity", taskToken, data, err,
			env.GetDataConverter(), defaultTestNamespace, cancelAllowed)
		env.handleActivityResult(activityID, request, activityHandle.activityType, env.GetDataConverter())
	}, false /* do not auto schedule workflow task, because activity might be still pending */)

	return nil
}

func (env *testWorkflowEnvironmentImpl) GetLogger() log.Logger {
	return env.logger
}

func (env *testWorkflowEnvironmentImpl) GetMetricsHandler() metrics.Handler {
	return env.metricsHandler
}

func (env *testWorkflowEnvironmentImpl) GetDataConverter() converter.DataConverter {
	return env.dataConverter
}

func (env *testWorkflowEnvironmentImpl) GetContextPropagators() []ContextPropagator {
	return env.contextPropagators
}

func (env *testWorkflowEnvironmentImpl) ExecuteActivity(parameters ExecuteActivityParams, callback ResultHandler) ActivityID {
	ensureDefaultRetryPolicy(&parameters)
	scheduleTaskAttr := &commandpb.ScheduleActivityTaskCommandAttributes{}
	scheduleID := env.nextID()
	if parameters.ActivityID == "" {
		scheduleTaskAttr.ActivityId = getStringID(scheduleID)
	} else {
		scheduleTaskAttr.ActivityId = parameters.ActivityID
	}
	activityID := ActivityID{id: scheduleTaskAttr.GetActivityId()}
	scheduleTaskAttr.ActivityType = &commonpb.ActivityType{Name: parameters.ActivityType.Name}
	scheduleTaskAttr.TaskQueue = &taskqueuepb.TaskQueue{Name: parameters.TaskQueueName, Kind: enumspb.TASK_QUEUE_KIND_NORMAL}
	scheduleTaskAttr.Input = parameters.Input
	scheduleTaskAttr.ScheduleToCloseTimeout = &parameters.ScheduleToCloseTimeout
	scheduleTaskAttr.StartToCloseTimeout = &parameters.StartToCloseTimeout
	scheduleTaskAttr.ScheduleToStartTimeout = &parameters.ScheduleToStartTimeout
	scheduleTaskAttr.HeartbeatTimeout = &parameters.HeartbeatTimeout
	scheduleTaskAttr.RetryPolicy = parameters.RetryPolicy
	scheduleTaskAttr.Header = parameters.Header
	err := env.validateActivityScheduleAttributes(scheduleTaskAttr, env.WorkflowInfo().WorkflowRunTimeout)
	if err != nil {
		callback(nil, err)
		return activityID
	}
	task := newTestActivityTask(
		env.workflowInfo.WorkflowExecution.ID,
		env.workflowInfo.WorkflowExecution.RunID,
		env.workflowInfo.WorkflowType.Name,
		env.workflowInfo.Namespace,
		scheduleTaskAttr,
	)

	taskHandler := env.newTestActivityTaskHandler(parameters.TaskQueueName, parameters.DataConverter)
	activityHandle := &testActivityHandle{callback: callback, activityType: parameters.ActivityType.Name}

	env.setActivityHandle(activityID.id, env.workflowInfo.WorkflowExecution.RunID, activityHandle)
	env.runningCount++
	// activity runs in separate goroutinue outside of workflow dispatcher
	// do callback in a defer to handle calls to runtime.Goexit inside the activity (which is done by t.FailNow)
	go func() {
		var result interface{}
		defer func() {
			panicErr := recover()
			if result == nil && panicErr == nil {
				failureErr := errors.New("activity called runtime.Goexit")
				result = &workflowservice.RespondActivityTaskFailedRequest{
					Failure: ConvertErrorToFailure(failureErr, env.GetDataConverter()),
				}
			} else if panicErr != nil {
				failureErr := newPanicError(fmt.Sprintf("%v", panicErr), "")
				result = &workflowservice.RespondActivityTaskFailedRequest{
					Failure: ConvertErrorToFailure(failureErr, env.GetDataConverter()),
				}
			}
			// post activity result to workflow dispatcher
			env.postCallback(func() {
				env.handleActivityResult(activityID, result, parameters.ActivityType.Name, parameters.DataConverter)
				env.runningCount--
			}, false /* do not auto schedule workflow task, because activity might be still pending */)
		}()
		result = env.executeActivityWithRetryForTest(taskHandler, parameters, task)
	}()

	return activityID
}

// Copy of the server function func (v *commandAttrValidator) validateActivityScheduleAttributes
func (env *testWorkflowEnvironmentImpl) validateActivityScheduleAttributes(
	attributes *commandpb.ScheduleActivityTaskCommandAttributes,
	runTimeout time.Duration,
) error {

	if attributes == nil {
		return serviceerror.NewInvalidArgument("ScheduleActivityTaskCommandAttributes is not set on command.")
	}

	defaultTaskQueueName := ""
	if _, err := env.validatedTaskQueue(attributes.TaskQueue, defaultTaskQueueName); err != nil {
		return err
	}

	if attributes.GetActivityId() == "" {
		return serviceerror.NewInvalidArgument("ActivityId is not set on command.")
	}

	if attributes.ActivityType == nil || attributes.ActivityType.GetName() == "" {
		return serviceerror.NewInvalidArgument("ActivityType is not set on command.")
	}

	if err := env.validateRetryPolicy(attributes.RetryPolicy); err != nil {
		return err
	}

	if len(attributes.GetActivityId()) > maxIDLengthLimit {
		return serviceerror.NewInvalidArgument("ActivityID exceeds length limit.")
	}

	if len(attributes.GetActivityType().GetName()) > maxIDLengthLimit {
		return serviceerror.NewInvalidArgument("ActivityType exceeds length limit.")
	}

	// Only attempt to deduce and fill in unspecified timeouts only when all timeouts are non-negative.
	if common.DurationValue(attributes.GetScheduleToCloseTimeout()) < 0 || common.DurationValue(attributes.GetScheduleToStartTimeout()) < 0 ||
		common.DurationValue(attributes.GetStartToCloseTimeout()) < 0 || common.DurationValue(attributes.GetHeartbeatTimeout()) < 0 {
		return serviceerror.NewInvalidArgument("A valid timeout may not be negative.")
	}

	validScheduleToClose := common.DurationValue(attributes.GetScheduleToCloseTimeout()) > 0
	validScheduleToStart := common.DurationValue(attributes.GetScheduleToStartTimeout()) > 0
	validStartToClose := common.DurationValue(attributes.GetStartToCloseTimeout()) > 0

	if validScheduleToClose {
		if validScheduleToStart {
			attributes.ScheduleToStartTimeout = common.MinDurationPtr(attributes.GetScheduleToStartTimeout(), attributes.GetScheduleToCloseTimeout())
		} else {
			attributes.ScheduleToStartTimeout = attributes.GetScheduleToCloseTimeout()
		}
		if validStartToClose {
			attributes.StartToCloseTimeout = common.MinDurationPtr(attributes.GetStartToCloseTimeout(), attributes.GetScheduleToCloseTimeout())
		} else {
			attributes.StartToCloseTimeout = attributes.GetScheduleToCloseTimeout()
		}
	} else if validStartToClose {
		// We are in !validScheduleToClose due to the first if above
		attributes.ScheduleToCloseTimeout = &runTimeout
		if !validScheduleToStart {
			attributes.ScheduleToStartTimeout = &runTimeout
		}
	} else {
		// Deduction failed as there's not enough information to fill in missing timeouts.
		return serviceerror.NewInvalidArgument("A valid StartToClose or ScheduleToCloseTimeout is not set on command.")
	}
	// ensure activity timeout never larger than workflow timeout
	if runTimeout > 0 {
		if common.DurationValue(attributes.GetScheduleToCloseTimeout()) > runTimeout {
			attributes.ScheduleToCloseTimeout = &runTimeout
		}
		if common.DurationValue(attributes.GetScheduleToStartTimeout()) > runTimeout {
			attributes.ScheduleToStartTimeout = &runTimeout
		}
		if common.DurationValue(attributes.GetStartToCloseTimeout()) > runTimeout {
			attributes.StartToCloseTimeout = &runTimeout
		}
		if common.DurationValue(attributes.GetHeartbeatTimeout()) > runTimeout {
			attributes.HeartbeatTimeout = &runTimeout
		}
	}
	attributes.HeartbeatTimeout = common.MinDurationPtr(attributes.GetHeartbeatTimeout(), attributes.GetScheduleToCloseTimeout())
	return nil
}

// Copy of the service func (v *commandAttrValidator) validatedTaskQueue
func (env *testWorkflowEnvironmentImpl) validatedTaskQueue(
	taskQueue *taskqueuepb.TaskQueue,
	defaultVal string,
) (*taskqueuepb.TaskQueue, error) {

	if taskQueue == nil {
		taskQueue = &taskqueuepb.TaskQueue{Kind: enumspb.TASK_QUEUE_KIND_NORMAL}
	}

	if taskQueue.GetName() == "" {
		if defaultVal == "" {
			return taskQueue, serviceerror.NewInvalidArgument("missing task queue name")
		}
		taskQueue.Name = defaultVal
		return taskQueue, nil
	}

	name := taskQueue.GetName()
	if len(name) > maxIDLengthLimit {
		return taskQueue, serviceerror.NewInvalidArgument(fmt.Sprintf("task queue name exceeds length limit of %v", maxIDLengthLimit))
	}

	if strings.HasPrefix(name, reservedTaskQueuePrefix) {
		return taskQueue, serviceerror.NewInvalidArgument(fmt.Sprintf("task queue name cannot start with reserved prefix %v", reservedTaskQueuePrefix))
	}

	return taskQueue, nil
}

// copy of the service func ValidateRetryPolicy(policy *commonpb.RetryPolicy)
func (env *testWorkflowEnvironmentImpl) validateRetryPolicy(policy *commonpb.RetryPolicy) error {
	if policy == nil {
		// nil policy is valid which means no retry
		return nil
	}

	if policy.GetMaximumAttempts() == 1 {
		// One maximum attempt effectively disable retries. Validating the
		// rest of the arguments is pointless
		return nil
	}
	if common.DurationValue(policy.GetInitialInterval()) < 0 {
		return serviceerror.NewInvalidArgument("InitialInterval cannot be negative on retry policy.")
	}
	if policy.GetBackoffCoefficient() < 1 {
		return serviceerror.NewInvalidArgument("BackoffCoefficient cannot be less than 1 on retry policy.")
	}
	if common.DurationValue(policy.GetMaximumInterval()) < 0 {
		return serviceerror.NewInvalidArgument("MaximumInterval cannot be negative on retry policy.")
	}
	if common.DurationValue(policy.GetMaximumInterval()) > 0 && common.DurationValue(policy.GetMaximumInterval()) < common.DurationValue(policy.GetInitialInterval()) {
		return serviceerror.NewInvalidArgument("MaximumInterval cannot be less than InitialInterval on retry policy.")
	}
	if policy.GetMaximumAttempts() < 0 {
		return serviceerror.NewInvalidArgument("MaximumAttempts cannot be negative on retry policy.")
	}
	return nil
}

func (env *testWorkflowEnvironmentImpl) getActivityHandle(activityID, runID string) (*testActivityHandle, bool) {
	handle, ok := env.activities[env.makeUniqueActivityID(activityID, runID)]
	return handle, ok
}

func (env *testWorkflowEnvironmentImpl) setActivityHandle(activityID, runID string, handle *testActivityHandle) {
	env.activities[env.makeUniqueActivityID(activityID, runID)] = handle
}

func (env *testWorkflowEnvironmentImpl) deleteHandle(activityID, runID string) {
	delete(env.activities, env.makeUniqueActivityID(activityID, runID))
}

func (env *testWorkflowEnvironmentImpl) makeUniqueActivityID(activityID, runID string) string {
	// ActivityID is unique per workflow, but different workflow could have same activityID.
	// Make the key unique globally as we share the same collection for all running workflows in test.
	return fmt.Sprintf("%v_%v", runID, activityID)
}

func (env *testWorkflowEnvironmentImpl) executeActivityWithRetryForTest(
	taskHandler ActivityTaskHandler,
	parameters ExecuteActivityParams,
	task *workflowservice.PollActivityTaskQueueResponse,
) (result interface{}) {
	var expireTime time.Time
	if parameters.ScheduleToCloseTimeout > 0 {
		expireTime = env.Now().Add(parameters.ScheduleToCloseTimeout)
	}

	for {
		var err error
		result, err = taskHandler.Execute(parameters.TaskQueueName, task)
		if err != nil {
			if err == context.DeadlineExceeded {
				return err
			}
			panic(err)
		}

		// check if a retry is needed
		if request, ok := result.(*workflowservice.RespondActivityTaskFailedRequest); ok && parameters.RetryPolicy != nil {
			p := fromProtoRetryPolicy(parameters.RetryPolicy)
			backoff := getRetryBackoffWithNowTime(p, task.GetAttempt(), ConvertFailureToError(request.GetFailure(), env.GetDataConverter()), env.Now(), expireTime)
			if backoff > 0 {
				// need a retry
				waitCh := make(chan struct{})

				// register the delayed call back first, otherwise other timers may be fired before the retry timer
				// is enqueued.
				env.registerDelayedCallback(func() {
					env.runningCount++
					task.Attempt = task.GetAttempt() + 1
					activityID := ActivityID{id: string(task.TaskToken)}
					if ah, ok := env.getActivityHandle(activityID.id, task.WorkflowExecution.RunId); ok {
						task.HeartbeatDetails = ah.heartbeatDetails
					}
					close(waitCh)
				}, backoff)
				env.postCallback(func() { env.runningCount-- }, false)

				<-waitCh
				continue
			}
		}

		// no retry
		break
	}

	return
}

func fromProtoRetryPolicy(p *commonpb.RetryPolicy) *RetryPolicy {
	return &RetryPolicy{
		InitialInterval:        common.DurationValue(p.GetInitialInterval()),
		BackoffCoefficient:     p.GetBackoffCoefficient(),
		MaximumInterval:        common.DurationValue(p.GetMaximumInterval()),
		MaximumAttempts:        p.GetMaximumAttempts(),
		NonRetryableErrorTypes: p.NonRetryableErrorTypes,
	}
}

func getRetryBackoffFromProtoRetryPolicy(prp *commonpb.RetryPolicy, attempt int32, err error, now, expireTime time.Time) time.Duration {
	if prp == nil {
		return noRetryBackoff
	}

	p := fromProtoRetryPolicy(prp)
	return getRetryBackoffWithNowTime(p, attempt, err, now, expireTime)
}

func ensureDefaultRetryPolicy(parameters *ExecuteActivityParams) {
	// ensure default retry policy
	if parameters.RetryPolicy == nil {
		parameters.RetryPolicy = &commonpb.RetryPolicy{}
	}

	if parameters.RetryPolicy.InitialInterval == nil || *parameters.RetryPolicy.InitialInterval == 0 {
		parameters.RetryPolicy.InitialInterval = common.DurationPtr(time.Second)
	}
	if parameters.RetryPolicy.MaximumInterval == nil || *parameters.RetryPolicy.MaximumInterval == 0 {
		parameters.RetryPolicy.MaximumInterval = common.DurationPtr(*parameters.RetryPolicy.InitialInterval)
	}
	if parameters.RetryPolicy.BackoffCoefficient == 0 {
		parameters.RetryPolicy.BackoffCoefficient = 2
	}

	// NOTE: the default MaximumAttempts for retry policy set by server is 0 which means unlimited retries.
	// However, unlimited retry with automatic fast forward clock in test framework will cause the CPU to spin and test
	// to go forever. So we need to set a reasonable default max attempts for unit test.
	if parameters.RetryPolicy.MaximumAttempts == 0 {
		parameters.RetryPolicy.MaximumAttempts = defaultMaximumAttemptsForUnitTest
	}
}

func (env *testWorkflowEnvironmentImpl) ExecuteLocalActivity(params ExecuteLocalActivityParams, callback LocalActivityResultHandler) LocalActivityID {
	activityID := getStringID(env.nextID())
	ae := &activityExecutor{name: getActivityFunctionName(env.registry, params.ActivityFn), fn: params.ActivityFn}
	if at, _ := getValidatedActivityFunction(params.ActivityFn, params.InputArgs, env.registry); at != nil {
		// local activity could be registered, if so use the registered name. This name is only used to find a mock.
		ae.name = at.Name
	}
	// We have to skip the interceptors on the first call because
	// ExecuteWithActualArgs is actually invoked twice to support a mock activity
	// function result
	ae.skipInterceptors = true
	aew := &activityExecutorWrapper{activityExecutor: ae, env: env}

	// substitute the local activity function so we could replace with mock if it is supplied.
	params.ActivityFn = func(ctx context.Context, inputArgs ...interface{}) (*commonpb.Payloads, error) {
		return aew.ExecuteWithActualArgs(ctx, params.InputArgs)
	}

	task := newLocalActivityTask(params, callback, activityID)
	taskHandler := localActivityTaskHandler{
		userContext:        env.workerOptions.BackgroundActivityContext,
		metricsHandler:     env.metricsHandler,
		logger:             env.logger,
		dataConverter:      env.dataConverter,
		contextPropagators: env.contextPropagators,
		interceptors:       env.registry.interceptors,
	}

	env.localActivities[activityID] = task
	env.runningCount++

	go func() {
		result := taskHandler.executeLocalActivityTask(task)
		env.postCallback(func() {
			env.handleLocalActivityResult(result)
			env.runningCount--
		}, false)
	}()

	return LocalActivityID{id: activityID}
}

func (env *testWorkflowEnvironmentImpl) RequestCancelLocalActivity(activityID LocalActivityID) {
	task, ok := env.localActivities[activityID.id]
	if !ok {
		env.logger.Debug("RequestCancelLocalActivity failed, LocalActivity not exists or already completed.", tagActivityID, activityID)
		return
	}
	env.logger.Debug("RequestCancelLocalActivity", tagActivityID, activityID)
	task.cancel()
}

func (env *testWorkflowEnvironmentImpl) handleActivityResult(activityID ActivityID, result interface{}, activityType string,
	dataConverter converter.DataConverter) {
	env.logger.Debug(fmt.Sprintf("handleActivityResult: %T.", result),
		tagActivityID, activityID, tagActivityType, activityType)
	activityInfo := env.getActivityInfo(activityID, activityType)
	if result == ErrActivityResultPending {
		// In case activity returns ErrActivityResultPending, the respond will be nil, and we don't need to do anything.
		// Activity will need to complete asynchronously using CompleteActivity().
		if env.onActivityCompletedListener != nil {
			env.onActivityCompletedListener(activityInfo, nil, ErrActivityResultPending)
		}
		return
	}

	// this is running in dispatcher
	activityHandle, ok := env.getActivityHandle(activityID.id, activityInfo.WorkflowExecution.RunID)
	if !ok {
		env.logger.Debug("handleActivityResult: ActivityID not exists, could be already completed or canceled.",
			tagActivityID, activityID)
		return
	}

	env.deleteHandle(activityID.id, activityInfo.WorkflowExecution.RunID)

	var blob *commonpb.Payloads
	var err error

	switch request := result.(type) {
	case *workflowservice.RespondActivityTaskCanceledRequest:
		details := newEncodedValues(request.Details, dataConverter)
		err = env.wrapActivityError(
			activityID,
			activityType,
			enumspb.RETRY_STATE_NON_RETRYABLE_FAILURE,
			NewCanceledError(details),
		)
		activityHandle.callback(nil, err)
	case *workflowservice.RespondActivityTaskFailedRequest:
		err = env.wrapActivityError(
			activityID,
			activityType,
			enumspb.RETRY_STATE_UNSPECIFIED,
			ConvertFailureToError(request.GetFailure(), dataConverter),
		)
		activityHandle.callback(nil, err)
	case *workflowservice.RespondActivityTaskCompletedRequest:
		blob = request.Result
		activityHandle.callback(blob, nil)
	default:
		if result == context.DeadlineExceeded {
			err = env.wrapActivityError(
				activityID,
				activityType,
				enumspb.RETRY_STATE_TIMEOUT,
				NewTimeoutError("Activity timeout", enumspb.TIMEOUT_TYPE_START_TO_CLOSE, context.DeadlineExceeded),
			)
			activityHandle.callback(nil, err)
		} else {
			panic(fmt.Sprintf("unsupported respond type %T", result))
		}
	}

	if env.onActivityCompletedListener != nil {
		if err != nil {
			env.onActivityCompletedListener(activityInfo, nil, err)
		} else {
			env.onActivityCompletedListener(activityInfo, newEncodedValue(blob, dataConverter), nil)
		}
	}

	env.startWorkflowTask()
}

func (env *testWorkflowEnvironmentImpl) wrapActivityError(activityID ActivityID, activityType string, retryState enumspb.RetryState, activityErr error) error {
	if activityErr == nil {
		return nil
	}

	return NewActivityError(
		0,
		0,
		env.identity,
		&commonpb.ActivityType{Name: activityType},
		activityID.id,
		retryState,
		activityErr,
	)
}

func (env *testWorkflowEnvironmentImpl) handleLocalActivityResult(result *localActivityResult) {
	activityID := ActivityID{id: result.task.activityID}
	activityType := getActivityFunctionName(env.registry, result.task.params.ActivityFn)
	env.logger.Debug(fmt.Sprintf("handleLocalActivityResult: Err: %v, Result: %v.", result.err, result.result),
		tagActivityID, activityID, tagActivityType, activityType)

	activityInfo := env.getActivityInfo(activityID, activityType)
	task, ok := env.localActivities[activityID.id]
	if !ok {
		env.logger.Debug("handleLocalActivityResult: ActivityID not exists, could be already completed or canceled.",
			tagActivityID, activityID)
		return
	}
	delete(env.localActivities, activityID.id)
	// If error is present do not return value
	if result.err != nil && result.result != nil {
		result.result = nil
	}
	// Always return CanceledError for canceled tasks
	if task.canceled {
		var canceledErr *CanceledError
		if !errors.As(result.err, &canceledErr) {
			result.err = NewCanceledError()
			result.result = nil
		}
	}
	lar := &LocalActivityResultWrapper{
		Err:     env.wrapActivityError(activityID, activityType, enumspb.RETRY_STATE_UNSPECIFIED, result.err),
		Result:  result.result,
		Backoff: noRetryBackoff,
		Attempt: 1,
	}
	if result.task.retryPolicy != nil && result.err != nil {
		lar.Backoff = getRetryBackoff(result, env.Now(), env.dataConverter)
		lar.Attempt = task.attempt
	}
	task.callback(lar)
	var canceledErr *CanceledError
	if errors.As(lar.Err, &canceledErr) {
		if env.onLocalActivityCanceledListener != nil {
			env.onLocalActivityCanceledListener(activityInfo)
		}
	} else if env.onLocalActivityCompletedListener != nil {
		env.onLocalActivityCompletedListener(activityInfo, newEncodedValue(result.result, env.GetDataConverter()), nil)
	}
	env.startWorkflowTask()
}

// runBeforeMockCallReturns is registered as mock call's RunFn by *mock.Call.Run(fn). It will be called by testify's
// mock.MethodCalled() before it returns.
func (env *testWorkflowEnvironmentImpl) runBeforeMockCallReturns(call *MockCallWrapper, args mock.Arguments) {
	var waitDuration time.Duration
	if call.waitDuration != nil {
		waitDuration = call.waitDuration()
	}
	if waitDuration > 0 {
		// we want this mock call to block until the wait duration is elapsed (on workflow clock).
		waitCh := make(chan time.Time)
		env.registerDelayedCallback(func() {
			env.runningCount++  // increase runningCount as the mock call is ready to resume.
			waitCh <- env.Now() // this will unblock mock call
		}, waitDuration)

		// make sure decrease runningCount after delayed callback is posted
		env.postCallback(func() {
			env.runningCount-- // reduce runningCount, since this mock call is about to be blocked.
		}, false)
		<-waitCh // this will block until mock clock move forward by waitDuration
	}

	// run the actual runFn if it was setup
	if call.runFn != nil {
		call.runFn(args)
	}
}

// Execute executes the activity code.
func (a *activityExecutorWrapper) Execute(ctx context.Context, input *commonpb.Payloads) (*commonpb.Payloads, error) {
	activityInfo := GetActivityInfo(ctx)
	// If the activity was cancelled before it starts here, we do not execute and
	// instead return cancelled
	a.env.locker.Lock()
	_, handleExists := a.env.getActivityHandle(activityInfo.ActivityID, activityInfo.WorkflowExecution.RunID)
	a.env.locker.Unlock()
	if !handleExists {
		return nil, NewCanceledError()
	}

	dc := getDataConverterFromActivityCtx(ctx)
	if a.env.onActivityStartedListener != nil {
		waitCh := make(chan struct{})
		a.env.postCallback(func() {
			a.env.onActivityStartedListener(&activityInfo, ctx, newEncodedValues(input, dc))
			close(waitCh)
		}, false)
		<-waitCh // wait until listener returns
	}

	m := &mockWrapper{env: a.env, name: a.name, fn: a.fn, isWorkflow: false, dataConverter: dc}
	if mockRet := m.getMockReturn(ctx, input); mockRet != nil {
		return m.executeMock(ctx, input, mockRet)
	}

	return a.activityExecutor.Execute(ctx, input)
}

// ExecuteWithActualArgs executes the activity code.
func (a *activityExecutorWrapper) ExecuteWithActualArgs(ctx context.Context, inputArgs []interface{}) (*commonpb.Payloads, error) {
	activityInfo := GetActivityInfo(ctx)
	if a.env.onLocalActivityStartedListener != nil {
		waitCh := make(chan struct{})
		a.env.postCallback(func() {
			a.env.onLocalActivityStartedListener(&activityInfo, ctx, inputArgs)
			close(waitCh)
		}, false)
		<-waitCh
	}

	m := &mockWrapper{env: a.env, name: a.name, fn: a.fn, isWorkflow: false}
	if mockRet := m.getMockReturnWithActualArgs(ctx, inputArgs); mockRet != nil {
		// check if mock returns function which must match to the actual function.
		if mockFn := m.getMockFn(mockRet); mockFn != nil {
			executor := &activityExecutor{name: m.name, fn: mockFn}
			return executor.ExecuteWithActualArgs(ctx, inputArgs)
		}
		return m.getMockValue(mockRet)
	}

	return a.activityExecutor.ExecuteWithActualArgs(ctx, inputArgs)
}

// Execute executes the workflow code.
func (w *workflowExecutorWrapper) Execute(ctx Context, input *commonpb.Payloads) (result *commonpb.Payloads, err error) {
	env := w.env
	if env.isChildWorkflow() && env.onChildWorkflowStartedListener != nil {
		env.onChildWorkflowStartedListener(GetWorkflowInfo(ctx), ctx, newEncodedValues(input, w.env.GetDataConverter()))
	}

	if !env.isChildWorkflow() {
		// This is to prevent auto-forwarding mock clock before main workflow starts. For child workflow, we increase
		// the counter in env.ExecuteChildWorkflow(). We cannot do it here for child workflow, because we need to make
		// sure the counter is increased before returning from ExecuteChildWorkflow().
		env.runningCount++
	}

	m := &mockWrapper{env: env, name: w.workflowType, fn: w.fn, isWorkflow: true, dataConverter: env.GetDataConverter()}
	// This method is called by workflow's dispatcher. In this test suite, it is run in the main loop. We cannot block
	// the main loop, but the mock could block if it is configured to wait. So we need to use a separate goroutinue to
	// run the mock, and resume after mock call returns.
	mockReadyChannel := NewChannel(ctx)
	// make a copy of the context for getMockReturn() call to avoid race condition
	_, ctxCopy, err := newWorkflowContext(w.env, nil)
	if err != nil {
		return nil, err
	}
	go func() {
		// getMockReturn could block if mock is configured to wait. The returned mockRet is what has been configured
		// for the mock by using MockCallWrapper.Return(). The mockRet could be mock values or mock function. We process
		// the returned mockRet by calling executeMock() later in the main thread after it is send over via mockReadyChannel.
		mockRet := m.getMockReturn(ctxCopy, input)
		env.postCallback(func() {
			mockReadyChannel.SendAsync(mockRet)
		}, true /* true to trigger the dispatcher for this workflow so it resume from mockReadyChannel block*/)
	}()

	var mockRet mock.Arguments
	// This will block workflow dispatcher (on temporal channel), which the dispatcher understand and will return from
	// ExecuteUntilAllBlocked() so the main loop is not blocked. The dispatcher will unblock when getMockReturn() returns.
	mockReadyChannel.Receive(ctx, &mockRet)

	// reduce runningCount to allow auto-forwarding mock clock after current workflow dispatcher run is blocked (aka
	// ExecuteUntilAllBlocked() returns).
	env.runningCount--

	childWE := env.workflowInfo.WorkflowExecution
	var startedErr error
	if mockRet != nil {
		// workflow was mocked.
		result, err = m.executeMock(ctx, input, mockRet)
		if env.isChildWorkflow() && err == ErrMockStartChildWorkflowFailed {
			childWE, startedErr = WorkflowExecution{}, err
		}
	}

	if env.isChildWorkflow() && env.startedHandler != nil /* startedHandler could be nil for retry */ {
		// notify parent that child workflow is started
		env.parentEnv.postCallback(func() {
			env.startedHandler(childWE, startedErr)
		}, true)
	}

	if mockRet != nil {
		return result, err
	}

	// no mock, so call the actual workflow
	return w.workflowExecutor.Execute(ctx, input)
}

func (m *mockWrapper) getCtxArg(ctx interface{}) []interface{} {
	fnType := reflect.TypeOf(m.fn)
	if fnType.NumIn() > 0 {
		if (!m.isWorkflow && isActivityContext(fnType.In(0))) ||
			(m.isWorkflow && isWorkflowContext(fnType.In(0))) {
			return []interface{}{ctx}
		}
	}
	return nil
}

func (m *mockWrapper) getMockReturn(ctx interface{}, input *commonpb.Payloads) (retArgs mock.Arguments) {
	if _, ok := m.env.expectedMockCalls[m.name]; !ok {
		// no mock
		return nil
	}

	fnType := reflect.TypeOf(m.fn)
	reflectArgs, err := decodeArgs(m.dataConverter, fnType, input)
	if err != nil {
		panic(fmt.Sprintf("Decode error: %v in %v of type %T", err.Error(), m.name, m.fn))
	}
	realArgs := m.getCtxArg(ctx)
	for _, arg := range reflectArgs {
		realArgs = append(realArgs, arg.Interface())
	}

	return m.env.mock.MethodCalled(m.name, realArgs...)
}

func (m *mockWrapper) getMockReturnWithActualArgs(ctx interface{}, inputArgs []interface{}) (retArgs mock.Arguments) {
	if _, ok := m.env.expectedMockCalls[m.name]; !ok {
		// no mock
		return nil
	}

	realArgs := m.getCtxArg(ctx)
	realArgs = append(realArgs, inputArgs...)
	return m.env.mock.MethodCalled(m.name, realArgs...)
}

func (m *mockWrapper) getMockFn(mockRet mock.Arguments) interface{} {
	fnName := m.name
	mockRetLen := len(mockRet)
	if mockRetLen == 0 {
		panic(fmt.Sprintf("mock of %v has no returns", fnName))
	}

	fnType := reflect.TypeOf(m.fn)
	// check if mock returns function which must match to the actual function.
	mockFn := mockRet.Get(0)
	mockFnType := reflect.TypeOf(mockFn)
	if mockFnType != nil && mockFnType.Kind() == reflect.Func {
		if mockFnType != fnType {
			fnName, _ := getFunctionName(m.fn)
			// mockDummyActivity is used to register mocks by name
			if fnName != "mockDummyActivity" {
				panic(fmt.Sprintf("mock of %v has incorrect return function, expected %v, but actual is %v",
					fnName, fnType, mockFnType))
			}
		}
		return mockFn
	}
	return nil
}

func (m *mockWrapper) getMockValue(mockRet mock.Arguments) (*commonpb.Payloads, error) {
	fnName := m.name
	mockRetLen := len(mockRet)
	fnType := reflect.TypeOf(m.fn)
	// check if mockRet have same types as function's return types
	if mockRetLen != fnType.NumOut() {
		panic(fmt.Sprintf("mock of %v has incorrect number of returns, expected %d, but actual is %d",
			fnName, fnType.NumOut(), mockRetLen))
	}
	// we already verified function either has 1 return value (error) or 2 return values (result, error)
	var retErr error
	mockErr := mockRet[mockRetLen-1] // last mock return must be error
	if mockErr == nil {
		retErr = nil
	} else if err, ok := mockErr.(error); ok {
		retErr = err
	} else {
		panic(fmt.Sprintf("mock of %v has incorrect return type, expected error, but actual is %T (%v)",
			fnName, mockErr, mockErr))
	}

	switch mockRetLen {
	case 1:
		return nil, retErr
	case 2:
		expectedType := fnType.Out(0)
		mockResult := mockRet[0]
		if mockResult == nil {
			switch expectedType.Kind() {
			case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Array:
				// these are supported nil-able types. (reflect.Chan, reflect.Func are nil-able, but not supported)
				return nil, retErr
			default:
				panic(fmt.Sprintf("mock of %v has incorrect return type, expected %v, but actual is %T (%v)",
					fnName, expectedType, mockResult, mockResult))
			}
		} else {
			if !reflect.TypeOf(mockResult).AssignableTo(expectedType) {
				panic(fmt.Sprintf("mock of %v has incorrect return type, expected %v, but actual is %T (%v)",
					fnName, expectedType, mockResult, mockResult))
			}
			result, encodeErr := encodeArg(m.env.GetDataConverter(), mockResult)
			if encodeErr != nil {
				panic(fmt.Sprintf("encode result from mock of %v failed: %v", fnName, encodeErr))
			}
			return result, retErr
		}
	default:
		// this will never happen, panic just in case
		panic("mock should either have 1 return value (error) or 2 return values (result, error)")
	}
}

func (m *mockWrapper) executeMock(ctx interface{}, input *commonpb.Payloads, mockRet mock.Arguments) (result *commonpb.Payloads, err error) {
	// have to handle panics here to support calling ExecuteChildWorkflow(...).GetChildWorkflowExecution().Get(...)
	// when a child is mocked.
	defer func() {
		if r := recover(); r != nil {
			st := getStackTrace("executeMock", "panic", 4)
			err = newPanicError(r, st)
		}
	}()

	fnName := m.name
	// check if mock returns function which must match to the actual function.
	if mockFn := m.getMockFn(mockRet); mockFn != nil {
		// we found a mock function that matches to actual function, so call that mockFn
		if m.isWorkflow {
			executor := &workflowExecutor{workflowType: fnName, fn: mockFn}
			return executor.Execute(ctx.(Context), input)
		}
		executor := &activityExecutor{name: fnName, fn: mockFn}
		return executor.Execute(ctx.(context.Context), input)
	}

	return m.getMockValue(mockRet)
}

func (env *testWorkflowEnvironmentImpl) newTestActivityTaskHandler(taskQueue string, dataConverter converter.DataConverter) ActivityTaskHandler {
	setWorkerOptionsDefaults(&env.workerOptions)
	params := workerExecutionParameters{
		TaskQueue:          taskQueue,
		Identity:           env.identity,
		MetricsHandler:     env.metricsHandler,
		Logger:             env.logger,
		UserContext:        env.workerOptions.BackgroundActivityContext,
		DataConverter:      dataConverter,
		WorkerStopChannel:  env.workerStopChannel,
		ContextPropagators: env.contextPropagators,
	}
	ensureRequiredParams(&params)
	if params.UserContext == nil {
		params.UserContext = context.Background()
	}
	if env.workerOptions.EnableSessionWorker && env.sessionEnvironment == nil {
		env.sessionEnvironment = newTestSessionEnvironment(env, &params, env.workerOptions.MaxConcurrentSessionExecutionSize)
	}
	params.UserContext = context.WithValue(params.UserContext, sessionEnvironmentContextKey, env.sessionEnvironment)
	registry := env.registry
	if len(registry.getRegisteredActivities()) == 0 {
		panic(fmt.Sprintf("no activity is registered for taskqueue '%v'", taskQueue))
	}

	getActivity := func(name string) activity {
		tlsa, ok := env.taskQueueSpecificActivities[name]
		if ok {
			_, ok := tlsa.taskQueues[taskQueue]
			if !ok {
				// activity are bind to specific task queue but not to current task queue
				return nil
			}
		}

		activity, ok := registry.GetActivity(name)
		if !ok {
			return nil
		}
		ae := &activityExecutor{name: activity.ActivityType().Name, fn: activity.GetFunction()}

		if env.sessionEnvironment != nil {
			// Special handling for session creation and completion activities.
			// If real creation activity is used, it will block timers from autofiring.
			if ae.name == sessionCreationActivityName {
				ae.fn = sessionCreationActivityForTest
			}
			if ae.name == sessionCompletionActivityName {
				ae.fn = sessionCompletionActivityForTest
			}
		}
		return &activityExecutorWrapper{activityExecutor: ae, env: env}
	}

	taskHandler := newActivityTaskHandlerWithCustomProvider(env.service, params, registry, getActivity)
	return taskHandler
}

func newTestActivityTask(workflowID, runID, workflowTypeName, namespace string,
	attr *commandpb.ScheduleActivityTaskCommandAttributes) *workflowservice.PollActivityTaskQueueResponse {
	activityID := attr.GetActivityId()
	now := time.Now()
	task := &workflowservice.PollActivityTaskQueueResponse{
		Attempt: 1,
		WorkflowExecution: &commonpb.WorkflowExecution{
			WorkflowId: workflowID,
			RunId:      runID,
		},
		ActivityId:             activityID,
		TaskToken:              []byte(activityID), // use activityID as TaskToken so we can map TaskToken in heartbeat calls.
		ActivityType:           &commonpb.ActivityType{Name: attr.GetActivityType().GetName()},
		Input:                  attr.GetInput(),
		ScheduledTime:          &now,
		ScheduleToCloseTimeout: attr.GetScheduleToCloseTimeout(),
		StartedTime:            &now,
		StartToCloseTimeout:    attr.GetStartToCloseTimeout(),
		HeartbeatTimeout:       attr.GetHeartbeatTimeout(),
		WorkflowType: &commonpb.WorkflowType{
			Name: workflowTypeName,
		},
		WorkflowNamespace: namespace,
		Header:            attr.GetHeader(),
	}
	return task
}

func (env *testWorkflowEnvironmentImpl) newTimer(d time.Duration, callback ResultHandler, notifyListener bool) *TimerID {
	nextID := env.nextID()
	timerInfo := &TimerID{id: getStringID(nextID)}
	timer := env.mockClock.AfterFunc(d, func() {
		delete(env.timers, timerInfo.id)
		env.postCallback(func() {
			callback(nil, nil)
			if notifyListener && env.onTimerFiredListener != nil {
				env.onTimerFiredListener(timerInfo.id)
			}
		}, true)
	})
	env.timers[timerInfo.id] = &testTimerHandle{
		env:            env,
		callback:       callback,
		timer:          timer,
		mockTimeToFire: env.mockClock.Now().Add(d),
		wallTimeToFire: env.wallClock.Now().Add(d),
		duration:       d,
		timerID:        nextID,
	}
	if notifyListener && env.onTimerScheduledListener != nil {
		env.onTimerScheduledListener(timerInfo.id, d)
	}
	return timerInfo
}

func (env *testWorkflowEnvironmentImpl) NewTimer(d time.Duration, callback ResultHandler) *TimerID {
	return env.newTimer(d, callback, true)
}

func (env *testWorkflowEnvironmentImpl) Now() time.Time {
	return env.mockClock.Now()
}

func (env *testWorkflowEnvironmentImpl) WorkflowInfo() *WorkflowInfo {
	return env.workflowInfo
}

func (env *testWorkflowEnvironmentImpl) RegisterWorkflow(w interface{}) {
	env.registry.RegisterWorkflow(w)
}

func (env *testWorkflowEnvironmentImpl) RegisterWorkflowWithOptions(w interface{}, options RegisterWorkflowOptions) {
	env.registry.RegisterWorkflowWithOptions(w, options)
}

func (env *testWorkflowEnvironmentImpl) RegisterActivity(a interface{}) {
	env.registry.RegisterActivityWithOptions(a, RegisterActivityOptions{DisableAlreadyRegisteredCheck: true})
}

func (env *testWorkflowEnvironmentImpl) RegisterActivityWithOptions(a interface{}, options RegisterActivityOptions) {
	options.DisableAlreadyRegisteredCheck = true
	env.registry.RegisterActivityWithOptions(a, options)
}

func (env *testWorkflowEnvironmentImpl) RegisterCancelHandler(handler func()) {
	env.workflowCancelHandler = handler
}

func (env *testWorkflowEnvironmentImpl) RegisterSignalHandler(
	handler func(name string, input *commonpb.Payloads, header *commonpb.Header) error,
) {
	env.signalHandler = handler
}

func (env *testWorkflowEnvironmentImpl) RegisterQueryHandler(
	handler func(string, *commonpb.Payloads, *commonpb.Header) (*commonpb.Payloads, error),
) {
	env.queryHandler = handler
}

func (env *testWorkflowEnvironmentImpl) RequestCancelChildWorkflow(_, workflowID string) {
	if childHandle, ok := env.runningWorkflows[workflowID]; ok && !childHandle.handled {
		// current workflow is a parent workflow, and we are canceling a child workflow
		childEnv := childHandle.env
		childEnv.cancelWorkflow(func(result *commonpb.Payloads, err error) {})
		return
	}
}

func (env *testWorkflowEnvironmentImpl) RequestCancelExternalWorkflow(namespace, workflowID, runID string, callback ResultHandler) {
	if env.workflowInfo.WorkflowExecution.ID == workflowID {
		// cancel current workflow
		env.workflowCancelHandler()
		// check if current workflow is a child workflow
		if env.isChildWorkflow() && env.onChildWorkflowCanceledListener != nil {
			env.postCallback(func() {
				env.onChildWorkflowCanceledListener(env.workflowInfo)
			}, false)
		}
		return
	} else if childHandle, ok := env.runningWorkflows[workflowID]; ok && !childHandle.handled {
		// current workflow is a parent workflow, and we are canceling a child workflow
		if !childHandle.params.WaitForCancellation {
			childHandle.env.Complete(nil, ErrCanceled)
		}
		childEnv := childHandle.env
		env.postCallback(func() {
			callback(nil, nil)
		}, true)
		childEnv.cancelWorkflow(callback)
		return
	}

	// target workflow is not child workflow, we need the mock. The mock needs to be called in a separate goroutinue
	// so it can block and wait on the requested delay time (if configured). If we run it in main thread, and the mock
	// configured to delay, it will block the main loop which stops the world.
	env.runningCount++
	go func() {
		args := []interface{}{namespace, workflowID, runID}
		// below call will panic if mock is not properly setup.
		mockRet := env.mock.MethodCalled(mockMethodForRequestCancelExternalWorkflow, args...)
		m := &mockWrapper{name: mockMethodForRequestCancelExternalWorkflow, fn: mockFnRequestCancelExternalWorkflow}
		var err error
		if mockFn := m.getMockFn(mockRet); mockFn != nil {
			_, err = executeFunctionWithContext(context.TODO(), mockFn, args)
		} else {
			_, err = m.getMockValue(mockRet)
		}
		env.postCallback(func() {
			callback(nil, err)
			env.runningCount--
		}, true)
	}()
}

func (env *testWorkflowEnvironmentImpl) IsReplaying() bool {
	// this test environment never replay
	return false
}

func (env *testWorkflowEnvironmentImpl) SignalExternalWorkflow(
	namespace string,
	workflowID string,
	runID string,
	signalName string,
	input *commonpb.Payloads,
	arg interface{},
	header *commonpb.Header,
	childWorkflowOnly bool,
	callback ResultHandler,
) {
	// check if target workflow is a known workflow
	if childHandle, ok := env.runningWorkflows[workflowID]; ok {
		// target workflow is a child
		childEnv := childHandle.env
		if childEnv.isWorkflowCompleted {
			// child already completed (NOTE: we have only one failed cause now)
			err := newUnknownExternalWorkflowExecutionError()
			callback(nil, err)
		} else {
			err := childEnv.signalHandler(signalName, input, header)
			callback(nil, err)
		}
		childEnv.postCallback(func() {}, true) // resume child workflow since a signal is sent.
		return
	}

	// here we signal a child workflow but we cannot find it
	if childWorkflowOnly {
		err := newUnknownExternalWorkflowExecutionError()
		callback(nil, err)
		return
	}

	// target workflow is not child workflow, we need the mock. The mock needs to be called in a separate goroutinue
	// so it can block and wait on the requested delay time (if configured). If we run it in main thread, and the mock
	// configured to delay, it will block the main loop which stops the world.
	env.runningCount++
	go func() {
		args := []interface{}{namespace, workflowID, runID, signalName, arg}
		// below call will panic if mock is not properly setup.
		mockRet := env.mock.MethodCalled(mockMethodForSignalExternalWorkflow, args...)
		m := &mockWrapper{name: mockMethodForSignalExternalWorkflow, fn: mockFnSignalExternalWorkflow}
		var err error
		if mockFn := m.getMockFn(mockRet); mockFn != nil {
			_, err = executeFunctionWithContext(context.TODO(), mockFn, args)
		} else {
			_, err = m.getMockValue(mockRet)
		}
		env.postCallback(func() {
			callback(nil, err)
			env.runningCount--
		}, true)
	}()
}

func (env *testWorkflowEnvironmentImpl) ExecuteChildWorkflow(params ExecuteWorkflowParams, callback ResultHandler, startedHandler func(r WorkflowExecution, e error)) {
	env.executeChildWorkflowWithDelay(0, params, callback, startedHandler)
}

func (env *testWorkflowEnvironmentImpl) executeChildWorkflowWithDelay(delayStart time.Duration, params ExecuteWorkflowParams, callback ResultHandler, startedHandler func(r WorkflowExecution, e error)) {
	childEnv, err := env.newTestWorkflowEnvironmentForChild(&params, callback, startedHandler)
	if err != nil {
		env.logger.Info("ExecuteChildWorkflow failed", tagError, err)
		callback(nil, err)
		startedHandler(WorkflowExecution{}, err)
		return
	}

	env.logger.Info("ExecuteChildWorkflow", tagWorkflowType, params.WorkflowType.Name)
	env.runningCount++

	// run child workflow in separate goroutinue
	go childEnv.executeWorkflowInternal(delayStart, params.WorkflowType.Name, params.Input)
}

func (env *testWorkflowEnvironmentImpl) SideEffect(f func() (*commonpb.Payloads, error), callback ResultHandler) {
	callback(f())
}

func (env *testWorkflowEnvironmentImpl) GetVersion(changeID string, minSupported, maxSupported Version) (retVersion Version) {
	if mockVersion, ok := env.getMockedVersion(changeID, changeID, minSupported, maxSupported); ok {
		// GetVersion for changeID is mocked
		_ = env.UpsertSearchAttributes(createSearchAttributesForChangeVersion(changeID, mockVersion, env.changeVersions))
		env.changeVersions[changeID] = mockVersion
		return mockVersion
	}
	if mockVersion, ok := env.getMockedVersion(mock.Anything, changeID, minSupported, maxSupported); ok {
		// GetVersion is mocked with any changeID.
		_ = env.UpsertSearchAttributes(createSearchAttributesForChangeVersion(changeID, mockVersion, env.changeVersions))
		env.changeVersions[changeID] = mockVersion
		return mockVersion
	}

	// no mock setup, so call regular path
	if version, ok := env.changeVersions[changeID]; ok {
		validateVersion(changeID, version, minSupported, maxSupported)
		return version
	}
	_ = env.UpsertSearchAttributes(createSearchAttributesForChangeVersion(changeID, maxSupported, env.changeVersions))
	env.changeVersions[changeID] = maxSupported
	return maxSupported
}

func (env *testWorkflowEnvironmentImpl) getMockedVersion(mockedChangeID, changeID string, minSupported, maxSupported Version) (Version, bool) {
	mockMethod := getMockMethodForGetVersion(mockedChangeID)
	if _, ok := env.expectedMockCalls[mockMethod]; !ok {
		// mock not found
		return DefaultVersion, false
	}

	args := []interface{}{changeID, minSupported, maxSupported}
	// below call will panic if mock is not properly setup.
	mockRet := env.mock.MethodCalled(mockMethod, args...)
	m := &mockWrapper{name: mockMethodForGetVersion, fn: mockFnGetVersion}
	if mockFn := m.getMockFn(mockRet); mockFn != nil {
		var reflectArgs []reflect.Value
		// Add context if first param
		if fnType := reflect.TypeOf(mockFn); fnType.NumIn() > 0 && isActivityContext(fnType.In(0)) {
			reflectArgs = append(reflectArgs, reflect.ValueOf(context.TODO()))
		}
		for _, arg := range args {
			reflectArgs = append(reflectArgs, reflect.ValueOf(arg))
		}
		reflectValues := reflect.ValueOf(mockFn).Call(reflectArgs)
		if len(reflectValues) != 1 || !reflect.TypeOf(reflectValues[0].Interface()).AssignableTo(reflect.TypeOf(DefaultVersion)) {
			panic(fmt.Sprintf("mock of GetVersion has incorrect return type, expected workflow.Version, but actual is %T (%v)",
				reflectValues[0].Interface(), reflectValues[0].Interface()))
		}
		return reflectValues[0].Interface().(Version), true
	}

	if len(mockRet) != 1 || !reflect.TypeOf(mockRet[0]).AssignableTo(reflect.TypeOf(DefaultVersion)) {
		panic(fmt.Sprintf("mock of GetVersion has incorrect return type, expected workflow.Version, but actual is %T (%v)",
			mockRet[0], mockRet[0]))
	}
	return mockRet[0].(Version), true
}

func getMockMethodForGetVersion(changeID string) string {
	return fmt.Sprintf("%v_%v", mockMethodForGetVersion, changeID)
}

func (env *testWorkflowEnvironmentImpl) UpsertSearchAttributes(attributes map[string]interface{}) error {
	attr, err := validateAndSerializeSearchAttributes(attributes)

	env.workflowInfo.SearchAttributes = mergeSearchAttributes(env.workflowInfo.SearchAttributes, attr)

	mockMethod := mockMethodForUpsertSearchAttributes
	if _, ok := env.expectedMockCalls[mockMethod]; !ok {
		// mock not found
		return err
	}

	args := []interface{}{attributes}
	env.mock.MethodCalled(mockMethod, args...)

	return err
}

func (env *testWorkflowEnvironmentImpl) MutableSideEffect(_ string, f func() interface{}, _ func(a, b interface{}) bool) converter.EncodedValue {
	return newEncodedValue(env.encodeValue(f()), env.GetDataConverter())
}

func (env *testWorkflowEnvironmentImpl) AddSession(sessionInfo *SessionInfo) {
	env.openSessions[sessionInfo.SessionID] = sessionInfo
}

func (env *testWorkflowEnvironmentImpl) RemoveSession(sessionID string) {
	delete(env.openSessions, sessionID)
}

func (env *testWorkflowEnvironmentImpl) encodeValue(value interface{}) *commonpb.Payloads {
	blob, err := env.GetDataConverter().ToPayloads(value)
	if err != nil {
		panic(err)
	}
	return blob
}

func (env *testWorkflowEnvironmentImpl) nextID() int64 {
	activityID := env.counterID
	env.counterID++
	return activityID
}

func (env *testWorkflowEnvironmentImpl) getActivityInfo(activityID ActivityID, activityType string) *ActivityInfo {
	return &ActivityInfo{
		ActivityID:        activityID.id,
		ActivityType:      ActivityType{Name: activityType},
		TaskToken:         []byte(activityID.id),
		WorkflowExecution: env.workflowInfo.WorkflowExecution,
		Attempt:           1,
	}
}

func (env *testWorkflowEnvironmentImpl) cancelWorkflow(callback ResultHandler) {
	env.postCallback(func() {
		// RequestCancelWorkflow needs to be run in main thread
		env.RequestCancelExternalWorkflow(
			env.workflowInfo.Namespace,
			env.workflowInfo.WorkflowExecution.ID,
			env.workflowInfo.WorkflowExecution.RunID,
			callback,
		)
	}, true)
}

func (env *testWorkflowEnvironmentImpl) signalWorkflow(name string, input interface{}, startWorkflowTask bool) {
	data, err := encodeArg(env.GetDataConverter(), input)
	if err != nil {
		panic(err)
	}
	env.postCallback(func() {
		// Do not send any headers on test invocations
		_ = env.signalHandler(name, data, nil)
	}, startWorkflowTask)
}

func (env *testWorkflowEnvironmentImpl) signalWorkflowByID(workflowID, signalName string, input interface{}) error {
	data, err := encodeArg(env.GetDataConverter(), input)
	if err != nil {
		panic(err)
	}

	if workflowHandle, ok := env.runningWorkflows[workflowID]; ok {
		if workflowHandle.handled {
			return serviceerror.NewNotFound(fmt.Sprintf("Workflow %v already completed", workflowID))
		}
		workflowHandle.env.postCallback(func() {
			// Do not send any headers on test invocations
			_ = workflowHandle.env.signalHandler(signalName, data, nil)
		}, true)
		return nil
	}

	return serviceerror.NewNotFound(fmt.Sprintf("Workflow %v not exists", workflowID))
}

func (env *testWorkflowEnvironmentImpl) queryWorkflow(queryType string, args ...interface{}) (converter.EncodedValue, error) {
	data, err := encodeArgs(env.GetDataConverter(), args)
	if err != nil {
		return nil, err
	}
	// Do not send any headers on test invocations
	blob, err := env.queryHandler(queryType, data, nil)
	if err != nil {
		return nil, err
	}
	return newEncodedValue(blob, env.GetDataConverter()), nil
}

func (env *testWorkflowEnvironmentImpl) queryWorkflowByID(workflowID, queryType string, args ...interface{}) (converter.EncodedValue, error) {
	if workflowHandle, ok := env.runningWorkflows[workflowID]; ok {
		data, err := encodeArgs(workflowHandle.env.GetDataConverter(), args)
		if err != nil {
			return nil, err
		}
		// Do not send any headers on test invocations
		blob, err := workflowHandle.env.queryHandler(queryType, data, nil)
		if err != nil {
			return nil, err
		}
		return newEncodedValue(blob, workflowHandle.env.GetDataConverter()), nil
	}
	return nil, serviceerror.NewNotFound(fmt.Sprintf("Workflow %v not exists", workflowID))
}

func (env *testWorkflowEnvironmentImpl) getMockRunFn(callWrapper *MockCallWrapper) func(args mock.Arguments) {
	env.locker.Lock()
	defer env.locker.Unlock()

	env.expectedMockCalls[callWrapper.call.Method] = struct{}{}
	return func(args mock.Arguments) {
		env.runBeforeMockCallReturns(callWrapper, args)
	}
}

func (env *testWorkflowEnvironmentImpl) setLastCompletionResult(result interface{}) {
	data, err := encodeArg(env.GetDataConverter(), result)
	if err != nil {
		panic(err)
	}
	env.workflowInfo.lastCompletionResult = data
}

func (env *testWorkflowEnvironmentImpl) setLastError(err error) {
	env.workflowInfo.lastFailure = ConvertErrorToFailure(err, env.dataConverter)
}

func (env *testWorkflowEnvironmentImpl) setHeartbeatDetails(details interface{}) {
	data, err := encodeArg(env.GetDataConverter(), details)
	if err != nil {
		panic(err)
	}
	env.heartbeatDetails = data
}

func (env *testWorkflowEnvironmentImpl) GetRegistry() *registry {
	return env.registry
}

func (env *testWorkflowEnvironmentImpl) setStartWorkflowOptions(options StartWorkflowOptions) {
	wf := env.workflowInfo
	if options.WorkflowExecutionTimeout > 0 {
		wf.WorkflowExecutionTimeout = options.WorkflowExecutionTimeout
	}
	if options.WorkflowRunTimeout > 0 {
		wf.WorkflowRunTimeout = options.WorkflowRunTimeout
	}
	if options.WorkflowTaskTimeout > 0 {
		wf.WorkflowTaskTimeout = options.WorkflowTaskTimeout
	}
	if len(options.ID) > 0 {
		wf.WorkflowExecution.ID = options.ID
	}
	if len(options.TaskQueue) > 0 {
		wf.TaskQueueName = options.TaskQueue
	}
}

func newTestSessionEnvironment(testWorkflowEnvironment *testWorkflowEnvironmentImpl,
	params *workerExecutionParameters, concurrentSessionExecutionSize int) *testSessionEnvironmentImpl {
	resourceID := params.SessionResourceID
	if resourceID == "" {
		resourceID = "testResourceID"
	}
	if concurrentSessionExecutionSize == 0 {
		concurrentSessionExecutionSize = defaultMaxConcurrentSessionExecutionSize
	}

	return &testSessionEnvironmentImpl{
		sessionEnvironmentImpl:  newSessionEnvironment(resourceID, concurrentSessionExecutionSize).(*sessionEnvironmentImpl),
		testWorkflowEnvironment: testWorkflowEnvironment,
	}
}

func (t *testSessionEnvironmentImpl) SignalCreationResponse(_ context.Context, sessionID string) error {
	t.testWorkflowEnvironment.signalWorkflow(sessionID, t.sessionEnvironmentImpl.getCreationResponse(), true)
	return nil
}

// function signature for mock SignalExternalWorkflow
func mockFnSignalExternalWorkflow(string, string, string, string, interface{}) error {
	return nil
}

// function signature for mock RequestCancelExternalWorkflow
func mockFnRequestCancelExternalWorkflow(string, string, string) error {
	return nil
}

// function signature for mock GetVersion
func mockFnGetVersion(string, Version, Version) Version {
	return DefaultVersion
}

// make sure interface is implemented
var _ WorkflowEnvironment = (*testWorkflowEnvironmentImpl)(nil)
