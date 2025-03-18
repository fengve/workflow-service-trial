package api_test

// Command to run this test only
// go test -v service/workflow_service/api/service_test.go service/workflow_service/api/workflow_test.go

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/code"
	"github.com/sugerio/workflow-service-trial/shared"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type WorkflowTestSuit struct {
	suite.Suite
}

func Test_WorkflowTestSuit(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuit))
}

func (s *WorkflowTestSuit) Test() {
	s.T().Run("TestWorkflow Create Update List Get Delete", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create new workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "test_files/request_create_workflow.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)
		assert.Equal(newWorkflow.SugerOrgId, organization.ID)

		// List workflows
		request_ListWorkflows := events.APIGatewayProxyRequest{
			HTTPMethod:     http.MethodGet,
			Path:           fmt.Sprintf("/workflow/org/%s/workflow", organization.ID),
			Headers:        map[string]string{"Content-Type": "application/json"},
			Body:           "",
			RequestContext: api.AuthorizerRequestContext,
		}
		response_ListWorkflows, err := testFiberLambda.Proxy(request_ListWorkflows)
		assert.Nil(err)
		var listWorkflowResponse structs.ListWorkflowsResponse
		err = json.Unmarshal([]byte(response_ListWorkflows.Body), &listWorkflowResponse)
		assert.Nil(err, fmt.Sprint("response body:", response_ListWorkflows.Body))
		assert.Equal(len(listWorkflowResponse.Data), 1)

		// Update the workflow with new name.
		updateWorkflowRequest := *newWorkflow
		updateWorkflowRequest.Name = "update work flow name"
		updateWorkflowRequestJson, err := json.Marshal(updateWorkflowRequest)
		assert.Nil(err)
		request_UpdateWorkflow := events.APIGatewayProxyRequest{
			HTTPMethod:     http.MethodPatch,
			Path:           fmt.Sprintf("/workflow/org/%s/workflow/%s", organization.ID, newWorkflow.ID),
			Headers:        map[string]string{"Content-Type": "application/json"},
			Body:           string(updateWorkflowRequestJson),
			RequestContext: api.AuthorizerRequestContext,
		}
		response_UpdateWorkflow, err := testFiberLambda.Proxy(request_UpdateWorkflow)
		assert.Nil(err)
		var updateWorkflowResponse structs.UpdateWorkflowResponse
		err = json.Unmarshal([]byte(response_UpdateWorkflow.Body), &updateWorkflowResponse)
		assert.Nil(err, fmt.Sprint("response body:", response_UpdateWorkflow.Body))
		assert.NotNil(updateWorkflowResponse.Data)
		assert.Equal(updateWorkflowResponse.Data.Name, "update work flow name")

		// Get workflow
		request_GetWorkflow := events.APIGatewayProxyRequest{
			HTTPMethod:     http.MethodGet,
			Path:           fmt.Sprintf("/workflow/org/%s/workflow/%s", organization.ID, newWorkflow.ID),
			Headers:        map[string]string{"Content-Type": "application/json"},
			Body:           "",
			RequestContext: api.AuthorizerRequestContext,
		}
		response_GetWorkflow, err := testFiberLambda.Proxy(request_GetWorkflow)
		assert.Nil(err)
		var getWorkflowResponse structs.GetWorkflowResponse
		err = json.Unmarshal([]byte(response_GetWorkflow.Body), &getWorkflowResponse)
		assert.Nil(err)
		assert.Equal(getWorkflowResponse.Data.ID, newWorkflow.ID)

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)

		// Get workflow not found
		response_GetWorkflow, err = testFiberLambda.Proxy(request_GetWorkflow)
		assert.Nil(err)
		assert.Equal("no such workflow", response_GetWorkflow.Body)
	})

	s.T().Run("TestWorkflow Activate ListActive Deactive workflows", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create new workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "test_files/request_create_workflow.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)
		assert.Equal(newWorkflow.SugerOrgId, organization.ID)

		// List active workflow ids
		request_ListActiveWorkflowIds := events.APIGatewayProxyRequest{
			HTTPMethod:     http.MethodGet,
			Path:           fmt.Sprintf("/workflow/org/%s/workflow/active", organization.ID),
			Headers:        map[string]string{"Content-Type": "application/json"},
			Body:           "",
			RequestContext: api.AuthorizerRequestContext,
		}
		response_ListActiveWorkflowIds, err := testFiberLambda.Proxy(request_ListActiveWorkflowIds)
		assert.Nil(err)
		var listActiveWorkflowIdsResponse structs.ListActiveWorkflowIdsResponse
		err = json.Unmarshal([]byte(response_ListActiveWorkflowIds.Body), &listActiveWorkflowIdsResponse)
		// no active workflow since we just created it, not activated yet.
		assert.Equal(0, len(listActiveWorkflowIdsResponse.Data))

		// Activate the workflow
		err = api.ActivateWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)

		// List active workflow ids
		request_ListActiveWorkflowIds = events.APIGatewayProxyRequest{
			HTTPMethod:     http.MethodGet,
			Path:           fmt.Sprintf("/workflow/org/%s/workflow/active", organization.ID),
			Headers:        map[string]string{"Content-Type": "application/json"},
			Body:           "",
			RequestContext: api.AuthorizerRequestContext,
		}
		response_ListActiveWorkflowIds, err = testFiberLambda.Proxy(request_ListActiveWorkflowIds)
		assert.Nil(err)
		err = json.Unmarshal([]byte(response_ListActiveWorkflowIds.Body), &listActiveWorkflowIdsResponse)
		assert.Equal(1, len(listActiveWorkflowIdsResponse.Data))

		// Deactivate the workflow
		err = api.DeactivateWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)

		// List active workflow ids again
		response_ListActiveWorkflowIds, err = testFiberLambda.Proxy(request_ListActiveWorkflowIds)
		assert.Nil(err)
		err = json.Unmarshal([]byte(response_ListActiveWorkflowIds.Body), &listActiveWorkflowIdsResponse)
		assert.Equal(0, len(listActiveWorkflowIdsResponse.Data))

		// Delete the workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)
	})

	s.T().Run("Test SugerOrgId Enforced", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create new workflow, nodes in the create workflow request have invalid orgId
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "test_files/request_create_workflow_limit.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)
		assert.Equal(newWorkflow.SugerOrgId, organization.ID)

		// Get workflow
		getWorkflow, err := api.GetWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)
		assert.NotNil(getWorkflow)

		// OrgId in the nodes should be updated to the organization's id
		assert.Equal(getWorkflow.SugerOrgId, organization.ID)
		for _, node := range getWorkflow.Nodes {
			assert.Equal(node.SugerOrgId, organization.ID)
		}

		// Update workflow with a different orgId in the nodes
		newWorkflow.Name = "updated workflow name"
		for _, node := range newWorkflow.Nodes {
			node.SugerOrgId = "some-random-org-id"
		}
		updateWorkflow, err := api.UpdateWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotNil(updateWorkflow)
		assert.Equal("updated workflow name", updateWorkflow.Name)

		// Check orgId in the workflow and its nodes, should be updated to the organization's id
		for _, node := range updateWorkflow.Nodes {
			assert.Equal(node.SugerOrgId, organization.ID)
		}

		// Get workflow again
		getWorkflow, err = api.GetWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)
		assert.NotNil(getWorkflow)

		// Check orgId in the workflow and its nodes, should be same as the organization's id
		assert.Equal(getWorkflow.SugerOrgId, organization.ID)
		for _, node := range getWorkflow.Nodes {
			assert.Equal(node.SugerOrgId, organization.ID)
		}

	})

	s.T().Run("Test Disable Node", func(t *testing.T) {
		t.Parallel()
		if environment.Env != shared.ENV_LOCAL_TEST {
			t.Skip()
		}
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create new workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "test_files/disable-node.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)

		// Manually run the workflow
		executionID, err := api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotEmpty(executionID)

		// Get the execution
		execution, err := api.GetWorkflowExecution_Testing(testFiberLambda, organization.ID, executionID)
		assert.Nil(err)
		// 3 nodes, 2 result nodes, 1 disable node
		assert.Equal(2, len(execution.Data.ResultData.RunData))

		// Delete the workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)
	})

	s.T().Run("Test Execution Data from previous node", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		newWorkflow, err := api.CreateWorkflow_Testing(testFiberLambda, organization.ID, "test_files/workflow_execution_data_from_previous_node.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)

		executionID, err := api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotEmpty(executionID)

		// Get the execution
		execution, err := api.GetWorkflowExecution_Testing(testFiberLambda, organization.ID, executionID)
		assert.Nil(err)
		assert.NotNil(execution)

		stdout := execution.Data.ResultData.RunData["second"][0].Data["main"][0][0]["json"].(map[string]interface{})["stdout"]
		dates := strings.Split(stdout.(string), "\n")
		// command: `echo {{ $json.stdout }} \n echo {{ $('first').item.json.stdout }}`
		// $json.stdout == $('first').item.json.stdout
		assert.Equal(dates[0], dates[1])
	})

	s.T().Run("Test Execution Data from previous node (CodeNode)", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		newWorkflow, err := api.CreateWorkflow_Testing(testFiberLambda, organization.ID, "test_files/workflow_execution_data_from_earlier_node_code.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)

		executionID, err := api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotEmpty(executionID)

		// Get the execution
		execution, err := api.GetWorkflowExecution_Testing(testFiberLambda, organization.ID, executionID)
		assert.Nil(err)
		assert.NotNil(execution)
		taskData := execution.Data.ResultData.RunData["Code1"]
		outputItems := taskData[0].Data["main"][0]
		assert.Equal("first", outputItems[0]["json"].(string))
		assert.Equal("second", outputItems[1]["json"].(string))
		assert.Equal("third", outputItems[2]["json"].(string))
	})

	s.T().Run("Test finished workflow execution status Running", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		newWorkflow, err := api.CreateWorkflow_Testing(testFiberLambda, organization.ID, "test_files/workflow_execution_empty_input_to_merge_node.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)

		executionID, err := api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotEmpty(executionID)

		// Get the execution
		execution, err := api.GetWorkflowExecution_Testing(testFiberLambda, organization.ID, executionID)
		assert.Nil(err)
		assert.NotNil(execution)
	})

	s.T().Run("Test Workflow Execution Retry", func(t *testing.T) {

		t.Parallel()
		if environment.Env != shared.ENV_LOCAL_TEST {
			t.Skip()
		}
		assert := require.New(s.T())

		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		newWorkflow, err := api.CreateWorkflow_Testing(testFiberLambda, organization.ID, "test_files/workflow_execution_with_error.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)

		executionID, err := api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotEmpty(executionID)

		// Get the execution
		execution, err := api.GetWorkflowExecution_Testing(testFiberLambda, organization.ID, executionID)
		assert.Nil(err)
		assert.NotNil(execution)
		assert.Equal("pre1", execution.Data.ResultData.LastNodeExecuted)
		assert.Equal(structs.WorkflowExecutionStatus_Failed, execution.Status)

		// fix and update workflow
		for _, n := range newWorkflow.Nodes {
			if n.Name == "pre1" {
				n.Parameters["command"] = "echo $(date)"
			}
		}
		_, _ = api.UpdateWorkflow_Testing(testFiberLambda, newWorkflow)

		retryExecutionId, err := api.RetryExecution_Testing(testFiberLambda, organization.ID, executionID)
		assert.Nil(err)
		assert.NotNil(retryExecutionId)

		// Get the execution
		execution, err = api.GetWorkflowExecution_Testing(testFiberLambda, organization.ID, retryExecutionId)
		assert.Nil(err)
		assert.NotNil(execution)
		assert.Equal("second", execution.Data.ResultData.LastNodeExecuted)
		assert.Equal(structs.WorkflowExecutionStatus_Success, execution.Status)
		assert.Equal(5, len(execution.Data.ResultData.RunData))
	})

	s.T().Run("Test nodes execution order", func(t *testing.T) {
		t.Parallel()
		if environment.Env != shared.ENV_LOCAL_TEST {
			t.Skip()
		}
		assert := require.New(s.T())

		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		newWorkflow, err := api.CreateWorkflow_Testing(testFiberLambda, organization.ID, "test_files/workflow_execution_data_from_previous_node.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)

		executionID, err := api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotEmpty(executionID)

		// Get the execution
		execution, err := api.GetWorkflowExecution_Testing(testFiberLambda, organization.ID, executionID)
		assert.Nil(err)
		assert.NotNil(execution)

		assert.True(execution.Data.ResultData.RunData["pre1"][0].StartTime < execution.Data.ResultData.RunData["pre2"][0].StartTime)

		newWorkflow, err = api.CreateWorkflow_Testing(testFiberLambda, organization.ID, "test_files/workflow_execution_order.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)

		executionID, err = api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotEmpty(executionID)

		// Get the execution
		execution, err = api.GetWorkflowExecution_Testing(testFiberLambda, organization.ID, executionID)
		assert.Nil(err)
		assert.NotNil(execution)

		assert.True(execution.Data.ResultData.RunData["pre1"][0].StartTime > execution.Data.ResultData.RunData["pre2"][0].StartTime)
	})
}

func (s *WorkflowTestSuit) TestGetFromTrigger() {
	s.T().Run("Test nodes execution order", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		assert.NotEmpty(organization)

		// Create CreateWorkflow for test
		newWorkflow, err := api.CreateWorkflow_Testing(testFiberLambda, organization.ID, "test_files/request_create_workflow_form.json")
		assert.Nil(err)
		assert.NotNil(newWorkflow)
		assert.Equal(newWorkflow.SugerOrgId, organization.ID)
		assert.Equal(len(newWorkflow.Nodes), 1)

		request_WorkflowsEntity := events.APIGatewayProxyRequest{
			HTTPMethod:     http.MethodGet,
			Path:           fmt.Sprintf("/workflow/public/form/%s/%s", organization.ID, newWorkflow.ID),
			RequestContext: api.AuthorizerRequestContext,
			Headers:        map[string]string{"Content-Type": "application/json"},
			Body:           "",
		}

		resWorkflowsEntity, err := testFiberLambda.Proxy(request_WorkflowsEntity)
		assert.Nil(err)
		assert.Equal(fiber.StatusOK, resWorkflowsEntity.StatusCode)
		assert.NotNil(resWorkflowsEntity)

		var workflowFromResponse structs.GetWorkflowFromResponse
		err = json.Unmarshal([]byte(resWorkflowsEntity.Body), &workflowFromResponse)
		assert.Nil(err, fmt.Sprint("response body:", resWorkflowsEntity.Body))

		parameters, err := json.Marshal(newWorkflow.Nodes[0].Parameters)
		assert.Nil(err)

		var workflowFrom2 structs.WorkflowFrom
		err = json.Unmarshal(parameters, &workflowFrom2)
		assert.Nil(err)

		// http response  == json file?
		assert.Equal(workflowFromResponse.Parameters, &workflowFrom2)

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)

		// Get workflow not found
		resWorkflowsEntity, err = testFiberLambda.Proxy(request_WorkflowsEntity)
		assert.Nil(err)
		assert.Equal("no such workflow", resWorkflowsEntity.Body)
		assert.Equal(fiber.StatusInternalServerError, resWorkflowsEntity.StatusCode)
	})
}

func (s *WorkflowTestSuit) TestPostFromTrigger() {

}
