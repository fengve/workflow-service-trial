package api_test

// Command to run this test only
// go test -v service/workflow_service/api/service_test.go service/workflow_service/api/dynamic_parameter_test.go

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/teris-io/shortid"

	sharedBigquery "github.com/sugerio/workflow-service-trial/integration/bigquery"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/code"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type DynamicParameterTestSuit struct {
	suite.Suite
}

func Test_DynamicParameterTestSuit(t *testing.T) {
	suite.Run(t, new(DynamicParameterTestSuit))
}

func (s *DynamicParameterTestSuit) Test() {
	s.T().Run("Test API resource-locator-results", func(t *testing.T) {
		t.Skip("flaky - need to investigate")
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
			"currentNodeParameters[projectId][__rl]":        "true",
			"currentNodeParameters[projectId][mode]":        "list",
			"currentNodeParameters[projectId][value]":       "suger-dev",
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

	s.T().Run("Test API options", func(t *testing.T) {
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

	s.T().Run("Test API resource-mapper-fields", func(t *testing.T) {
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

		// Test Node: Google Sheets
		request := events.APIGatewayProxyRequest{}
		request.HTTPMethod = http.MethodGet
		request.Path = fmt.Sprintf("/workflow/org/%s/workflow/dynamic-node-parameters/resource-mapper-fields", sugerOrgId)
		request.Headers = map[string]string{"Content-Type": "application/json"}
		// This request param has been modified for test, because GoogleSheets Node has not been supported for now.
		// Teal nodeTypeAndVersion is:
		// "nodeTypeAndVersion[name]":                            "n8n-nodes-base.googleSheets",
		// "nodeTypeAndVersion[version]":                         "4.3",
		request.QueryStringParameters = map[string]string{
			"nodeTypeAndVersion[name]":                            "n8n-nodes-base.googleBigQuery",
			"nodeTypeAndVersion[version]":                         "2",
			"currentNodeParameters[resource]":                     "sheet",
			"currentNodeParameters[operation]":                    "append",
			"currentNodeParameters[documentId][__rl]":             "true",
			"currentNodeParameters[documentId][value]":            "1uYNNOAxKkOeBUfcaZIG0vgr8tb7ScC3XcaHiNPdVhrM",
			"currentNodeParameters[documentId][mode]":             "list",
			"currentNodeParameters[documentId][cachedResultName]": "n8n test spreadsheet",
			"currentNodeParameters[documentId][cachedResultUrl]":  "https://docs.google.com/spreadsheets/d/1uYNNOAxKkOeBUfcaZIG0vgr8tb7ScC3XcaHiNPdVhrM/edit?usp=drivesdk",
			"currentNodeParameters[sheetName][__rl]":              "true",
			"currentNodeParameters[sheetName][value]":             "gid=0",
			"currentNodeParameters[sheetName][mode]":              "list",
			"currentNodeParameters[sheetName][cachedResultName]":  "Sheet1",
			"currentNodeParameters[sheetName][cachedResultUrl]":   "https://docs.google.com/spreadsheets/d/1uYNNOAxKkOeBUfcaZIG0vgr8tb7ScC3XcaHiNPdVhrM/edit#gid=0",
			"currentNodeParameters[columns][mappingMode]":         "defineBelow",
			"currentNodeParameters[columns][value]":               "defineBelow",
			"path":                                                "parameters.columns",
			"methodName":                                          "getMappingColumns",
			"sugerOrgId":                                          sugerOrgId,
		}
		request.RequestContext = api.AuthorizerRequestContext

		response, err := testFiberLambda.Proxy(request)
		assert.Nil(err)
		var resp structs.GetDynamicNodeParametersResponse_ResourceMapperFields
		err = json.Unmarshal([]byte(response.Body), &resp)
		assert.Nil(err)
		assert.NotEmpty(resp.Data)
	})
}
