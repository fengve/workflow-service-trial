package core_test

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

func TestNormalizeItems(t *testing.T) {

	t.Run("Normalize items", func(t *testing.T) {
		assert := require.New(t)

		items := structs.NodeData{
			{
				"id":    1,
				"field": "value2",
			},
			{
				"id":    2,
				"field": "value2",
			},
			{
				"json": map[string]interface{}{
					"id":    3,
					"field": "value3",
				},
			},
		}

		items = core.NormalizeItems(items)
		assert.Equal(1, items[0]["json"].(map[string]interface{})["id"])
		assert.Equal("value2", items[1]["json"].(map[string]interface{})["field"])
		assert.Equal("value3", items[2]["json"].(map[string]interface{})["field"])

	})

}

func TestStandardize(t *testing.T) {

	t.Run("Standardize", func(t *testing.T) {
		assert := require.New(t)

		items := structs.NodeData{
			{
				"id":    1,
				"field": goja.Undefined(),
			},
			{
				"id":    2,
				"field": goja.Null(), // should keep null
			},
			{
				"json": map[string]interface{}{
					"id":    3,
					"field": func() {},
				},
			},
		}

		items = core.StandardizeOutput(items)
		assert.Nil(items[0]["field"])
		assert.Equal(goja.Null(), items[1]["field"])
		assert.Nil(items[2]["field"])

	})
}
