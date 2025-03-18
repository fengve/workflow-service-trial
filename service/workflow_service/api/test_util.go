package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	fiberAdapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/sugerio/workflow-service-trial/rds-db/lib"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

var AuthorizerRequestContext = events.APIGatewayProxyRequestContext{
	Authorizer: map[string]interface{}{
		"email":              "chengjun@suger.io",
		"integrationLatency": 454,
		"principalId":        "vQAUJlvfT",
	},
}

// Create the workflow entity for testing via testFiberLambda
func CreateWorkflow_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	orgId string,
	requestBodyFilePath string,
) (*structs.WorkflowEntity, error) {
	requestBodyBytes, err := os.ReadFile(requestBodyFilePath)
	if err != nil {
		return nil, err
	}
	workflowEntity := structs.WorkflowEntity{}
	err = json.Unmarshal(requestBodyBytes, &workflowEntity)
	if err != nil {
		return nil, err
	}
	request := events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodPost,
		Path:           fmt.Sprintf("/workflow/org/%s/workflow", orgId),
		Headers:        map[string]string{"Content-Type": "application/json"},
		Body:           string(requestBodyBytes),
		RequestContext: AuthorizerRequestContext,
	}
	response, err := testFiberLambda.Proxy(request)
	if err != nil {
		return nil, err
	}
	var createWorkflowResponse structs.GetWorkflowResponse
	err = json.Unmarshal([]byte(response.Body), &createWorkflowResponse)
	if err != nil {
		return nil, err
	}

	return createWorkflowResponse.Data, nil
}

// Get the workflow entity for testing via testFiberLambda
func GetWorkflow_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	orgId string,
	workflowId string,
) (*structs.WorkflowEntity, error) {
	request := events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodGet,
		Path:           fmt.Sprintf("/workflow/org/%s/workflow/%s", orgId, workflowId),
		Headers:        map[string]string{"Content-Type": "application/json"},
		RequestContext: AuthorizerRequestContext,
	}
	response, err := testFiberLambda.Proxy(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get workflow: %s", response.Body)
	}
	var getWorkflowResponse structs.GetWorkflowResponse
	err = json.Unmarshal([]byte(response.Body), &getWorkflowResponse)
	if err != nil {
		return nil, err
	}
	return getWorkflowResponse.Data, nil
}

// Manual run the workflow entity for testing via testFiberLambda
// Return the execution ID of the manual run workflow if success.
func ManualRunWorkflow_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	workflowEntity *structs.WorkflowEntity,
) (string, error) {
	workflowManualRunResponse, err := ManualRunWorkflowFullResponse_Testing(testFiberLambda, workflowEntity)
	if err != nil {
		return "", err
	}

	if workflowManualRunResponse.Data == nil || workflowManualRunResponse.Data.ExecutionId == "" {
		return "", fmt.Errorf("failed to manual run workflow, Data null or Data.ExecutionId null")
	}

	return workflowManualRunResponse.Data.ExecutionId, nil
}

func RetryExecution_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	sugerOrgId string,
	executionId string,
) (string, error) {
	request := events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodPost,
		Path:           fmt.Sprintf("/workflow/org/%s/workflow/execution/%s/retry", sugerOrgId, executionId),
		Headers:        map[string]string{"Content-Type": "application/json"},
		Body:           "",
		RequestContext: AuthorizerRequestContext,
	}
	response, err := testFiberLambda.Proxy(request)
	if err != nil {
		return "", err
	}
	var result map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &result)
	return result["id"].(string), nil
}

// Return full response if success.
func ManualRunWorkflowFullResponse_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	workflowEntity *structs.WorkflowEntity,
) (*structs.WorkflowManualRunResponse, error) {
	if workflowEntity == nil || workflowEntity.SugerOrgId == "" || workflowEntity.ID == "" {
		return nil, fmt.Errorf("invalid workflow entity")
	}
	workflowManualRunRequest := structs.WorkflowManualRunRequest{
		WorkflowData: workflowEntity,
	}
	workflowManualRunRequestBytes, err := json.Marshal(workflowManualRunRequest)
	if err != nil {
		return nil, err
	}
	request := events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodPost,
		Path:           fmt.Sprintf("/workflow/org/%s/workflow/%s/run", workflowEntity.SugerOrgId, workflowEntity.ID),
		Headers:        map[string]string{"Content-Type": "application/json"},
		Body:           string(workflowManualRunRequestBytes),
		RequestContext: AuthorizerRequestContext,
	}
	response, err := testFiberLambda.Proxy(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to manual run workflow: %s", response.Body)
	}
	var workflowManualRunResponse structs.WorkflowManualRunResponse
	err = json.Unmarshal([]byte(response.Body), &workflowManualRunResponse)
	if err != nil {
		return nil, err
	}
	return &workflowManualRunResponse, nil
}

