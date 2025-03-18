package google_bigquery

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"cloud.google.com/go/bigquery"
	sharedBigquery "github.com/sugerio/workflow-service-trial/integration/bigquery"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	"google.golang.org/api/iterator"
)

const (
	// Category is the category of ManualTriggerNode.
	Category = structs.CategoryExecutor

	// Name is the name of ManualTriggerNode.
	Name = "n8n-nodes-base.googleBigQuery"
)

var (
	//go:embed node.json
	rawJson []byte

	//go:embed googleBigQuery.svg
	icon []byte
)

type (
	GoogleBigQueryExecutor struct {
		spec *structs.WorkflowNodeSpec
	}

	SourceParam struct {
		Mode  string `json:"mode"`
		Value string `json:"value"`
	}

	ExecuteQueryOptions struct {
		DefaultDataset     string `json:"defaultDataset,omitempty"`
		DryRun             bool   `json:"dryRun,omitempty"`
		RawOutput          bool   `json:"rawOutput,omitempty"`
		IncludeSchema      bool   `json:"includeSchema,omitempty"`
		Location           string `json:"location,omitempty"`
		MaximumBytesBilled string `json:"maximumBytesBilled,omitempty"`
		MaxResults         int    `json:"maxResults,omitempty"`
		TimeoutMs          int    `json:"timeoutMs,omitempty"`
		UseLegacySql       bool   `json:"useLegacySql,omitempty"`
	}

	InsertOptions struct {
		BatchSize           int    `json:"batchSize,omitempty"`
		IgnoreUnknownValues bool   `json:"ignoreUnknownValues,omitempty"`
		SkipInvalidRows     bool   `json:"skipInvalidRows,omitempty"`
		TemplateSuffix      string `json:"templateSuffix,omitempty"`
		TraceId             string `json:"traceId,omitempty"`
	}

	InsertField struct {
		FieldId    string      `json:"fieldId,omitempty"`
		FieldValue interface{} `json:"fieldValue,omitempty"`
	}

	InsertDataSaver struct {
		data map[string]interface{}
	}
)

func init() {
	gbq := &GoogleBigQueryExecutor{
		spec: &structs.WorkflowNodeSpec{},
	}
	gbq.spec.JsonConfig = rawJson
	gbq.spec.GenerateSpec()

	core.Register(gbq)
	core.RegisterEmbedIcons(Name, icon)
}

func (gbq *GoogleBigQueryExecutor) Category() structs.NodeObjectCategory {
	return Category
}

func (gbq *GoogleBigQueryExecutor) Name() string {
	return Name
}

func (gbq *GoogleBigQueryExecutor) DefaultSpec() interface{} {
	return gbq.spec
}

func (gbq *GoogleBigQueryExecutor) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	// Create BigqueryClient with projectId
	sugerOrgId := input.Params.SugerOrgId

	projectId, err := core.GetNodeParameterAsBasicType(Name, "projectId.value", "",
		input, 0)
	if projectId == "" || err != nil {
		return core.GenerateFailedResponse(Name, fmt.Errorf("projectId must not be null"))
	}

	integration, err := sharedBigquery.GetBigqueryIntegration(sugerOrgId, core.GetRdsDbQueries(), core.GetAwsSdkClients())
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}
	client, err := sharedBigquery.NewBigqueryClient(integration, projectId)

	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}
	defer client.Close()

	operation, err := core.GetNodeParameterAsBasicType(Name, "operation", "executeQuery",
		input, 0)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}

	switch operation {
	case "executeQuery":
		return executeQuery(ctx, client, input, projectId)
	case "insert":
		return insert(ctx, client, input)
	}

	return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{})
}

func (gbq *GoogleBigQueryExecutor) Methods() *structs.NodeMethods {
	methods := &structs.NodeMethods{}
	methods.ListSearch = map[string]structs.NodeMethodListSearch{
		"searchProjects": searchProjects,
		"searchDatasets": searchDatasets,
		"searchTables":   searchTables,
	}
	methods.LoadOptions = map[string]structs.NodeMethodLoadOptions{
		"getDatasets": getDatasets,
		"getSchema":   getSchema,
	}
	methods.ResourceMapping = map[string]structs.NodeMethodResourceMapping{
		"getMappingColumns": getMappingColumns,
	}
	return methods
}

