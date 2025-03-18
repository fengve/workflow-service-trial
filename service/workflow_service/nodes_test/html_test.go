package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go  service/workflow_service/nodes_test/html_test.go

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	htmlnode "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/html"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type HtmlTestSuite struct {
	suite.Suite
}

func Test_Html(t *testing.T) {
	suite.Run(t, new(HtmlTestSuite))
}

func (s *HtmlTestSuite) Test() {
	s.T().Run("TestHtmlGenerate", func(t *testing.T) {
		assert := require.New(s.T())

		_ = htmlnode.HtmlExecutor{}
		executor := core.NewExecutor(htmlnode.Name)
		node := executor.GetNode()
		assert.Equal(9, len(node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.Properties))
		// iconUrl
		assert.Equal("/icons/embed/n8n-nodes-base.html/html.svg", node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.IconUrl)
		assert.Equal("json", node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.Properties[3].Default)
	})

	s.T().Run("TestHtmlExecute", func(t *testing.T) {
		assert := require.New(s.T())

		inputData := structs.NodeData{
			structs.NodeSingleData{
				"json": map[string]interface{}{
					"id":    1,
					"data":  "data1",
					"style": "color:red;",
				},
			},
			structs.NodeSingleData{
				"json": map[string]interface{}{
					"id":    2,
					"data":  "data2",
					"style": "color:blue;",
				},
			},
			structs.NodeSingleData{
				"json": map[string]interface{}{
					"id":    3,
					"data":  nil,
					"style": "color:grey;",
				},
			},
		}
		{ // empty
			htmlNode := &htmlnode.HtmlExecutor{}
			input := &structs.NodeExecuteInput{
				Params: &structs.WorkflowNode{},
				Data:   []structs.NodeData{inputData},
			}
			result := htmlNode.Execute(context.Background(), input)
			assert.Equal(3, len(result.ExecutorData[0]))
		}

		htmlGenerated := []string{
			"<!DOCTYPE html>\n\n<html>\n  <head>\n    <meta charset=\"UTF-8\" />\n    <title>This is title 1</title>\n  </head>\n  <body>\n    <div class=\"container\">\n      <h1>This is an H1 heading</h1>\n      <h2>This is an H2 heading</h2>\n      <p>This is a paragraph 1</p>\n      <p>This is a paragraph 1</p>\n      <p>Total paragraph 3</p>\n    </div>\n  </body>\n</html>\n\n<style>\n  .container {\n    background-color: #ffffff;\n    text-align: center;\n    padding: 16px;\n    border-radius: 8px;\n  }\n\n  h1 {\n    color: #ff6d5a;\n    font-size: 24px;\n    font-weight: bold;\n    padding: 8px;\n    color:red;\n  }\n\n  h2 {\n    color: #909399;\n    font-size: 18px;\n    font-weight: bold;\n    padding: 8px;\n  }\n</style>\n\n<script>\n  console.log(\"Hello World!\", data1);\n</script>\n",
			"<!DOCTYPE html>\n\n<html>\n  <head>\n    <meta charset=\"UTF-8\" />\n    <title>This is title 2</title>\n  </head>\n  <body>\n    <div class=\"container\">\n      <h1>This is an H1 heading</h1>\n      <h2>This is an H2 heading</h2>\n      <p>This is a paragraph 1</p>\n      <p>This is a paragraph 2</p>\n      <p>Total paragraph 3</p>\n    </div>\n  </body>\n</html>\n\n<style>\n  .container {\n    background-color: #ffffff;\n    text-align: center;\n    padding: 16px;\n    border-radius: 8px;\n  }\n\n  h1 {\n    color: #ff6d5a;\n    font-size: 24px;\n    font-weight: bold;\n    padding: 8px;\n    color:blue;\n  }\n\n  h2 {\n    color: #909399;\n    font-size: 18px;\n    font-weight: bold;\n    padding: 8px;\n  }\n</style>\n\n<script>\n  console.log(\"Hello World!\", data2);\n</script>\n",
			"<!DOCTYPE html>\n\n<html>\n  <head>\n    <meta charset=\"UTF-8\" />\n    <title>This is title 3</title>\n  </head>\n  <body>\n    <div class=\"container\">\n      <h1>This is an H1 heading</h1>\n      <h2>This is an H2 heading</h2>\n      <p>This is a paragraph 1</p>\n      <p>This is a paragraph 3</p>\n      <p>Total paragraph 3</p>\n    </div>\n  </body>\n</html>\n\n<style>\n  .container {\n    background-color: #ffffff;\n    text-align: center;\n    padding: 16px;\n    border-radius: 8px;\n  }\n\n  h1 {\n    color: #ff6d5a;\n    font-size: 24px;\n    font-weight: bold;\n    padding: 8px;\n    color:grey;\n  }\n\n  h2 {\n    color: #909399;\n    font-size: 18px;\n    font-weight: bold;\n    padding: 8px;\n  }\n</style>\n\n<script>\n  console.log(\"Hello World!\", );\n</script>\n"}

		// generate html
		params := structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/html-params_generate_html.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &params)
		assert.Nil(err)

		html, err := os.ReadFile("./test_files/html-params_generate_html.html")
		assert.Nil(err)

		params.Parameters["html"] = string(html)

		htmlNode := &htmlnode.HtmlExecutor{}
		input := &structs.NodeExecuteInput{
			Params: &params,
			Data:   []structs.NodeData{inputData},
		}
		result := htmlNode.Execute(context.Background(), input)
		assert.Equal([]structs.WorkflowNodeExecutionError(nil), result.Errors)
		assert.Equal(structs.WorkflowExecutionStatus_Success, result.ExecutionStatus)
		assert.Equal(3, len(result.ExecutorData[0]))

		assert.Equal(htmlGenerated[0],
			result.ExecutorData[0][0]["json"].(map[string]interface{})["html"])

		assert.Equal(htmlGenerated[1],
			result.ExecutorData[0][1]["json"].(map[string]interface{})["html"])

		assert.Equal(htmlGenerated[2],
			result.ExecutorData[0][2]["json"].(map[string]interface{})["html"])

		nextInputData := result.ExecutorData[0]

		// extract html
		params = structs.WorkflowNode{}
		testFile, err = os.ReadFile("./test_files/html-params_extract_html.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &params)
		assert.Nil(err)

		htmlNode = &htmlnode.HtmlExecutor{}
		input = &structs.NodeExecuteInput{
			Params: &params,
			Data:   []structs.NodeData{nextInputData},
		}

		result = htmlNode.Execute(context.Background(), input)
		assert.Equal([]structs.WorkflowNodeExecutionError(nil), result.Errors)
		assert.Equal(structs.WorkflowExecutionStatus_Success, result.ExecutionStatus)
		assert.Equal(3, len(result.ExecutorData[0]))

		assert.Equal("This is title 1",
			result.ExecutorData[0][0]["json"].(map[string]interface{})["title"])

		assert.Equal("This is title 2",
			result.ExecutorData[0][1]["json"].(map[string]interface{})["title"])

		assert.Equal("This is a paragraph 1",
			result.ExecutorData[0][2]["json"].(map[string]interface{})["first"])

		assert.Equal(
			[]string{
				"This is a paragraph 1",
				"This is a paragraph 3",
				"Total paragraph 3",
			},
			result.ExecutorData[0][2]["json"].(map[string]interface{})["list"])
	})

	s.T().Run("Test Html Workflow Create and Execute", func(t *testing.T) {
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/html-e2e.json")
		assert.NotNil(newWorkflow)
		assert.Nil(err)

		// Manual run workflow
		executionId, err := api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotEmpty(executionId)

		// Get execution
		execution, err := api.GetWorkflowExecution_Testing(testFiberLambda, organization.ID, executionId)
		assert.Nil(err)
		assert.NotNil(execution)

		assert.Equal(structs.WorkflowExecutionStatus_Success, execution.Status)
		assert.Equal(structs.WorkflowExecutionMode_Manual, execution.Mode)
		assert.NotNil(execution.Data)
		assert.NotNil(execution.Data.ResultData)
		assert.Len(execution.Data.ResultData.RunData, 5)

		assert.Nil(err)
		resHtml := execution.Data.ResultData.RunData["HTML1"][0].Data["main"][0][0]
		resEmail := execution.Data.ResultData.RunData["Suger Email"][0].Data["main"][0][0]

		jsonMap := resHtml["json"].(map[string]interface{})
		assert.Equal("CANCEL ENTITLEMENT", jsonMap["name"])
		assert.Equal([]interface{}{
			"Entitlement Details",
			"Channel AWS",
			"Product Teleport Access Platform for AWS",
			"Offer Gravitational, Inc. private offer for Zapier Inc (acct: 996097627176) d1b90d58-c2d9-42ce-9a4a-60a522d70c4c",
			"Buyer Zapier Inc.",
			"Entitlement ID QrtyTXyaM",
			"External ID agmt-4wj0hqzrdyqplohnjen30eq2i",
			"Create Date 2023-07-30",
			"Start Date 2023-03-17",
			"End Date 2024-03-17",
		}, jsonMap["fields"])

		assert.NotEmpty(resEmail["json"].(map[string]interface{})["toEmail"])

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)
	})
}
