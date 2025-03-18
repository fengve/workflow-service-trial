package webhook

import (
	"context"
	_ "embed"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	"github.com/valyala/fasthttp"
)

const (
	// Category is the category of WebhookNode.
	Category = structs.CategoryTrigger

	// Name is the name of WebhookNode.
	Name = "n8n-nodes-base.webhook"
)

var (
	//go:embed node.json
	rawJson []byte

	//go:embed webhook.svg
	icon []byte
)

type Webhook struct {
	spec *structs.WorkflowNodeSpec
}

type webhookOutput struct {
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
	Query   map[string]string `json:"query"`
}

func init() {
	wh := &Webhook{
		spec: &structs.WorkflowNodeSpec{},
	}
	wh.spec.JsonConfig = rawJson
	wh.spec.GenerateSpec()
	core.Register(wh)
	core.RegisterEmbedIcons(Name, icon)
}

func (wh *Webhook) Category() structs.NodeObjectCategory {
	return Category
}

func (wh *Webhook) Name() string {
	return Name
}

func (wh *Webhook) DefaultSpec() interface{} {
	return wh.spec
}

func (wh *Webhook) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	request := input.AdditionalData.HttpRequest
	returnItem := wh.generateOutput(request)
	return core.GenerateSuccessResponse(structs.NodeData{
		{
			"json": returnItem,
		},
	}, []structs.NodeData{})
}

func (wh *Webhook) generateOutput(request *fasthttp.Request) webhookOutput {
	headers := make(map[string]string)
	request.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})

	params := make(map[string]string)
	request.URI().QueryArgs().VisitAll(func(key, value []byte) {
		params[string(key)] = string(value)
	})

	return webhookOutput{
		Body:    string(request.Body()),
		Headers: headers,
		Query:   params,
	}
}
