package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/gofiber/fiber/v2"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	"github.com/valyala/fasthttp"
)

func (service *WorkflowService) HandleWebhook(ctx *fiber.Ctx) error {
	method := ctx.Method()
	workflowId := ctx.Params("workflowId")
	nodeId := ctx.Params("nodeId")
	webhookId := ctx.Query("webhookId")
	isTest := ctx.QueryBool("isTest", false)
	if len(webhookId) == 0 {
		return HandleBadRequestErrorWithTrace(
			ctx, errors.New("missing required parameter: webhookId"))
	}

	// Confirm the AWS SNS subscription
	if strings.Contains(string(ctx.Body()), "SubscriptionConfirmation") {
		event := structs.AwsSnsSubscriptionConfirmationEvent{}
		if err := json.Unmarshal(ctx.Body(), &event); err != nil {
			return HandleBadRequestErrorWithTrace(ctx, err)
		}

		confirmInput := sns.ConfirmSubscriptionInput{
			Token:    &event.Token,
			TopicArn: &event.TopicArn,
		}
		output, err := service.awsSdkClients.GetSnsClient().ConfirmSubscription(ctx.UserContext(), &confirmInput)
		if err != nil {
			return HandleInternalServerErrorWithTrace(ctx, err)
		}
		return ctx.Status(fiber.StatusOK).JSON(output)
	}

	// Load webhook entity
	webhookEntity, err := core.GetWebhookEntity(ctx.UserContext(), workflowId, webhookId, isTest)
	if err != nil {
		return HandleNotFoundErrorWithTrace(ctx, err)
	}
	// Verify the http method matches
	if webhookEntity.Method != method {
		return HandleBadRequestErrorWithTrace(
			ctx, errors.New("the http method is not allowed for this webhook"))
	}

	// Load workflow entity
	workflowEntity, err := core.GetWorkflowEntityById(ctx.UserContext(), workflowId)
	if err != nil {
		return HandleNotFoundErrorWithTrace(ctx, err)
	}

	// Find webhook node
	webhookNode := workflowEntity.GetNodeById(nodeId)
	if webhookNode == nil {
		return HandleNotFoundErrorWithTrace(ctx, errors.New("no such webhook node in the workflow"))
	}
	if webhookNode.WebhookId != webhookId {
		return HandleBadRequestErrorWithTrace(
			ctx, errors.New("the webhookId is not associated with the nodeId"))
	}

	// Parse webhook node options
	options, err := core.ParseWebhookNodeOptions(webhookNode)
	if err != nil {
		_ = service.Logger.Log("message", "Error in workflow",
			"err", err, "workflowId", workflowId, "webhookId", webhookId)
	}
	for _, entry := range options.ResponseHeaders.Entries {
		ctx.Set(entry.Name, entry.Value)
	}

	// Determine respond mode (return immediately or wait for workflow result)
	responseMode, err := webhookNode.GetWebhookResponseMode()
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}

	additionalData := core.GetBaseAdditionalData()
	// Copy the request from fiber ctx.
	httpRequest := fasthttp.Request{}
	ctx.Request().CopyTo(&httpRequest)
	additionalData.HttpRequest = &httpRequest

	switch responseMode {
	case structs.WebhookResponseMode_OnReceived:
		// Return immediately while executing a workflow in background via goroutine
		// A known issue in test suite: the goroutine of TestA may report error when TestB is running
		executionId, _, err := core.RunWorkflow(
			ctx.UserContext(),
			"",
			workflowEntity,
			additionalData,
			structs.WorkflowExecutionMode_Webhook,
			webhookNode)
		if err != nil {
			return HandleInternalServerErrorWithTrace(ctx, err)
		}
		if options.NoResponseBody {
			return ctx.Status(fiber.StatusOK).SendString("")
		} else if options.ResponseData != "" {
			return ctx.Status(fiber.StatusOK).SendString(options.ResponseData)
		} else {
			return ctx.Status(fiber.StatusOK).JSON(map[string]string{
				"executionId": executionId,
				"message":     "Workflow was started",
			})
		}

	case structs.WebhookResponseMode_LastNode:
		// Execute a workflow in foreground via await
		executionId, executingWorkflowData, err := core.RunWorkflow(
			ctx.UserContext(),
			"",
			workflowEntity,
			additionalData,
			structs.WorkflowExecutionMode_Webhook,
			webhookNode)
		if err != nil {
			return HandleInternalServerErrorWithTrace(ctx, err)
		}

		executionData, err := executingWorkflowData.WorkflowExecutionRun.Wait(ctx.UserContext(), nil)
		if err != nil {
			_ = service.Logger.Log(
				"msg", "Error in workflow",
				"err", err,
				"workflowId", workflowId)
			return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]string{
				"message": "Error in workflow",
			})
		}

		return sendResponseUsingLastNodeResult(ctx, executionId, executionData, webhookNode, options)

	case structs.WebhookResponseMode_ResponseNode:
		// Like lastNode mode but may respond earlier during workflow execution using the hook function.
		// Though workflow execution is single-threaded, still make a thread-safe hook function.
		sendOnce := sync.Once{}
		var sendResponseErr error
		responseSendChan := make(chan struct{})

		sendResponseFunc := func(_ context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response) {
			sendOnce.Do(func() {
				for _, headerKeyBytes := range response.Header.PeekKeys() {
					ctx.Set(string(headerKeyBytes), string(response.Header.Peek(string(headerKeyBytes))))
				}
				sendResponseErr = ctx.Status(response.Header.StatusCode()).Send(response.Body())
				// Nofity the waiting execution
				responseSendChan <- struct{}{}
			})
		}

		additionalData.Hooks.HookFunctions.SendResponse = append(
			additionalData.Hooks.HookFunctions.SendResponse, sendResponseFunc)

		// Execute a workflow in foreground via await
		executionId, executingWorkflowData, err := core.RunWorkflow(
			ctx.UserContext(),
			"",
			workflowEntity,
			additionalData,
			structs.WorkflowExecutionMode_Webhook,
			webhookNode)
		if err != nil {
			return HandleInternalServerErrorWithTrace(ctx, err)
		}

		// Await execution result
		data, err := executingWorkflowData.WorkflowExecutionRun.Wait(ctx.UserContext(), responseSendChan)
		// TODO It only covers the workflow DB error. Need to get node execution error.
		if err != nil {
			_ = service.Logger.Log(
				"msg", "Error in workflow",
				"err", err,
				"workflowId", workflowId)
			sendOnce.Do(func() {
				sendResponseErr = ctx.Status(fiber.StatusInternalServerError).JSON(map[string]string{
					"executionId": executionId,
					"message":     "Error in workflow",
				})
			})
		}

		// Check workflow execution data error
		if data != nil && data.ResultData.Error != "" {
			_ = service.Logger.Log(
				"msg", "Error in workflow",
				"err", data.ResultData.Error,
				"workflowId", workflowId,
				"nodeName", data.ResultData.LastNodeExecuted)
			sendOnce.Do(func() {
				sendResponseErr = ctx.Status(fiber.StatusInternalServerError).JSON(
					map[string]string{
						"executionId": executionId,
						"message":     data.ResultData.Error,
					})
			})
		}

		// Use default response if not responded yet
		sendOnce.Do(func() {
			sendResponseErr = ctx.Status(fiber.StatusOK).JSON(map[string]string{
				"executionId": executionId,
				"message":     "Workflow executed successfully",
			})
		})
		return sendResponseErr

	default:
		err := fmt.Errorf("workflowId=%s, unknown responseMode: %s", workflowId, responseMode)
		return HandleInternalServerErrorWithTrace(ctx, err)
	}
}

