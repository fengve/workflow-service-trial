package core

import (
	_ "embed"
	"fmt"
	"strings"

	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

// TODO: use this to control the global keyword
// var AllowedGlobalKeywordMap = map[string]bool{}

const CodeLanguage = "javaScript"

var (
	BuiltInFunctions = map[string]interface{}{
		"$if": func(condition bool, trueValue, falseValue interface{}) interface{} {
			if condition {
				return trueValue
			}
			return falseValue
		},

		"$max": func(a, b int) int {
			if a > b {
				return a
			}
			return b
		},

		"$min": func(a, b int) int {
			if a < b {
				return a
			}
			return b
		},
	}
)

// SandboxContext is the context that will be passed to the Sandbox
// Note it can not be shared between Sandboxes as we don't export anything from the VM for now
type SandboxContext struct {
	Items     structs.NodeData
	ItemIndex int
	Params    map[string]interface{}
	Variables map[string]interface{}
	Functions map[string]interface{}
	RunData   map[string][]*structs.WorkflowExecutionTaskData
}

func (sc *SandboxContext) SetupCtxForRunCode(s *Sandbox) {
	var item structs.NodeSingleData = nil
	// Default variables
	if len(sc.Items) > 0 {
		item = sc.Items[sc.ItemIndex]
	}
	inputObj := map[string]interface{}{
		"all": func() structs.NodeData {
			return sc.Items
		},
		"item": item,
		"first": func() structs.NodeSingleData {
			if len(sc.Items) > 0 {
				return sc.Items[0]
			}
			return nil
		},
		"last": func() structs.NodeSingleData {
			if len(sc.Items) > 0 {
				return sc.Items[len(sc.Items)-1]
			}
			return nil
		},
		"params": sc.Params,
	}
	s.VM.Set("$input", inputObj)
	s.VM.Set("$item", item)
	if item != nil {
		s.VM.Set("$json", item["json"])
	} else {
		s.VM.Set("$json", nil)
	}

	// setup Functions
	for k, v := range sc.Functions {
		s.VM.Set(k, v)
	}

	// setup Variables
	for k, v := range sc.Variables {
		s.VM.Set(k, v)
	}

	nodesResults := make(map[string]interface{})
	nodeOutputItems := sc.getNodeOutputItems()
	for nodeName := range nodeOutputItems {
		items := nodeOutputItems[nodeName]
		var itemData structs.NodeSingleData
		if len(items) > sc.ItemIndex {
			itemData = items[sc.ItemIndex]
		}
		nodesResults[nodeName] = map[string]interface{}{
			"item": itemData,
			"all": func() structs.NodeData {
				return items
			},
			"first": func() structs.NodeSingleData {
				if len(items) > 0 {
					return items[0]
				}
				return nil
			},
			"last": func() structs.NodeSingleData {
				if len(items) > 0 {
					return items[len(items)-1]
				}
				return nil
			},
		}
	}
	s.VM.Set("$", func(name string) interface{} {
		return nodesResults[name]
	})
}

func (sc *SandboxContext) SetupCtxForRunCodeAllItems(s *Sandbox) {
	// Default variables
	inputObj := map[string]interface{}{
		"all": func() structs.NodeData {
			return sc.Items
		},
		"items": sc.Items,
		"first": func() structs.NodeSingleData {
			if len(sc.Items) > 0 {
				return sc.Items[0]
			}
			return nil
		},
		"last": func() structs.NodeSingleData {
			if len(sc.Items) > 0 {
				return sc.Items[len(sc.Items)-1]
			}
			return nil
		},
		"params": sc.Params,
	}
	s.VM.Set("$input", inputObj)
	s.VM.Set("$items", sc.Items)

	// setup Functions
	for k, v := range sc.Functions {
		s.VM.Set(k, v)
	}
	// setup Variables
	for k, v := range sc.Variables {
		s.VM.Set(k, v)
	}

	nodeOutputItems := sc.getNodeOutputItems()
	nodesResults := make(map[string]interface{})
	for nodeName := range sc.getNodeOutputItems() {
		items := nodeOutputItems[nodeName]
		nodesResults[nodeName] = map[string]interface{}{
			"all": func() structs.NodeData {
				return items
			},
			"items": items,
			"first": func() structs.NodeSingleData {
				if len(items) > 0 {
					return items[0]
				}
				return nil
			},
			"last": func() structs.NodeSingleData {
				if len(items) > 0 {
					return items[len(items)-1]
				}
				return nil
			},
		}
	}
	s.VM.Set("$", func(name string) interface{} {
		return nodesResults[name]
	})
}

func (sc *SandboxContext) getNodeOutputItems() map[string]structs.NodeData {
	nodesResults := make(map[string]structs.NodeData)
	for nodeName := range sc.RunData {
		taskDataList := sc.RunData[nodeName]
		if len(taskDataList) > 0 {
			taskData := taskDataList[len(taskDataList)-1]
			if taskData != nil && taskData.Data != nil {
				outputItems := taskData.Data["main"]
				if len(outputItems) > 0 && outputItems[0] != nil {
					items := outputItems[0]
					nodesResults[nodeName] = items
				}
			}
		}
	}
	return nodesResults
}

// The default timeout for code execution in sandbox is 180 seconds (3 minutes).
const TimeoutDefault = 180 * 1000 // ms
type Sandbox struct {
	Name    string
	Lang    string
	JsCode  string
	Context *SandboxContext
	VM      *goja.Runtime
	Timeout time.Duration
}

func newGoja() (*goja.Runtime, *require.RequireModule) {
	registry := new(require.Registry)
	vm := goja.New()
	req := registry.Enable(vm)
	console.Enable(vm)

	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	// https://github.com/dop251/goja#mapping-struct-field-and-method-names
	// use this if we need optionally uncapitalises
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	return vm, req

}

func (s *Sandbox) Initialize() {
	s.VM, _ = newGoja()
	if s.Timeout <= 0 {
		s.Timeout = TimeoutDefault * time.Millisecond
	}
	if s.Name == "" {
		s.Name = "Main"
	}

	if len(s.Lang) > 0 && s.Lang != CodeLanguage {
		panic("Only support js language")
	}
}

const RunCodeWrapperFmt = `(()=>{return %s
	})()`

// RunCode run code with itemIndex
// number result type -> int64 or float64
// no return statement needed
func (s *Sandbox) RunCode(code string, itemIndex int) (interface{}, error) {
	code = strings.Trim(code, "\n\t ")
	// see test `Evaluate expression first line //` in expression_test.go
	if !(strings.HasPrefix(code, "//") || strings.HasPrefix(code, "/*")) {
		code = fmt.Sprintf(RunCodeWrapperFmt, code)
	}

	if s.Context != nil {
		s.Context.ItemIndex = itemIndex
		s.Context.SetupCtxForRunCode(s)
	}

	// -------- timeout ---------
	s.SetupTimeout()

	// --------- run script -------
	v, err := s.VM.RunScript(s.Name, code)
	if err != nil {
		err = HandleJavaScriptError(err)
		return nil, err

	}
	if v == nil {
		return v, nil
	}
	returnData := v.Export()

	if returnData, ok := returnData.(map[string]interface{}); ok {
		return StandardizeJavaScriptObject(returnData), nil

	}

	if returnData, ok := returnData.(structs.NodeSingleData); ok {
		return StandardizeJavaScriptObject(returnData), nil

	}

	return returnData, nil
}

func (s *Sandbox) SetupTimeout() *time.Timer {
	// -------- timeout ---------
	if s.Timeout > 0 {
		s.VM.ClearInterrupt()
		// set timeout
		return time.AfterFunc(s.Timeout, func() {
			s.VM.Interrupt("Code run timeout")
		})
	}
	return nil
}

const ScriptWrapperFmt = `(()=>{%s
	})()`

// return statement needed
func (s *Sandbox) RunCodeAllItems() (structs.NodeData, error) {
	if s.Context != nil {
		s.Context.SetupCtxForRunCodeAllItems(s)
	}

	script := fmt.Sprintf(ScriptWrapperFmt, s.JsCode)

	// -------- timeout ---------
	timer := s.SetupTimeout()

	// --------- run script -------
	v, err := s.VM.RunScript(s.Name, script)
	// Cancel the timer if the script runs successfully
	timer.Stop()

	if err != nil {
		err = HandleJavaScriptError(err)
		return nil, err
	}

	// --------- run script -------

	if v == nil {
		return nil, fmt.Errorf("Code doesn't return items properly")
	}
	returnData, err := s.ValidateRunCodeAllItems(v.Export())

	if err != nil {
		return nil, err
	}
	return StandardizeOutput(returnData), nil

}

func (s *Sandbox) RunCodeEachItem() {
	// TODO
}

func (s *Sandbox) ValidateRunCodeAllItems(res interface{}) (structs.NodeData, error) {
	if res == nil {
		return nil, fmt.Errorf("Code doesn't return items properly")
	}
	if res, ok := res.(structs.NodeData); ok {
		return res, nil
	}

	if res, ok := res.(structs.NodeSingleData); ok {
		return structs.NodeData{res}, nil
	}

	if res, ok := res.([]interface{}); ok {

		nodeData := make(structs.NodeData, len(res))
		for i, v := range res {
			nodeData[i], ok = v.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("Code return invalid item [index %d] type %T", i, v)
			}
		}
		return nodeData, nil

	}

	if _, ok := res.(map[string]interface{}); ok {
		// note here is different from the n8n original code
		return structs.NodeData{res.(map[string]interface{})}, nil

	}
	return nil, fmt.Errorf("Code return invalid type %T", res)

}

// there's no need to validateItem in ValidateRunCodeAllItems anymore
// since we will NormalizeItems after RunCodeAllItems in Code.Execute anyway.
// maybe this can be used for RunCodeEachItem
func (s *Sandbox) ValidateItem(item map[string]interface{}) {
	// TODO
}
