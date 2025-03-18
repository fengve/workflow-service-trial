package core

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	sharedLog "github.com/sugerio/workflow-service-trial/shared/log"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	"golang.org/x/exp/maps"
)

var (
	activeExecutions = &ActiveExecutions{
		waitGroup:         sync.WaitGroup{},
		logger:            sharedLog.GetLogger(context.Background()),
		currentExecutions: make(map[string]*structs.ExecutingWorkflowData),
		lock:              sync.Mutex{},
	}
)

type ActiveExecutions struct {
	waitGroup         sync.WaitGroup
	logger            sharedLog.Logger
	currentExecutions map[string]*structs.ExecutingWorkflowData
	lock              sync.Mutex
}

func GetActiveExecutions() *ActiveExecutions {
	return activeExecutions
}

func ShutdownActiveExecutions() {
	if activeExecutions == nil {
		return
	}

	// Send done signal if all complete
	done := make(chan struct{})
	go func() {
		defer close(done)
		activeExecutions.waitGroup.Wait()
	}()

	// Wait for goroutines, only exit when either done or timeout of 30 seconds is reached.
	timeout := time.After(30 * time.Second)
	select {
	case <-done:
		activeExecutions.logger.Info("ActiveExecutions is closed.")
	case <-timeout:
		activeExecutions.logger.Info("ActiveExecutions is forced to close after timeout.")
	}
}

func (activeExecutions *ActiveExecutions) CurrentExecutions() map[string]*structs.ExecutingWorkflowData {
	activeExecutions.lock.Lock()
	defer activeExecutions.lock.Unlock()
	return maps.Clone(activeExecutions.currentExecutions)
}

func (activeExecutions *ActiveExecutions) AddExecution(
	ctx context.Context,
	executionData *structs.WorkflowExecutionDataProcess,
	executionId int) (int, error) {
	now := time.Now()
	execution := structs.ExecutingWorkflowData{
		ExecutionData: executionData,
		StartedAt:     &now,
		Status:        structs.WorkflowExecutionStatus_New,
	}

	if executionId == 0 {
		// Is a new execution so save in DB
		executionEntity, err := CreateWorkflowExecutionAndData(ctx, &execution)
		if err != nil {
			return 0, err
		}
		executionId = int(executionEntity.ID)
	} else {
		// Is an existing execution we want to finish so update in DB
		workflowExecution, err := GetWorkflowExecution(ctx, int32(executionId))
		if err != nil {
			return 0, err
		}
		workflowExecution.Data = executionData.ExecutionData
		workflowExecution.WaitTill = nil
		workflowExecution.Status = structs.WorkflowExecutionStatus_Running
		err = UpdateWorkflowExecutionEntityAndData(ctx, int32(executionId), workflowExecution)
		if err != nil {
			return 0, err
		}
	}

	execution.Status = structs.WorkflowExecutionStatus_Running
	activeExecutions.setExecution(strconv.Itoa(executionId), &execution)
	return executionId, nil
}

func (activeExecutions *ActiveExecutions) AddTestWebhookExecution(
	executionData *structs.WorkflowExecutionDataProcess) (string, error) {
	now := time.Now()
	execution := structs.ExecutingWorkflowData{
		ExecutionData: executionData,
		StartedAt:     &now,
		Status:        structs.WorkflowExecutionStatus_New,
	}
	execution.Status = structs.WorkflowExecutionStatus_Running
	executionId := uuid.New().String()
	activeExecutions.setExecution(executionId, &execution)
	return executionId, nil
}

func (activeExecutions *ActiveExecutions) ExecuteAsync(
	ctx context.Context,
	executionId string,
	executable func(ctx context.Context) (*structs.WorkflowRunExecutionData, error),
) *structs.ExecutingWorkflowData {
	executionCtx, cancelFunc := context.WithCancel(context.Background())

	waitErr := make(chan error)
	waitData := make(chan *structs.WorkflowRunExecutionData)
	activeExecutions.waitGroup.Add(1)
	go func() {
		// Handle panic
		defer func() {
			if r := recover(); r != nil {
				activeExecutions.removeExecution(executionId)
				activeExecutions.logger.Info("ExecuteAsync panic", r)
				waitErr <- fmt.Errorf("execute async panic, %v", r)
			}
		}()
		defer activeExecutions.waitGroup.Done()
		data, err := executable(executionCtx)
		activeExecutions.removeExecution(executionId)
		if err != nil {
			waitErr <- err
			return
		}
		waitData <- data
	}()

	cancelableRun := structs.WorkflowExecutionCancelableRun{
		Ctx:          executionCtx,
		Cancel:       cancelFunc,
		WaitErrChan:  waitErr,
		WaitDataChan: waitData,
	}

	execution, ok := activeExecutions.getExecution(executionId)
	if !ok {
		return nil
	}
	execution.WorkflowExecutionRun = &cancelableRun
	return execution
}

func (activeExecutions *ActiveExecutions) StopExecution(
	executionId string) *structs.ExecutingWorkflowData {
	execution, ok := activeExecutions.getExecution(executionId)
	if !ok {
		return nil
	}
	execution.WorkflowExecutionRun.Cancel()
	activeExecutions.removeExecution(executionId)
	return execution
}

func (activeExecutions *ActiveExecutions) removeExecution(executionId string) {
	activeExecutions.lock.Lock()
	defer activeExecutions.lock.Unlock()

	delete(activeExecutions.currentExecutions, executionId)
}

func (activeExecutions *ActiveExecutions) getExecution(
	executionId string) (*structs.ExecutingWorkflowData, bool) {
	activeExecutions.lock.Lock()
	defer activeExecutions.lock.Unlock()

	executionWorkflowData, ok := activeExecutions.currentExecutions[executionId]
	return executionWorkflowData, ok
}

func (activeExecutions *ActiveExecutions) setExecution(
	executionId string, execution *structs.ExecutingWorkflowData) {
	activeExecutions.lock.Lock()
	defer activeExecutions.lock.Unlock()

	activeExecutions.currentExecutions[executionId] = execution
}
