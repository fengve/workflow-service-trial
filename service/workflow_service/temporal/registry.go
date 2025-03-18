package temporal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sugerio/workflow-service-trial/shared/structs"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/schedule_trigger"
	"github.com/sugerio/workflow-service-trial/shared/log"
	sharedTemporal "github.com/sugerio/workflow-service-trial/shared/temporal"
)

const (
	DeadlockDetectionTimeout         = 20 * time.Minute
	MaxHeartbeatThrottleInterval     = 5 * time.Minute
	DefaultHeartbeatThrottleInterval = 1 * time.Minute

	// Temporal service task queue for schedule trigger.
	TaskQueue = "schedule-trigger"
)

// InitiateTemporalWorker registers the temporal workflows & activities for schedule-trigger.
func InitiateTemporalWorker(temporalClient client.Client) (worker.Worker, error) {
	workerOptions := worker.Options{
		MaxConcurrentActivityExecutionSize:     200,
		MaxConcurrentWorkflowTaskExecutionSize: 100,
		DeadlockDetectionTimeout:               DeadlockDetectionTimeout,
		MaxHeartbeatThrottleInterval:           MaxHeartbeatThrottleInterval,
		DefaultHeartbeatThrottleInterval:       DefaultHeartbeatThrottleInterval,
	}

	w := worker.New(temporalClient, TaskQueue, workerOptions)

	w.RegisterActivity(Activity_ScheduleTriggerExecuteWorkflow)
	w.RegisterWorkflow(Workflow_ScheduleTrigger)

	w.RegisterActivity(Activity_UnregisterTestWebhooks)
	w.RegisterWorkflow(Workflow_UnregisterTestWebhooks)

	err := w.Start()
	return w, err
}

// Sets up all workflows for schedule triggers.
func SetupAllTemporalWorkflows_ScheduleTrigger(ctx context.Context) error {
	logger := log.GetLogger(ctx)
	// Terminate all open temporal workflows for schedule trigger.
	err := TerminateAllTemporalWorkflows_ScheduleTrigger(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to terminate all open temporal workflows: %v", err))
		// Don't return error, continue with the next steps.
	}

	workflowEntities, err := core.ListAllActiveWorkflowEntities(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to list all active workflow entities: %v", err))
		return err
	}

	for _, workflowEntity := range workflowEntities {
		err := SetupTemporalWorkflow_ScheduleTrigger(ctx, &workflowEntity)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed toSetupTemporalWorkflow_ScheduleTrigger: %v", err))
			// Don't return error, continue with the next structs.
		}
	}

	return nil
}

// Terminate all open temporal workflows for schedule trigger.
func TerminateAllTemporalWorkflows_ScheduleTrigger(ctx context.Context) error {
	logger := log.GetLogger(ctx)
	// List all open temporal workflows for schedule trigger.
	temporalWorkflowExecutions, err := sharedTemporal.ListOpenWorkflowExecutionsByType(
		ctx, core.GetTemporalClient(), "Workflow_ScheduleTrigger")
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to list open temporal workflows: %v", err))
		return err
	}

	// Terminate each open temporal structs.
	for _, temporalWorkflowExecution := range temporalWorkflowExecutions {
		if temporalWorkflowExecution.GetExecution() == nil ||
			temporalWorkflowExecution.GetExecution().GetWorkflowId() == "" {
			// Skip if the temporal workflow ID is empty.
			continue
		}
		err := sharedTemporal.TerminateWorkflowIfOpen(
			ctx,
			core.GetTemporalClient(),
			temporalWorkflowExecution.GetExecution().GetWorkflowId(),
			"terminate workflow for schedule trigger")
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to terminate open temporal workflow: %v", err))
			// Don't return error, continue with the next structs.
			continue
		}
	}

	return nil
}

// Set up one temporal workflow for schedule trigger.
// If the temporal workflow already exists, it will be terminated and recreated.
func SetupTemporalWorkflow_ScheduleTrigger(ctx context.Context, workflowEntity *structs.WorkflowEntity) error {
	if workflowEntity == nil || workflowEntity.SugerOrgId == "" || workflowEntity.ID == "" {
		return errors.New("workflowEntity is nil or missing required fields")
	}

	// find all schedule triggers from structs.
	for index := range workflowEntity.Nodes {
		node := workflowEntity.Nodes[index]
		if node.Disabled {
			continue
		}

		if node.Type != schedule_trigger.Name {
			// Skip if the node is not a schedule trigger.
			continue
		}

		// create trigger
		nodeObj := core.MustNewNode(node.Type)
		triggerObj, ok := nodeObj.(core.TriggerObject)
		if !ok {
			continue
		}

		ctx := context.WithValue(
			ctx,
			sharedTemporal.CommonPropagateContextKey,
			sharedTemporal.CommonCtxPropagation{Environment: core.GetEnvironment()})

		scheduleSpec := triggerObj.Trigger(ctx, &node)
		if scheduleSpec == "" {
			// Skip if the schedule spec is empty.
			continue
		}

		// Only start the temporal workflow for the first schedule.
		// We don't support multiple cron schedules for the same structs.
		temporalWorkflowOptions := GetTemporalWorkflowOptions_ScheduleTrigger(
			workflowEntity.SugerOrgId, workflowEntity.ID, scheduleSpec)
		_, err := sharedTemporal.StartWorkflow_Override(
			ctx,
			core.GetTemporalClient(),
			&temporalWorkflowOptions,
			Workflow_ScheduleTrigger,
			workflowEntity.SugerOrgId,
			workflowEntity.ID)
		if err != nil {
			return err
		}

		// Only start the temporal workflow for the first found schedule trigger node.
		// We don't support multiple schedule trigger nodes for the same structs.
		break
	}

	return nil
}

// Terminate one temporal workflow for schedule trigger by given the workflow entity.
// If there is no active temporal workflow for the given workflow entity, it will be skipped and just return nil.
func TerminateTemporalWorkflow_ScheduleTrigger(ctx context.Context, workflowEntity *structs.WorkflowEntity) error {
	if workflowEntity == nil || workflowEntity.SugerOrgId == "" || workflowEntity.ID == "" {
		return errors.New("workflowEntity is nil or missing required fields")
	}

	temporalWorkflowId := GetTemporalWorkflowId_ScheduleTrigger(workflowEntity.SugerOrgId, workflowEntity.ID)
	// Terminate the temporal workflow by temporal workflow ID.
	err := sharedTemporal.TerminateWorkflowIfOpen(
		ctx,
		core.GetTemporalClient(),
		temporalWorkflowId,
		"terminate workflow for schedule trigger")
	if err != nil {
		return err
	}

	return nil
}

// Start a temporal workflow of UnregisterTestWebhook.
func StartTemporalWorkflow_UnregisterTestWebhooks(ctx context.Context, orgId, workflowId string) error {
	logger := log.GetLogger(ctx)

	workflowOptions := GetTemporalWorkflowOptions_UnregisterTestWebhooks(orgId, workflowId)
	_, err := sharedTemporal.StartWorkflow_Override(
		ctx, core.GetTemporalClient(), &workflowOptions, Workflow_UnregisterTestWebhooks, workflowId)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to start temporal workflow of UnregisterTestWebhooks: %v", err))
		return err
	}

	return nil
}
