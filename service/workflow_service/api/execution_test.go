package api_test

// Command to run this test only
// go test -v service/workflow_service/api/service_test.go service/workflow_service/api/execution_test.go

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/code"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type ExecutionTestSuite struct {
	suite.Suite
}

func Test_ExecutionTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutionTestSuite))
}

func (s *ExecutionTestSuite) Test() {
	s.T().Run("TestListWorkflowExecutions returns empty for no data", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/workflow_execution_simplest.json")
		assert.Nil(err)
		api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)

		request, err := GetAPIGatewayProxyRequest_CreateOrganization()
		assert.Nil(err)
		request.HTTPMethod = http.MethodGet
		request.Path = fmt.Sprintf("/workflow/org/%s/workflow/%s/execution?filter={}", organization.ID, newWorkflow.ID)
		response, err := testFiberLambda.Proxy(request)

		assert.Nil(err)
		assert.Equal(200, response.StatusCode)
		var responseData structs.ListWorkflowExecutionsResponse
		err = json.Unmarshal([]byte(response.Body), &responseData)
		assert.Nil(err, fmt.Sprint("response body:", response.Body))
		assert.Equal(1, int(responseData.Data.Count))
		assert.Equal(false, responseData.Data.Estimated)
		assert.Equal(1, len(responseData.Data.Results))
	})

	s.T().Run("TestListWorkflowExecutions returns results", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/workflow_execution_simplest.json")
		executionId1 := createWorkflowExecutionAndData_Testing(assert, newWorkflow)
		executionId2 := createWorkflowExecutionAndData_Testing(assert, newWorkflow)

		request, err := GetAPIGatewayProxyRequest_CreateOrganization()
		assert.Nil(err)
		request.HTTPMethod = http.MethodGet
		request.Path = fmt.Sprintf("/workflow/org/%s/workflow/%s/execution?filter={}", organization.ID, newWorkflow.ID)
		response, err := testFiberLambda.Proxy(request)

		assert.Nil(err)
		assert.Equal(200, response.StatusCode)
		var responseData structs.ListWorkflowExecutionsResponse
		err = json.Unmarshal([]byte(response.Body), &responseData)
		assert.Nil(err, fmt.Sprint("response body:", response.Body))
		assert.Equal(2, int(responseData.Data.Count))
		assert.Equal(false, responseData.Data.Estimated)
		assert.Equal(2, len(responseData.Data.Results))
		assert.Equal(fmt.Sprint(executionId1), responseData.Data.Results[0].Id)
		assert.Equal(fmt.Sprint(executionId2), responseData.Data.Results[1].Id)
	})

	s.T().Run("TestGetWorkflowExecution", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/workflow_execution_simplest.json")
		executionId := createWorkflowExecutionAndData_Testing(assert, newWorkflow)

		request, err := GetAPIGatewayProxyRequest_CreateOrganization()
		assert.Nil(err)
		request.HTTPMethod = http.MethodGet
		request.Path = fmt.Sprintf("/workflow/org/%s/workflow/execution/%d", organization.ID, executionId)
		response, err := testFiberLambda.Proxy(request)

		assert.Nil(err)
		assert.Equal(200, response.StatusCode)
		var responseData structs.GetWorkflowExecutionResponse
		err = json.Unmarshal([]byte(response.Body), &responseData)
		assert.Nil(err, fmt.Sprint("response body:", response.Body))
		assert.Equal(fmt.Sprint(executionId), responseData.Data.Id)
	})

	s.T().Run("TestDeleteWorkflowExecutions deletes one with specified id", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/workflow_execution_simplest.json")
		// Create two executions
		executionId := createWorkflowExecutionAndData_Testing(assert, newWorkflow)
		executionId2nd := createWorkflowExecutionAndData_Testing(assert, newWorkflow)
		assert.NotEqual(executionId, executionId2nd)

		request, err := GetAPIGatewayProxyRequest_CreateOrganization()
		assert.Nil(err)
		request.HTTPMethod = http.MethodPost
		request.Path = fmt.Sprintf("/workflow/org/%s/workflow/%s/execution/delete?filter={}", organization.ID, newWorkflow.ID)
		request.Headers = map[string]string{"Content-Type": "application/json"}
		request.Body = fmt.Sprintf("{\"ids\": [\"%d\"]}", executionId)
		response, err := testFiberLambda.Proxy(request)

		assert.Nil(err)
		assert.Equal(200, response.StatusCode)
		executionEntities, err := rdsDbQueries.ListWorkflowExecutionEntitiesByWorkflowId(
			context.Background(),
			rdsDbLib.ListWorkflowExecutionEntitiesByWorkflowIdParams{
				WorkflowId: newWorkflow.ID,
				Limit:      10,
			})
		assert.Nil(err)
		// Deleted only one of the two
		assert.Equal(1, len(executionEntities))
		assert.Equal(executionId2nd, executionEntities[0].ID)
	})

	s.T().Run("TestListWorkflowCurrentExecutions returns empty for no data", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/workflow_execution_simplest.json")

		request, err := GetAPIGatewayProxyRequest_CreateOrganization()
		assert.Nil(err)
		request.HTTPMethod = http.MethodGet
		request.Path = fmt.Sprintf("/workflow/org/%s/workflow/%s/executions-current", organization.ID, newWorkflow.ID)
		response, err := testFiberLambda.Proxy(request)

		assert.Nil(err)
		assert.Equal(200, response.StatusCode)
		var responseData structs.ListWorkflowCurrentExecutionsResponse
		err = json.Unmarshal([]byte(response.Body), &responseData)
		assert.Nil(err, fmt.Sprint("response body:", response.Body))
		assert.Equal(0, len(responseData.Data))
	})

	s.T().Run("TestListWorkflowCurrentExecutions can filter by workflowId", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		ctx := context.Background()
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// prepare two workflows and their executions
		newWorkflow1, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/workflow_execution_simplest.json")
		assert.Nil(err)
		newWorkflow2, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/workflow_execution_simplest.json")
		executionId1 := addActiveExecution_Testing(assert, ctx, newWorkflow1.ID)
		executionId2 := addActiveExecution_Testing(assert, ctx, newWorkflow2.ID)

		// query for 1st workflow can only view executions of 1st workflow
		{
			request, err := GetAPIGatewayProxyRequest_CreateOrganization()
			assert.Nil(err)
			request.HTTPMethod = http.MethodGet
			request.Path = fmt.Sprintf("/workflow/org/%s/workflow/%s/executions-current", organization.ID, newWorkflow1.ID)
			response, err := testFiberLambda.Proxy(request)

			assert.Nil(err)
			assert.Equal(200, response.StatusCode)
			var responseData structs.ListWorkflowCurrentExecutionsResponse
			err = json.Unmarshal([]byte(response.Body), &responseData)
			assert.Nil(err, fmt.Sprint("response body:", response.Body))
			assert.Equal(1, len(responseData.Data))
			assert.Equal(strconv.Itoa(executionId1), responseData.Data[0].Id)
			assert.Equal(newWorkflow1.ID, responseData.Data[0].WorkflowId)
		}

		// query for 2nd workflow can only view executions of 2nd workflow
		{
			request, err := GetAPIGatewayProxyRequest_CreateOrganization()
			assert.Nil(err)
			request.HTTPMethod = http.MethodGet
			request.Path = fmt.Sprintf("/workflow/org/%s/workflow/%s/executions-current", organization.ID, newWorkflow2.ID)
			response, err := testFiberLambda.Proxy(request)

			assert.Nil(err)
			assert.Equal(200, response.StatusCode)
			var responseData structs.ListWorkflowCurrentExecutionsResponse
			err = json.Unmarshal([]byte(response.Body), &responseData)
			assert.Nil(err, fmt.Sprint("response body:", response.Body))
			assert.Equal(1, len(responseData.Data))
			assert.Equal(strconv.Itoa(executionId2), responseData.Data[0].Id)
			assert.Equal(newWorkflow2.ID, responseData.Data[0].WorkflowId)
		}
	})
}
