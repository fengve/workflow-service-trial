package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go service/workflow_service/nodes_test/http_request_test.go

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	httprequestNode "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/http_request"
	"github.com/sugerio/workflow-service-trial/shared"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type HttpRequestTestSuite struct {
	suite.Suite
}

func Test_HttpRequestNode(t *testing.T) {
	suite.Run(t, new(HttpRequestTestSuite))
}

func (s *HttpRequestTestSuite) Test() {
	s.T().Run("POST GET PUT PATCH DELETE", func(t *testing.T) {
		t.Parallel()
		if environment.Env != shared.ENV_LOCAL_TEST {
			t.Skip("Skip test for non-local environment since it take too long to run.")
		}
		assert := require.New(s.T())

		// POST
		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/http-request-params-post.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)

		node := &httprequestNode.HttpRequestExecutor{}
		input := &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{},
				},
			},
		}
		result := node.Execute(context.Background(), input)
		assert.NotEmpty(result.ExecutorData[0])
		id := result.ExecutorData[0][0]["json"].(map[string]interface{})["id"]

		// GET
		np = &structs.WorkflowNode{}
		testFile, err = os.ReadFile("./test_files/http-request-params-get.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)
		input = &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id": id,
						},
					},
				},
			},
		}
		result = node.Execute(context.Background(), input)
		assert.Equal(1, len(result.ExecutorData[0]))

		// PUT
		np = &structs.WorkflowNode{}
		testFile, err = os.ReadFile("./test_files/http-request-params-put.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)
		input = &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id": id,
						},
					},
				},
			},
		}
		result = node.Execute(context.Background(), input)
		assert.Equal(1, len(result.ExecutorData[0]))

		// PATCH
		np = &structs.WorkflowNode{}
		testFile, err = os.ReadFile("./test_files/http-request-params-patch.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)
		input = &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id": id,
						},
					},
				},
			},
		}
		result = node.Execute(context.Background(), input)
		assert.Equal(1, len(result.ExecutorData[0]))

		// DELETE
		np = &structs.WorkflowNode{}
		testFile, err = os.ReadFile("./test_files/http-request-params-delete.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)
		input = &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id": id,
						},
					},
				},
			},
		}
		result = node.Execute(context.Background(), input)
		assert.Equal(1, len(result.ExecutorData[0]))
	})

	s.T().Run("GET Parallel", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// GET
		node := &httprequestNode.HttpRequestExecutor{}
		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/http-request-params-get.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)
		input := &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id": "3",
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id":      "4",
							"timeout": 10,
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id":      "5",
							"timeout": 10,
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id":      "6",
							"timeout": 5,
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id": "7",
						},
					},
				},
			},
		}
		result := node.Execute(context.Background(), input)
		assert.Equal(5, len(result.ExecutorData[0]))
	})

	s.T().Run("Forbidden Url", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// GET
		node := &httprequestNode.HttpRequestExecutor{}
		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/http-request-params-get.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)
		np.Parameters["url"] = "http://localhost:3000/api/xxx"
		input := &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id": "1",
						},
					},
				},
			},
		}
		result := node.Execute(context.Background(), input)
		assert.Equal(result.ExecutionStatus, structs.WorkflowExecutionStatus_Failed)

		np.Parameters["url"] = "http://10.0.2.15:3000/api/xxx"
		result = node.Execute(context.Background(), input)
		assert.Equal(result.ExecutionStatus, structs.WorkflowExecutionStatus_Failed)

		np.Parameters["url"] = "http://temporal-frontend.temporal.svc:7233/api/xxx"
		result = node.Execute(context.Background(), input)
		assert.Equal(result.ExecutionStatus, structs.WorkflowExecutionStatus_Failed)
	})

	s.T().Run("POST Form-data and GET Binary", func(t *testing.T) {
		if environment.Env != shared.ENV_LOCAL_TEST {
			t.Skip("Skip test because the test server was run in local.")
		}
		assert := require.New(s.T())

		// POST form-data
		node := &httprequestNode.HttpRequestExecutor{}
		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/http-request-params-post-formdata.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)

		// update Authorization header with secret
		clients := core.GetAwsSdkClients()
		secret, err := clients.GetSecretFromSecretManager("", "org_w43Vc6UfM_test_api_key")
		assert.NoError(err)

		np.Parameters["headerParameters"] = map[string]interface{}{
			"parameters": []map[string]interface{}{
				{
					"name":  "Authorization",
					"value": fmt.Sprintf("Key %s", *secret),
				},
			},
		}

		inputBytes, err := os.ReadFile("./test_files/http-request-params-post-formdata-input.json")
		assert.NoError(err)
		var inputData []structs.NodeData
		err = json.Unmarshal(inputBytes, &inputData)
		assert.NoError(err)

		input := &structs.NodeExecuteInput{
			Params: np,
			Data:   inputData,
		}
		result := node.Execute(context.Background(), input)
		assert.Len(result.ExecutorData, 1)
		assert.Len(result.ExecutorData[0], 1)
		execDataJson, ok := result.ExecutorData[0][0]["json"].(map[string]interface{})
		assert.True(ok)
		assert.EqualValues("minimal.pdf", execDataJson["name"])
		assert.Equal(result.ExecutionStatus, structs.WorkflowExecutionStatus_Success)
		signedUrl, ok := execDataJson["signedUrl"].(string)
		assert.True(ok)
		assert.NotEmpty(signedUrl)

		// GET binary of the file just uploaded.
		node = &httprequestNode.HttpRequestExecutor{}
		np = &structs.WorkflowNode{}
		testFile, err = os.ReadFile("./test_files/http-request-params-get.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)
		// Parse url query params in the url and convert them into NodeExecuteInput.params.queryParameters
		url, err := url.Parse(signedUrl)
		assert.Nil(err)
		queryValues := url.Query()
		queryParams := []map[string]interface{}{}
		for key, values := range queryValues {
			var value interface{}
			if len(values) == 1 {
				value = values[0]
			} else {
				value = values
			}
			queryParams = append(queryParams, map[string]interface{}{
				"name":  key,
				"value": value,
			})
		}
		// Remove query params from url.
		url.RawQuery = ""
		np.Parameters["url"] = url.String()
		np.Parameters["queryParameters"] = map[string]interface{}{
			"parameters": queryParams,
		}
		input = &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{},
					},
				},
			},
		}
		result = node.Execute(context.Background(), input)
		assert.Equal(result.ExecutionStatus, structs.WorkflowExecutionStatus_Success)
		assert.Len(result.ExecutorData, 1)
		assert.Len(result.ExecutorData[0], 1)
		binaryDatas, ok := result.ExecutorData[0][0]["binary"].(map[string]structs.WorkflowBinaryData)
		assert.True(ok)
		binaryData := binaryDatas["data"]
		assert.EqualValues("minimal.pdf", binaryData.FileName)
	})

	s.T().Run("Workflow Create Execute", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// Create workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/http-request-create-workflow.json")
		assert.NotNil(newWorkflow)
		assert.Nil(err)

		// Manual run workflow
		executionId, err := api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotEmpty(executionId)

		// Get execution
		execution, err := api.GetWorkflowExecution_Testing(testFiberLambda, organization.ID, executionId)
		assert.Nil(err)
		assert.NotNil(execution)
		assert.Equal(structs.WorkflowExecutionStatus_Success, execution.Status)
		assert.Equal(structs.WorkflowExecutionMode_Manual, execution.Mode)
		assert.NotNil(execution.Data)
		assert.NotNil(execution.Data.ResultData)
		assert.Len(execution.Data.ResultData.RunData, 2)

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)
	})
}
