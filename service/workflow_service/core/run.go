package core

import (
	"context"
	"errors"
	"strconv"

	sharedlog "github.com/sugerio/workflow-service-trial/shared/log"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

// RunWorkflow should be placed in WorkflowRunner (like n8n), place here temporarily
func RunWorkflow(
	ctx context.Context,
	userId string,
	workflowEntity *structs.WorkflowEntity,
	additionalData *structs.WorkflowExecuteAdditionalData,
	mode structs.WorkflowExecutionMode,
	startNode *structs.WorkflowNode,
) (string, *structs.ExecutingWorkflowData, error) {
	if workflowEntity == nil {
		return "", nil, errors.New("workflow is required")
	}

	executionData := structs.WorkflowExecutionDataProcess{
		ExecutionMode: structs.WorkflowExecutionMode_Webhook,
		ExecutionData: nil, // empty by default
		SessionId:     "",  // empty by default
		WorkflowData:  workflowEntity,
		UserId:        userId,
	}

	executionId, err := GetActiveExecutions().AddExecution(ctx, &executionData, 0)
	if err != nil {
		return "", nil, err
	}

	hooks := GetWorkflowHooksMain(strconv.Itoa(executionId))
	hooks.HookFunctions.SendResponse = append(hooks.HookFunctions.SendResponse, additionalData.Hooks.HookFunctions.SendResponse...)
	additionalData.Hooks = hooks
	additionalData.Hooks.Mode = executionData.ExecutionMode
	additionalData.Hooks.RetryOf = executionData.RetryOf
	additionalData.Hooks.WorkflowData = workflowEntity

	executingWorkflowData := GetActiveExecutions().ExecuteAsync(
		ctx,
		strconv.Itoa(executionId),
		func(ctx context.Context) (*structs.WorkflowRunExecutionData, error) {
			workflowExecute := NewWorkflowExecute(ctx, additionalData, mode)
			err := workflowExecute.RunFromNode(ctx, workflowEntity, startNode)
			if err != nil {
				sharedlog.GetLogger(ctx).Error("Failed to execute the workflow",
					"workflowId", workflowEntity.ID,
					"err", err)
				return nil, err
			}
			return workflowExecute.RunExecutionData, nil
		},
	)

	return strconv.Itoa(executionId), executingWorkflowData, nil
}
