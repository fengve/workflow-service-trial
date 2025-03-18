package core_test

// Command to run this test file only.
// go test -v service/workflow_service/core/init_test.go service/workflow_service/core/sandbox_test.go

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

func TestSandbox(t *testing.T) {

	t.Run("Sandbox create and init", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `console.log('Hello, World!')`
		sandbox := core.Sandbox{
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		assert.Equal(sandbox.Timeout, 180*time.Second)

		res, err := sandbox.VM.RunString(code)
		assert.Nil(res.Export())
		assert.Nil(err)

	})

	t.Run("Sandbox run code empty", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		sandbox := core.Sandbox{}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		// number
		res, err := sandbox.RunCode(``, 1)

		assert.Nil(err)
		assert.Equal(res, nil)
	})

	t.Run("Sandbox run code with comments", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		sandbox := core.Sandbox{}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		// number
		res, err := sandbox.RunCode(`// 123
		123`, 1)

		assert.Nil(err)
		assert.Equal(int64(123), res)
	})

	t.Run("Sandbox run code with no context", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		sandbox := core.Sandbox{}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		// number
		res, err := sandbox.RunCode(`123`, 1)

		assert.Nil(err)
		assert.Equal(res, int64(123))

		res, err = sandbox.RunCode(`3.14`, 1)

		assert.Nil(err)
		assert.Equal(res, float64(3.14))

		// string
		res, err = sandbox.RunCode(`"Hello, World!"`, 1)

		assert.Nil(err)
		assert.Equal(res, "Hello, World!")

		// boolean
		res, err = sandbox.RunCode(`1==1`, 1)

		assert.Nil(err)
		assert.Equal(res, true)

		// array
		res, err = sandbox.RunCode(`[1,2,3]`, 1)

		assert.Nil(err)
		assert.Equal(res, []interface{}{int64(1), int64(2), int64(3)})

	})

	t.Run("Sandbox run code obj", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		sandbox := core.Sandbox{}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		// object
		res, err := sandbox.RunCode(`{"a":1,"b":2}`, 1)

		assert.Nil(err)
		assert.Equal(res, map[string]interface{}{"a": int64(1), "b": int64(2)})
	})

	t.Run("Sandbox run code array", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		sandbox := core.Sandbox{}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		// array
		res, err := sandbox.RunCode(`[1,{},"3"]`, 1)

		assert.Nil(err)
		assert.Equal(res, []any{int64(1), map[string]interface{}{}, "3"})
	})

	t.Run("Sandbox run code with context", func(t *testing.T) {
		assert := require.New(t)
		sandbox := core.Sandbox{
			Context: &core.SandboxContext{
				Items: structs.NodeData{
					{"json": map[string]interface{}{"a": 1}},
					{"json": map[string]interface{}{"a": 2}},
				},
			},
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		// test run input
		res, err := sandbox.RunCode(`$input.item`, 0)

		assert.Nil(err)
		assert.Equal(res, map[string]interface{}{"json": map[string]interface{}{"a": 1}})

		// test run input
		res, err = sandbox.RunCode(`$input.all()`, 0)

		assert.Nil(err)
		assert.Equal(len(res.(structs.NodeData)), 2)

		// test run with index 0
		res, err = sandbox.RunCode(`$json.a`, 0)

		assert.Nil(err)
		assert.Equal(res, int64(1))

		// test run with index 1
		res, err = sandbox.RunCode(`$json.a`, 1)

		assert.Nil(err)
		assert.Equal(res, int64(2))

	})

	t.Run("Sandbox run code with context and builtin funcs", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		sandbox := core.Sandbox{
			Context: &core.SandboxContext{
				Items: structs.NodeData{
					{"json": map[string]interface{}{"a": 1}},
					{"json": map[string]interface{}{"a": 2}},
				},

				Functions: core.BuiltInFunctions,
			},
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		// test if
		res, err := sandbox.RunCode(`$if(1==2, "true value","false value")`, 0)

		assert.Nil(err)
		assert.Equal(res, "false value")

		// test max
		res, err = sandbox.RunCode(`$max(1, 2)`, 0)

		assert.Nil(err)
		assert.Equal(res, int64(2))

		// test min
		res, err = sandbox.RunCode(`$min(1, 2)`, 0)

		assert.Nil(err)
		assert.Equal(res, int64(1))

		// test regular code
		res, err = sandbox.RunCode(`$if($json.a + 1 == $input.last().json.a, "true value","false value")`, 0)

		assert.Nil(err)
		assert.Equal(res, "true value")

	})

	t.Run("Evaluate expression if method", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)

		code := `$if(false, $json.foo?.filter(x=>!!x), "foo")`
		sandbox := core.Sandbox{
			Context: &core.SandboxContext{
				Items: structs.NodeData{
					{"json": map[string]interface{}{"a": 1}},
					{"json": map[string]interface{}{"a": 2}},
				},

				Functions: core.BuiltInFunctions,
			},
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		// test if
		res, err := sandbox.RunCode(code, 0)

		assert.Nil(err)
		assert.Equal(`foo`, res)
	})

	t.Run("Sandbox run code with comments", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `// this is a comment
			return []	// return empty items
			/* end */
		`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.Nil(err)
		assert.NotNil(res)
	})

	t.Run("Sandbox run code with throw error 1", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `throw 'test error'`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "test error [line 1]")
		assert.Nil(res)
	})

	t.Run("Sandbox run code with throw error 2", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `throw new Error('test error')`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "Error: test error [line 1]")
		assert.Nil(res)
	})

	t.Run("Sandbox run code with ref error", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `console.log(item)	// item is not defined`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "ReferenceError: item is not defined [line 1]")
		assert.Nil(res)
	})

	t.Run("Sandbox run code with timeout", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `while(true){}`
		sandbox := core.Sandbox{
			Name:    "main",
			JsCode:  code,
			Timeout: 2 * time.Millisecond,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.Equal(sandbox.Timeout, 2*time.Millisecond)

		assert.Nil(res)
		assert.EqualError(err, "Code run timeout [line 1]")
	})

	t.Run("Sandbox run code with return {}", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `return {a: 1}`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.Nil(err)
		assert.NotNil(res)

		assert.Equal(res[0]["a"], int64(1))
	})

	t.Run("Sandbox run code with return nested {}", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `return { json:{a: 1}, error: null}`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.Nil(err)
		assert.NotNil(res)

		assert.Equal(res[0]["a"], nil)
		assert.Equal(res[0]["json"].(map[string]interface{})["a"], int64(1))
	})

	t.Run("Sandbox run code with return []", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `return [{a: 1}]`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.Nil(err)
		assert.NotNil(res)

		assert.Equal(res[0]["a"], int64(1))
	})

	t.Run("Sandbox run code with return [{},{},int]", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `return [{a: 1}, {a:2}, 3]`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		_, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "Code return invalid item [index 2] type int64")
	})

	t.Run("Sandbox run code with return func", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `return function(){}`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		_, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "Code return invalid type func(goja.FunctionCall) goja.Value")
	})

	t.Run("Sandbox run code with return obj has func inside", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `return {foo:"baz", bar: function(){}}`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.Nil(err)

		assert.Nil(res[0]["bar"])
		assert.Equal(res[0]["foo"], "baz")
	})

	t.Run("Sandbox run code with return int", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `return 1`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		_, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "Code return invalid type int64")
	})

	t.Run("Sandbox run code with no return", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `console.log('Hello, World!')`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "Code doesn't return items properly")

		assert.Nil(res)

	})

	t.Run("Sandbox built in variables", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `
			console.log($input)

			console.log($input.first())

			console.log($input.last())

			console.log($input.items)

			console.log($input.params)

			console.log($items)

			console.log($input.all())

			return [...$items, {json:{a:3}}]
		`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
			Context: &core.SandboxContext{
				Items: structs.NodeData{
					{"json": map[string]interface{}{"a": 1}},
					{"json": map[string]interface{}{"a": 2}},
				},
				Params: map[string]interface{}{
					"notice": true,
				},
			},
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		res, err := sandbox.RunCodeAllItems()

		assert.Nil(err)

		assert.Equal(res[0]["json"].(map[string]interface{})["a"], int(1))
		assert.Equal(res[1]["json"].(map[string]interface{})["a"], int(2))
		assert.Equal(res[2]["json"].(map[string]interface{})["a"], int64(3))

	})

	t.Run("Sandbox built in functions", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `
			var a = $max(1, 2)
			var b = $min(1, 2)
			var c = $if(1==2, "true value","false value")
			return {a,b,c}
 		`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
			Context: &core.SandboxContext{
				Functions: core.BuiltInFunctions,
			},
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		res, err := sandbox.RunCodeAllItems()

		assert.Nil(err)

		assert.Equal(res[0]["a"], int64(2))
		assert.Equal(res[0]["b"], int64(1))
		assert.Equal(res[0]["c"], "false value")

	})

	t.Run("Sandbox myNewField", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `// Loop over input items and add a new field called 'myNewField' to the JSON of each one
		for (const item of $input.all()) {
			  item.json.myNewField = 1;
		}
		return $input.all();`

		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
			Context: &core.SandboxContext{
				Items: structs.NodeData{
					{"json": map[string]interface{}{"a": 1}},
				},
				Functions: core.BuiltInFunctions,
			},
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		res, err := sandbox.RunCodeAllItems()

		assert.Nil(err)

		assert.Equal(res[0]["json"].(map[string]interface{})["a"], int(1))
		assert.Equal(res[0]["json"].(map[string]interface{})["myNewField"], int64(1))
	})
}

