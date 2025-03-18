package api_test

// Command to run this test only
// go test -v service/workflow_service/api/service_test.go service/workflow_service/api/node_test.go

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/code"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type NodeTestSuit struct {
	suite.Suite
}

func Test_NodeTestSuit(t *testing.T) {
	suite.Run(t, new(NodeTestSuit))
}

func (s *NodeTestSuit) Test() {
	s.T().Run("TestGetWorkflowNodesJson", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		request := events.APIGatewayProxyRequest{}
		request.HTTPMethod = http.MethodGet
		request.Path = "/workflow/public/nodes.json"
		request.Headers = map[string]string{"Content-Type": "application/json"}

		resp, err := testFiberLambda.Proxy(request)

		assert.Nil(err)
		assert.Equal(200, resp.StatusCode)
		var nodes []structs.WorkflowNodeDescriptionSpec
		err = json.Unmarshal([]byte(resp.Body), &nodes)
		assert.Nil(err)
		assert.Greater(len(nodes), 0)
		assert.NotEmpty(resp.Body)
		for _, node := range nodes {
			assert.NotEmpty(node.Name)
			assert.NotNil(node.Codex)
			assert.NotEmpty(node.Codex.Categories)
			assert.NotEmpty(node.Description)
			assert.NotEmpty(node.DisplayName)
			if node.IconUrl == "" {
				assert.NotEmpty(node.Icon)
			}
		}
	})

	s.T().Run("TestGetNodeIcons", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		request := events.APIGatewayProxyRequest{}
		request.HTTPMethod = http.MethodGet
		request.Path = "/workflow/public/nodes/icons/embed/n8n-nodes-base.code/code.svg"

		resp, err := testFiberLambda.Proxy(request)
		assert.Nil(err)
		assert.Equal(200, resp.StatusCode)
		assert.NotEmpty(resp.Body)

		// invalid path 400
		request = events.APIGatewayProxyRequest{}
		request.HTTPMethod = http.MethodGet
		request.Path = "/workflow/public/nodes/icons/embed/n8n-nodes-base.code"
		resp, err = testFiberLambda.Proxy(request)
		assert.Nil(err)
		assert.Equal(400, resp.StatusCode)
		assert.Equal("invalid path:embed/n8n-nodes-base.code", resp.Body)

		// not exist node name 404
		request = events.APIGatewayProxyRequest{}
		request.HTTPMethod = http.MethodGet
		request.Path = "/workflow/public/nodes/icons/embed/not-exist/node.png"
		resp, err = testFiberLambda.Proxy(request)
		assert.Nil(err)
		assert.Equal(404, resp.StatusCode)
		assert.Equal("icon not found:embed/not-exist/node.png", resp.Body)

		// invalid path 400
		request = events.APIGatewayProxyRequest{}
		request.HTTPMethod = http.MethodGet
		request.Path = "/workflow/public/nodes/icons/invalid/path"
		resp, err = testFiberLambda.Proxy(request)
		assert.Nil(err)
		assert.Equal(400, resp.StatusCode)
		assert.Equal("invalid path:invalid/path", resp.Body)

		// n8n icon path 400
		request = events.APIGatewayProxyRequest{}
		request.HTTPMethod = http.MethodGet
		request.Path = "/workflow/public/nodes/icons/n8n-nodes-base/dist/nodes/Suger/suger.svg"
		resp, err = testFiberLambda.Proxy(request)
		assert.Nil(err)
		assert.Equal(400, resp.StatusCode)
		assert.Equal("n8n node icon not supported:n8n-nodes-base/dist/nodes/Suger/suger.svg", resp.Body)
	})
}