// Delete the workflow entity for testing via testFiberLambda
func DeleteWorkflow_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	orgId string,
	workflowID string,
) error {
	request := events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodDelete,
		Path:           fmt.Sprintf("/workflow/org/%s/workflow/%s", orgId, workflowID),
		Headers:        map[string]string{"Content-Type": "application/json"},
		RequestContext: AuthorizerRequestContext,
	}
	response, err := testFiberLambda.Proxy(request)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("failed to delete workflow: %s", response.Body)
	}
	var deleteWorkflowResponse structs.DeleteWorkflowResponse
	err = json.Unmarshal([]byte(response.Body), &deleteWorkflowResponse)
	if err != nil {
		return err
	}
	if !deleteWorkflowResponse.Data {
		return fmt.Errorf("failed to delete workflow: %s", response.Body)
	}

	return nil
}

// Activate the workflow entity for testing via testFiberLambda
func ActivateWorkflow_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	orgId string,
	workflowID string,
) error {
	activateWorkflowRequest := structs.WorkflowEntity{
		ID:         workflowID,
		SugerOrgId: orgId,
		Active:     true,
	}
	activateWorkflowRequestBytes, err := json.Marshal(activateWorkflowRequest)
	if err != nil {
		return err
	}
	request := events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodPatch,
		Path:           fmt.Sprintf("/workflow/org/%s/workflow/%s", orgId, workflowID),
		Headers:        map[string]string{"Content-Type": "application/json"},
		Body:           string(activateWorkflowRequestBytes),
		RequestContext: AuthorizerRequestContext,
	}
	response, err := testFiberLambda.Proxy(request)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("failed to activate workflow: %s", response.Body)
	}
	var activateWorkflowResponse structs.UpdateWorkflowResponse
	err = json.Unmarshal([]byte(response.Body), &activateWorkflowResponse)
	if err != nil {
		return err
	}
	if activateWorkflowResponse.Data == nil || !activateWorkflowResponse.Data.Active {
		return fmt.Errorf("failed to activate workflow: %s", response.Body)
	}

	return nil
}

// Deactivate the workflow entity for testing via testFiberLambda
func DeactivateWorkflow_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	orgId string,
	workflowID string,
) error {
	deactivateWorkflowRequest := structs.WorkflowEntity{
		ID:         workflowID,
		SugerOrgId: orgId,
		Active:     false,
	}
	activateWorkflowRequestBytes, err := json.Marshal(deactivateWorkflowRequest)
	if err != nil {
		return err
	}
	request := events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodPatch,
		Path:           fmt.Sprintf("/workflow/org/%s/workflow/%s", orgId, workflowID),
		Headers:        map[string]string{"Content-Type": "application/json"},
		Body:           string(activateWorkflowRequestBytes),
		RequestContext: AuthorizerRequestContext,
	}
	response, err := testFiberLambda.Proxy(request)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("failed to deactivate workflow: %s", response.Body)
	}
	var deactivateWorkflowResponse structs.UpdateWorkflowResponse
	err = json.Unmarshal([]byte(response.Body), &deactivateWorkflowResponse)
	if err != nil {
		return err
	}
	if deactivateWorkflowResponse.Data == nil || deactivateWorkflowResponse.Data.Active {
		return fmt.Errorf("failed to deactivate workflow: %s", response.Body)
	}

	return nil
}

// Update the workflow entity for testing via testFiberLambda
func UpdateWorkflow_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	workflowEntity *structs.WorkflowEntity,
) (*structs.WorkflowEntity, error) {
	requestBodyBytes, err := json.Marshal(workflowEntity)
	if err != nil {
		return nil, err
	}
	request := events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodPatch,
		Path:           fmt.Sprintf("/workflow/org/%s/workflow/%s", workflowEntity.SugerOrgId, workflowEntity.ID),
		Headers:        map[string]string{"Content-Type": "application/json"},
		Body:           string(requestBodyBytes),
		RequestContext: AuthorizerRequestContext,
	}
	response, err := testFiberLambda.Proxy(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to update workflow: %s", response.Body)
	}
	var updateWorkflowResponse structs.UpdateWorkflowResponse
	err = json.Unmarshal([]byte(response.Body), &updateWorkflowResponse)
	if err != nil {
		return nil, err
	}
	if updateWorkflowResponse.Data == nil {
		return nil, fmt.Errorf("failed to update workflow: %s", response.Body)
	}

	return updateWorkflowResponse.Data, nil
}