func TestSandboxContext(t *testing.T) {

	t.Run("Sanbox context setup with no items", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		sandbox := core.Sandbox{
			Context: &core.SandboxContext{},
		}

		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		sandbox.Context.SetupCtxForRunCode(&sandbox)

		res, err := sandbox.VM.RunString("$item")
		assert.Nil(err)
		assert.Nil(res.Export())
	})

	t.Run("Sandbox context setup for RunCode", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		sandbox := core.Sandbox{
			Context: &core.SandboxContext{
				Items: structs.NodeData{
					{"json": map[string]interface{}{"a": 1}},
					{"json": map[string]interface{}{"a": 2}},
				},
				ItemIndex: 1,
				Params: map[string]interface{}{
					"notice": true,
				},
				Functions: map[string]interface{}{
					"$add": func(a, b int) int { return a + b },
				},
				Variables: map[string]interface{}{
					"var1": 1,
				},
			},
		}

		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		sandbox.Context.SetupCtxForRunCode(&sandbox)

		// inputs
		res, err := sandbox.VM.RunString("$input.item.json.a")
		assert.Nil(err)
		assert.Equal(res.Export(), int64(2))

		res, err = sandbox.VM.RunString("$json.a")
		assert.Nil(err)
		assert.Equal(res.Export(), int64(2))

		// params
		res, err = sandbox.VM.RunString("$input.params.notice")
		assert.Nil(err)
		assert.Equal(res.Export(), true)

		// function
		res, err = sandbox.VM.RunString("$add(1,2)")
		assert.Nil(err)
		assert.Equal(res.Export(), int64(3))

		assert.Nil(sandbox.VM.RunString("$notExist(1,2)"))

		// varibales
		res, err = sandbox.VM.RunString("var1*5")
		assert.Nil(err)
		assert.Equal(res.Export(), int64(5))

		assert.Nil(sandbox.VM.RunString("notExist"))

	})

	t.Run("Sandbox context setup for RunCodeAllItems", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		sandbox := core.Sandbox{
			Context: &core.SandboxContext{
				Items: structs.NodeData{
					{"json": map[string]interface{}{"a": 1}},
					{"json": map[string]interface{}{"a": 2}},
				},
				Params: map[string]interface{}{
					"notice": true,
				},
				Functions: map[string]interface{}{
					"$add": func(a, b int) int { return a + b },
				},
				Variables: map[string]interface{}{
					"var1": 1,
				},
			},
		}

		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		sandbox.Context.SetupCtxForRunCodeAllItems(&sandbox)

		// inputs
		res, err := sandbox.VM.RunString("$input.last().json.a")
		assert.Nil(err)
		assert.Equal(res.Export(), int64(2))

		res, err = sandbox.VM.RunString("$items[1].json.a")
		assert.Nil(err)
		assert.Equal(res.Export(), int64(2))

		// params
		res, err = sandbox.VM.RunString("$input.params.notice")
		assert.Nil(err)
		assert.Equal(res.Export(), true)

		// function
		res, err = sandbox.VM.RunString("$add(1,2)")
		assert.Nil(err)
		assert.Equal(res.Export(), int64(3))

		assert.Nil(sandbox.VM.RunString("$notExist(1,2)"))

		// varibales
		res, err = sandbox.VM.RunString("var1*5")
		assert.Nil(err)
		assert.Equal(res.Export(), int64(5))

		assert.Nil(sandbox.VM.RunString("notExist"))
	})

	t.Run("Sandbox context setup for RunCodeAllItems with builtin funcs", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		sandbox := core.Sandbox{
			Context: &core.SandboxContext{
				Items: structs.NodeData{
					{"json": map[string]interface{}{"a": 1}},
					{"json": map[string]interface{}{"a": 2}},
				},
				Params: map[string]interface{}{
					"notice": true,
				},
				Functions: core.BuiltInFunctions,
			},
		}

		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		sandbox.Context.SetupCtxForRunCodeAllItems(&sandbox)

		assert.NotNil(sandbox.VM.RunString("$input"))
		assert.NotNil(sandbox.VM.RunString("$items"))
		assert.Nil(sandbox.VM.RunString("$json"))
		assert.Nil(sandbox.VM.RunString("$item"))
		assert.NotNil(sandbox.VM.RunString("$input.all"))
		assert.NotNil(sandbox.VM.RunString("$input.first"))
		assert.NotNil(sandbox.VM.RunString("$input.last"))
		assert.NotNil(sandbox.VM.RunString("$input.params"))
		// builtin methods
		assert.NotNil(sandbox.VM.RunString("$max"))
		assert.NotNil(sandbox.VM.RunString("$min"))
		assert.NotNil(sandbox.VM.RunString("$if"))
		assert.Nil(sandbox.VM.RunString("$notExist"))

	})
	t.Run("Sandbox context setup for RunCode with builtin funcs", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		sandbox := core.Sandbox{
			Context: &core.SandboxContext{
				Items: structs.NodeData{
					{"json": map[string]interface{}{"a": 1}},
					{"json": map[string]interface{}{"a": 2}},
				},
				ItemIndex: 0,
				Params: map[string]interface{}{
					"notice": true,
				},
				Functions: core.BuiltInFunctions,
			},
		}

		sandbox.Initialize()

		assert.NotNil(sandbox.VM)

		sandbox.Context.SetupCtxForRunCode(&sandbox)

		assert.NotNil(sandbox.VM.RunString("$input"))
		assert.NotNil(sandbox.VM.RunString("$json"))
		assert.NotNil(sandbox.VM.RunString("$item"))
		assert.Nil(sandbox.VM.RunString("$items"))
		assert.NotNil(sandbox.VM.RunString("$input.all"))
		assert.NotNil(sandbox.VM.RunString("$input.first"))
		assert.NotNil(sandbox.VM.RunString("$input.last"))
		assert.NotNil(sandbox.VM.RunString("$input.params"))
		// builtin methods
		assert.NotNil(sandbox.VM.RunString("$max"))
		assert.NotNil(sandbox.VM.RunString("$min"))
		assert.NotNil(sandbox.VM.RunString("$if"))
		assert.Nil(sandbox.VM.RunString("$notExist"))

	})

}

func TestSandboxIsolation(t *testing.T) {

	t.Run("Sandbox fetch", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `fetch('https://google.com')`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "ReferenceError: fetch is not defined [line 1]")
		assert.Nil(res)

	})

	t.Run(("Sandbox ajax"), func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `new XMLHttpRequest()`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "ReferenceError: XMLHttpRequest is not defined [line 1]")
		assert.Nil(res)

	})

	t.Run("Sandbox http module", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `require('http')`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "GoError: Invalid module [line 1]")
		assert.Nil(res)

	})

	t.Run("Sandbox process", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `process.env`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "ReferenceError: process is not defined [line 1]")
		assert.Nil(res)

	})

	t.Run("Sandbox net module", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `require('net')`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "GoError: Invalid module [line 1]")
		assert.Nil(res)

	})

	t.Run("Sandbox fs module", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)
		const code = `require('fs')`
		sandbox := core.Sandbox{
			Name:   "main",
			JsCode: code,
		}
		sandbox.Initialize()

		assert.NotNil(sandbox.VM)
		res, err := sandbox.RunCodeAllItems()

		assert.EqualError(err, "GoError: Invalid module [line 1]")
		assert.Nil(res)
	})
}
