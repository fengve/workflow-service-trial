package temporal

import (
	"fmt"
	"time"

	temporalEnums "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	temporalWorkflow "go.temporal.io/sdk/workflow"

	"github.com/sugerio/workflow-service-trial/shared/log"
	sharedTemporal "github.com/sugerio/workflow-service-trial/shared/temporal"
)

const (
	StartToCloseTimeout = 5 * time.Minute
	InitialInterval     = 5 * time.Second
	MaximumAttempts     = 1 // Only try once
)

// The temporal workflow to schedule the schedule trigger.
// Attention: Never change the function name, it is used as WorkflowType in the temporal workflow service.
func Workflow_ScheduleTrigger(ctx temporalWorkflow.Context, orgId string, workflowId string) error {
	logger := log.GetLogger(ctx)
	logger.Info("Workflow_ScheduleTrigger", "orgId", orgId, "workflowId", workflowId)

	ctx = temporalWorkflow.WithActivityOptions(
		ctx,
		temporalWorkflow.ActivityOptions{
			StartToCloseTimeout: StartToCloseTimeout,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval: InitialInterval,
				MaximumAttempts: MaximumAttempts,
			},
		})

	err := temporalWorkflow.ExecuteActivity(
		ctx, Activity_ScheduleTriggerExecuteWorkflow, orgId, workflowId).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to run Activity_ScheduleTriggerExecuteWorkflow", "err", err)
		return err
	}

	return nil
}

func GetTemporalWorkflowOptions_ScheduleTrigger(
	orgId string, workflowId string, cronSchedule string) client.StartWorkflowOptions {
	temporalWorkflowId := GetTemporalWorkflowId_ScheduleTrigger(orgId, workflowId)
	return client.StartWorkflowOptions{
		ID:                    temporalWorkflowId,
		TaskQueue:             TaskQueue,
		WorkflowIDReusePolicy: temporalEnums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
		CronSchedule:          cronSchedule,
	}
}

// Get the temporal workflow ID for schedule trigger by given the workflow Id (the workflow entity ID in workflow service).
func GetTemporalWorkflowId_ScheduleTrigger(orgId string, workflowId string) string {
	return fmt.Sprintf(sharedTemporal.WorkflowIdTemplate_ScheduleTrigger, orgId, workflowId)
}

// The temporal workflow to unregister test webhooks of a workflow.
func Workflow_UnregisterTestWebhooks(ctx temporalWorkflow.Context, workflowId string) error {
	logger := temporalWorkflow.GetLogger(ctx)
	logger.Info("Workflow_UnregisterTestWebhook", "workflowId", workflowId)

	ctx = temporalWorkflow.WithActivityOptions(ctx, temporalWorkflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: 10 * time.Second,
			MaximumAttempts: 2,
		},
	})

	// Wait for 2 minutes to let the test webhook be triggered.
	temporalWorkflow.Sleep(ctx, 2*time.Minute)

	err := temporalWorkflow.ExecuteActivity(ctx, Activity_UnregisterTestWebhooks, workflowId).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to run Activity_UnregisterTestWebhook", "error", err)
		return err
	}

	return nil
}

// Get the WorkflowOptions for the workflow of UnregisterTestWebhook.
func GetTemporalWorkflowOptions_UnregisterTestWebhooks(
	orgId string, workflowId string) client.StartWorkflowOptions {
	return client.StartWorkflowOptions{
		ID:                       fmt.Sprintf(sharedTemporal.WorkflowIdTemplate_UnregisterTestWebhooks, orgId, workflowId),
		TaskQueue:                TaskQueue,
		WorkflowExecutionTimeout: 5 * time.Minute,
		WorkflowRunTimeout:       5 * time.Minute,
		WorkflowIDReusePolicy:    temporalEnums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
}
