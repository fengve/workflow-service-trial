package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go service/workflow_service/nodes_test/manual_trigger_test.go

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sharedLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/manual_trigger"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type ManualTriggerNodeTestSuite struct {
	suite.Suite
	organization *sharedLib.IdentityOrganization
}

func Test_ManualTriggerNode(t *testing.T) {
	suite.Run(t, new(ManualTriggerNodeTestSuite))
}

func (s *ManualTriggerNodeTestSuite) Test() {
	s.T().Run("TestManualTriggerSpec", func(t *testing.T) {
		assert := require.New(s.T())

		var manualTriggerSpec structs.WorkflowNodeDescriptionSpec
		testFile, err := os.ReadFile("./test_files/manual-trigger-node.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &manualTriggerSpec)
		assert.Nil(err)

		assert.Empty(manualTriggerSpec.EventTriggerDescription)
		assert.Equal(1, manualTriggerSpec.MaxNodes)
		manualDefaults := structs.DescriptionDefaults{
			Name:  "When clicking \"Execute Workflow\"",
			Color: "#909298",
		}
		assert.Equal(manualDefaults, manualTriggerSpec.Defaults)
		assert.Equal("This node is where a manual workflow execution starts. To make one, go back to the canvas and click ‘execute workflow’", manualTriggerSpec.Properties[0].DisplayName)
		assert.Equal("notice", manualTriggerSpec.Properties[0].Name)
		assert.Equal("notice", manualTriggerSpec.Properties[0].Type)
		assert.Empty(manualTriggerSpec.Properties[0].Default)
	})

	s.T().Run("TestManualTriggerGenerate", func(t *testing.T) {
		assert := require.New(s.T())

		_ = manual_trigger.ManualTrigger{}
		executor := core.NewExecutor(manual_trigger.Name)
		node := executor.GetNode()
		assert.Equal("Manual Trigger", node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.DisplayName)
	})

	s.T().Run("TestManualTriggerExecute", func(t *testing.T) {
		assert := require.New(s.T())

		mt := manual_trigger.ManualTrigger{}
		data := mt.Execute(context.Background(), nil).TriggerData
		raw, err := json.Marshal(data)
		assert.Nil(err)
		assert.Equal("[{\"json\":{}}]", string(raw))
	})
}
