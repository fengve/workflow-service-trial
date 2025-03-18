package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go service/workflow_service/nodes_test/switch_test.go

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	switchnode "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/switch"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type SwitchTestSuite struct {
	suite.Suite
}

func Test_Switch(t *testing.T) {
	suite.Run(t, new(SwitchTestSuite))
}

func (s *SwitchTestSuite) Test() {
	s.T().Run("TestSwitchSpec", func(t *testing.T) {
	})

	s.T().Run("TestSwitchGenerate", func(t *testing.T) {
		assert := require.New(s.T())

		_ = switchnode.SwitchExecutor{}
		executor := core.NewExecutor(switchnode.Name)
		node := executor.GetNode()
		assert.Equal(13, len(node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.Properties))
		assert.Equal("rules", node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.Properties[4].DisplayOptions.Show["mode"][0])
	})

	s.T().Run("TestSwitchExecute", func(t *testing.T) {

		assert := require.New(s.T())

		inputData := []structs.NodeData{
			{
				structs.NodeSingleData{
					"json": map[string]interface{}{
						"id":   1,
						"date": "2024-03-22T00:00:00",
					},
				},
				structs.NodeSingleData{
					"json": map[string]interface{}{
						"id":   2,
						"date": "2024-03-23T00:00:00",
					},
				},
				structs.NodeSingleData{
					"json": map[string]interface{}{
						"id":   3,
						"date": nil,
					},
				},
			},
		}
		// empty
		switchNode := &switchnode.SwitchExecutor{}
		input := &structs.NodeExecuteInput{
			Params: &structs.WorkflowNode{},
			Data:   inputData,
		}
		result := switchNode.Execute(context.Background(), input)
		assert.Equal(structs.WorkflowExecutionStatus_Success, result.ExecutionStatus)
		assert.Equal(0, len(result.ExecutorData))

		// rule mode 1
		params := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/switch-params.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &params)
		assert.Nil(err)
		rule, ok := params.Parameters["rules"].(map[string]interface{})
		assert.True(ok)
		rules, ok := rule["rules"].([]interface{})
		assert.Equal(3, len(rules))

		switchNode = &switchnode.SwitchExecutor{}
		input = &structs.NodeExecuteInput{
			Params: params,
			Data:   inputData,
		}
		result = switchNode.Execute(context.Background(), input)
		assert.Equal(structs.WorkflowExecutionStatus_Success, result.ExecutionStatus)
		assert.Equal(1, len(result.ExecutorData[0]))
		assert.Equal(0, len(result.ExecutorData[1]))
		assert.Equal(1, len(result.ExecutorData[2]))

		// rule mode 2
		params = &structs.WorkflowNode{}
		testFile, err = os.ReadFile("./test_files/switch-params2.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &params)
		assert.Nil(err)
		rule, ok = params.Parameters["rules"].(map[string]interface{})
		assert.True(ok)
		rules, ok = rule["rules"].([]interface{})
		assert.Equal(2, len(rules))

		switchNode = &switchnode.SwitchExecutor{}

		input = &structs.NodeExecuteInput{
			Params: params,
			Data:   inputData,
		}

		result = switchNode.Execute(context.Background(), input)
		assert.Equal(structs.WorkflowExecutionStatus_Success, result.ExecutionStatus)
		assert.Equal(1, len(result.ExecutorData[0]))
		assert.Equal(2, len(result.ExecutorData[1]))

		// expression mode
		params = &structs.WorkflowNode{}
		testFile, err = os.ReadFile("./test_files/switch-params3.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &params)
		assert.Nil(err)

		switchNode = &switchnode.SwitchExecutor{}

		input = &structs.NodeExecuteInput{
			Params: params,
			Data:   inputData,
		}

		result = switchNode.Execute(context.Background(), input)
		assert.Equal(structs.WorkflowExecutionStatus_Success, result.ExecutionStatus)
		assert.Equal(0, len(result.ExecutorData[0]))
		assert.Equal(1, len(result.ExecutorData[1]))
		assert.Equal(1, len(result.ExecutorData[2]))
		assert.Equal(1, len(result.ExecutorData[3]))
	})

	s.T().Run("Test Switch Workflow Create and Execute", func(t *testing.T) {

		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/switch-e2e.json")
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
		assert.Len(execution.Data.ResultData.RunData, 7)

		resData1 := execution.Data.ResultData.RunData["HTML CREATE OFFER"][0].Data["main"][0][0]
		resData2 := execution.Data.ResultData.RunData["HTML UPDATE OFFER"][0].Data["main"][0][0]
		resData3 := execution.Data.ResultData.RunData["HTML CREATE ENTITLEMENT"][0].Data["main"][0][0]
		resData4 := execution.Data.ResultData.RunData["HTML CANCEL ENTITLEMENT"][0].Data["main"][0][0]

		assert.Contains(resData1["json"].(map[string]interface{})["html"], "OFFER CREATE", "EVENT ID : 1")
		assert.Contains(resData2["json"].(map[string]interface{})["html"], "OFFER UPDATE", "EVENT ID : 2")
		assert.Contains(resData3["json"].(map[string]interface{})["html"], "ENTITLEMENT CREATE", "EVENT ID : 3")
		assert.Contains(resData4["json"].(map[string]interface{})["html"], "ENTITLEMENT CANCEL", "EVENT ID : 4")

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)
	})
}
