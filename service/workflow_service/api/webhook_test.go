package api_test

// Command to run this test only
// go test -v service/workflow_service/api/service_test.go service/workflow_service/api/webhook_test.go

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/code"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type WebhookTestSuite struct {
	suite.Suite
}

func Test_WebhookTestSuite(t *testing.T) {
	suite.Run(t, new(WebhookTestSuite))
}

func (s *WebhookTestSuite) TestHandleWebhook() {
	defaultRequestJson := "{\"msg\":\"content here\"}"

	s.T().Run("TestWebhook Call Webhook Mode of onReceived", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// Create new workflow which contains a webhook of OnReceived mode.
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/workflow_execution_webhook_with_onReceived.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)
		workflowId := newWorkflow.ID
		nodeId, webhookId, err := api.GetWebhookIdAndNodeIdInWorkflow(newWorkflow)
		assert.Nil(err)

		// Active the workflow
		err = api.ActivateWorkflow_Testing(testFiberLambda, organization.ID, workflowId)
		assert.Nil(err)

		// Call webhook api (isTest=false)
		webhookResponse, err := api.CallWebhookFullResponse_Testing(
			testFiberLambda, http.MethodPost, workflowId, nodeId, webhookId, false, defaultRequestJson)
		assert.Nil(err)
		// Check Result Body
		assert.Equal(200, webhookResponse.StatusCode)
		var webhookResponseBody map[string]interface{}
		err = json.Unmarshal([]byte(webhookResponse.Body), &webhookResponseBody)
		assert.Nil(err)
		assert.Equal("Workflow was started", webhookResponseBody["message"].(string))
		executionId := webhookResponseBody["executionId"].(string)
		assert.NotEmpty(executionId)
		// Check Result Header
		headerH1 := webhookResponse.MultiValueHeaders["H1"]
		assert.Equal(1, len(headerH1))
		assert.Equal("v1", headerH1[0])

		time.Sleep(3 * time.Second)

		// Get and Verify execution result
		execution, err := api.GetWorkflowExecution_Testing(testFiberLambda, organization.ID, executionId)
		assert.Nil(err)
		assert.NotNil(execution)
		assert.Equal(structs.WorkflowExecutionStatus_Success, execution.Status)
		assert.NotNil(execution.Data)
		assert.Equal(1, len(execution.Data.ResultData.RunData["Code"]))
		codeTaskResult := execution.Data.ResultData.RunData["Code"][0]
		assert.NotEmpty(codeTaskResult.Data["main"])
		assert.NotEmpty(codeTaskResult.Data["main"][0])
		codeResult := codeTaskResult.Data["main"][0][0]
		assert.NotEmpty(codeResult)
		codeResultJson, err := json.Marshal(codeResult["json"])
		assert.Nil(err)
		assert.Equal(defaultRequestJson, string(codeResultJson))

		// Call webhook api with invalid params
		// Invalid workflowId
		webhookResponse, err = api.CallWebhookFullResponse_Testing(
			testFiberLambda, http.MethodPost, "invalidWorkflowId", nodeId, webhookId, false, defaultRequestJson)
		assert.Nil(err)
		assert.Equal(404, webhookResponse.StatusCode)
		// Invalid nodeId
		webhookResponse, err = api.CallWebhookFullResponse_Testing(
			testFiberLambda, http.MethodPost, workflowId, "invalidNodeId", webhookId, false, defaultRequestJson)
		assert.Nil(err)
		assert.Equal(404, webhookResponse.StatusCode)
		// Invalid webhookId
		webhookResponse, err = api.CallWebhookFullResponse_Testing(
			testFiberLambda, http.MethodPost, workflowId, nodeId, "invalidWebhookId", false, defaultRequestJson)
		assert.Nil(err)
		assert.Equal(404, webhookResponse.StatusCode)
		// Invalid httpMethod
		webhookResponse, err = api.CallWebhookFullResponse_Testing(
			testFiberLambda, http.MethodGet, workflowId, nodeId, webhookId, false, defaultRequestJson)
		assert.Nil(err)
		assert.Equal(400, webhookResponse.StatusCode)

		// Deactive the workflow
		err = api.DeactivateWorkflow_Testing(testFiberLambda, organization.ID, workflowId)
		assert.Nil(err)

		// Call webhook api again, result should be 404 because webhook has been deleted
		webhookResponse, err = api.CallWebhookFullResponse_Testing(
			testFiberLambda, http.MethodPost, workflowId, nodeId, webhookId, false, defaultRequestJson)
		assert.Nil(err)
		assert.Equal(404, webhookResponse.StatusCode)

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, workflowId)
		assert.Nil(err)
	})

	s.T().Run("TestWebhook Call Webhook Mode of lastNode", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// Create new workflow
		workflowEntity, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/workflow_execution_webhook_with_lastNode.json")
		assert.Nil(err)
		assert.NotNil(workflowEntity)
		workflowId := workflowEntity.ID
		nodeId, webhookId, err := api.GetWebhookIdAndNodeIdInWorkflow(workflowEntity)
		assert.Nil(err)

		// Active the workflow
		err = api.ActivateWorkflow_Testing(testFiberLambda, organization.ID, workflowId)
		assert.Nil(err)

		// Call webhook api (isTest=false)
		webhookResponse, err := api.CallWebhook_Testing(
			testFiberLambda, http.MethodPost, workflowId, nodeId, webhookId, false, defaultRequestJson)
		assert.Nil(err)
		webhookResponseJson, err := json.Marshal(webhookResponse)
		assert.Nil(err)
		assert.Equal(defaultRequestJson, string(webhookResponseJson))
	})

	// When a webhook use mode of responseNode, the request will return after the respondToWebhook node executed.
	s.T().Run("TestWebhook Call Webhook Mode of responseNode", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// Create new workflow
		workflowEntity, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/workflow_execution_webhook_with_responseNode.json")
		assert.Nil(err)
		assert.NotNil(workflowEntity)
		workflowId := workflowEntity.ID
		nodeId, webhookId, err := api.GetWebhookIdAndNodeIdInWorkflow(workflowEntity)
		assert.Nil(err)

		// Active the workflow
		err = api.ActivateWorkflow_Testing(testFiberLambda, organization.ID, workflowId)
		assert.Nil(err)

		// Call webhook api (isTest=false)
		webhookResponse, err := api.CallWebhook_Testing(
			testFiberLambda, http.MethodPost, workflowId, nodeId, webhookId, false, defaultRequestJson)
		assert.Nil(err)
		webhookResponseJson, err := json.Marshal(webhookResponse)
		assert.Nil(err)
		assert.Equal(defaultRequestJson, string(webhookResponseJson))
	})

	// When a webhook use mode of responseNode but there is no respondToWebhook node
	// the request will wait until the workflow last node executed
	s.T().Run("TestWebhook Call Webhook Mode of responseNode absent", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// Create new workflow
		workflowEntity, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/workflow_execution_webhook_with_responseNode_absent.json")
		assert.Nil(err)
		assert.NotNil(workflowEntity)
		workflowId := workflowEntity.ID
		nodeId, webhookId, err := api.GetWebhookIdAndNodeIdInWorkflow(workflowEntity)
		assert.Nil(err)

		// Active the workflow
		err = api.ActivateWorkflow_Testing(testFiberLambda, organization.ID, workflowId)
		assert.Nil(err)

		// Call webhook api (isTest=false)
		webhookResponse, err := api.CallWebhook_Testing(
			testFiberLambda, http.MethodPost, workflowId, nodeId, webhookId, false, defaultRequestJson)
		assert.Nil(err)
		assert.Equal("Workflow executed successfully", webhookResponse["message"].(string))
		executionId := webhookResponse["executionId"].(string)
		assert.NotEmpty(executionId)
	})

	s.T().Run("TestWebhook Call Webhook Mode of lastNode Execution Failed", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// Create new workflow
		workflowEntity, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/workflow_execution_webhook_with_lastNode_failed.json")
		assert.Nil(err)
		assert.NotNil(workflowEntity)
		workflowId := workflowEntity.ID
		nodeId, webhookId, err := api.GetWebhookIdAndNodeIdInWorkflow(workflowEntity)
		assert.Nil(err)

		// Active the workflow
		err = api.ActivateWorkflow_Testing(testFiberLambda, organization.ID, workflowId)
		assert.Nil(err)

		// Call webhook api (isTest=false)
		webhookResponse, err := api.CallWebhookFullResponse_Testing(
			testFiberLambda, http.MethodPost, workflowId, nodeId, webhookId, false, defaultRequestJson)
		assert.Nil(err)
		// Verify result, code should be 500 because the workflow execute should got failed.
		assert.Equal(500, webhookResponse.StatusCode)
		var webhookResponseBody map[string]interface{}
		err = json.Unmarshal([]byte(webhookResponse.Body), &webhookResponseBody)
		assert.Nil(err)
		assert.Equal(
			"TypeError: Cannot read property 'split' of undefined or null [line 2]",
			webhookResponseBody["message"].(string))
	})
}
