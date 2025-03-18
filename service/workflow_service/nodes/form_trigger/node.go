package form_trigger

import (
	"context"
	_ "embed"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	"github.com/valyala/fasthttp"
)

const (
	// Category is the category of formTrigger.
	Category = structs.CategoryTrigger

	// Name is the name of formTrigger.
	Name = "n8n-nodes-base.formTrigger"
)

var (
	//go:embed node.json
	rawJson []byte

	//go:embed form.svg
	icon []byte
)

type FormTrigger struct {
	spec *structs.WorkflowNodeSpec
}

type formTriggerOutput struct {
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
	Query   map[string]string `json:"query"`
}

func init() {

	trigger := &FormTrigger{
		spec: &structs.WorkflowNodeSpec{},
	}
	trigger.spec.JsonConfig = rawJson
	trigger.spec.GenerateSpec()

	core.Register(trigger)
	core.RegisterEmbedIcons(Name, icon)
}

func (ft *FormTrigger) Category() structs.NodeObjectCategory {
	return Category
}

func (ft *FormTrigger) Name() string {
	return Name
}

func (ft *FormTrigger) DefaultSpec() interface{} {
	return ft.spec
}

func (ft *FormTrigger) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	request := input.AdditionalData.HttpRequest
	returnItem := ft.generateOutput(request)
	return core.GenerateSuccessResponse(structs.NodeData{
		{
			"json": returnItem,
		},
	}, []structs.NodeData{})
}

func (ft *FormTrigger) generateOutput(request *fasthttp.Request) formTriggerOutput {
	headers := make(map[string]string)
	request.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})

	params := make(map[string]string)
	request.URI().QueryArgs().VisitAll(func(key, value []byte) {
		params[string(key)] = string(value)
	})

	return formTriggerOutput{}
}
