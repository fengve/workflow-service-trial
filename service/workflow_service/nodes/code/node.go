package code

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (
	// Category is the category of CodeNode.
	Category = structs.CategoryExecutor

	// Name is the name of CodeNode.
	Name = "n8n-nodes-base.code"

	DefaultJsCode = ``
)

const (
	CodeNodeModeParameter = "mode"
	CodeParameterNameJs   = "jsCode"
	CodeLanguageJs        = "javaScript"
)

var (
	//go:embed node.json
	rawJson []byte

	//go:embed code.svg
	rawIcon []byte
)

type CodeExecutor struct {
	spec *structs.WorkflowNodeSpec
}

func init() {
	ce := &CodeExecutor{
		spec: &structs.WorkflowNodeSpec{},
	}
	ce.spec.JsonConfig = rawJson
	ce.spec.GenerateSpec()

	core.Register(ce)
	core.RegisterEmbedIcons(ce.spec.Name(), rawIcon)
}

func (ce *CodeExecutor) Category() structs.NodeObjectCategory {
	return Category
}

func (ce *CodeExecutor) Name() string {
	return Name
}

func (ce *CodeExecutor) DefaultSpec() interface{} {
	return ce.spec
}

func getSandbox(
	code string,
	items structs.NodeData,
	input *structs.NodeExecuteInput,
	itemIndex int,
) *core.Sandbox {

	runData := make(map[string][]*structs.WorkflowExecutionTaskData)
	if input.RunExecutionData != nil && input.RunExecutionData.ResultData != nil && input.RunExecutionData.ResultData.RunData != nil {
		runData = input.RunExecutionData.ResultData.RunData
	}

	context := core.SandboxContext{
		Items:     items,
		Params:    input.Params.Parameters,
		ItemIndex: itemIndex,
		RunData:   runData,
	}

	sandbox := core.Sandbox{
		Lang:    CodeLanguageJs,
		JsCode:  code,
		Context: &context,
	}

	sandbox.Initialize() // panic

	return &sandbox
}

func (ce *CodeExecutor) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	items := core.GetInputData(input.Data)

	// The Execute Once Option
	// If active, the node executes only once, with data from the first item it receives
	executeOnce := input.Params.ExecutionOnce
	if executeOnce {
		items = items[0:1]
	}

	nodeMode, err := core.GetNodeParameter(Name, "mode", "runOnceForAllItems", input, 0)

	if err != nil {
		nodeMode = "runOnceForAllItems"
	}

	code, err := core.GetNodeParameter(Name, CodeParameterNameJs, "", input, 0)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}

	codeStr, ok := code.(string)
	if !ok {
		return core.GenerateFailedResponse(Name, fmt.Errorf("code is not a string"))
	}

	sandbox := getSandbox(codeStr, items, input, 0)

	// ------- runOnceForAllItems ----------
	if nodeMode == "runOnceForAllItems" {

		returnData, err := sandbox.RunCodeAllItems()

		if err != nil {
			if core.ContinueOnFail(input.Params) {
				res := core.NewNodeSingleDataError(err, 0)
				returnData = structs.NodeData{res}
				return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{returnData})

			}
			return core.GenerateFailedResponse(Name, err)

		}

		returnData = core.NormalizeItems(returnData)

		return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{returnData})

	}

	// ------- runOnceForEachItem ----------
	if nodeMode == "runOnceForEachItem" {
		// TODO: runOnceForEachItem
		return core.GenerateFailedResponse(Name, fmt.Errorf(`"runOnceForEachItem" mode is not implement`))
	}

	return core.GenerateFailedResponse(Name, fmt.Errorf("Unknown error"))

}
