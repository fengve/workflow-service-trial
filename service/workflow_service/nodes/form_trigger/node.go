package form_trigger

import (
	"bytes"
	"context"
	_ "embed"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
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

	node := input.Params
	if bytes.Equal(request.Header.Method(), []byte("GET")) {
		return core.GenerateSuccessResponse(structs.NodeData{
			{
				"parameters": node.Parameters,
			},
		}, []structs.NodeData{})
	}

	// POST
	return core.GenerateSuccessResponse(structs.NodeData{
		{
			"parameters": node.Parameters,
		},
	}, []structs.NodeData{})
}