func (service *WorkflowService) GetFromTrigger(ctx *fiber.Ctx) error {
	orgId := ctx.Params("orgId")
	workflowId := ctx.Params("workflowId")
	if orgId == "" || workflowId == "" {
		return HandleBadRequestErrorWithTrace(ctx, fmt.Errorf("orgId or workflowId is empty"))
	}

	workflowEntity, err := core.GetWorkflowEntity(ctx.UserContext(), orgId, workflowId)
	if err != nil {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}

	err = fmt.Errorf("no such form data")
	if len(workflowEntity.Nodes) == 0 {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}

	if len(workflowEntity.Nodes[0].Parameters) == 0 {
		return HandleInternalServerErrorWithTrace(ctx, err)
	}

	//response := structs.GetWorkflowFromResponse{
	//	Parameters: workflowEntity.Nodes[0].Parameters,
	//}

	response := map[string]interface{}{
		"parameters": workflowEntity.Nodes[0].Parameters,
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (service *WorkflowService) SaveFromTriggerData(ctx *fiber.Ctx) error {

	return nil
}

// https://github.com/sugerio/workflow-service/blob/c1b5d949658247b19abfdb598cf4b427089cb099/packages/cli/src/WebhookHelpers.ts#L668
func sendResponseUsingLastNodeResult(
	ctx *fiber.Ctx,
	executionId string,
	executionData *structs.WorkflowRunExecutionData,
	webhookNode *structs.WorkflowNode,
	options *core.WebhookNodeOptions) error {

	// https://github.com/sugerio/workflow-service/blob/c1b5d949658247b19abfdb598cf4b427089cb099/packages/cli/src/WebhookHelpers.ts#L606
	if executionData == nil {
		return ctx.Status(fiber.StatusOK).JSON(
			map[string]string{
				"executionId": executionId,
				"message":     "Workflow executed successfully but no data was returned",
			})
	}

	// Check workflow execution data error
	if executionData.ResultData.Error != "" {
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			map[string]string{
				"executionId": executionId,
				"message":     executionData.ResultData.Error,
			})
	}

	lastExecutedNodeData := getDataLastExecutedNodeData(executionData)
	if lastExecutedNodeData == nil {
		return ctx.Status(fiber.StatusOK).JSON(
			map[string]string{
				"executionId": executionId,
				"message":     "Workflow executed successfully but the last node did not return any data",
			})
	}

	mainData, ok := lastExecutedNodeData.Data["main"]
	if !ok || len(mainData) == 0 {
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			map[string]string{
				"executionId": executionId,
				"message":     "unexpected mainData type or empty mainData",
			})
	}

	responseData, err := webhookNode.GetWebhookResponseData()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			map[string]string{
				"executionId": executionId,
				"message":     fmt.Sprintf("failed to GetWebhookResponseData: %v", err),
			})
	}

	switch responseData {
	case structs.WebhookResponseData_FirstEntryJson:
		entries := mainData[0]
		if len(entries) == 0 {
			return ctx.Status(fiber.StatusInternalServerError).JSON(
				map[string]string{
					"executionId": executionId,
					"message":     "no entries are found in mainData[0]",
				})
		}

		responsePropertyName := options.ResponsePropertyName
		resultJson := entries[0]["json"]
		if responsePropertyName != "" {
			if resultJsonValid, ok := resultJson.(map[string]interface{}); ok {
				resultJson = resultJsonValid[responsePropertyName]
			}
		}
		responseContentType := options.ResponseContentType
		if responseContentType != "" {
			ctx.Set("Content-Type", responseContentType)
		}

		resultRaw, err := json.Marshal(resultJson)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(
				map[string]string{
					"executionId": executionId,
					"message":     fmt.Sprintf("failed to marshal resultJson: %v", err),
				})
		}
		return ctx.Status(fiber.StatusOK).Send(resultRaw)

	case structs.WebhookResponseData_FirstEntryBinary:
		// TODO firstEntryBinary
		return HandleInternalServerErrorWithTrace(ctx, errors.New("unimplemented"))

	case structs.WebhookResponseData_NoData:
		return ctx.Status(fiber.StatusOK).SendString("")

	// n8n handles 'allEntries' in default branch
	default:
		resultItems := make([]interface{}, 0)
		for _, entry := range mainData[0] {
			resultItems = append(resultItems, entry["json"])
		}

		// Weird, n8n does not handle any node option here.
		return ctx.Status(fiber.StatusOK).JSON(resultItems)
	}
}

// n8n getDataLastExecutedNodeData
func getDataLastExecutedNodeData(
	data *structs.WorkflowRunExecutionData) *structs.WorkflowExecutionTaskData {
	if data == nil || data.ResultData == nil {
		return nil
	}
	lastNodeExecuted := data.ResultData.LastNodeExecuted
	if lastNodeExecuted == "" {
		return nil
	}
	runData := data.ResultData.RunData
	if runData == nil {
		return nil
	}
	lastNodeRunDataArray, ok := runData[lastNodeExecuted]
	if !ok || len(lastNodeRunDataArray) == 0 {
		return nil
	}
	lastNodeRunData := lastNodeRunDataArray[len(lastNodeRunDataArray)-1]
	// TODO pinData
	return lastNodeRunData
}

func (service *WorkflowService) RegisterRouteMethods_Webhook() {
	service.fiberApp.All("/workflow/public/webhook/workflow/:workflowId/node/:nodeId", service.HandleWebhook)

	formTriggerApi := service.fiberApp.Group("/workflow/public/form")
	formTriggerApi.Get("/:orgId/:workflowId", service.GetFromTrigger)
	formTriggerApi.Post("/:workflowId/form_trigger/:nodeId", service.SaveFromTriggerData)

}
