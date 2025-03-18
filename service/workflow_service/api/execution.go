package api

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

func (service *WorkflowService) RegisterRouteMethods_Execution() {
	app := service.fiberApp
	app.Get("/workflow/org/:orgId/workflow/:workflowId/execution",
		service.ListWorkflowExecutions)

	app.Post("/workflow/org/:orgId/workflow/:workflowId/execution/delete",
		service.DeleteWorkflowExecutions)

	app.Get("/workflow/org/:orgId/workflow/execution/:executionId",
		service.GetWorkflowExecution)

	app.Post("/workflow/org/:orgId/workflow/execution/:executionId/retry",
		service.RetryWorkflowExecution)

	app.Post("/workflow/org/:orgId/workflow/execution/:executionId/stop",
		service.StopWorkflowCurrentExecution)

	app.Get("/workflow/org/:orgId/workflow/:workflowId/executions-current",
		service.ListWorkflowCurrentExecutions)
}

func (service *WorkflowService) ListWorkflowExecutions(ctx *fiber.Ctx) error {
	orgId := ctx.Params("orgId")
	workflowId := ctx.Params("workflowId")
	if orgId == "" || workflowId == "" {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("orgId or workflowId is empty"))
	}

	filterStr := ctx.Query("filter")
	limit := ctx.QueryInt("limit", 100)
	//lastId := ctx.Query("lastId")
	//firstId := ctx.Query("firstId")
	// TODO filter is not really supported
	filter, err := service.decodeWorkflowExecutionsQueryFilter(workflowId, filterStr)
	if err != nil {
		return HandleBadRequestErrorWithTrace(ctx, err)
	}
	if filter.WorkflowId != "" && filter.WorkflowId != workflowId {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("the filter is for a different workflow"))
	}

	workflowEntity, err := core.GetWorkflowEntityById(ctx.UserContext(), workflowId)
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}
	if workflowEntity.SugerOrgId != orgId {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("the workflow does not belong to the organization"))
	}

	// total count
	count, err := service.rdsDbQueries.CountWorkflowExecutionEntitiesByWorkflowId(ctx.UserContext(), workflowId)
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}

	// list within limit
	executionEntities, err := service.rdsDbQueries.ListWorkflowExecutionEntitiesByWorkflowId(
		ctx.UserContext(),
		rdsDbLib.ListWorkflowExecutionEntitiesByWorkflowIdParams{
			WorkflowId: workflowId,
			Limit:      int32(limit),
			Offset:     0, // TODO pagination
		})
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}

	var executionSummaries []structs.WorkflowExecutionSummary
	for idx := range executionEntities {
		executionEntity := executionEntities[idx]
		executionSummaries = append(executionSummaries, structs.WorkflowExecutionSummary{
			Id:                  fmt.Sprint(executionEntity.ID),
			ExecutionError:      nil, // unused
			Finished:            executionEntity.Finished,
			LastNodeExecuted:    "", // unused
			Mode:                structs.WorkflowExecutionMode(executionEntity.Mode),
			NodeExecutionStatus: nil, // unused
			RetryOf:             executionEntity.RetryOf.String,
			RetrySuccessId:      executionEntity.RetrySuccessId.String,
			Status:              structs.WorkflowExecutionStatus(executionEntity.Status.String),
			StartedAt:           &executionEntity.StartedAt,
			StoppedAt:           core.ConvertNullTimeToStandardTimePointer(executionEntity.StoppedAt),
			WaitTill:            core.ConvertNullTimeToStandardTimePointer(executionEntity.WaitTill),
			WorkflowId:          workflowId,
			WorkflowName:        workflowEntity.Name,
		})
	}

	// Compose response
	response := structs.ListWorkflowExecutionsResponse{
		Data: &structs.ListWorkflowExecutionsResponseData{
			Count:     count,
			Results:   executionSummaries,
			Estimated: false,
		},
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (service *WorkflowService) GetWorkflowExecution(ctx *fiber.Ctx) error {
	orgId := ctx.Params("orgId")
	executionId, err := ctx.ParamsInt("executionId")
	if err != nil {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("executionId is not a valid integer"))
	}
	if orgId == "" || executionId == 0 {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("orgId or executionId is empty"))
	}

	workflowExecution, err := core.GetWorkflowExecution(ctx.UserContext(), int32(executionId))
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}

	err = service.validateExecutionOwnership(ctx.UserContext(), orgId, "", workflowExecution)
	if err != nil {
		return HandleBadRequestErrorWithTrace(ctx, err)
	}

	response := structs.GetWorkflowExecutionResponse{
		Data: workflowExecution,
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (service *WorkflowService) RetryWorkflowExecution(ctx *fiber.Ctx) error {
	orgId := ctx.Params("orgId")
	executionId, err := ctx.ParamsInt("executionId")
	if err != nil {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("executionId is not a valid integer"))
	}
	if orgId == "" || executionId == 0 {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("orgId or executionId is empty"))
	}

	workflowExecution, err := core.GetWorkflowExecution(ctx.UserContext(), int32(executionId))
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}
	err = service.validateExecutionOwnership(ctx.UserContext(), orgId, "", workflowExecution)
	if err != nil {
		return HandleBadRequestErrorWithTrace(ctx, err)
	}

	if workflowExecution.Finished {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("the execution succeeded, so it cannot be retried"))
	}

	// currently ignore reload workflow param, always reload latest workflow
	workflowEntity, err := core.GetWorkflowEntity(ctx.UserContext(), orgId, workflowExecution.WorkflowId)
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}
	// adjust last failed execution data, remove failed results prepare to run it again
	if resultData := workflowExecution.Data.ResultData; resultData != nil {
		resultData.Error = ""
		// remove last failed result
		taskDataList := resultData.RunData[resultData.LastNodeExecuted]
		if length := len(taskDataList); length > 0 && taskDataList[length-1].Error != nil {
			resultData.RunData[resultData.LastNodeExecuted] = taskDataList[:length-1]
		}
		// find last failed node, run as start node
		var startNode *structs.WorkflowNode
		for _, node := range workflowEntity.Nodes {
			if node.Name == resultData.LastNodeExecuted {
				startNode = &node
				break
			}
		}
		if startNode == nil {
			return HandleBadRequestErrorWithTrace(ctx, errors.New("start node not found since workflow changed"))
		}

		sources, ok := workflowExecution.Data.ExecutionData.WaitingExecutionSource[startNode.Name]
		if !ok {
			return HandleBadRequestErrorWithTrace(ctx, errors.New("start node source not found since workflow changed"))
		}
		// find input data of the last failed node. push node and input data to execution stack
		prevRunResultList := make([]structs.NodeData, 0)
		for _, source := range sources {
			outputs := workflowExecution.Data.ExecutionData.WaitingExecution[source.PreviousNode]
			if len(outputs) == 0 {
				prevRunResultList = append(prevRunResultList, structs.NodeData{})
			} else if len(outputs) > source.PreviousNodeOutput {
				prevRunResultList = append(prevRunResultList, outputs[source.PreviousNodeOutput])
			} else {
				return HandleBadRequestErrorWithTrace(ctx, errors.New("previous node outputs inconsistent"))
			}
		}

		workflowExecution.Data.ExecutionData.NodeExecutionStack.PushFront(&structs.NodeExecutionStackData{
			Node:          startNode,
			RunResultList: prevRunResultList,
		})
	}

	executionData := structs.WorkflowExecutionDataProcess{
		ExecutionMode: structs.WorkflowExecutionMode_Retry,
		ExecutionData: workflowExecution.Data,
		SessionId:     "", // empty by default
		WorkflowData:  workflowEntity,
		UserId:        GetContextUserId(ctx),
		RetryOf:       strconv.Itoa(executionId),
	}

	retryExecutionId, err := core.GetActiveExecutions().AddExecution(ctx.UserContext(), &executionData, 0)
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}

	additionalData := &structs.WorkflowExecuteAdditionalData{}
	additionalData.Hooks = core.GetWorkflowHooksMain(strconv.Itoa(retryExecutionId))
	additionalData.Hooks.Mode = executionData.ExecutionMode
	additionalData.Hooks.RetryOf = strconv.Itoa(executionId)
	additionalData.Hooks.WorkflowData = workflowEntity
	var retryRunData *structs.Run
	additionalData.Hooks.HookFunctions.WorkflowExecuteAfter = append(
		additionalData.Hooks.HookFunctions.WorkflowExecuteAfter,
		func(ctx context.Context, hooks *structs.WorkflowHooks, fullRunData *structs.Run) {
			retryRunData = fullRunData
		})

	retryExecution := core.NewWorkflowExecute(ctx.UserContext(), additionalData, structs.WorkflowExecutionMode_Retry)
	retryExecution.RunExecutionData = workflowExecution.Data

	err = retryExecution.Run(ctx.UserContext(), workflowEntity)
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}

	response := map[string]interface{}{
		"data": retryRunData.Finished,
		"id":   strconv.Itoa(retryExecutionId),
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (service *WorkflowService) DeleteWorkflowExecutions(ctx *fiber.Ctx) error {
	orgId := ctx.Params("orgId")
	workflowId := ctx.Params("workflowId")
	if orgId == "" || workflowId == "" {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("orgId, workflowId or executionId is empty"))
	}

	params := structs.DeleteWorkflowExecutionsRequest{}
	if err := ctx.BodyParser(&params); err != nil {
		return HandleBadRequestErrorWithTrace(ctx, err)
	}

	// The execution ids are required to delete the execution data.
	// We don't allow to delete all the executions for the given workflow by default.
	if len(params.Ids) == 0 {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("ids is empty"))
	}
	executionIds := []int32{}
	for _, id := range params.Ids {
		executionId, err := strconv.Atoi(id)
		if err != nil {
			return HandleBadRequestErrorWithTrace(ctx, err)
		}
		executionIds = append(executionIds, int32(executionId))
	}

	// Batch delete the execution data.
	err := service.rdsDbQueries.BatchDeleteWorkflowExecutionData(
		ctx.UserContext(),
		rdsDbLib.BatchDeleteWorkflowExecutionDataParams{
			WorkflowID:   workflowId,
			ExecutionIds: executionIds,
		})
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}

	// Batch delete the execution entities.
	err = service.rdsDbQueries.BatchDeleteWorkflowExecutionEntities(
		ctx.UserContext(),
		rdsDbLib.BatchDeleteWorkflowExecutionEntitiesParams{
			WorkflowId:   workflowId,
			ExecutionIds: executionIds,
		})
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).SendString("")
}

