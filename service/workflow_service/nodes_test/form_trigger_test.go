package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go service/workflow_service/nodes_test/form_trigger_test.go

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	formTriggerNode "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/form_trigger"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type FormTriggerTestSuite struct {
	suite.Suite
}

func Test_FormTriggerNode(t *testing.T) {
	suite.Run(t, new(IfTestSuite))
}

func (s *FormTriggerTestSuite) Test() {
	s.T().Run("TestFormTriggerExecute", func(t *testing.T) {
		assert := require.New(s.T())

		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/form-trigger.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)
		conditions, ok := np.Parameters["conditions"].(map[string]interface{})
		assert.True(ok)
		condition, ok := conditions["conditions"].([]interface{})
		assert.Equal(3, len(condition))

		executor := core.NewExecutor(formTriggerNode.Name)
		node := executor.GetNode()
		assert.Equal("Form Trigger", node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.DisplayName)

	})

}
