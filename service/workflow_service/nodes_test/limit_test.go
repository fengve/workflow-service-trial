package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go service/workflow_service/nodes_test/limit_test.go

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sharedLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	limitnode "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/limit"
	"github.com/sugerio/workflow-service-trial/shared"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type LimitNodeTestSuite struct {
	suite.Suite
	organization *sharedLib.IdentityOrganization
}

func Test_LimitNode(t *testing.T) {
	suite.Run(t, new(LimitNodeTestSuite))
}

func (s *LimitNodeTestSuite) SetupSuite() {
	// Create Organization for test
	s.organization = structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
}

func (s *LimitNodeTestSuite) Test() {
	s.T().Run("TestLimitExecute", func(t *testing.T) {
		assert := require.New(s.T())

		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/limit-params.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)

		limitNode := &limitnode.LimitExecutor{}
		input := &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "a",
							"count": 2,
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "b",
							"count": 3,
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "c",
							"count": 4,
						},
					},
				},
			},
		}
		result := limitNode.Execute(context.Background(), input)
		assert.Equal(2, len(result.ExecutorData[0]))
		item1 := result.ExecutorData[0][0]
		assert.Equal("a", item1["json"].(map[string]interface{})["name"])
		assert.Equal(2, item1["json"].(map[string]interface{})["count"])
		item2 := result.ExecutorData[0][1]
		assert.Equal("b", item2["json"].(map[string]interface{})["name"])
		assert.Equal(3, item2["json"].(map[string]interface{})["count"])

	})
}

func (s *LimitNodeTestSuite) Test_Limit_Workflow() {
	s.T().Run("TestLimit E2E ManualRun", func(t *testing.T) {
		if environment.Env != shared.ENV_LOCAL_TEST {
			t.Skip()
		}
		assert := require.New(s.T())

		// Create Workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, s.organization.ID, "./test_files/limit-create-workflow.json")
		assert.Nil(err)
		workflowID := newWorkflow.ID

		// Manual run workflow
		executionId, err := api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotEmpty(executionId)

		// get workflow execution
		execution, err := api.GetWorkflowExecution_Testing(testFiberLambda, s.organization.ID, executionId)
		assert.Nil(err)
		assert.Equal(structs.WorkflowExecutionStatus_Success, execution.Status)
		// check result
		assert.NotNil(execution.Data)
		assert.NotNil(execution.Data.ResultData)
		assert.Equal(1, len(execution.Data.ResultData.RunData["Limit"]))
		mainResult := execution.Data.ResultData.RunData["Limit"][0].Data
		resultVal := mainResult["main"]
		assert.Len(resultVal, 1)
		result := resultVal[0]
		assert.Len(result, 2)

		// check value
		item1 := result[0]
		assert.Equal("value0", item1["json"].(map[string]interface{})["key"])
		assert.Equal(0.0, item1["json"].(map[string]interface{})["val"])
		item2 := result[1]
		assert.Equal("value1", item2["json"].(map[string]interface{})["key"])
		assert.Equal(1.0, item2["json"].(map[string]interface{})["val"])

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, s.organization.ID, workflowID)
		assert.Nil(err)
	})

	s.T().Run("TestLimit E2E Activate Deactivate", func(t *testing.T) {
		t.Parallel()
		if environment.Env != shared.ENV_LOCAL_TEST {
			t.Skip()
		}
		assert := require.New(s.T())

		// Create Workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, s.organization.ID, "./test_files/limit-cron-create-workflow.json")
		assert.Nil(err)
		workflowID := newWorkflow.ID
		assert.NotEmpty(workflowID)

		// Activate the workflow
		err = api.ActivateWorkflow_Testing(testFiberLambda, s.organization.ID, workflowID)
		assert.Nil(err)

		// Wait for 10 seconds to let the cron job run.
		time.Sleep(5 * time.Second)

		// Check Execution Entity
		count, err := rdsDbQueries.CountWorkflowExecutionEntitiesByWorkflowId(context.Background(), workflowID)
		assert.Nil(err)
		assert.Equal(count, int64(1))

		// Deactivate the workflow
		err = api.DeactivateWorkflow_Testing(testFiberLambda, s.organization.ID, workflowID)
		assert.Nil(err)

		time.Sleep(5 * time.Second)

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, s.organization.ID, workflowID)
		assert.Nil(err)
	})
}
