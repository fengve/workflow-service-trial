package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go  service/workflow_service/nodes_test/google_bigquery_test.go

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/teris-io/shortid"

	sharedBigquery "github.com/sugerio/workflow-service-trial/integration/bigquery"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	googlebigquerynode "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/google_bigquery"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type GoogleBigqueryTestSuite struct {
	suite.Suite
}

func Test_GoogleBigqueryTestSuite(t *testing.T) {
	suite.Run(t, new(GoogleBigqueryTestSuite))
}

func (s *GoogleBigqueryTestSuite) TestGoogleBigquery() {

	s.T().Run("Test GoogleBigquery Execute Query", func(t *testing.T) {
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create Bigquery Integration for test
		_, err := sharedBigquery.CreateBigqueryIntegration_Testing(
			organization.ID, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		// Read node param file
		workflowNode := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/google-bigquery-params-executequery.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &workflowNode)
		assert.Nil(err)
		workflowNode.SugerOrgId = organization.ID

		// Check test file project info
		projectInfo, ok := workflowNode.Parameters["projectId"].(map[string]interface{})
		assert.True(ok)
		projecId := projectInfo["value"]
		assert.Equal("suger-stag", projecId)

		bigqueryNode := &googlebigquerynode.GoogleBigQueryExecutor{}
		input := &structs.NodeExecuteInput{
			Params: workflowNode,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"limit": 100,
						},
					},
				},
			},
		}

		result := bigqueryNode.Execute(context.Background(), input)
		assert.NotEmpty(result.ExecutorData)
		assert.Len(result.ExecutorData, 1)
		assert.Len(result.ExecutorData[0], 100)

		// Update the sqlQuery
		workflowNode.Parameters["sqlQuery"] = "=DROP TABLE IF EXISTS suger_stag_bigquery_test.offer_report_{{ new Date().toISOString().slice(0, 10) }}"
		input = &structs.NodeExecuteInput{
			Params: workflowNode,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"limit": 100,
						},
					},
				},
			},
		}
		result = bigqueryNode.Execute(context.Background(), input)
		// No execute result since the table to drop does not exist.
		assert.Nil(result.ExecutorData)
	})

	s.T().Run("Test GoogleBigquery Insert by mode of autoMap", func(t *testing.T) {
		assert := require.New(s.T())
		// Set up short ID generator.
		sid, err := shortid.New(2, shortid.DefaultABC, 2342)
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize short ID generator: %v", err))
		}

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create Bigquery Integration for test
		_, err = sharedBigquery.CreateBigqueryIntegration_Testing(
			organization.ID, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		//Insert mode: autoMap
		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/google-bigquery-params-insert.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)
		np.SugerOrgId = organization.ID

		bigqueryNode := &googlebigquerynode.GoogleBigQueryExecutor{}
		randomInt := rand.Intn(1_000_000)
		input := &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"Name":  "test" + strconv.Itoa(randomInt),
							"Size":  float64(randomInt) + 0.1,
							"Count": 1,
						},
					},
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"Name":  "test" + strconv.Itoa(randomInt),
							"Size":  float64(randomInt) + 0.1,
							"Count": 2,
						},
					},
				},
			},
		}

		result := bigqueryNode.Execute(context.Background(), input)
		assert.NotEmpty(result.ExecutorData)
	})

	s.T().Run("Test GoogleBigquery Insert by mode of define", func(t *testing.T) {
		assert := require.New(s.T())
		// Set up short ID generator.
		sid, err := shortid.New(2, shortid.DefaultABC, 2342)
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize short ID generator: %v", err))
		}

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create Bigquery Integration for test
		_, err = sharedBigquery.CreateBigqueryIntegration_Testing(
			organization.ID, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		//Insert mode: define
		np := &structs.WorkflowNode{}
		testFile, err := os.ReadFile("./test_files/google-bigquery-params-insert2.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &np)
		assert.Nil(err)
		np.SugerOrgId = organization.ID

		bigqueryNode := &googlebigquerynode.GoogleBigQueryExecutor{}
		randomInt := rand.Intn(1_000_000)
		input := &structs.NodeExecuteInput{
			Params: np,
			Data: []structs.NodeData{
				{
					structs.NodeSingleData{
						"json": map[string]interface{}{
							"arg1": "test" + strconv.Itoa(randomInt),
							"arg2": float64(randomInt) + 0.1,
						},
					},
				},
			},
		}
		result := bigqueryNode.Execute(context.Background(), input)
		assert.NotEmpty(result.ExecutorData)
	})

	s.T().Run("Test GoogleBigquery Method of searchProjects", func(t *testing.T) {
		assert := require.New(s.T())
		// Set up short ID generator.
		sid, err := shortid.New(2, shortid.DefaultABC, 2342)
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize short ID generator: %v", err))
		}

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		sugerOrgId := organization.ID

		// Create Bigquery Integration for test
		_, err = sharedBigquery.CreateBigqueryIntegration_Testing(
			sugerOrgId, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		// Test Node: Google Bigquery
		request := events.APIGatewayProxyRequest{}
		request.HTTPMethod = http.MethodGet
		request.Path = fmt.Sprintf("/workflow/org/%s/workflow/dynamic-node-parameters/resource-locator-results", sugerOrgId)
		request.Headers = map[string]string{"Content-Type": "application/json"}
		request.QueryStringParameters = map[string]string{
			"nodeTypeAndVersion[name]":                      "n8n-nodes-base.googleBigQuery",
			"nodeTypeAndVersion[version]":                   "2",
			"path":                                          "parameters.projectId",
			"methodName":                                    "searchProjects",
			"currentNodeParameters[resource]":               "database",
			"currentNodeParameters[operation]":              "executeQuery",
			"currentNodeParameters[options][includeSchema]": "false",
			"sugerOrgId":                                    sugerOrgId,
			"filter":                                        "sta",
		}
		request.RequestContext = api.AuthorizerRequestContext

		response, err := testFiberLambda.Proxy(request)
		assert.Nil(err)
		var resp structs.GetDynamicNodeParametersResponse_ResourceLocatorResults
		err = json.Unmarshal([]byte(response.Body), &resp)
		assert.Nil(err)
		assert.NotEmpty(resp.Data.Results)
	})

	s.T().Run("Test GoogleBigquery Method of searchDatasets", func(t *testing.T) {
		assert := require.New(s.T())
		// Set up short ID generator.
		sid, err := shortid.New(2, shortid.DefaultABC, 2342)
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize short ID generator: %v", err))
		}

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		sugerOrgId := organization.ID

		// Create Bigquery Integration for test
		_, err = sharedBigquery.CreateBigqueryIntegration_Testing(
			sugerOrgId, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		// Test Node: Google Bigquery
		request := events.APIGatewayProxyRequest{}
		request.HTTPMethod = http.MethodGet
		request.Path = fmt.Sprintf("/workflow/org/%s/workflow/dynamic-node-parameters/resource-locator-results", sugerOrgId)
		request.Headers = map[string]string{"Content-Type": "application/json"}
		request.QueryStringParameters = map[string]string{
			"nodeTypeAndVersion[name]":                      "n8n-nodes-base.googleBigQuery",
			"nodeTypeAndVersion[version]":                   "2",
			"path":                                          "parameters.datasetId",
			"methodName":                                    "searchDatasets",
			"currentNodeParameters[resource]":               "database",
			"currentNodeParameters[operation]":              "insert",
			"currentNodeParameters[projectId][__rl]":        "true",
			"currentNodeParameters[projectId][mode]":        "list",
			"currentNodeParameters[projectId][value]":       "suger-stag",
			"currentNodeParameters[sqlQuery]":               "SELECT * FROM `suger-stag.suger_stag_bigquery_test.simple-test-table` LIMIT 100",
			"currentNodeParameters[options][includeSchema]": "false",
			"currentNodeParameters[options][location]":      "US",
			"sugerOrgId":                                    sugerOrgId,
		}
		request.RequestContext = api.AuthorizerRequestContext

		response, err := testFiberLambda.Proxy(request)
		assert.Nil(err)
		var resp structs.GetDynamicNodeParametersResponse_ResourceLocatorResults
		err = json.Unmarshal([]byte(response.Body), &resp)
		assert.Nil(err)
		assert.NotEmpty(resp.Data.Results)
	})

	s.T().Run("Test GoogleBigquery Method of searchTables", func(t *testing.T) {
		assert := require.New(s.T())
		// Set up short ID generator.
		sid, err := shortid.New(2, shortid.DefaultABC, 2342)
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize short ID generator: %v", err))
		}

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		sugerOrgId := organization.ID

		// Create Bigquery Integration for test
		_, err = sharedBigquery.CreateBigqueryIntegration_Testing(
			sugerOrgId, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		// Test Node: Google Bigquery
		request := events.APIGatewayProxyRequest{}
		request.HTTPMethod = http.MethodGet
		request.Path = fmt.Sprintf("/workflow/org/%s/workflow/dynamic-node-parameters/resource-locator-results", sugerOrgId)
		request.Headers = map[string]string{"Content-Type": "application/json"}
		request.QueryStringParameters = map[string]string{
			"nodeTypeAndVersion[name]":                      "n8n-nodes-base.googleBigQuery",
			"nodeTypeAndVersion[version]":                   "2",
			"path":                                          "parameters.tableId",
			"methodName":                                    "searchTables",
			"currentNodeParameters[resource]":               "database",
			"currentNodeParameters[operation]":              "insert",
			"currentNodeParameters[projectId][__rl]":        "true",
			"currentNodeParameters[projectId][mode]":        "list",
			"currentNodeParameters[projectId][value]":       "suger-stag",
			"currentNodeParameters[datasetId][__rl]":        "true",
			"currentNodeParameters[datasetId][mode]":        "list",
			"currentNodeParameters[datasetId][value]":       "suger_stag_bigquery_test",
			"currentNodeParameters[sqlQuery]":               "SELECT * FROM `suger-stag.suger_stag_bigquery_test.simple-test-table` LIMIT 100",
			"currentNodeParameters[options][includeSchema]": "false",
			"currentNodeParameters[options][location]":      "US",
			"sugerOrgId": sugerOrgId,
		}
		request.RequestContext = api.AuthorizerRequestContext

		response, err := testFiberLambda.Proxy(request)
		assert.Nil(err)
		var resp structs.GetDynamicNodeParametersResponse_ResourceLocatorResults
		err = json.Unmarshal([]byte(response.Body), &resp)
		assert.Nil(err)
		assert.NotEmpty(resp.Data.Results)
	})

	s.T().Run("Test GoogleBigquery Method of getDatasets", func(t *testing.T) {
		assert := require.New(s.T())
		// Set up short ID generator.
		sid, err := shortid.New(2, shortid.DefaultABC, 2342)
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize short ID generator: %v", err))
		}

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		sugerOrgId := organization.ID

		// Create Bigquery Integration for test
		_, err = sharedBigquery.CreateBigqueryIntegration_Testing(
			sugerOrgId, rdsDbQueries, awsSdkClients)
		assert.Nil(err)
		// Test Node: Google Bigquery
		request := events.APIGatewayProxyRequest{}
		request.HTTPMethod = http.MethodGet
		request.Path = fmt.Sprintf("/workflow/org/%s/workflow/dynamic-node-parameters/options", sugerOrgId)
		request.Headers = map[string]string{"Content-Type": "application/json"}
		request.QueryStringParameters = map[string]string{
			"nodeTypeAndVersion[name]":                       "n8n-nodes-base.googleBigQuery",
			"nodeTypeAndVersion[version]":                    "2",
			"path":                                           "parameters.options.defaultDataset",
			"methodName":                                     "getDatasets",
			"currentNodeParameters[resource]":                "database",
			"currentNodeParameters[operation]":               "executeQuery",
			"currentNodeParameters[projectId][__rl]":         "true",
			"currentNodeParameters[projectId][mode]":         "list",
			"currentNodeParameters[projectId][value]":        "suger-stag",
			"currentNodeParameters[options][defaultDataset]": "",
			"sugerOrgId":                                     sugerOrgId,
		}
		request.RequestContext = api.AuthorizerRequestContext

		response, err := testFiberLambda.Proxy(request)
		assert.Nil(err)
		var resp structs.GetDynamicNodeParametersResponse_Options
		err = json.Unmarshal([]byte(response.Body), &resp)
		assert.Nil(err)
		assert.NotEmpty(resp.Data)
	})

	s.T().Run("Test GoogleBigquery Method of getSchema", func(t *testing.T) {
		assert := require.New(s.T())
		// Set up short ID generator.
		sid, err := shortid.New(2, shortid.DefaultABC, 2342)
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize short ID generator: %v", err))
		}

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		sugerOrgId := organization.ID

		// Create Bigquery Integration for test
		_, err = sharedBigquery.CreateBigqueryIntegration_Testing(
			sugerOrgId, rdsDbQueries, awsSdkClients)
		assert.Nil(err)
		// Test Node: Google Bigquery
		request := events.APIGatewayProxyRequest{}
		request.HTTPMethod = http.MethodGet
		request.Path = fmt.Sprintf("/workflow/org/%s/workflow/dynamic-node-parameters/options", sugerOrgId)
		request.Headers = map[string]string{"Content-Type": "application/json"}
		request.QueryStringParameters = map[string]string{
			"nodeTypeAndVersion[name]":                               "n8n-nodes-base.googleBigQuery",
			"nodeTypeAndVersion[version]":                            "2",
			"path":                                                   "parameters.fieldsUi.values[0].fieldId",
			"methodName":                                             "getSchema",
			"currentNodeParameters[resource]":                        "database",
			"currentNodeParameters[operation]":                       "insert",
			"currentNodeParameters[projectId][__rl]":                 "true",
			"currentNodeParameters[projectId][mode]":                 "list",
			"currentNodeParameters[projectId][value]":                "suger-stag",
			"currentNodeParameters[datasetId][__rl]":                 "true",
			"currentNodeParameters[datasetId][value]":                "suger_stag_bigquery_test",
			"currentNodeParameters[datasetId][mode]":                 "id",
			"currentNodeParameters[tableId][__rl]":                   "true",
			"currentNodeParameters[tableId][value]":                  "simple-test-table",
			"currentNodeParameters[tableId][mode]":                   "id",
			"currentNodeParameters[dataMode]":                        "define",
			"currentNodeParameters[fieldsUi][values][0][fieldId]":    "",
			"currentNodeParameters[fieldsUi][values][0][fieldValue]": "",
			"currentNodeParameters[options][batchSize]":              "100",
			"currentNodeParameters[options][skipInvalidRows]":        "true",
			"sugerOrgId": sugerOrgId,
		}
		request.RequestContext = api.AuthorizerRequestContext

		response, err := testFiberLambda.Proxy(request)
		assert.Nil(err)
		var resp structs.GetDynamicNodeParametersResponse_Options
		err = json.Unmarshal([]byte(response.Body), &resp)
		assert.Nil(err)
		assert.NotEmpty(resp.Data)
	})

	s.T().Run("Test GoogleBigquery Workflow Create and Execute", func(t *testing.T) {
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create Bigquery Integration for test
		_, err := sharedBigquery.CreateBigqueryIntegration_Testing(
			organization.ID, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		// Create workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/google-bigquery-create-workflow.json")
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

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)
	})
}
