package core_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"

	"github.com/sqlc-dev/pqtype"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type (
	UtilTestSuit struct {
		suite.Suite
	}
)

func Test_Util(t *testing.T) {
	suite.Run(t, new(UtilTestSuit))
}

func (s *UtilTestSuit) Test() {
	s.T().Run("TestConvertWorkflowEntityToReq", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		manualRunReq := &structs.WorkflowManualRunRequest{}
		testFileList := []string{
			"./test_files/workflow_execution_email.json",
		}
		ctx := context.Background()

		for _, file := range testFileList {
			testFile, err := os.ReadFile(file)
			assert.Nil(err)

			err = json.Unmarshal(testFile, manualRunReq)

			entity, err := rdsDbQueries.CreateWorkflowEntity(
				ctx,
				lib.CreateWorkflowEntityParams{
					Name:         manualRunReq.WorkflowData.Name,
					Active:       manualRunReq.WorkflowData.Active,
					Nodes:        json.RawMessage(core.JsonStr(manualRunReq.WorkflowData.Nodes)),
					Connections:  json.RawMessage(core.JsonStr(manualRunReq.WorkflowData.Connections)),
					Settings:     pqtype.NullRawMessage{json.RawMessage(core.JsonStr(manualRunReq.WorkflowData.Settings)), true},
					StaticData:   pqtype.NullRawMessage{Valid: false},
					PinData:      pqtype.NullRawMessage{json.RawMessage("{}"), true},
					VersionId:    sql.NullString{manualRunReq.WorkflowData.VersionId, true}, // new versionid.
					TriggerCount: 0,
					ID:           manualRunReq.WorkflowData.ID,
					Meta:         pqtype.NullRawMessage{Valid: false},
					SugerOrgId:   "",
				})
			assert.Nil(err)
			assert.Equal("YHXuzUV1vj6Mno9b", entity.ID)

			req, err := core.GetWorkflowEntityById(ctx, "YHXuzUV1vj6Mno9b")
			assert.Nil(err)
			assert.NotNil(req)
		}
	})

	s.T().Run("TestIsArray", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		assert.True(core.IsArray([]interface{}{1, 2, 3}))
		assert.True(core.IsArray([]int{1, 2, 3}))
		assert.True(core.IsArray([]map[string]interface{}{
			{"id": 1},
			{"id": 3},
			{"id": 3},
		}))
		assert.False(core.IsArray(map[string]interface{}{"a": 1, "b": 2}))

	})

	s.T().Run("TestConvertToInterfaceArray", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		res, err := core.ConvertToInterfaceArray([]interface{}{1, 2, 3})
		assert.Nil(err)
		assert.Equal([]interface{}{1, 2, 3}, res)

		res, err = core.ConvertToInterfaceArray([]int{1, 2, 3})

		assert.Nil(err)
		assert.Equal([]interface{}{1, 2, 3}, res)

		res, err = core.ConvertToInterfaceArray([]string{"a", "b"})

		assert.Nil(err)
		assert.Equal([]interface{}{"a", "b"}, res)

		res, err = core.ConvertToInterfaceArray([]map[string]interface{}{
			{"id": 1},
			{"id": 3},
			{"id": 3},
		})
		expected := []interface{}{
			map[string]interface{}{"id": 1},
			map[string]interface{}{"id": 3},
			map[string]interface{}{"id": 3},
		}
		assert.Nil(err)
		assert.Equal(expected, res)

		res, err = core.ConvertToInterfaceArray(structs.NodeData{
			{"id": 1},
			{"id": 2},
			{"json": map[string]interface{}{
				"id": 3,
			}},
		})
		expected = []interface{}{
			map[string]interface{}{"id": 1},
			map[string]interface{}{"id": 2},
			map[string]interface{}{"json": map[string]interface{}{
				"id": 3,
			}},
		}
		assert.Nil(err)
		assert.Equal(expected, res)

		// errors

		res, err = core.ConvertToInterfaceArray(map[string]interface{}{"a": 1, "b": 2})

		assert.Error(err)
		assert.Nil(res)

	})

	s.T().Run("TestConvertToInterfaceMap", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		res, err := core.ConvertToInterfaceMap(map[string]interface{}{
			"foo": "bar",
		})
		assert.Nil(err)
		assert.Equal(map[string]interface{}{
			"foo": "bar",
		}, res)

		res, err = core.ConvertToInterfaceMap(map[string]map[string]interface{}{
			"json": {
				"id":   1,
				"data": "value",
			},
		})

		assert.Nil(err)
		assert.Equal(map[string]interface{}{
			"json": map[string]interface{}{
				"id":   1,
				"data": "value",
			},
		}, res)

		res, err = core.ConvertToInterfaceMap(structs.NodeSingleData{
			"json": map[string]interface{}{
				"id":   1,
				"data": "value",
			},
		})

		assert.Nil(err)
		assert.Equal(map[string]interface{}{
			"json": map[string]interface{}{
				"id":   1,
				"data": "value",
			},
		}, res)

		// errors
		res, err = core.ConvertToInterfaceMap([]interface{}{1, 2, 3})

		assert.Error(err)
		assert.Nil(res)

		res, err = core.ConvertToInterfaceMap([]map[string]interface{}{
			{"id": 1},
			{"id": 2},
			{"id": 3},
		})

		assert.Error(err)
		assert.Nil(res)
	})

	s.T().Run("TestConvertToFloat64", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		res, err := core.ConvertToFloat64(1)
		assert.Nil(err)
		assert.Equal(float64(1), res)

		res, err = core.ConvertToFloat64(int64(1))
		assert.Nil(err)
		assert.Equal(float64(1), res)

		res, err = core.ConvertToFloat64(1.1)
		assert.Nil(err)
		assert.Equal(float64(1.1), res)

		// errors
		res, err = core.ConvertToFloat64("1")
		assert.Error(err)
		assert.Equal(float64(0), res)

		res, err = core.ConvertToFloat64([]interface{}{1, 2, 3})
		assert.Error(err)
		assert.Equal(float64(0), res)

		res, err = core.ConvertToFloat64(map[string]interface{}{"a": 1, "b": 2})
		assert.Error(err)
		assert.Equal(float64(0), res)

	})

	s.T().Run("TestGetMapValueByPath", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// success
		res, ok := core.GetMapValueByPath(map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		}, "foo.bar")

		assert.True(ok)
		assert.Equal("baz", res)

		res, ok = core.GetMapValueByPath(map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		}, "foo")

		assert.True(ok)
		assert.Equal(map[string]interface{}{
			"bar": "baz",
		}, res)

		res, ok = core.GetMapValueByPath(map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": []interface{}{"baz", "quz"},
			},
		}, "foo.bar")

		assert.True(ok)
		assert.Equal([]interface{}{
			"baz", "quz",
		}, res)

		// fail
		res, ok = core.GetMapValueByPath(map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		}, "foo.baz")

		assert.False(ok)
		assert.Nil(res)

		res, ok = core.GetMapValueByPath(map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		}, "bar")

		assert.False(ok)
		assert.Nil(res)

		res, ok = core.GetMapValueByPath(map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		}, "bar.baz")

		assert.False(ok)
		assert.Nil(res)

		res, ok = core.GetMapValueByPath(map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		}, "foo.bar.baz.qux")

		assert.False(ok)
		assert.Nil(res)

	})

	s.T().Run("Test GetMapValueByPath not change the input outside variable", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		params := map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		}

		res, ok := core.GetMapValueByPath(params, "foo.bar")
		assert.True(ok)
		assert.Equal("baz", res)

		// check if the original map is not changed
		assert.Equal(map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		}, params)
	})

	s.T().Run("Test_GetValueFromMap GetValueFromMapWithDefault", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// a nil map will not panic
		// key not found, ok will be false
		val, ok := core.GetValueFromMap[string](nil, "key")
		assert.False(ok)
		assert.Equal("", val)

		// Get value from map, key found.
		val, ok = core.GetValueFromMap[string](map[string]any{"key": "value"}, "key")
		assert.True(ok)
		assert.Equal("value", val)

		// Get value from map, with default value, key found.
		val = core.GetValueFromMapWithDefault(map[string]interface{}{"key": "value"}, "key", "default")
		assert.Equal("value", val)

		// Get value from map, with default value, key not found.
		val = core.GetValueFromMapWithDefault(map[string]interface{}{"key": "value"}, "key1", "default")
		assert.Equal("default", val)
	})

	s.T().Run("TestUnflattenString", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		fileBytes, err := os.ReadFile("./test_files/workflow_execution_data_4342.json")
		assert.Nil(err)
		source := string(fileBytes)

		unflattenResult, err := core.UnflattenString(source)
		assert.Nil(err)
		assert.NotEmpty(unflattenResult)

		runExecutionData := structs.WorkflowRunExecutionData{}
		err = structs.UnmarshalOmitEmpty([]byte(unflattenResult), &runExecutionData)
		assert.Nil(err)
		assert.Equal("Suger Slack", runExecutionData.ResultData.LastNodeExecuted)
	})

	s.T().Run("Test_GetHtmlContentFromTemplate", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		template := `<!DOCTYPE html>
<html>
<head>
	<title>{{.Title}}</title>
</head>
<body>
	<h1>{{.Header}}</h1>
	<p>{{.Content}}</p>
</body>
</html>`

		data := map[string]interface{}{
			"Title":   "Hello, World!",
			"Header":  "Welcome to the world of Go Templates",
			"Content": "This is a sample content",
		}
		result, err := core.GetHtmlContentFromTemplate(template, data)
		assert.Nil(err)
		expectedResult := `<!DOCTYPE html>
<html>
<head>
	<title>Hello, World!</title>
</head>
<body>
	<h1>Welcome to the world of Go Templates</h1>
	<p>This is a sample content</p>
</body>
</html>`
		assert.Equal(expectedResult, result)
	})

	s.T().Run("Test_GetHtmlContentFromTemplate_No_Template", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		template := "Hello, World!"
		result, err := core.GetHtmlContentFromTemplate(template, nil)
		assert.Nil(err)
		assert.Equal("Hello, World!", result)
	})

	s.T().Run("Test_GetHtmlContentFromTemplate_Embedded_Template", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		template := `{{ define "Tmpl" }}<h1>This is a template</h1>{{ end }}{{ template "Tmpl" . }}`
		result, err := core.GetHtmlContentFromTemplate(template, nil)
		assert.Nil(err)
		assert.Equal("<h1>This is a template</h1>", result)
	})

	s.T().Run("Test_GetHtmlContentFromTemplate_With_Array", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		template := `This is {{ index .data 0 "json" "Title"}}`
		result, err := core.GetHtmlContentFromTemplate(template, map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"json": map[string]interface{}{
						"Title": "Testing Template With Array",
					},
				},
			},
		})
		assert.Nil(err)
		assert.Equal("This is Testing Template With Array", result)
	})
}
