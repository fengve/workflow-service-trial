package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go service/workflow_service/nodes_test/filter_test.go

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	filternode "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/filter"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type FilterTestSuite struct {
	suite.Suite
}

func Test_FilterNode(t *testing.T) {
	suite.Run(t, new(FilterTestSuite))
}

func (s *FilterTestSuite) Test() {
	s.T().Run("TestFilterExecute", func(t *testing.T) {
		assert := require.New(s.T())

		params := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/filter-params.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &params)
		assert.Nil(err)
		conditions, ok := params.Parameters["conditions"].(map[string]interface{})
		assert.True(ok)
		condition, ok := conditions["conditions"].([]interface{})
		assert.Equal(3, len(condition))

		filterNode := &filternode.FilterExecutor{}
		input := &structs.NodeExecuteInput{
			Params: params,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id":   1,
							"data": "test1",
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id":   2,
							"data": "test2",
						},
					},
				},
			},
		}
		result := filterNode.Execute(context.Background(), input)
		assert.Equal(1, len(result.ExecutorData[0]))

		params = &structs.WorkflowNode{}
		testFile, err = os.ReadFile("./test_files/filter-params2.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &params)
		assert.Nil(err)
		conditions, ok = params.Parameters["conditions"].(map[string]interface{})
		assert.True(ok)
		condition, ok = conditions["conditions"].([]interface{})
		assert.Equal(2, len(condition))

		filterNode = &filternode.FilterExecutor{}
		input = &structs.NodeExecuteInput{
			Params: params,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id": 1,
							"data": []map[string]interface{}{
								{
									"item": "test1",
								},
								{
									"item": "test2",
								},
							},
							"test": map[string]interface{}{
								"item": "test1",
							},
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id": 2,
							"data": []map[string]interface{}{
								{
									"item": "test3",
								},
								{
									"item": "test4",
								},
							},
							"test": map[string]interface{}{
								"item": "test1",
							},
						},
					},
				},
			},
		}
		result = filterNode.Execute(context.Background(), input)
		assert.Equal(1, len(result.ExecutorData[0]))
		assert.Equal(1, len(result.ExecutorData[1]))
	})

	s.T().Run("Test Filter Workflow Create and Execute", func(t *testing.T) {
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/filter-e2e.json")
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
		assert.Len(execution.Data.ResultData.RunData, 5)
		resData1 := execution.Data.ResultData.RunData["Filter"][0].Data["main"][0]
		resData2 := execution.Data.ResultData.RunData["Filter"][1].Data["main"][0]
		assert.Len(resData1, 2)
		assert.Len(resData2, 2)

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)
	})
}
