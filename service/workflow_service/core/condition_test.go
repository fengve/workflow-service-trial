package core_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
)

func TestExecuteFilterCondition(t *testing.T) {

	t.Run("Boolean", func(t *testing.T) {
		assert := require.New(t)

		var TestCases = []struct {
			LeftValue  interface{}
			RightValue interface{}
			Operation  string
			Expected   bool
			Error      error
		}{
			// exists
			{LeftValue: true, Operation: "exists", Expected: true},
			{LeftValue: false, Operation: "exists", Expected: true},
			{LeftValue: nil, Operation: "exists", Expected: false},
			// notExists
			{LeftValue: nil, Operation: "notExists", Expected: true},
			{LeftValue: true, Operation: "notExists", Expected: false},
			{LeftValue: false, Operation: "notExists", Expected: false},
			// equals
			{LeftValue: true, RightValue: true, Operation: "equals", Expected: true},
			{LeftValue: nil, RightValue: nil, Operation: "equals", Expected: true},
			{LeftValue: false, RightValue: false, Operation: "equals", Expected: true},
			{LeftValue: true, RightValue: true, Operation: "equals", Expected: true},
			{LeftValue: false, RightValue: nil, Operation: "equals", Expected: false},
			// notEquals
			{LeftValue: true, RightValue: false, Operation: "notEquals", Expected: true},
			{LeftValue: nil, RightValue: false, Operation: "notEquals", Expected: true},
			{LeftValue: nil, RightValue: nil, Operation: "notEquals", Expected: false},
			{LeftValue: true, RightValue: true, Operation: "notEquals", Expected: false},
			{LeftValue: false, RightValue: false, Operation: "notEquals", Expected: false},
			// true
			{LeftValue: true, Operation: "true", Expected: true},
			{LeftValue: nil, Operation: "true", Expected: false},
			{LeftValue: false, Operation: "true", Expected: false},
			// false
			{LeftValue: false, Operation: "false", Expected: true},
			{LeftValue: nil, Operation: "false", Expected: true},
			{LeftValue: true, Operation: "false", Expected: false},
		}

		for _, tc := range TestCases {
			t.Run(fmt.Sprintf("%s %v %v %v", tc.Operation, tc.LeftValue, tc.RightValue, tc.Expected), func(t *testing.T) {
				res, err := core.ExecuteFilterCondition(
					false, 0, 0, core.FilterConditionValue{
						LeftValue:  tc.LeftValue,
						RightValue: tc.RightValue,
						Operator: core.ConditionOperator{
							Type:      "boolean",
							Operation: tc.Operation,
						},
					},
				)

				if tc.Error != nil {
					assert.Error(err, tc.Error.Error())
				} else {
					assert.Nil(err)
				}
				assert.Equal(tc.Expected, res)
			})
		}
	})

	t.Run("String", func(t *testing.T) {
		assert := require.New(t)

		var TestCases = []struct {
			LeftValue  interface{}
			RightValue interface{}
			Operation  string
			Expected   bool
			Error      error
		}{
			// exists
			{LeftValue: "", Operation: "exists", Expected: true},
			{LeftValue: "123", Operation: "exists", Expected: true},
			{LeftValue: nil, Operation: "exists", Expected: false},
			// notExists
			{LeftValue: nil, Operation: "notExists", Expected: true},
			{LeftValue: "true", Operation: "notExists", Expected: false},
			{LeftValue: "", Operation: "notExists", Expected: false},
			// equals
			{LeftValue: "test", RightValue: "test", Operation: "equals", Expected: true},
			{LeftValue: "", RightValue: "", Operation: "equals", Expected: true},
			{LeftValue: nil, RightValue: nil, Operation: "equals", Expected: true},
			{LeftValue: "test", RightValue: "test ", Operation: "equals", Expected: false},
			{LeftValue: nil, RightValue: "test", Operation: "equals", Expected: false},
			// notEquals
			{LeftValue: "test", RightValue: "test ", Operation: "notEquals", Expected: true},
			{LeftValue: nil, RightValue: "test ", Operation: "notEquals", Expected: true},
			{LeftValue: "test", RightValue: "test", Operation: "notEquals", Expected: false},
			{LeftValue: "", RightValue: "", Operation: "notEquals", Expected: false},
			{LeftValue: nil, RightValue: nil, Operation: "notEquals", Expected: false},
			// contains
			{LeftValue: "", RightValue: "", Operation: "contains", Expected: true},
			{LeftValue: "1234567", RightValue: "456", Operation: "contains", Expected: true},
			{LeftValue: nil, RightValue: nil, Operation: "contains", Expected: false},
			{LeftValue: "test", RightValue: nil, Operation: "contains", Expected: false},
			// notContains
			{LeftValue: "test", RightValue: "123", Operation: "notContains", Expected: true},
			{LeftValue: nil, RightValue: nil, Operation: "notContains", Expected: false},
			{LeftValue: nil, RightValue: "", Operation: "notContains", Expected: false},
			{LeftValue: "", RightValue: nil, Operation: "notContains", Expected: false},
			{LeftValue: "", RightValue: "", Operation: "notContains", Expected: false},
			// startsWith
			{LeftValue: "test", RightValue: "te", Operation: "startsWith", Expected: true},
			{LeftValue: "test", RightValue: "es", Operation: "startsWith", Expected: false},
			{LeftValue: "", RightValue: "", Operation: "startsWith", Expected: true},
			{LeftValue: nil, RightValue: nil, Operation: "startsWith", Expected: false},
			{LeftValue: nil, RightValue: "", Operation: "startsWith", Expected: false},
			{LeftValue: "", RightValue: nil, Operation: "startsWith", Expected: false},
			// endsWith
			{LeftValue: "test", RightValue: "st", Operation: "endsWith", Expected: true},
			{LeftValue: "test", RightValue: "es", Operation: "endsWith", Expected: false},
			{LeftValue: "", RightValue: "", Operation: "endsWith", Expected: true},
			{LeftValue: nil, RightValue: nil, Operation: "endsWith", Expected: false},
			{LeftValue: nil, RightValue: "", Operation: "endsWith", Expected: false},
			{LeftValue: "", RightValue: nil, Operation: "endsWith", Expected: false},
			// notStartsWith
			{LeftValue: "test", RightValue: "es", Operation: "notStartsWith", Expected: true},
			{LeftValue: "test", RightValue: "te", Operation: "notStartsWith", Expected: false},
			{LeftValue: "", RightValue: "", Operation: "notStartsWith", Expected: false},
			{LeftValue: nil, RightValue: nil, Operation: "notStartsWith", Expected: false},
			{LeftValue: nil, RightValue: "", Operation: "notStartsWith", Expected: false},
			{LeftValue: "", RightValue: nil, Operation: "notStartsWith", Expected: false},
			// notEndsWith
			{LeftValue: "test", RightValue: "es", Operation: "notEndsWith", Expected: true},
			{LeftValue: "test", RightValue: "st", Operation: "notEndsWith", Expected: false},
			{LeftValue: "", RightValue: "", Operation: "notEndsWith", Expected: false},
			{LeftValue: nil, RightValue: nil, Operation: "notEndsWith", Expected: false},
			{LeftValue: nil, RightValue: "", Operation: "notEndsWith", Expected: false},
			{LeftValue: "", RightValue: nil, Operation: "notEndsWith", Expected: false},
			// regex
			{LeftValue: "test", RightValue: "es", Operation: "regex", Expected: true},
			{LeftValue: "test", RightValue: "es$", Operation: "regex", Expected: false},

			// notRegex
			{LeftValue: "test", RightValue: "es$", Operation: "notRegex", Expected: true},
			{LeftValue: "test", RightValue: "es", Operation: "notRegex", Expected: false},
		}

		for _, tc := range TestCases {
			t.Run(fmt.Sprintf("%s %v %v %v", tc.Operation, tc.LeftValue, tc.RightValue, tc.Expected), func(t *testing.T) {

				res, err := core.ExecuteFilterCondition(
					false, 0, 0, core.FilterConditionValue{
						LeftValue:  tc.LeftValue,
						RightValue: tc.RightValue,
						Operator: core.ConditionOperator{
							Type:      "string",
							Operation: tc.Operation,
						},
					},
				)

				if tc.Error != nil {
					assert.Error(err, tc.Error.Error())
				} else {
					assert.Nil(err)
				}
				assert.Equal(tc.Expected, res)
			})
		}

	})

	t.Run("Number", func(t *testing.T) {
		assert := require.New(t)

		var TestCases = []struct {
			LeftValue  interface{}
			RightValue interface{}
			Operation  string
			Expected   bool
			Error      error
		}{
			// exists
			{LeftValue: 0, Operation: "exists", Expected: true},
			{LeftValue: -123, Operation: "exists", Expected: true},
			{LeftValue: "", Operation: "exists", Expected: true},
			{LeftValue: nil, Operation: "exists", Expected: false},
			// notExists
			{LeftValue: nil, Operation: "notExists", Expected: true},
			{LeftValue: 0, Operation: "notExists", Expected: false},
			{LeftValue: -123, Operation: "notExists", Expected: false},
			{LeftValue: "", Operation: "notExists", Expected: false},
			// equals
			{LeftValue: 123, RightValue: 123, Operation: "equals", Expected: true},
			{LeftValue: "123", RightValue: "123", Operation: "equals", Expected: true},
			{LeftValue: "123", RightValue: 123, Operation: "equals", Expected: true},
			{LeftValue: 3.14, RightValue: 3.14, Operation: "equals", Expected: true},
			{LeftValue: 0, RightValue: 0, Operation: "equals", Expected: true},
			{LeftValue: -1234567.1234567, RightValue: -1234567.1234567, Operation: "equals", Expected: true},
			// TODO big number for goja
			{LeftValue: nil, RightValue: nil, Operation: "equals", Expected: true},
			{LeftValue: nil, RightValue: 1, Operation: "equals", Expected: false},
			{LeftValue: 1, RightValue: 2, Operation: "equals", Expected: false},
			// notEquals
			{LeftValue: 123, RightValue: 1, Operation: "notEquals", Expected: true},
			{LeftValue: nil, RightValue: 1, Operation: "notEquals", Expected: true},
			{LeftValue: 123, RightValue: 123, Operation: "notEquals", Expected: false},
			{LeftValue: "123", RightValue: 123, Operation: "notEquals", Expected: false},
			{LeftValue: nil, RightValue: nil, Operation: "notEquals", Expected: false},
			{LeftValue: "123", RightValue: "123", Operation: "notEquals", Expected: false},
			// gt
			{LeftValue: 123, RightValue: 1, Operation: "gt", Expected: true},
			{LeftValue: 123, RightValue: "1", Operation: "gt", Expected: true},
			{LeftValue: 123, RightValue: 123, Operation: "gt", Expected: false},
			{LeftValue: 123, RightValue: 1000, Operation: "gt", Expected: false},
			{LeftValue: 1, RightValue: nil, Operation: "gt", Expected: false},
			{LeftValue: nil, RightValue: 1, Operation: "gt", Expected: false},
			{LeftValue: nil, RightValue: nil, Operation: "gt", Expected: false},
			// gte
			{LeftValue: 123, RightValue: 1, Operation: "gte", Expected: true},
			{LeftValue: 123, RightValue: 123, Operation: "gte", Expected: true},
			{LeftValue: 123, RightValue: "123", Operation: "gte", Expected: true},
			{LeftValue: 123, RightValue: 1000, Operation: "gte", Expected: false},
			{LeftValue: 1, RightValue: nil, Operation: "gte", Expected: false},
			{LeftValue: nil, RightValue: 1, Operation: "gte", Expected: false},
			{LeftValue: nil, RightValue: nil, Operation: "gte", Expected: false},
			// lt
			{LeftValue: 123, RightValue: 1, Operation: "lt", Expected: false},
			{LeftValue: 123, RightValue: 123, Operation: "lt", Expected: false},
			{LeftValue: 123, RightValue: 1000, Operation: "lt", Expected: true},
			{LeftValue: 123, RightValue: "1000", Operation: "lt", Expected: true},
			{LeftValue: 1, RightValue: nil, Operation: "lt", Expected: false},
			{LeftValue: nil, RightValue: 1, Operation: "lt", Expected: false},
			{LeftValue: nil, RightValue: nil, Operation: "lt", Expected: false},
			// lte
			{LeftValue: 123, RightValue: 1, Operation: "lte", Expected: false},
			{LeftValue: 123, RightValue: 123, Operation: "lte", Expected: true},
			{LeftValue: 123, RightValue: 1000, Operation: "lte", Expected: true},
			{LeftValue: 123, RightValue: "123", Operation: "lte", Expected: true},
			{LeftValue: 1, RightValue: nil, Operation: "lte", Expected: false},
			{LeftValue: nil, RightValue: 1, Operation: "lte", Expected: false},
			{LeftValue: nil, RightValue: nil, Operation: "lte", Expected: false},
		}

		for _, tc := range TestCases {
			t.Run(fmt.Sprintf("%s %v %v %v", tc.Operation, tc.LeftValue, tc.RightValue, tc.Expected), func(t *testing.T) {

				res, err := core.ExecuteFilterCondition(
					false, 0, 0, core.FilterConditionValue{
						LeftValue:  tc.LeftValue,
						RightValue: tc.RightValue,
						Operator: core.ConditionOperator{
							Type:      "number",
							Operation: tc.Operation,
						},
					},
				)

				if tc.Error != nil {
					assert.Error(err, tc.Error.Error())
				} else {
					assert.Nil(err)
				}
				assert.Equal(tc.Expected, res)
			})
		}
	})

	t.Run("Array", func(t *testing.T) {
		assert := require.New(t)

		var TestCases = []struct {
			LeftValue  interface{}
			RightValue interface{}
			Operation  string
			Expected   bool
			Error      error
		}{
			// exists
			{LeftValue: []interface{}{}, Operation: "exists", Expected: true},
			{LeftValue: []int{1, 2, 3}, Operation: "exists", Expected: true},
			{LeftValue: "[]", Operation: "exists", Expected: true},
			{LeftValue: nil, Operation: "exists", Expected: false},
			// notExists
			{LeftValue: nil, Operation: "notExists", Expected: true},
			{LeftValue: "[]", Operation: "notExists", Expected: false},
			{LeftValue: []int{1, 2, 3}, Operation: "notExists", Expected: false},
			{LeftValue: []interface{}{}, Operation: "notExists", Expected: false},
			// contains
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 1, Operation: "contains", Expected: true},
			{LeftValue: []interface{}{"1", 2, 3}, RightValue: "1", Operation: "contains", Expected: true},
			{LeftValue: []interface{}{1, 2, 3, nil}, RightValue: nil, Operation: "contains", Expected: true},
			{LeftValue: []interface{}{[]interface{}{1, 2, 3}, 2, 3}, RightValue: []interface{}{1, 2, 3}, Operation: "contains", Expected: true},
			// Note: n8n return false but we return true ^
			{LeftValue: []interface{}{map[string]interface{}{"id": 1}, 2, 3}, RightValue: map[string]interface{}{"id": 1}, Operation: "contains", Expected: true},
			// Note: n8n return false but we return true ^
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 4, Operation: "contains", Expected: false},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: []interface{}{1, 2}, Operation: "contains", Expected: false},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: []interface{}{1, 2, 3}, Operation: "contains", Expected: false},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: []interface{}{1, 2, 3, 4}, Operation: "contains", Expected: false},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: nil, Operation: "contains", Expected: false},
			{LeftValue: nil, RightValue: nil, Operation: "contains", Expected: false},
			// notContains
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 4, Operation: "notContains", Expected: true},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: []interface{}{1, 2}, Operation: "notContains", Expected: true},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: []interface{}{1, 2, 3}, Operation: "notContains", Expected: true},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: []interface{}{1, 2, 3, 4}, Operation: "notContains", Expected: true},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: nil, Operation: "notContains", Expected: true},
			{LeftValue: nil, RightValue: nil, Operation: "notContains", Expected: false},
			// lengthEquals
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 3, Operation: "lengthEquals", Expected: true},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 2, Operation: "lengthEquals", Expected: false},
			{LeftValue: []interface{}{}, RightValue: 0, Operation: "lengthEquals", Expected: true},
			{LeftValue: nil, RightValue: 0, Operation: "lengthEquals", Expected: false},
			// lengthNotEquals
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 3, Operation: "lengthNotEquals", Expected: false},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 2, Operation: "lengthNotEquals", Expected: true},
			{LeftValue: []interface{}{}, RightValue: 0, Operation: "lengthNotEquals", Expected: false},
			{LeftValue: nil, RightValue: 0, Operation: "lengthNotEquals", Expected: false},
			// lengthGt
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 2, Operation: "lengthGt", Expected: true},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 3, Operation: "lengthGt", Expected: false},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 4, Operation: "lengthGt", Expected: false},
			{LeftValue: []interface{}{}, RightValue: 0, Operation: "lengthGt", Expected: false},
			{LeftValue: nil, RightValue: 0, Operation: "lengthGt", Expected: false},
			// lengthLt
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 2, Operation: "lengthLt", Expected: false},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 3, Operation: "lengthLt", Expected: false},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 4, Operation: "lengthLt", Expected: true},
			{LeftValue: []interface{}{}, RightValue: 0, Operation: "lengthLt", Expected: false},
			{LeftValue: nil, RightValue: 0, Operation: "lengthLt", Expected: false},
			// lengthGte
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 2, Operation: "lengthGte", Expected: true},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 3, Operation: "lengthGte", Expected: true},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 4, Operation: "lengthGte", Expected: false},
			{LeftValue: []interface{}{}, RightValue: 0, Operation: "lengthGte", Expected: true},
			{LeftValue: nil, RightValue: 0, Operation: "lengthGte", Expected: false},
			// lengthLte
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 2, Operation: "lengthLte", Expected: false},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 3, Operation: "lengthLte", Expected: true},
			{LeftValue: []interface{}{1, 2, 3}, RightValue: 4, Operation: "lengthLte", Expected: true},
			{LeftValue: []interface{}{}, RightValue: 0, Operation: "lengthLte", Expected: true},
			{LeftValue: nil, RightValue: 0, Operation: "lengthLte", Expected: false},
			// empty
			{LeftValue: []interface{}{}, Operation: "empty", Expected: true},
			{LeftValue: "[]", Operation: "empty", Expected: true},
			{LeftValue: nil, Operation: "empty", Expected: true},
			{LeftValue: []interface{}{1, 2, 3}, Operation: "empty", Expected: false},
			// notEmpty
			{LeftValue: []interface{}{1, 2, 3}, Operation: "notEmpty", Expected: true},
			{LeftValue: []interface{}{}, Operation: "notEmpty", Expected: false},
			{LeftValue: nil, Operation: "notEmpty", Expected: false},
		}

		for _, tc := range TestCases {
			t.Run(fmt.Sprintf("%s %v %v %v", tc.Operation, tc.LeftValue, tc.RightValue, tc.Expected), func(t *testing.T) {

				res, err := core.ExecuteFilterCondition(
					false, 0, 0, core.FilterConditionValue{
						LeftValue:  tc.LeftValue,
						RightValue: tc.RightValue,
						Operator: core.ConditionOperator{
							Type:      "array",
							Operation: tc.Operation,
						},
					},
				)

				if tc.Error != nil {
					assert.Error(err, tc.Error.Error())
				} else {
					assert.Nil(err)
				}

				assert.Equal(tc.Expected, res)
			})
		}

	})

	t.Run("Object", func(t *testing.T) {
		assert := require.New(t)

		var TestCases = []struct {
			LeftValue  interface{}
			RightValue interface{}
			Operation  string
			Expected   bool
			Error      error
		}{
			// exists
			{LeftValue: map[string]interface{}{}, Operation: "exists", Expected: true},
			{LeftValue: map[string]interface{}{"data": "foo"}, Operation: "exists", Expected: true},
			{LeftValue: "{}", Operation: "exists", Expected: true},
			{LeftValue: nil, Operation: "exists", Expected: false},
			// notExists
			{LeftValue: nil, Operation: "notExists", Expected: true},
			{LeftValue: "{}", Operation: "notExists", Expected: false},
			{LeftValue: map[string]interface{}{"data": "foo"}, Operation: "notExists", Expected: false},
			{LeftValue: map[string]interface{}{}, Operation: "notExists", Expected: false},
			// empty
			{LeftValue: map[string]interface{}{}, Operation: "empty", Expected: true},
			{LeftValue: nil, Operation: "empty", Expected: false}, // note: n8n return false if leftValue is nil
			{LeftValue: map[string]interface{}{"data": "foo"}, Operation: "empty", Expected: false},
			{LeftValue: "{}", Operation: "empty", Expected: false, Error: core.TypeError},
			// notEmpty
			{LeftValue: map[string]interface{}{"data": "foo"}, Operation: "notEmpty", Expected: true},
			{LeftValue: map[string]interface{}{}, Operation: "notEmpty", Expected: false},
			{LeftValue: nil, Operation: "notEmpty", Expected: false},
		}

		for _, tc := range TestCases {
			t.Run(fmt.Sprintf("%s %v %v %v", tc.Operation, tc.LeftValue, tc.RightValue, tc.Expected), func(t *testing.T) {

				res, err := core.ExecuteFilterCondition(
					false, 0, 0, core.FilterConditionValue{
						LeftValue:  tc.LeftValue,
						RightValue: tc.RightValue,
						Operator: core.ConditionOperator{
							Type:      "object",
							Operation: tc.Operation,
						},
					},
				)
				if tc.Error != nil {
					assert.Error(err, tc.Error.Error())
				} else {
					assert.Nil(err)
				}
				assert.Equal(tc.Expected, res)
			})
		}

	})

	t.Run("dateTime", func(t *testing.T) {

		assert := require.New(t)

		var TestCases = []struct {
			LeftValue  interface{}
			RightValue interface{}
			Operation  string
			Expected   bool
			Error      error
		}{
			// exists
			{LeftValue: "2024-01-01T00:00:00Z", Operation: "exists", Expected: true},
			{LeftValue: "2024-01-01T00:00:00+00:00", Operation: "exists", Expected: true},
			{LeftValue: "2024-01-01T00:00:00", Operation: "exists", Expected: true},
			{LeftValue: nil, Operation: "exists", Expected: false},
			// notExists
			{LeftValue: nil, Operation: "notExists", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", Operation: "notExists", Expected: false},
			{LeftValue: "2024-01-01T00:00:00+00:00", Operation: "notExists", Expected: false},
			{LeftValue: "2024-01-01T00:00:00", Operation: "notExists", Expected: false},
			// equals
			{LeftValue: nil, RightValue: nil, Operation: "equals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00Z", Operation: "equals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00+00:00", Operation: "equals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00", Operation: "equals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01 00:00:00", Operation: "equals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01", Operation: "equals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00+00:00", Operation: "equals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:01Z", Operation: "equals", Expected: false},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00+08:00", Operation: "equals", Expected: false},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:01", Operation: "equals", Expected: false},
			{LeftValue: nil, RightValue: "2024-01-01T00:00:00Z", Operation: "equals", Expected: false},
			// notEquals
			{LeftValue: nil, RightValue: "2024-01-01T00:00:00Z", Operation: "notEquals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: nil, Operation: "notEquals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:01Z", Operation: "notEquals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:01", Operation: "notEquals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:01+00:00", Operation: "notEquals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:01+00:00", Operation: "notEquals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00Z", Operation: "notEquals", Expected: false},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00+00:00", Operation: "notEquals", Expected: false},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00", Operation: "notEquals", Expected: false},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00+00:00", Operation: "notEquals", Expected: false},
			{LeftValue: nil, RightValue: nil, Operation: "notEquals", Expected: false},
			// after
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2020-01-01T00:00:00Z", Operation: "after", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00Z", Operation: "after", Expected: false},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2025-01-01T00:00:00Z", Operation: "after", Expected: false},
			{LeftValue: nil, RightValue: "2024-01-01T00:00:00Z", Operation: "after", Expected: false},

			// before
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2025-01-01T00:00:00Z", Operation: "before", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2020-01-01T00:00:00Z", Operation: "before", Expected: false},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00Z", Operation: "before", Expected: false},
			{LeftValue: nil, RightValue: "2024-01-01T00:00:00Z", Operation: "before", Expected: false},

			// afterEquals
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2020-01-01T00:00:00Z", Operation: "afterEquals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00Z", Operation: "afterEquals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2025-01-01T00:00:00Z", Operation: "afterEquals", Expected: false},
			{LeftValue: nil, RightValue: "2024-01-01T00:00:00Z", Operation: "afterEquals", Expected: false},

			// beforeEquals
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2020-01-01T00:00:00Z", Operation: "beforeEquals", Expected: false},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2024-01-01T00:00:00Z", Operation: "beforeEquals", Expected: true},
			{LeftValue: "2024-01-01T00:00:00Z", RightValue: "2025-01-01T00:00:00Z", Operation: "beforeEquals", Expected: true},
			{LeftValue: nil, RightValue: "2024-01-01T00:00:00Z", Operation: "beforeEquals", Expected: false},
		}

		for _, tc := range TestCases {
			t.Run(fmt.Sprintf("%s %v %v %v", tc.Operation, tc.LeftValue, tc.RightValue, tc.Expected), func(t *testing.T) {

				res, err := core.ExecuteFilterCondition(
					false, 0, 0, core.FilterConditionValue{
						LeftValue:  tc.LeftValue,
						RightValue: tc.RightValue,
						Operator: core.ConditionOperator{
							Type:      "dateTime",
							Operation: tc.Operation,
						},
					},
				)
				if tc.Error != nil {
					assert.Error(err, tc.Error.Error())
				} else {
					assert.Nil(err)
				}
				assert.Equal(tc.Expected, res)
			})
		}

	})

}
