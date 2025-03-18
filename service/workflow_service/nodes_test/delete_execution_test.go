package nodes_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/delete_execution"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

func (s *NodeTestSuite) TestDeleteExecution() {
	s.T().Run("TestDeleteExecutionSpec", func(t *testing.T) {
		assert := require.New(s.T())

		var deleteSpec structs.WorkflowNodeDescriptionSpec
		testFile, err := os.ReadFile("./test_files/delete-execution-node.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &deleteSpec)
		assert.Nil(err)

		assert.Equal("Delete Execution", deleteSpec.DisplayName)
		assert.Equal("n8n-nodes-base.deleteExecution", deleteSpec.Name)
		assert.Equal("fa:trash", deleteSpec.Icon)
		assert.Equal("transform", deleteSpec.Group[0])
		assert.Equal(float64(1), deleteSpec.Version)
		assert.Equal("Delete the execution record", deleteSpec.Description)
		assert.Equal("Delete Execution", deleteSpec.Defaults.Name)
		assert.Equal("main", deleteSpec.Inputs[0])
		assert.Empty(deleteSpec.Outputs)
		assert.Empty(deleteSpec.Properties)
	})

	s.T().Run("TestDeleteExecutionGenerate", func(t *testing.T) {
		assert := require.New(s.T())

		_ = delete_execution.DeleteExecutionExecutor{}
		executor := core.NewExecutor(delete_execution.Name)
		node := executor.GetNode()
		assert.Equal("Delete Execution", node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.Defaults.Name)
	})

	s.T().Run("TestDeleteExecutionExecute", func(t *testing.T) {
		assert := require.New(s.T())

		dtt := delete_execution.DeleteExecutionExecutor{}
		data := dtt.Execute(context.Background(), nil).ExecutorData
		raw, err := json.Marshal(data)
		assert.Nil(err)
		assert.Equal("[]", string(raw))
	})
}