func executeQuery(
	ctx context.Context,
	client *bigquery.Client, input *structs.NodeExecuteInput, projectId string) *structs.NodeExecutionResult {
	result := structs.NodeData{}
	items := core.GetInputData(input.Data)

	for index, _ := range items {
		sqlQuery, err := core.GetNodeParameterAsBasicType(Name, "sqlQuery", "",
			input, index)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}

		options, err := core.GetNodeParameterAsType(Name, "options", ExecuteQueryOptions{},
			input, index)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}

		singleResult, err := handelSingleExecuteQuery(ctx, client, projectId, sqlQuery, options)
		if err != nil {
			if core.ContinueOnFail(input.Params) {
				errorResult := core.NewNodeSingleDataError(err, index)
				result = append(result, errorResult)
				continue
			}
			return core.GenerateFailedResponse(Name, err)
		}
		result = append(result, singleResult[:]...)
	}

	if len(result) == 0 {
		defaultResult := structs.NodeSingleData{
			"json": map[string]interface{}{
				"success": true,
			},
		}
		result = append(result, defaultResult)
	}

	return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
}

// Execute Query and get result
func handelSingleExecuteQuery(
	ctx context.Context,
	client *bigquery.Client, projectId string, sqlQuery string, options *ExecuteQueryOptions) (structs.NodeData, error) {
	query := client.Query(sqlQuery)

	result := structs.NodeData{}

	// DefaultDataset
	if options.DefaultDataset != "" {
		query.QueryConfig.DefaultProjectID = projectId
		query.QueryConfig.DefaultDatasetID = options.DefaultDataset
	}
	// UseLegacySql
	query.QueryConfig.UseLegacySQL = options.UseLegacySql
	// DryRun
	query.QueryConfig.DryRun = options.DryRun
	dryRun := options.DryRun
	// MaximumBytesBilled
	maximumBytesBilled, err := strconv.ParseInt(options.MaximumBytesBilled, 10, 32)
	if err != nil {
		query.QueryConfig.MaxBytesBilled = maximumBytesBilled
	}
	// Location
	location := options.Location
	if location != "" {
		query.Location = location
	}
	// TODO: MaxResults
	// TODO: TimeoutMs

	// rawOutput
	rawOutput := options.RawOutput
	// includeSchema
	includeSchema := options.IncludeSchema

	// Execute the query.
	job, err := query.Run(ctx)
	if err != nil {
		// msg := fmt.Sprintf("failed to execute query. item index: %d sql: %s. %s", index, sqlQuery, err.Error())
		return nil, err
	}

	// Read execute result
	jobId := job.ID()
	it, err := job.Read(ctx)
	if err != nil {
		return nil, err
	}
	// Iterate through the results.
	rowData := rawOutput || dryRun || false
	fmt.Printf("query jobId %s totalRows %d\n", jobId, it.TotalRows)
	for {
		schame := it.Schema
		var values []bigquery.Value
		err := it.Next(&values)
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println("read query result by iterator error", err)
			continue
		}
		res, err := prepareQueryOutput(values, schame, rowData, includeSchema)
		if err != nil {
			fmt.Println("prepare query output error", err)
			continue
		}
		result = append(result, res)
	}

	return result, nil
}

func prepareQueryOutput(
	values []bigquery.Value, schema bigquery.Schema, rawOutput bool, includeSchema bool) (structs.NodeSingleData, error) {
	res := structs.NodeSingleData{}
	json := map[string]interface{}{}

	if rawOutput {
		json["rawData"] = values
		json["success"] = true
	} else {
		if values != nil && schema != nil {
			for i, value := range values {
				fieldSchema := schema[i]
				// TODO: handle fieldSchema.Type == "RECORD"
				json[fieldSchema.Name] = value
			}
		} else {
			json["success"] = true
		}

		if schema != nil && includeSchema {
			json["_schema"] = schema
		}
	}
	res["json"] = json
	return res, nil
}

