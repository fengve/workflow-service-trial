package google_bigquery

import (
	"context"
	"fmt"
	"strings"

	sharedBigquery "github.com/sugerio/workflow-service-trial/integration/bigquery"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

// Search projets of current account
func searchProjects(
	ctx context.Context,
	sugerOrgId string,
	nodeParameters map[string]interface{},
	filter string,
	paginationToken string) (*structs.GetDynamicNodeParametersResponse_ResourceLocatorResults, error) {
	response := &structs.GetDynamicNodeParametersResponse_ResourceLocatorResults{}
	// call bigquery api
	integration, err := sharedBigquery.GetBigqueryIntegration(sugerOrgId, core.GetRdsDbQueries(), core.GetAwsSdkClients())
	if err != nil {
		return nil, err
	}
	projects, err := sharedBigquery.ListProjects(integration)
	if err != nil {
		return response, err
	}
	items := []structs.WorkflowNodeListSearchItem{}
	for _, project := range projects.Projects {
		valid := filter == "" || strings.Contains(project.FriendlyName, filter) || strings.Contains(project.ID, filter)
		if valid {
			items = append(items, structs.WorkflowNodeListSearchItem{
				Name:  project.FriendlyName,
				Value: project.ID,
			})
		}
	}
	response.Data = &structs.WorkflowNodeListSearchResult{
		Results: items,
	}
	return response, nil
}

// Search Datasets of the project
func searchDatasets(
	ctx context.Context,
	sugerOrgId string,
	nodeParameters map[string]interface{},
	filter string,
	paginationToken string) (*structs.GetDynamicNodeParametersResponse_ResourceLocatorResults, error) {
	response := &structs.GetDynamicNodeParametersResponse_ResourceLocatorResults{}
	projectId := getSourceId("projectId", nodeParameters)
	if projectId == "" {
		return nil, fmt.Errorf("projectId must not be null")
	}
	// create bigquery client
	integration, err := sharedBigquery.GetBigqueryIntegration(sugerOrgId, core.GetRdsDbQueries(), core.GetAwsSdkClients())
	if err != nil {
		return nil, err
	}
	client, err := sharedBigquery.NewBigqueryClient(integration, projectId)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	// call bigquery api
	items := []structs.WorkflowNodeListSearchItem{}
	datasetIterator := client.Datasets(ctx)
	for {
		dataset, err := datasetIterator.Next()
		if err != nil {
			break
		}
		valid := filter == "" || strings.Contains(dataset.DatasetID, filter)
		if valid {
			items = append(items, structs.WorkflowNodeListSearchItem{
				Name:  dataset.DatasetID,
				Value: dataset.DatasetID,
			})
		}

	}
	response.Data = &structs.WorkflowNodeListSearchResult{
		Results: items,
	}
	return response, nil
}

// Search Tables of the dataset
func searchTables(
	ctx context.Context,
	sugerOrgId string,
	nodeParameters map[string]interface{},
	filter string,
	paginationToken string) (*structs.GetDynamicNodeParametersResponse_ResourceLocatorResults, error) {
	response := &structs.GetDynamicNodeParametersResponse_ResourceLocatorResults{}
	projectId := getSourceId("projectId", nodeParameters)
	if projectId == "" {
		return nil, fmt.Errorf("projectId must not be null")
	}
	datasetId := getSourceId("datasetId", nodeParameters)
	if datasetId == "" {
		return nil, fmt.Errorf("datasetId must not be null")
	}
	// create bigquery client
	integration, err := sharedBigquery.GetBigqueryIntegration(sugerOrgId, core.GetRdsDbQueries(), core.GetAwsSdkClients())
	if err != nil {
		return nil, err
	}
	client, err := sharedBigquery.NewBigqueryClient(integration, projectId)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	// call bigquery api
	items := []structs.WorkflowNodeListSearchItem{}
	tableIterator := client.Dataset(datasetId).Tables(ctx)
	for {
		table, err := tableIterator.Next()
		if err != nil {
			break
		}
		valid := filter == "" || strings.Contains(table.TableID, filter)
		if valid {
			items = append(items, structs.WorkflowNodeListSearchItem{
				Name:  table.TableID,
				Value: table.TableID,
			})
		}
	}
	response.Data = &structs.WorkflowNodeListSearchResult{
		Results: items,
	}
	return response, nil
}

// Get Datasets select options
func getDatasets(
	ctx context.Context,
	sugerOrgId string, nodeParameters map[string]interface{}) (*structs.GetDynamicNodeParametersResponse_Options, error) {
	response := &structs.GetDynamicNodeParametersResponse_Options{}
	projectId := getSourceId("projectId", nodeParameters)
	if projectId == "" {
		return nil, fmt.Errorf("projectId must not be null")
	}
	// create bigquery client
	integration, err := sharedBigquery.GetBigqueryIntegration(sugerOrgId, core.GetRdsDbQueries(), core.GetAwsSdkClients())
	if err != nil {
		return nil, err
	}
	client, err := sharedBigquery.NewBigqueryClient(integration, projectId)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	// call bigquery api
	options := []structs.WorkflowNodePropertyOptions{}
	datasetIterator := client.Datasets(ctx)
	for {
		dataset, err := datasetIterator.Next()
		if err != nil {
			break
		}
		options = append(options, structs.WorkflowNodePropertyOptions{
			Name:  dataset.DatasetID,
			Value: dataset.DatasetID,
		})

	}
	response.Data = options
	return response, nil
}

// Get table schema
func getSchema(
	ctx context.Context,
	sugerOrgId string,
	nodeParameters map[string]interface{}) (*structs.GetDynamicNodeParametersResponse_Options, error) {
	response := &structs.GetDynamicNodeParametersResponse_Options{}
	projectId := getSourceId("projectId", nodeParameters)
	if projectId == "" {
		return nil, fmt.Errorf("projectId must not be null")
	}
	datasetId := getSourceId("datasetId", nodeParameters)
	if datasetId == "" {
		return nil, fmt.Errorf("datasetId must not be null")
	}
	tableId := getSourceId("tableId", nodeParameters)
	if tableId == "" {
		return nil, fmt.Errorf("tableId must not be null")
	}

	// create bigquery client
	integration, err := sharedBigquery.GetBigqueryIntegration(sugerOrgId, core.GetRdsDbQueries(), core.GetAwsSdkClients())
	if err != nil {
		return nil, err
	}
	client, err := sharedBigquery.NewBigqueryClient(integration, projectId)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	// call bigquery api
	md, err := client.Dataset(datasetId).Table(tableId).Metadata(ctx)
	if err != nil {
		return nil, err
	}

	options := []structs.WorkflowNodePropertyOptions{}
	for _, field := range md.Schema {
		options = append(options, structs.WorkflowNodePropertyOptions{
			Name:        field.Name,
			Value:       field.Name,
			Description: fmt.Sprintf("type: %s required: %t", field.Type, field.Required),
		})
	}
	response.Data = options
	return response, nil
}

// This func is just for test. GoogleBigquery Node does not have this method.
// It is a method of GoogleSheets Node actually.
func getMappingColumns(
	ctx context.Context,
	sugerOrgId string,
	nodeParameters map[string]interface{}) (*structs.GetDynamicNodeParametersResponse_ResourceMapperFields, error) {
	response := &structs.GetDynamicNodeParametersResponse_ResourceMapperFields{}
	// call google sheets api

	// use fake data
	fields := []structs.WorkflowResourceMapperField{
		{
			ID:          "Name",
			DisplayName: "Name",
			Type:        "string",
		},
		{
			ID:          "Count",
			DisplayName: "Count",
			Type:        "string",
		},
		{
			ID:          "Num",
			DisplayName: "Num",
			Type:        "string",
		},
	}
	response.Data = &structs.WorkflowResourceMapperFields{
		Fields: fields,
	}
	return response, nil
}
