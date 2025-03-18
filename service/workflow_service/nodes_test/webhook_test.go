package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go  service/workflow_service/nodes_test/webhook_test.go

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/webhook"
	"github.com/sugerio/workflow-service-trial/shared"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type WebhookTestSuite struct {
	suite.Suite
}

func Test_Webhook(t *testing.T) {
	suite.Run(t, new(WebhookTestSuite))
}

func (s *WebhookTestSuite) Test() {
	s.T().Run("TestWebhookSpec", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		var nodeSpec structs.WorkflowNodeDescriptionSpec
		testFile, err := os.ReadFile("./test_files/webhook-node.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &nodeSpec)
		assert.Nil(err)

		assert.Equal("Waiting for you to call the Test URL", nodeSpec.EventTriggerDescription)
		assert.Equal(0, nodeSpec.MaxNodes)
	})

	s.T().Run("TestWebhookGenerate", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		_ = webhook.Webhook{}
		executor := core.NewExecutor(webhook.Name)
		node := executor.GetNode()
		assert.Equal("Webhook", node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.DisplayName)
	})

	s.T().Run("TestWebhook Call Test Webhook of ManuallyRun Workflow", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// Create Workflow for test
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/webhook-e2e.json")
		assert.Nil(err)
		workflowID := newWorkflow.ID

		// Manual run the workflow
		workflowRunResponse, err := api.ManualRunWorkflowFullResponse_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.True(workflowRunResponse.Data.WaitingForWebhook)

		nodeId, webhookId, err := api.GetWebhookIdAndNodeIdInWorkflow(newWorkflow)
		assert.Nil(err)

		// Check db, there should be 1 records of the webhook
		webhookEitities, err := api.GetWebhookEntities(workflowID, webhookId)
		assert.Nil(err)
		assert.Equal(1, len(webhookEitities))

		// Call Test Webhook (isTest=true)
		response, err := api.CallWebhook_Testing(testFiberLambda, http.MethodPost, workflowID, nodeId, webhookId,
			true, "{\"data\": \"test\"}")
		assert.Nil(err)
		assert.NotNil(response)
		webhookResponseJson, err := json.Marshal(response)
		assert.Nil(err)
		assert.Equal(string(webhookResponseJson), "{\"data\":\"test\"}")

		// Delete Test Webhook
		deleteWebhookResult, err := api.DeleteTestWebhook_Testing(testFiberLambda, organization.ID, workflowID)
		assert.Nil(err)
		assert.True(deleteWebhookResult)

		// Check db, there should be 0 records of the webhook
		webhookEitities, err = api.GetWebhookEntities(workflowID, webhookId)
		assert.Nil(err)
		assert.Equal(0, len(webhookEitities))

		// Call Test Webhook again, it's result should be 404
		_, err = api.CallWebhook_Testing(testFiberLambda, http.MethodPost, workflowID, nodeId, webhookId,
			true, "{\"data\": \"test\"}")
		assert.NotNil(err)

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, workflowID)
		assert.Nil(err)
	})

	s.T().Run("TestWebhook Call Webhook of Active Workflow", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// Create Workflow for test
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/webhook-e2e2.json")
		assert.Nil(err)
		workflowID := newWorkflow.ID

		// Active the workflow
		err = api.ActivateWorkflow_Testing(testFiberLambda, organization.ID, workflowID)
		assert.Nil(err)

		nodeId, webhookId, err := api.GetWebhookIdAndNodeIdInWorkflow(newWorkflow)
		assert.Nil(err)

		// Check db, there should be 1 records of the webhook
		webhookEitities, err := api.GetWebhookEntities(workflowID, webhookId)
		assert.Nil(err)
		assert.Equal(1, len(webhookEitities))

		// Call Webhook (isTest=false)
		response, err := api.CallWebhook_Testing(testFiberLambda, http.MethodPost, workflowID, nodeId, webhookId,
			false, "{\"data\": \"test\"}")
		assert.Nil(err)
		assert.NotNil(response)
		webhookResponseJson, err := json.Marshal(response)
		assert.Nil(err)
		assert.Equal("{\"data\":\"test\"}", string(webhookResponseJson))

		// Run test webhook call at the same time
		// Manual run the workflow
		workflowRunResponse, err := api.ManualRunWorkflowFullResponse_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.True(workflowRunResponse.Data.WaitingForWebhook)
		// Call Test Webhook (isTest=true)
		response, err = api.CallWebhook_Testing(testFiberLambda, http.MethodPost, workflowID, nodeId, webhookId,
			true, "{\"data\": \"test\"}")
		assert.Nil(err)
		assert.NotNil(response)
		webhookResponseJson, err = json.Marshal(response)
		assert.Nil(err)
		assert.Equal(string(webhookResponseJson), "{\"data\":\"test\"}")

		// Check db, there should be 2 records of the webhook
		webhookEitities, err = api.GetWebhookEntities(workflowID, webhookId)
		assert.Nil(err)
		assert.Equal(2, len(webhookEitities))

		// Delete Test Webhook
		deleteWebhookResult, err := api.DeleteTestWebhook_Testing(testFiberLambda, organization.ID, workflowID)
		assert.Nil(err)
		assert.True(deleteWebhookResult)

		// Check db, there should be 1 records of the webhook
		webhookEitities, err = api.GetWebhookEntities(workflowID, webhookId)
		assert.Nil(err)
		assert.Equal(1, len(webhookEitities))

		// Deactive the workflow
		err = api.DeactivateWorkflow_Testing(testFiberLambda, organization.ID, workflowID)
		assert.Nil(err)

		// Check db, there should be 0 records of the webhook
		webhookEitities, err = api.GetWebhookEntities(workflowID, webhookId)
		assert.Nil(err)
		assert.Equal(0, len(webhookEitities))

		// Call webhook api again, result should be 404 because webhook has been deleted
		_, err = api.CallWebhook_Testing(testFiberLambda, http.MethodPost, workflowID, nodeId, webhookId,
			false, "{\"data\": \"test\"}")
		assert.NotNil(err)

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, workflowID)
		assert.Nil(err)
	})

	s.T().Run("TestWebhook Call Webhook of Multi Webhook Workflow", func(t *testing.T) {
		t.Parallel()
		if environment.Env != shared.ENV_LOCAL_TEST {
			t.Skip()
		}

		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// Create two webhook Workflows for test
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/webhook-multiple-e2e.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)

		// Active the workflow
		err = api.ActivateWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)

		nodeToWebhook, err := api.GetAllWebhookIdAndNodeIdInWorkflow(newWorkflow)
		assert.Nil(err)
		assert.Equal(2, len(nodeToWebhook))

		// Get the workflow execution. it should be empty because the webhook is not called yet.
		count, err := rdsDbQueries.CountWorkflowExecutionEntitiesByWorkflowId(context.Background(), newWorkflow.ID)
		assert.Nil(err)
		assert.Equal(count, int64(0))

		// Call Webhook
		for nodeID, webhookID := range nodeToWebhook {
			_, err = api.CallWebhook_Testing(testFiberLambda, http.MethodPost, newWorkflow.ID, nodeID, webhookID, false, "{\"data\": \"test\"}")
			assert.Nil(err)
		}
		count, err = rdsDbQueries.CountWorkflowExecutionEntitiesByWorkflowId(context.Background(), newWorkflow.ID)
		assert.Nil(err)
		assert.Equal(count, int64(2))

		// Delete workflows
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)
	})
}
