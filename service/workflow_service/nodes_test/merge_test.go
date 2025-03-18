package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go service/workflow_service/nodes_test/merge_test.go

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	mergenode "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/merge"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type MergeTestSuite struct {
	suite.Suite
}

func Test_MergeTestSuite(t *testing.T) {
	suite.Run(t, new(MergeTestSuite))
}

func (s *MergeTestSuite) Test() {
	s.T().Run("Test Merge mode of append", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/merge-params-append.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)

		mergeNode := &mergenode.MergeExecutor{}
		input := &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "a",
							"count": 1,
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "b",
							"count": 2,
						},
					},
				},
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "c",
							"count": 1,
							"num":   1.1,
							"extra": map[string]interface{}{
								"extra2": 1,
							},
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "d",
							"count": 2,
							"num":   2.2,
							"extra": map[string]interface{}{
								"extra2": 2,
							},
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "e",
							"count": 3,
							"num":   3.3,
							"extra": map[string]interface{}{
								"extra2": 3,
							},
						},
					},
				},
			},
		}

		result := mergeNode.Execute(context.Background(), input)
		assert.Equal(5, len(result.ExecutorData[0]))
	})

	s.T().Run("Test Merge mode of chooseBranch", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/merge-params-choose-branch.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)

		mergeNode := &mergenode.MergeExecutor{}
		input := &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "a",
							"count": 1,
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "b",
							"count": 2,
						},
					},
				},
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "c",
							"count": 1,
							"num":   1.1,
							"extra": map[string]interface{}{
								"extra2": 1,
							},
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "d",
							"count": 2,
							"num":   2.2,
							"extra": map[string]interface{}{
								"extra2": 2,
							},
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "e",
							"count": 3,
							"num":   3.3,
							"extra": map[string]interface{}{
								"extra2": 3,
							},
						},
					},
				},
			},
		}

		result := mergeNode.Execute(context.Background(), input)
		assert.Equal(2, len(result.ExecutorData[0]))

		np.Parameters["output"] = "input2"
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(3, len(result.ExecutorData[0]))

		np.Parameters["output"] = "empty"
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(1, len(result.ExecutorData[0]))
	})

	s.T().Run("Test Merge mode of comine", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// combinationMode: mergeByFields, field: name
		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/merge-params-combine.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)

		mergeNode := &mergenode.MergeExecutor{}
		input := &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "a",
							"count": 1,
							"extra": map[string]interface{}{
								"extra1": 1,
							},
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "b",
							"count": 2,
							"extra": map[string]interface{}{
								"extra1": 2,
							},
						},
					},
				},
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "c",
							"count": 1,
							"num":   1.1,
							"extra": map[string]interface{}{
								"extra2": 1,
							},
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "d",
							"count": 2,
							"num":   2.2,
							"extra": map[string]interface{}{
								"extra2": 2,
							},
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "e",
							"count": 3,
							"num":   3.3,
							"extra": map[string]interface{}{
								"extra2": 3,
							},
						},
					},
				},
			},
		}

		// result should be empty
		result := mergeNode.Execute(context.Background(), input)
		assert.Empty(result.ExecutorData[0])

		// combinationMode: mergeByFields, field: count, joinMode: keepMatches
		mergeByFields := map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"field1": "count",
					"field2": "count",
				},
			},
		}
		np.Parameters["mergeByFields"] = mergeByFields
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(2, len(result.ExecutorData[0]))

		// combinationMode: mergeByFields, field: count, joinMode: keepMatches, options.clashHandling.values.resolveClash=preferInput1
		np.Parameters["options"] = map[string]interface{}{
			"clashHandling": map[string]interface{}{
				"values": map[string]interface{}{
					"resolveClash": "preferInput1",
				},
			},
		}
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(2, len(result.ExecutorData[0]))

		// combinationMode: mergeByFields, field: count, joinMode: keepMatches, options.clashHandling.values.mergeMode=shallowMerge
		np.Parameters["options"] = map[string]interface{}{
			"clashHandling": map[string]interface{}{
				"values": map[string]interface{}{
					"mergeMode": "shallowMerge",
				},
			},
		}
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(2, len(result.ExecutorData[0]))

		np.Parameters["options"] = map[string]interface{}{}
		// combinationMode: mergeByFields, field: count, joinMode: keepNonMatches
		np.Parameters["joinMode"] = "keepNonMatches"
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(1, len(result.ExecutorData[0]))

		// combinationMode: mergeByFields, field: count, joinMode: keepNonMatches, outputDataFrom: input1
		np.Parameters["joinMode"] = "keepNonMatches"
		np.Parameters["outputDataFrom"] = "input1"
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(0, len(result.ExecutorData[0]))

		// combinationMode: mergeByFields, field: count, joinMode: keepEverything
		np.Parameters["joinMode"] = "keepEverything"
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(3, len(result.ExecutorData[0]))

		// combinationMode: mergeByFields, field: count, joinMode: enrichInput1
		np.Parameters["joinMode"] = "enrichInput1"
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(2, len(result.ExecutorData[0]))

		// combinationMode: mergeByFields, field: count, joinMode: enrichInput2
		np.Parameters["joinMode"] = "enrichInput2"
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(3, len(result.ExecutorData[0]))

		// combinationMode: mergeByPosition
		np.Parameters["combinationMode"] = "mergeByPosition"
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(2, len(result.ExecutorData[0]))

		// combinationMode: multiplex
		np.Parameters["combinationMode"] = "multiplex"
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(6, len(result.ExecutorData[0]))
	})

	s.T().Run("Test Merge mode of comine with wrong data format", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/merge-params-combine.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)

		mergeNode := &mergenode.MergeExecutor{}
		input := &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"name":  "a",
							"count": 1,
						},
					},
				},
				{
					structs.NodeSingleData{
						"json": []map[string]interface{}{
							{
								"name":  "c",
								"count": 2,
							},
						},
					},
				},
			},
		}

		// combinationMode: mergeByPosition
		np.Parameters["combinationMode"] = "mergeByPosition"
		result := mergeNode.Execute(context.Background(), input)
		// Result should be data2 as default, because the format of data1 and data2 are not match.
		assert.Equal(1, len(result.ExecutorData[0]))
		resultRaw, err := json.Marshal(result.ExecutorData[0])
		assert.Nil(err)
		assert.Equal("[{\"json\":[{\"count\":2,\"name\":\"c\"}]}]", string(resultRaw))

		// If set the clashHandling to preferInput1, Result shoule be data1.
		np.Parameters["options"] = map[string]interface{}{
			"clashHandling": map[string]interface{}{
				"values": map[string]interface{}{
					"resolveClash": "preferInput1",
				},
			},
		}
		result = mergeNode.Execute(context.Background(), input)
		assert.Equal(1, len(result.ExecutorData[0]))
		resultRaw, err = json.Marshal(result.ExecutorData[0])
		assert.Nil(err)
		assert.Equal("[{\"json\":{\"count\":1,\"name\":\"a\"}}]", string(resultRaw))
	})

	s.T().Run("Test Merge Workflow Create and Execute", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// Create workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/merge-create-workflow.json")
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
		assert.Len(execution.Data.ResultData.RunData, 6)
		assert.Len(execution.Data.ResultData.RunData["Merge"][0].Data["main"][0], 5)
		assert.Len(execution.Data.ResultData.RunData["Merge1"][0].Data["main"][0], 2)
		assert.Len(execution.Data.ResultData.RunData["Merge2"][0].Data["main"][0], 3)

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)
	})
}
