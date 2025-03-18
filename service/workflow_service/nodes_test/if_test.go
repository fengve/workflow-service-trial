package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go service/workflow_service/nodes_test/if_test.go

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	ifnode "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/if"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type IfTestSuite struct {
	suite.Suite
}

func Test_IfNode(t *testing.T) {
	suite.Run(t, new(IfTestSuite))
}

func (s *IfTestSuite) Test() {
	s.T().Run("TestIfExecute", func(t *testing.T) {
		assert := require.New(s.T())

		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/if-params.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)
		conditions, ok := np.Parameters["conditions"].(map[string]interface{})
		assert.True(ok)
		condition, ok := conditions["conditions"].([]interface{})
		assert.Equal(3, len(condition))

		ifNode := &ifnode.IfExecutor{}
		input := &structs.NodeExecuteInput{
			Params: np,
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
		result := ifNode.Execute(context.Background(), input)
		assert.Equal(2, len(result.ExecutorData[0]))

		np = &structs.WorkflowNode{}
		testFile, err = os.ReadFile("./test_files/if-params2.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)
		conditions, ok = np.Parameters["conditions"].(map[string]interface{})
		assert.True(ok)
		condition, ok = conditions["conditions"].([]interface{})
		assert.Equal(4, len(condition))

		ifNode = &ifnode.IfExecutor{}
		input = &structs.NodeExecuteInput{
			Params: np,
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
		result = ifNode.Execute(context.Background(), input)
		assert.Equal(1, len(result.ExecutorData[0]))
		assert.Equal(1, len(result.ExecutorData[1]))
	})

	s.T().Run("Test If Workflow Create and Execute", func(t *testing.T) {
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/if-e2e.json")
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
