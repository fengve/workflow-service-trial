package core_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

func TestEvaluateExpression(t *testing.T) {
	assert := require.New(t)

	t.Run("Evaluate expression fixed", func(t *testing.T) {
		expression := "123"
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal("123", res)
	})

	t.Run("Evaluate expression fixed 2", func(t *testing.T) {
		expression := "={123}"
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal("{123}", res)
	})

	t.Run("Evaluate expression 1+2", func(t *testing.T) {
		expression := "={{1 + 2}}"
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal(int64(3), res)
	})

	t.Run("Evaluate expression 1+2", func(t *testing.T) {
		expression := "={{1 + 2}}"
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal(int64(3), res)
	})

	t.Run("Evaluate expression 123", func(t *testing.T) {
		expression := "=1{{1 + 1}}3"
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal("123", res)
	})

	t.Run("Evaluate expression 123123", func(t *testing.T) {
		expression := `=1{{1 + 1}}3{{12+"3"}}`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal("123123", res)
	})

	t.Run("Evaluate expression array", func(t *testing.T) {
		expression := `={{[1,"2"]}}`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal([]interface{}{int64(1), "2"}, res)
	})

	t.Run("Evaluate expression obj", func(t *testing.T) {
		expression := `={{ {a:1,b:"2"} }}`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal(map[string]interface{}{"a": int64(1), "b": "2"}, res)
	})

	t.Run("Evaluate expression obj 2", func(t *testing.T) {
		expression := `={{{a:1,b:"2"} }}`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal(map[string]interface{}{"a": int64(1), "b": "2"}, res)
	})

	t.Run("Evaluate expression nil", func(t *testing.T) {
		expression := `={{{a:1,b:"2"}.c }}`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal(nil, res)
	})

	t.Run("Evaluate expression //", func(t *testing.T) {
		expression := `={{
			123 // 456
			}}`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal(int64(123), res)
	})

	t.Run("Evaluate expression first line //", func(t *testing.T) {
		expression := `={{ // 456
			123
			}}`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal(int64(123), res)
	})

	t.Run("Evaluate expression with comments", func(t *testing.T) {
		expression := `={{
			// 456
			// 789
			/* abc */
			123	// no return
			}}`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal(int64(123), res)
	})

	t.Run("Evaluate expression function", func(t *testing.T) {
		expression := `={{ ()=>1 }}`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.IsType(func(goja.FunctionCall) goja.Value { return nil }, res)
	})

	t.Run("Evaluate expression function", func(t *testing.T) {
		expression := `={{ ()=>1 }}`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.IsType(func(goja.FunctionCall) goja.Value { return nil }, res)
	})

	t.Run("Evaluate expression Date", func(t *testing.T) {
		expression := `={{ new Date() }}`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.IsType(time.Time{}, res)
	})

	t.Run("Evaluate expression Date str", func(t *testing.T) {
		expression := `={{new Date('2000-01-01T00:00:00Z')}} 123`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}
		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal("2000-01-01T00:00:00Z 123", res)
	})

	t.Run("Evaluate expression stringify", func(t *testing.T) {
		expression := `={{ 1 }} {{"2"}} {{ {a:3} }} {{false}} {{[5,6]}}
 {{new Date('2000-01-01T00:00:00Z')}}`
		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}

		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal("1 2 {\"a\":3} false [5,6]\n 2000-01-01T00:00:00Z", res)

	})

	t.Run("Evaluate expression itemIndex", func(t *testing.T) {
		exp1 := `={{ !$json?.data?.pageInfo?.hasNextPage }}`
		exp2 := `={{ !$item?.json?.data?.pageInfo?.hasNextPage }}`
		jsonStr1 := `{"data":{"pageInfo":{"hasNextPage":true}}}`
		jsonStr2 := `{"data":{"pageInfo":{}}}`
		var obj1, obj2 map[string]interface{}
		json.Unmarshal([]byte(jsonStr1), &obj1)
		json.Unmarshal([]byte(jsonStr2), &obj2)

		sb := core.Sandbox{
			Context: &core.SandboxContext{
				Items: []map[string]interface{}{
					{"json": obj1},
					{"json": obj2},
				},
			},
		}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}

		res, err := eval.EvaluateExpression(exp1, 0)
		assert.Nil(err)
		assert.Equal(false, res)

		res, err = eval.EvaluateExpression(exp2, 1)
		assert.Nil(err)
		assert.Equal(true, res)

	})

	t.Run("Evaluate expression $input", func(t *testing.T) {
		exp1 := `={{ $json }}`
		sb := core.Sandbox{
			Context: &core.SandboxContext{
				Items: []map[string]interface{}{
					{"json": map[string]interface{}{"a": 1}},
					{"json": map[string]interface{}{"b": 2}},
				},
			},
		}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}

		res, err := eval.EvaluateExpression(exp1, 0)
		assert.Nil(err)
		assert.Equal((map[string]interface{}{"a": 1}), res)

		exp2 := `={{ $item.json }} 123`
		res, err = eval.EvaluateExpression(exp2, 0)
		assert.Nil(err)
		assert.Equal(("{\"a\":1} 123"), res)
	})

	t.Run("Evaluate expression builtin funcs", func(t *testing.T) {
		expression := `={{ [1,2,3].map(x=> $max(x,6)) }} {{ [1,2,3].map(x=> $if(x%2==0, "even", "odd")) }}`
		sb := core.Sandbox{
			Context: &core.SandboxContext{
				Functions: core.BuiltInFunctions,
			},
		}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}

		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal(`[6,6,6] ["odd","even","odd"]`, res)

	})

	t.Run("Evaluate expression builtin funcs", func(t *testing.T) {
		expression := `={{ [1,2,3].map(x=> $max(x,6)) }} {{ [1,2,3].map(x=> $if(x%2==0, "even", "odd")) }}`
		sb := core.Sandbox{
			Context: &core.SandboxContext{
				Functions: core.BuiltInFunctions,
			},
		}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}

		res, err := eval.EvaluateExpression(expression, 0)
		assert.Nil(err)
		assert.Equal(`[6,6,6] ["odd","even","odd"]`, res)

	})

	t.Run("html", func(t *testing.T) {
		assert := require.New(t)

		sb := core.Sandbox{
			Context: &core.SandboxContext{
				Items: []map[string]interface{}{
					{"json": map[string]interface{}{"id": 1, "style": "color: red"}},
					{"json": map[string]interface{}{"id": 2, "style": "color: blue"}},
				},
				Functions: core.BuiltInFunctions,
			},
		}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}

		// read from html_template.hbs
		htmlBytes, err := os.ReadFile("test_files/html_template.hbs")
		assert.Nil(err)
		html := string(htmlBytes)

		htmlExpectBytes, err := os.ReadFile("test_files/html_template_expect.html")
		assert.Nil(err)
		htmlExpect := string(htmlExpectBytes)

		val, err := eval.EvaluateExpression("="+html, 0)
		assert.Nil(err)
		assert.Equal(htmlExpect, val)

		val, err = eval.EvaluateExpression("="+html, 1)
		assert.Nil(err)
		assert.NotEqual(htmlExpect, val)

	})

	t.Run("html error", func(t *testing.T) {
		assert := require.New(t)

		sb := core.Sandbox{
			Context: &core.SandboxContext{
				Items: []map[string]interface{}{
					{"json": map[string]interface{}{"id": 1, "style": "color: red"}},
					{"json": map[string]interface{}{"id": 2, "style": "color: blue"}},
				},
				Functions: core.BuiltInFunctions,
			},
		}
		sb.Initialize()

		eval := core.ExpressionEvaluator{
			Sandbox: &sb,
		}

		// read from html_template.hbs
		htmlBytes, err := os.ReadFile("test_files/html_template_error.hbs")
		assert.Nil(err)
		html := string(htmlBytes)

		htmlExpectBytes, err := os.ReadFile("test_files/html_template_error_expect.html")
		assert.Nil(err)
		htmlExpect := string(htmlExpectBytes)

		val, err := eval.EvaluateExpression("="+html, 0)
		assert.Nil(err)
		assert.Equal(htmlExpect, val)

		val, err = eval.EvaluateExpression("="+html, 1)
		assert.Nil(err)
		assert.NotEqual(htmlExpect, val)

	})

}

