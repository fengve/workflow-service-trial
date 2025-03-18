package temporal

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/sugerio/workflow-service-trial/shared/structs"

	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/log"
)

func Activity_ScheduleTriggerExecuteWorkflow(ctx context.Context, orgID, workflowID string) error {
	logger := log.GetLogger(ctx)

	workflowEntity, err := core.GetWorkflowEntity(ctx, orgID, workflowID)
	if err != nil {
		logger.Error("Failed to get workflow entity", "err", err)
		ScheduleTriggerSaveExecutionError(ctx, workflowID)
		return err
	}

	if workflowEntity == nil {
		return nil
	}

	// create WorkflowExecute
	additionalData, executionId, err := core.GetAdditionalDataWithHooks(
		ctx, structs.WorkflowExecutionMode_Trigger, workflowEntity, "")
	if err != nil {
		logger.Error("Runner flow: workflowId %s create WorkflowExecute failed with err", workflowEntity.ID, err)
		ScheduleTriggerSaveExecutionError(ctx, workflowID)
		return err
	}
	workFlowExecute := core.NewWorkflowExecute(ctx, additionalData, structs.WorkflowExecutionMode_Trigger)

	// 2. run WorkflowExec.
	err = workFlowExecute.Run(ctx, workflowEntity)
	if err != nil {
		logger.Error(fmt.Sprintf(
			"Runner flow: workflowId %s executionId %d run failed with err %v", workflowEntity.ID, executionId, err))
		return err
	}

	return nil
}

// This activity is only used to save the execution error before the workflow is executed.
func ScheduleTriggerSaveExecutionError(ctx context.Context, workflowID string) {
	logger := log.GetLogger(ctx)

	entity, err := core.GetRdsDbQueries().CreateWorkflowExecutionEntity(
		ctx,
		rdsDbLib.CreateWorkflowExecutionEntityParams{
			Finished:       false,
			Mode:           string(structs.WorkflowExecutionMode_Trigger),
			RetryOf:        sql.NullString{},
			RetrySuccessId: sql.NullString{},
			Status:         sql.NullString{String: string(structs.WorkflowExecutionStatus_Failed), Valid: true},
			WorkflowId:     workflowID,
		})
	if err != nil {
		logger.Error("Failed to create workflow execution entity", "err", err)
		// Don't return error here since we are already in a failed state.
		return
	}

	_, err = core.GetRdsDbQueries().CreateWorkflowExecutionData(
		ctx,
		rdsDbLib.CreateWorkflowExecutionDataParams{
			ExecutionId: entity.ID,
			// TODO: Save the error message in the data field.
		})
	if err != nil {
		logger.Error("Failed to create workflow execution data", "err", err)
		// Don't return error here since we are already in a failed state.
		return
	}
	return
}

// Temporal Activity to unregister test webhooks of the given workflow
func Activity_UnregisterTestWebhooks(ctx context.Context, workflowId string) error {
	logger := log.GetLogger(ctx)

	err := core.UnregisterWebhook(ctx, workflowId, true)
	if err != nil {
		logger.Error("Failed to unregister test webhook", "err", err)
		return err
	}
	return nil
}