func (service *WorkflowService) ListWorkflowCurrentExecutions(ctx *fiber.Ctx) error {
	orgId := ctx.Params("orgId")
	workflowId := ctx.Params("workflowId")
	if orgId == "" || workflowId == "" {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("orgId or workflowId is empty"))
	}

	filterStr := ctx.Query("filter")
	_, err := service.decodeWorkflowExecutionsQueryFilter(workflowId, filterStr)
	if err != nil {
		return HandleBadRequestErrorWithTrace(ctx, err)
	}

	err = service.validateWorkflowOwnership(ctx.UserContext(), orgId, workflowId)
	if err != nil {
		return HandleBadRequestErrorWithTrace(ctx, err)
	}

	// Get current executions in memory
	var executionSummaries []structs.WorkflowExecutionSummary
	for executionId, executingWorkflowData := range core.GetActiveExecutions().CurrentExecutions() {
		executionData := executingWorkflowData.ExecutionData
		// Filter by workflowId
		if executionData.WorkflowData.ID != workflowId {
			continue
		}
		executionSummaries = append(executionSummaries, structs.WorkflowExecutionSummary{
			Id:         fmt.Sprint(executionId),
			RetryOf:    executionData.RetryOf,
			StartedAt:  executingWorkflowData.StartedAt,
			Mode:       executionData.ExecutionMode,
			WorkflowId: executionData.WorkflowData.ID,
			Status:     executingWorkflowData.Status,
		})
	}
	// TODO sort

	response := structs.ListWorkflowCurrentExecutionsResponse{
		Data: executionSummaries,
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (service *WorkflowService) StopWorkflowCurrentExecution(ctx *fiber.Ctx) error {
	orgId := ctx.Params("orgId")
	executionId, err := ctx.ParamsInt("executionId")
	if err != nil {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("executionId is not a valid integer"))
	}
	if orgId == "" || executionId == 0 {
		return HandleBadRequestErrorWithTrace(ctx, errors.New("orgId or executionId is empty"))
	}

	workflowExecution, err := core.GetWorkflowExecution(ctx.UserContext(), int32(executionId))
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}
	if err := service.validateExecutionOwnership(ctx.UserContext(), orgId, "", workflowExecution); err != nil {
		return HandleBadRequestErrorWithTrace(ctx, err)
	}

	executingWorkflowData := core.GetActiveExecutions().StopExecution(strconv.Itoa(executionId))

	response := structs.StopWorkflowExecutionResponse{
		Data: &structs.WorkflowExecutionStopData{
			Finished:  false,
			Mode:      executingWorkflowData.ExecutionData.ExecutionMode,
			StartedAt: executingWorkflowData.StartedAt,
			StoppedAt: nil,
			Status:    executingWorkflowData.Status,
		},
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (service *WorkflowService) decodeWorkflowExecutionsQueryFilter(
	workflowId string, filterStr string) (structs.WorkflowExecutionsQueryFilter, error) {
	filter := structs.WorkflowExecutionsQueryFilter{}
	err := core.UnmarshalOmitEmpty([]byte(filterStr), &filter)
	if err != nil {
		return structs.WorkflowExecutionsQueryFilter{}, err
	}
	// Ensure the filter is for the given workflow.
	filter.WorkflowId = workflowId

	return filter, nil
}

func (service *WorkflowService) validateWorkflowOwnership(
	ctx context.Context,
	sugerOrgId string, workflowId string,
) error {
	if sugerOrgId == "" {
		return errors.New("sugerOrgId is required")
	}
	if workflowId == "" {
		return errors.New("workflowId is required")
	}
	workflowEntity, err := core.GetWorkflowEntityById(ctx, workflowId)
	if err != nil {
		return err
	}
	if workflowEntity.SugerOrgId != sugerOrgId {
		return errors.New("the workflow does not belong to the organization")
	}

	return nil
}

func (service *WorkflowService) validateExecutionOwnership(
	ctx context.Context,
	sugerOrgId string, workflowId string,
	execution *structs.WorkflowExecution,
) error {
	if sugerOrgId == "" {
		return errors.New("sugerOrgId is required")
	}
	if execution == nil {
		return errors.New("execution is required")
	}

	if workflowId != "" && workflowId != execution.WorkflowId {
		return errors.New("the execution does not belong to the workflow")
	}

	workflowEntity := execution.WorkflowData
	if workflowEntity == nil {
		return errors.New("WorkflowData is nil")
	}
	if workflowEntity.SugerOrgId != "" && workflowEntity.SugerOrgId != sugerOrgId {
		return errors.New("the workflow & execution does not belong to the organization")
	} else if workflowEntity.SugerOrgId == "" {
		// If the SugerOrgId field is missing, then get the workflow entity from the database
		workflowEntity, err := core.GetWorkflowEntityById(ctx, execution.WorkflowId)
		if err != nil {
			return err
		}
		if workflowEntity.SugerOrgId != sugerOrgId {
			return fmt.Errorf("the workflow %s & execution %s does not belong to the organization %s", workflowEntity.ID, execution.Id, sugerOrgId)
		}
	}

	return nil
}