func TestNewExpressionEvaluator(t *testing.T) {

	t.Run("NewExpressionEvaluator", func(t *testing.T) {
		assert := require.New(t)

		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.NewExpressionEvaluator(&core.SandboxContext{
			Items: []map[string]interface{}{
				{"json": map[string]interface{}{"a": 1}},
				{"json": map[string]interface{}{"b": 2}},
			},
		},
		)

		assert.NotNil(eval)

		val, err := eval.EvaluateExpression(`={{ $json }}`, 0)

		assert.Nil(err)
		assert.Equal((map[string]interface{}{"a": 1}), val)

	})

	t.Run("NewExpressionEvaluator with nil ctx", func(t *testing.T) {
		assert := require.New(t)

		sb := core.Sandbox{}
		sb.Initialize()

		eval := core.NewExpressionEvaluator(nil)

		assert.NotNil(eval)

		// err
		val, err := eval.EvaluateExpression(`={{ $json }}`, 0)

		assert.Error(err, " $json is not defined at Main:1:14(1)")
		assert.Equal(nil, val)

		// success
		val, err = eval.EvaluateExpression(`=12{{ 1+2 }}`, 0)

		assert.Nil(err)
		assert.Equal("123", val)

	})
}

func TestGetParameterValue(t *testing.T) {

	t.Run("expression nil or empty", func(t *testing.T) {
		assert := require.New(t)

		ctx := &structs.NodeExecuteInput{
			Data: nil,
			Params: &structs.WorkflowNode{
				Parameters: map[string]interface{}{},
			},
		}

		res, err := core.GetParameterValue(
			[]map[string]interface{}{
				{

					"leftValue":  nil,
					"rightValue": "",
				},
			},
			"test",
			ctx, 0, false,
		)

		assert.Nil(err)

		assert.Len(res, 1)
		actual := res.([]interface{})[0].(map[string]interface{})
		assert.Equal(nil, actual["leftValue"])
		assert.Equal("", actual["rightValue"])

	})

	t.Run("expression static string", func(t *testing.T) {
		assert := require.New(t)

		ctx := &structs.NodeExecuteInput{
			Data: nil,
			Params: &structs.WorkflowNode{
				Parameters: map[string]interface{}{},
			},
		}

		res, err := core.GetParameterValue(
			[]map[string]interface{}{
				{
					"leftValue":  "{123}",
					"rightValue": "={123}",
				},
			},
			"test",
			ctx, 0, false,
		)

		assert.Nil(err)

		assert.Len(res, 1)
		actual := res.([]interface{})[0].(map[string]interface{})
		assert.Equal("{123}", actual["leftValue"])
		assert.Equal("{123}", actual["rightValue"])

	})

	t.Run("expression num and string", func(t *testing.T) {
		assert := require.New(t)

		ctx := &structs.NodeExecuteInput{
			Data: nil,
			Params: &structs.WorkflowNode{
				Parameters: map[string]interface{}{},
			},
		}

		res, err := core.GetParameterValue(
			[]map[string]interface{}{
				{
					"leftValue":  "={{123}}{{456}}",
					"rightValue": `={{123456}}`,
				},
			},
			"test",
			ctx, 0, false,
		)

		assert.Nil(err)

		assert.Len(res, 1)
		actual := res.([]interface{})[0].(map[string]interface{})
		assert.Equal("123456", actual["leftValue"])
		assert.Equal(int64(123456), actual["rightValue"])

	})

	t.Run("expression items and index", func(t *testing.T) {
		assert := require.New(t)

		ctx := &structs.NodeExecuteInput{
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id":   1,
							"data": "test1",
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"id":   2,
							"data": "test2",
						},
					},
				},
			},
			Params: &structs.WorkflowNode{
				Parameters: map[string]interface{}{},
			},
		}

		res, err := core.GetParameterValue(
			[]map[string]interface{}{
				{
					"leftValue":  `={{$item['json'].data}}`,
					"rightValue": `={{$json.id}}`,
				},
			},
			"test",
			ctx, 0, false,
		)

		assert.Nil(err)

		assert.Len(res, 1)
		actual := res.([]interface{})[0].(map[string]interface{})
		assert.Equal("test1", actual["leftValue"])
		assert.Equal(int64(1), actual["rightValue"])

		res, err = core.GetParameterValue(
			[]map[string]interface{}{
				{
					"leftValue":  `={{$item['json'].data}}`,
					"rightValue": `={{$json.id}}`,
				},
			},
			"test",
			ctx, 1, false,
		)
		assert.Nil(err)

		assert.Len(res, 1)
		actual = res.([]interface{})[0].(map[string]interface{})
		assert.Equal("test2", actual["leftValue"])
		assert.Equal(int64(2), actual["rightValue"])

	})

	t.Run("expression err", func(t *testing.T) {
		assert := require.New(t)

		ctx := &structs.NodeExecuteInput{
			Data: nil,
			Params: &structs.WorkflowNode{
				Parameters: map[string]interface{}{},
			},
		}

		res, err := core.GetParameterValue(
			[]map[string]interface{}{
				{
					"leftValue":  "={{ notExist }}",
					"rightValue": `={{throw "err"}}`,
				},
			},
			"test",
			ctx, 0, false,
		)

		assert.Error(err, "ReferenceError: notExist is not defined [line 1]")

		assert.Nil(res)

	})

}