func insert(
	ctx context.Context,
	client *bigquery.Client, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	var result structs.NodeData

	datasetId, err := core.GetNodeParameterAsBasicType(Name, "datasetId.value", "",
		input, 0)
	if datasetId == "" || err != nil {
		return core.GenerateFailedResponse(Name, fmt.Errorf("datasetId must not be null"))
	}

	tableId, err := core.GetNodeParameterAsBasicType(Name, "tableId.value", "",
		input, 0)
	if tableId == "" || err != nil {
		return core.GenerateFailedResponse(Name, fmt.Errorf("tableId must not be null"))
	}

	// Get table schema
	md, err := client.Dataset(datasetId).Table(tableId).Metadata(ctx)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}
	schema := md.Schema
	if schema == nil {
		return core.GenerateFailedResponse(Name, fmt.Errorf("table %s has no defined schema", tableId))
	}

	dataMode, err := core.GetNodeParameterAsBasicType(Name, "dataMode", "autoMap",
		input, 0)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}
	// Read options
	options, err := core.GetNodeParameterAsType(Name, "options", InsertOptions{},
		input, 0)

	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}

	batchSize := options.BatchSize
	if batchSize == 0 {
		batchSize = 100
	}

	// Prepare insert data
	items := core.GetInputData(input.Data)

	rows := []map[string]interface{}{}
	for itemIndex, item := range items {
		row := map[string]interface{}{}
		if dataMode == "autoMap" {
			json := item["json"].(map[string]interface{})
			for _, field := range schema {
				if val, ok := json[field.Name]; ok {
					row[field.Name] = val
				}
			}
		} else if dataMode == "define" {
			insertFields, err := core.GetNodeParameterAsType(Name, "fieldsUi.values", []InsertField{},
				input, itemIndex)
			if err != nil {
				if core.ContinueOnFail(input.Params) {
					errorResult := core.NewNodeSingleDataError(err, itemIndex)
					result = append(result, errorResult)
					continue
				}
				return core.GenerateFailedResponse(Name, err)
			}
			for _, field := range *insertFields {
				row[field.FieldId] = field.FieldValue
			}
		}
		err := checkInsertDataBySchema(row, schema)
		if err != nil {
			if core.ContinueOnFail(input.Params) {
				errorResult := core.NewNodeSingleDataError(err, itemIndex)
				result = append(result, errorResult)
				continue
			}
			return core.GenerateFailedResponse(Name, err)
		}
		rows = append(rows, row)
	}

	// Batch insert
	ins := client.Dataset(datasetId).Table(tableId).Inserter()
	ins.SkipInvalidRows = options.SkipInvalidRows
	ins.IgnoreUnknownValues = options.IgnoreUnknownValues
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}

		batchRows := rows[i:end]
		insertDataSavers := make([]bigquery.ValueSaver, len(batchRows))
		for idx, mapVal := range batchRows {
			insertDataSavers[idx] = &InsertDataSaver{data: mapVal}
		}

		err := ins.Put(ctx, insertDataSavers)
		if err == nil {
			insertResult := map[string]interface{}{
				"index":   i,
				"success": true,
			}
			result = append(result, insertResult)
			continue
		}

		// Handle insert error
		errOutput := err
		errorMsgs := []string{}
		if putMultiError, ok := err.(bigquery.PutMultiError); ok {
			for _, e := range putMultiError {
				errorMsgs = append(errorMsgs, e.Error())
			}
			errOutput = fmt.Errorf("multi row insert error. %s", strings.Join(errorMsgs, ";"))
		}
		errorResult := map[string]interface{}{
			"index":   i,
			"success": false,
			"error":   errOutput.Error(),
		}

		if core.ContinueOnFail(input.Params) {
			result = append(result, errorResult)
			continue
		}

		return core.GenerateFailedResponse(Name, errOutput)
	}

	return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
}

// Check insert data by schema
func checkInsertDataBySchema(row map[string]interface{}, schema bigquery.Schema) error {
	for _, field := range schema {
		name := field.Name
		if val, ok := row[name]; field.Required && (!ok || val == "undefined") {
			return fmt.Errorf("the property '%s' is required, please define it", name)
		}
		if field.Type != bigquery.StringFieldType && row[name] == "" {
			delete(row, name)
		}
		// TODO: handle RecordFieldType
	}
	return nil
}

// Read sourceId such as projectId,datasetId,tableId
//
//	{
//			"projectId": {
//				"__rl": "true",
//				"mode": "list",
//				"value":"suger-dev"
//			},
//	}
func getSourceId(sourceName string, parameters map[string]interface{}) string {
	raw, ok := parameters[sourceName]
	if raw == nil || !ok {
		return ""
	}
	sourceParam := &SourceParam{}
	data, err := json.Marshal(raw)
	if err != nil {
		return ""
	}
	err = json.Unmarshal(data, sourceParam)
	if err != nil {
		return ""
	}
	return sourceParam.Value
}

func (dsv *InsertDataSaver) Save() (map[string]bigquery.Value, string, error) {
	row := make(map[string]bigquery.Value, len(dsv.data))
	for key, val := range dsv.data {
		row[key] = val
	}
	// TODO use unique insertID
	return row, "", nil
}