// Get the Execution for testing via testFiberLambda
func GetWorkflowExecution_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	orgId string,
	executionId string,
) (*structs.WorkflowExecution, error) {
	request := events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodGet,
		Path:           fmt.Sprintf("/workflow/org/%s/workflow/execution/%s", orgId, executionId),
		Headers:        map[string]string{"Content-Type": "application/json"},
		RequestContext: AuthorizerRequestContext,
	}
	response, err := testFiberLambda.Proxy(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get execution: %s", response.Body)
	}
	var getExecutionResponse structs.GetWorkflowExecutionResponse
	err = json.Unmarshal([]byte(response.Body), &getExecutionResponse)
	if err != nil {
		return nil, err
	}
	if getExecutionResponse.Data == nil {
		return nil, fmt.Errorf("failed to get execution: %s", response.Body)
	}

	return getExecutionResponse.Data, nil
}

// Call webhook API
func CallWebhook_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	httpMethod string,
	workflowId string,
	nodeId string,
	webhookId string,
	isTest bool,
	body string,
) (map[string]interface{}, error) {
	request := events.APIGatewayProxyRequest{
		HTTPMethod:            httpMethod,
		Path:                  fmt.Sprintf("/workflow/public/webhook/workflow/%s/node/%s", workflowId, nodeId),
		QueryStringParameters: map[string]string{"webhookId": webhookId, "isTest": strconv.FormatBool(isTest)},
		Body:                  body,
		Headers:               map[string]string{"Content-Type": "application/json"},
		RequestContext:        AuthorizerRequestContext,
	}
	response, err := testFiberLambda.Proxy(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to call webhook: %s", response.Body)
	}
	var result map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Call webhook API
func CallWebhookFullResponse_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	httpMethod string,
	workflowId string,
	nodeId string,
	webhookId string,
	isTest bool,
	body string,
) (events.APIGatewayProxyResponse, error) {
	request := events.APIGatewayProxyRequest{
		HTTPMethod:            httpMethod,
		Path:                  fmt.Sprintf("/workflow/public/webhook/workflow/%s/node/%s", workflowId, nodeId),
		QueryStringParameters: map[string]string{"webhookId": webhookId, "isTest": strconv.FormatBool(isTest)},
		Body:                  body,
		Headers:               map[string]string{"Content-Type": "application/json"},
		RequestContext:        AuthorizerRequestContext,
	}
	return testFiberLambda.Proxy(request)
}

// Delete test webhook
func DeleteTestWebhook_Testing(
	testFiberLambda *fiberAdapter.FiberLambda,
	orgId string,
	workflowId string,
) (bool, error) {
	request := events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodDelete,
		Path:           fmt.Sprintf("/workflow/org/%s/workflow/%s/test-webhook", orgId, workflowId),
		Headers:        map[string]string{"Content-Type": "application/json"},
		RequestContext: AuthorizerRequestContext,
	}
	response, err := testFiberLambda.Proxy(request)
	if err != nil {
		return false, err
	}
	if response.StatusCode != 200 {
		return false, fmt.Errorf("failed to delete test webhook: %s", response.Body)
	}

	var deleteWebhookResponse structs.DeleteWorkflowTestWebhookResponse
	err = json.Unmarshal([]byte(response.Body), &deleteWebhookResponse)
	if err != nil {
		return false, err
	}
	return deleteWebhookResponse.Data, nil
}

// Get nodeId and webhookId of the first webhook node in the workflowEntity.
func GetWebhookIdAndNodeIdInWorkflow(workflowEntity *structs.WorkflowEntity) (string, string, error) {
	if workflowEntity == nil || len(workflowEntity.Nodes) == 0 {
		return "", "", fmt.Errorf("workflowEntity is nil or empty")
	}

	for _, node := range workflowEntity.Nodes {
		if node.Type == "n8n-nodes-base.webhook" {
			return node.ID, node.WebhookId, nil
		}
	}

	return "", "", fmt.Errorf("webhook node not found")
}

// Get nodeId and webhookId of all webhook nodes in the workflowEntity.
func GetAllWebhookIdAndNodeIdInWorkflow(workflowEntity *structs.WorkflowEntity) (map[string]string, error) {
	if workflowEntity == nil || len(workflowEntity.Nodes) == 0 {
		return nil, fmt.Errorf("workflowEntity is nil or empty")
	}
	result := map[string]string{}
	for _, node := range workflowEntity.Nodes {
		if node.Type == "n8n-nodes-base.webhook" {
			result[node.ID] = node.WebhookId
		}
	}
	return result, nil
}

func GetWebhookEntities(workflowId, webhookId string) ([]lib.WorkflowWebhookEntity, error) {
	return core.GetRdsDbQueries().ListWebhookEntities(
		context.Background(),
		rdsDbLib.ListWebhookEntitiesParams{
			WorkflowId: workflowId,
			WebhookId:  sql.NullString{String: webhookId, Valid: true},
		})
}
