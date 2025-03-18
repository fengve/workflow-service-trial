package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go  service/workflow_service/nodes_test/respond_to_webhook_test.go

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/respond_to_webhook"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	"github.com/valyala/fasthttp"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
)

type RespondToWebhookTestSuite struct {
	suite.Suite
}

func Test_RespondToWebhook(t *testing.T) {
	suite.Run(t, new(RespondToWebhookTestSuite))
}

func (s *RespondToWebhookTestSuite) Test() {
	s.T().Run("TestRespondToWebhookSpec", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		var nodeSpec structs.WorkflowNodeDescriptionSpec
		testFile, err := os.ReadFile("./test_files/respond-to-webhook.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &nodeSpec)
		assert.Nil(err)
		assert.Equal("", nodeSpec.EventTriggerDescription)
		assert.Equal(0, nodeSpec.MaxNodes)
	})

	s.T().Run("TestRespondToWebhookGenerate", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		executor := core.NewExecutor(respond_to_webhook.Name)
		node := executor.GetNode()

		assert.Equal("Respond to Webhook", node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.DisplayName)
	})

	s.T().Run("TestRespondToWebhookExecute", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		var recorded *fasthttp.Response
		node := respond_to_webhook.RespondToWebhook{}
		result := node.Execute(context.Background(), &structs.NodeExecuteInput{
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"webhookId": "123",
						},
					},
				},
			},
			Params: &structs.WorkflowNode{},
			AdditionalData: &structs.WorkflowExecuteAdditionalData{
				Hooks: structs.WorkflowHooks{
					HookFunctions: structs.WorkflowExecuteHooks{
						SendResponse: []func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response){
							func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response) {
								recorded = response
							},
						},
					},
				},
			},
		})

		raw, err := json.Marshal(result.TriggerData)
		assert.Nil(err)
		assert.Equal("[{\"json\":{\"webhookId\":\"123\"}}]", string(raw))
		raw, err = json.Marshal(result.ExecutorData)
		assert.Nil(err)
		assert.Equal("[[{\"json\":{\"webhookId\":\"123\"}}]]", string(raw))
		assert.Equal([]byte("{\"webhookId\":\"123\"}"), recorded.Body())
		assert.Equal(200, recorded.StatusCode())
		assert.Equal("application/json; charset=utf-8", string(recorded.Header.Peek("Content-Type")))
	})

	s.T().Run("TestRespondToWebhookExecute respondWith firstIncomingItem", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		var recorded *fasthttp.Response
		node := respond_to_webhook.RespondToWebhook{}
		node.Execute(context.Background(), &structs.NodeExecuteInput{
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"webhookId": "123",
						},
					},
				},
			},
			Params: &structs.WorkflowNode{
				Parameters: map[string]interface{}{
					"respondWith": "firstIncomingItem",
				},
			},
			AdditionalData: &structs.WorkflowExecuteAdditionalData{
				Hooks: structs.WorkflowHooks{
					HookFunctions: structs.WorkflowExecuteHooks{
						SendResponse: []func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response){
							func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response) {
								recorded = response
							},
						},
					},
				},
			},
		})

		assert.Equal([]byte("{\"webhookId\":\"123\"}"), recorded.Body())
		assert.Equal(200, recorded.StatusCode())
		assert.Equal("application/json; charset=utf-8", string(recorded.Header.Peek("Content-Type")))
	})

	s.T().Run("TestRespondToWebhookExecute respondWith allIncomingItems", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		var recorded *fasthttp.Response
		node := respond_to_webhook.RespondToWebhook{}
		node.Execute(context.Background(), &structs.NodeExecuteInput{
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"webhookId": "123",
					},
				},
			},
			Params: &structs.WorkflowNode{
				Parameters: map[string]interface{}{
					"respondWith": "allIncomingItems",
				},
			},
			AdditionalData: &structs.WorkflowExecuteAdditionalData{
				Hooks: structs.WorkflowHooks{
					HookFunctions: structs.WorkflowExecuteHooks{
						SendResponse: []func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response){
							func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response) {
								recorded = response
							},
						},
					},
				},
			},
		})

		assert.Equal([]byte("[{\"webhookId\":\"123\"}]"), recorded.Body())
		assert.Equal(200, recorded.StatusCode())
		assert.Equal("application/json; charset=utf-8", string(recorded.Header.Peek("Content-Type")))
	})

	s.T().Run("TestRespondToWebhookExecute respondWith json", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		var recorded *fasthttp.Response
		node := respond_to_webhook.RespondToWebhook{}
		node.Execute(context.Background(), &structs.NodeExecuteInput{
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{},
				},
			},
			Params: &structs.WorkflowNode{
				Parameters: map[string]interface{}{
					"respondWith":  "json",
					"responseBody": "{\"number\":1}",
				},
			},
			AdditionalData: &structs.WorkflowExecuteAdditionalData{
				Hooks: structs.WorkflowHooks{
					HookFunctions: structs.WorkflowExecuteHooks{
						SendResponse: []func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response){
							func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response) {
								recorded = response
							},
						},
					},
				},
			},
		})

		assert.Equal([]byte("{\"number\":1}"), recorded.Body())
		assert.Equal(200, recorded.StatusCode())
		assert.Equal("application/json; charset=utf-8", string(recorded.Header.Peek("Content-Type")))
	})

	s.T().Run("TestRespondToWebhookExecute respondWith text", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		var recorded *fasthttp.Response
		node := respond_to_webhook.RespondToWebhook{}
		node.Execute(context.Background(), &structs.NodeExecuteInput{
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{},
				},
			},
			Params: &structs.WorkflowNode{
				Parameters: map[string]interface{}{
					"respondWith":  "text",
					"responseBody": "Workflow completed",
				},
			},
			AdditionalData: &structs.WorkflowExecuteAdditionalData{
				Hooks: structs.WorkflowHooks{
					HookFunctions: structs.WorkflowExecuteHooks{
						SendResponse: []func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response){
							func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response) {
								recorded = response
							},
						},
					},
				},
			},
		})

		assert.Equal([]byte("Workflow completed"), recorded.Body())
		assert.Equal(200, recorded.StatusCode())
		assert.Equal("text/html; charset=utf-8", string(recorded.Header.Peek("Content-Type")))
	})

	s.T().Run("TestRespondToWebhookExecute respondWith redirect", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		var recorded *fasthttp.Response
		node := respond_to_webhook.RespondToWebhook{}
		node.Execute(context.Background(), &structs.NodeExecuteInput{
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{},
				},
			},
			Params: &structs.WorkflowNode{
				Parameters: map[string]interface{}{
					"respondWith": "redirect",
					"redirectURL": "localhost",
				},
			},
			AdditionalData: &structs.WorkflowExecuteAdditionalData{
				Hooks: structs.WorkflowHooks{
					HookFunctions: structs.WorkflowExecuteHooks{
						SendResponse: []func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response){
							func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response) {
								recorded = response
							},
						},
					},
				},
			},
		})

		assert.Equal([]byte(""), recorded.Body())
		assert.Equal(307, recorded.StatusCode())
		assert.Equal("localhost", string(recorded.Header.Peek("location")))
	})

	s.T().Run("TestRespondToWebhookExecute respondWith noData", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		var recorded *fasthttp.Response
		node := respond_to_webhook.RespondToWebhook{}
		result := node.Execute(context.Background(), &structs.NodeExecuteInput{
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{},
				},
			},
			Params: &structs.WorkflowNode{
				Parameters: map[string]interface{}{
					"respondWith": "noData",
				},
			},
			AdditionalData: &structs.WorkflowExecuteAdditionalData{
				Hooks: structs.WorkflowHooks{
					HookFunctions: structs.WorkflowExecuteHooks{
						SendResponse: []func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response){
							func(ctx context.Context, hooks *structs.WorkflowHooks, response *fasthttp.Response) {
								recorded = response
							},
						},
					},
				},
			},
		})

		raw, err := json.Marshal(result.ExecutorData)
		assert.Nil(err)
		assert.Equal("[[{\"json\":{}}]]", string(raw))
		assert.Equal([]byte(""), recorded.Body())
		assert.Equal(200, recorded.StatusCode())
	})

	s.T().Run("TestRespondToWebhookExecute with unknown respondWith value", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		node := respond_to_webhook.RespondToWebhook{}
		result := node.Execute(context.Background(), &structs.NodeExecuteInput{
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"webhookId": "123",
					},
				},
			},
			Params: &structs.WorkflowNode{
				Parameters: map[string]interface{}{
					"respondWith": "t",
				},
			},
		})

		assert.Equal(structs.WorkflowNodeExecutionError{
			Name:    respond_to_webhook.Name,
			Message: "Unknown value of respondWith parameter: t",
		}, result.Errors[0])
		assert.Nil(result.TriggerData)
		assert.Nil(result.ExecutorData)
	})

	s.T().Run("TestRespondToWebhookExecute with node not found", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		node := respond_to_webhook.RespondToWebhook{}
		result := node.Execute(context.Background(), &structs.NodeExecuteInput{
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"webhookId": "123",
					},
				},
			},
		})

		assert.Equal(structs.WorkflowNodeExecutionError{
			Name:    respond_to_webhook.Name,
			Message: "node not found",
		}, result.Errors[0])
		assert.Nil(result.TriggerData)
		assert.Nil(result.ExecutorData)
	})
}
