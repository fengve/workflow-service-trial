package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go service/workflow_service/nodes_test/aggregate_test.go

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	aggregatenode "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/aggregate"
	"github.com/sugerio/workflow-service-trial/shared"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type AggregateTestSuite struct {
	suite.Suite
	organization *rdsDbLib.IdentityOrganization
}

func Test_AggregateNode(t *testing.T) {
	suite.Run(t, new(AggregateTestSuite))
}

// run before the tests in the suite are run
func (s *AggregateTestSuite) SetupSuite() {
	// Create a new organization
	s.organization = structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
}

func (s *AggregateTestSuite) Test() {
	s.T().Run("TestAggregateExecute aggregateAllItemData", func(t *testing.T) {
		assert := require.New(s.T())

		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/aggregate-params.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)

		node := &aggregatenode.AggregateExecutor{}
		input := &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "a",
							"count": 1,
							"num":   1.1,
							"arr":   []int{1, 2},
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "b",
							"count": 2,
							"arr":   []int{2, 3},
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "c",
							"count": 3,
						},
					},
				},
			},
		}
		result := node.Execute(context.Background(), input)
		assert.Equal(1, len(result.ExecutorData[0]))
		resultData := result.ExecutorData[0][0]
		list := resultData["json"].(map[string]interface{})["data"]
		assert.IsType([]map[string]interface{}{}, list)
		assert.Equal(3, len(list.([]map[string]interface{})))
	})

	s.T().Run("TestAggregateExecute aggregateIndividualFields", func(t *testing.T) {
		assert := require.New(s.T())

		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/aggregate-params2.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)

		node := &aggregatenode.AggregateExecutor{}
		input := &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "a",
							"count": 1,
							"num":   1.1,
							"arr":   []int{1, 2},
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "b",
							"count": 2,
							"arr":   []int{2, 3},
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "c",
							"count": 3,
						},
					},
				},
			},
		}
		result := node.Execute(context.Background(), input)
		resultData := result.ExecutorData[0][0]
		jsonObject := resultData["json"].(map[string]interface{})
		assert.Equal(3, len(jsonObject))
	})
}

func (s *AggregateTestSuite) Test_AggregateWorkflow() {
	s.T().Run("TestAggregateWorkflow E2E ManualRun", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// Create Workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, s.organization.ID, "./test_files/aggregate-create-workflow.json")
		assert.NotNil(newWorkflow)
		assert.Nil(err)
		workflowID := newWorkflow.ID

		// Manual run workflow
		executionId, err := api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotEmpty(executionId)

		// Check the execution result
		execution, err := api.GetWorkflowExecution_Testing(testFiberLambda, s.organization.ID, executionId)
		assert.Nil(err)
		assert.NotNil(execution)
		assert.NotNil(execution.Data)
		assert.NotNil(execution.Data.ResultData)
		aggregateResult := execution.Data.ResultData.RunData["Aggregate"]
		assert.Len(aggregateResult, 2)
		for _, result := range aggregateResult {
			resultArray, _ := result.Data["main"]
			assert.Len(resultArray, 1)
			resultItem := resultArray[0]
			resultItemFirst := resultItem[0]
			jsonVal, _ := resultItemFirst["json"].(map[string]interface{})
			keys, _ := jsonVal["key"].([]any)
			assert.Len(keys, 5)
		}

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, s.organization.ID, workflowID)
		assert.Nil(err)
	})

	s.T().Run("TestAggregateWorkflow E2E Activate Deactivate", func(t *testing.T) {
		t.Parallel()
		// Only run this test in local test environment since it may trigger the cron workflow in temporal service.
		if environment.Env != shared.ENV_LOCAL_TEST {
			t.Skip()
		}
		assert := require.New(s.T())
		// Create Workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, s.organization.ID, "./test_files/aggregate-cron-create-workflow.json")
		assert.NotNil(newWorkflow)
		assert.Nil(err)
		workflowID := newWorkflow.ID

		// Activate the workflow
		err = api.ActivateWorkflow_Testing(testFiberLambda, s.organization.ID, workflowID)
		assert.Nil(err)

		// Wait for 8 seconds to let the cron job run.
		time.Sleep(8 * time.Second)

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
