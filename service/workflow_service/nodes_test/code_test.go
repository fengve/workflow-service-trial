package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go service/workflow_service/nodes_test/code_test.go

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	codenode "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/code"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type CodeTestSuite struct {
	suite.Suite
}

func Test_Code(t *testing.T) {
	suite.Run(t, new(CodeTestSuite))
}

func (s *CodeTestSuite) Test() {
	s.T().Run("TestCodeSpec", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		var codeSpec structs.WorkflowNodeDescriptionSpec
		testFile, err := os.ReadFile("./test_files/code.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &codeSpec)
		assert.Nil(err)

		assert.Equal(11, len(codeSpec.Properties))
		assert.Equal("runOnceForAllItems", codeSpec.Properties[0].Default)
	})

	s.T().Run("TestCodeGenerate", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		_ = codenode.CodeExecutor{}
		executor := core.NewExecutor(codenode.Name)
		node := executor.GetNode()
		assert.Equal("Code", node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.Defaults.Name)
		assert.Equal("/icons/embed/n8n-nodes-base.code/code.svg", node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.IconUrl)
	})

	s.T().Run("TestCodeFirstOfEmptyList", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		executor := codenode.CodeExecutor{}

		rawData := `{"parameters":{"jsCode": "return [{\"json\": {\"first\": $input.first().json, \"last\": $input.last().json }}]"}}`
		params := structs.WorkflowNode{}
		err := json.Unmarshal([]byte(rawData), &params)
		assert.Nil(err)

		input := &structs.NodeExecuteInput{
			Data: []structs.NodeData{
				{},
			},
			Params: &params,
		}

		result := executor.Execute(context.Background(), input)
		assert.Equal(structs.WorkflowExecutionStatus_Success, result.ExecutionStatus)
		jsonField := result.ExecutorData[0][0]["json"].(map[string]interface{})
		assert.Nil(jsonField["first"])
		assert.Nil(jsonField["last"])
	})

	s.T().Run("TestCodeExecute", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		executor := codenode.CodeExecutor{}

		// Test success
		{
			rawData := `{"parameters":{"jsCode": "return {data: 1}"}}`
			params := structs.WorkflowNode{}
			err := json.Unmarshal([]byte(rawData), &params)
			assert.Nil(err)
			input := &structs.NodeExecuteInput{
				Data: []structs.NodeData{
					structs.NodeData{
						structs.NodeSingleData{},
					},
				},
				Params: &params,
			}
			result := executor.Execute(context.Background(), input)
			assert.Equal(structs.WorkflowExecutionStatus_Success, result.ExecutionStatus)
			assert.Equal(structs.NodeData{map[string]interface{}{"json": map[string]interface{}{"data": int64(1)}}}, result.ExecutorData[0])
		}

		// Test failure
		{
			rawData := `{"parameters":{"jsCode": "throw new Error('error')"}}`
			params := structs.WorkflowNode{}
			err := json.Unmarshal([]byte(rawData), &params)
			assert.Nil(err)
			input := &structs.NodeExecuteInput{
				Data: []structs.NodeData{
					structs.NodeData{
						structs.NodeSingleData{},
					},
				},
				Params: &params,
			}
			result := executor.Execute(context.Background(), input)
			assert.Equal(structs.WorkflowExecutionStatus_Failed, result.ExecutionStatus)
			assert.Equal("Error: error [line 1]", result.Errors[0].Message)
		}

	})
}
