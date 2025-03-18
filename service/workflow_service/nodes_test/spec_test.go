package nodes_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type NodeTestSuite struct {
	suite.Suite
}

func Test_ReportTestSuite(t *testing.T) {
	suite.Run(t, new(NodeTestSuite))
}

func (s *NodeTestSuite) TestSpec() {
	s.T().Run("TestAllNodesSpec", func(t *testing.T) {
		assert := require.New(s.T())

		var specs []structs.WorkflowNodeDescriptionSpec
		testFile, err := os.ReadFile("./test_files/nodes.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &specs)
		assert.Nil(err)
	})
}
