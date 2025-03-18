package html

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (
	// Category is the category of HtmlNode.
	Category = structs.CategoryExecutor

	// Name is the name of HtmlNode.
	Name = "n8n-nodes-base.html"
)

const (
	// CodeNodeModeParameter = "mode"
	// CodeParameterNameJs   = "jsCode"
	// CodeLanguageJs        = "javaScript"
	OperationGenerateHtmlTemplate = "generateHtmlTemplate"
	OperationExtractHtmlContent   = "extractHtmlContent"
	OperationConvertToHtmlTable   = "convertToHtmlTable" // TODO
)

var (
	//go:embed node.json
	rawJson []byte

	//go:embed html.svg
	rawIcon []byte
)

type (
	HtmlExecutor struct {
		spec *structs.WorkflowNodeSpec
	}

	ParameterExtractionValueOptions struct {
		Key         string `json:"key"`
		CssSelector string `json:"cssSelector"`
		ReturnValue string `json:returnValue` // attribute, text, html, value
		ReturnArray bool   `json:returnArray`

		Attribute string `json:attribute`
	}

	ParameterOptions struct {
		TrimValues  bool `json:trimValues`
		CleanUpText bool `json:cleanUpText`
	}
)

func toExtractionValueOptions(raw map[string]interface{}) (*ParameterExtractionValueOptions, error) {
	options, err := core.ConvertInterfaceToType[ParameterExtractionValueOptions](raw)
	if err != nil {
		return nil, err
	}

	if options != nil && options.ReturnValue == "" {
		options.ReturnValue = "text"
	}
	return options, nil
}

func init() {
	he := &HtmlExecutor{
		spec: &structs.WorkflowNodeSpec{},
	}
	he.spec.JsonConfig = rawJson
	he.spec.GenerateSpec()

	core.Register(he)
	core.RegisterEmbedIcons(he.spec.Name(), rawIcon)
}

func (he *HtmlExecutor) Category() structs.NodeObjectCategory {
	return Category
}

func (he *HtmlExecutor) Name() string {
	return Name
}

func (he *HtmlExecutor) DefaultSpec() interface{} {
	return he.spec
}

func (he *HtmlExecutor) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	items := core.GetInputData(input.Data)

	operation, err := core.GetNodeParameterAsBasicType(Name, "operation", "", input, 0)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}
	if operation == "" {
		return core.GenerateEmptyResponse()
	}

	if operation == OperationConvertToHtmlTable && len(items) > 0 {
		// --------------------- Convert to HTML Table ---------------------
		// TODO
		// table := ""

		// optionsRaw, _ := core.GetNodeParameter(Name, "options", nil, input, 0)

		// // TODO toOptions (map[string]interface{})

		// tableStyle := ""
		// headerStyle := ""
		// cellStyle := ""

		// if !options["customStyling"].(bool) {
		// 	tableStyle = "style='border-spacing:0; font-family:helvetica,arial,sans-serif'"
		// 	headerStyle = "style='margin:0; padding:7px 20px 7px 0px; border-bottom:1px solid #eee; text-align:left; color:#888; font-weight:normal'"
		// 	cellStyle = "style='margin:0; padding:7px 20px 7px 0px; border-bottom:1px solid #eee'"
		// }

		// tableAttributes := options["tableAttributes"].(string)
		// headerAttributes := options["headerAttributes"].(string)

		// itemsData := make([]map[string]interface{}, 0)
		// itemsKeys := make(map[string]bool)

		// for _, entry := range items {
		// 	entryData := entry.(map[string]interface{})
		// 	itemsData = append(itemsData, entryData["json"].(map[string]interface{}))

		// 	for key := range entryData["json"].(map[string]interface{}) {
		// 		itemsKeys[key] = true
		// 	}
		// }

		// headers := make([]string, 0, len(itemsKeys))
		// for key := range itemsKeys {
		// 	headers = append(headers, key)
		// }

		// table += "<table " + tableStyle + " " + tableAttributes + ">"

		// if options["caption"] != nil {
		// 	table += "<caption>" + options["caption"].(string) + "</caption>"
		// }

		// table += "<thead " + headerStyle + " " + headerAttributes + ">"
		// table += "<tr>"
		// for _, header := range headers {
		// 	table += "<th>" + capitalizeHeader(header, options["capitalize"].(bool)) + "</th>"
		// }
		// table += "</tr>"
		// table += "</thead>"

		// table += "<tbody>"
		// for entryIndex, entry := range itemsData {
		// 	rowsAttributes := getNodeParameter(input, "options.rowsAttributes", entryIndex, "").(string)

		// 	table += "<tr " + rowsAttributes + ">"

		// 	cellsAttributes := getNodeParameter(input, "options.cellAttributes", entryIndex, "").(string)

		// 	for _, header := range headers {
		// 		td := "<td " + cellStyle + " " + cellsAttributes + ">"

		// 		if value, ok := entry[header].(bool); ok {
		// 			isChecked := ""
		// 			if value {
		// 				isChecked = "checked=\"checked\""
		// 			}
		// 			td += "<input type=\"checkbox\" " + isChecked + "/>"
		// 		} else {
		// 			td += entry[header].(string)
		// 		}
		// 		td += "</td>"
		// 		table += td
		// 	}
		// 	table += "</tr>"
		// }

		// table += "</tbody>"
		// table += "</table>"

		// return []*structs.NodeExecutionResult{
		// 	{
		// 		Json: map[string]interface{}{
		// 			"table": table,
		// 		},
		// 		PairedItem: func() []map[string]interface{} {
		// 			pairedItems := make([]map[string]interface{}, len(items))
		// 			for index := range items {
		// 				pairedItems[index] = map[string]interface{}{
		// 					"item": index,
		// 				}
		// 			}
		// 			return pairedItems
		// 		}(),
		// 	},
		// }
	}

	returnData := structs.NodeData{}

	sbc := core.SandboxContext{
		Items:     items,
		Params:    input.Params.Parameters,
		Functions: core.BuiltInFunctions,
	}
	eval := core.NewExpressionEvaluator(&sbc)

ItemLoop:
	for itemIndex, item := range items {

		if operation == OperationGenerateHtmlTemplate {
			// --------------------- Generate HTML Template ---------------------
			html, err := core.GetNodeParameter(Name, "html", "", input, itemIndex)
			if err != nil {
				if core.ContinueOnFail(input.Params) {
					returnData = append(returnData, core.NewNodeSingleDataError(err, itemIndex))
					continue ItemLoop
				}
				return core.GenerateFailedResponse(Name, err)
			}

			htmlResult, err := eval.EvaluateExpression("="+html.(string), itemIndex)

			if err != nil {
				if core.ContinueOnFail(input.Params) {
					returnData = append(returnData, core.NewNodeSingleDataError(err, itemIndex))
					continue ItemLoop
				}
				return core.GenerateFailedResponse(Name, err)
			}

			newItem := structs.NodeSingleData(
				map[string]interface{}{
					"json": map[string]interface{}{
						"html": htmlResult.(string),
					},
				},
			)

			returnData = append(returnData, newItem)
		} else if operation == OperationExtractHtmlContent {
			// --------------------- Extract HTML Content ---------------------
			sourceData, err := core.GetNodeParameter(Name, "sourceData", "", input, itemIndex)
			if err != nil {
				return core.GenerateFailedResponse(Name, err)
			}

			dataPropertyName, err := core.GetNodeParameter(Name, "dataPropertyName", "", input, itemIndex)
			if err != nil {
				return core.GenerateFailedResponse(Name, err)
			}
			dataPropertyPath, ok := dataPropertyName.(string)
			if !ok && core.IsArray(dataPropertyName) {

				if dataPropertyNameArr, ok := dataPropertyName.([]interface{}); ok {
					strArr := make([]string, len(dataPropertyNameArr))
					for i, v := range dataPropertyNameArr {
						strArr[i] = v.(string)
					}
					dataPropertyPath = strings.Join(strArr, ".")
				} else {
					return core.GenerateFailedResponse(Name, fmt.Errorf("dataPropertyName is not a string or array of strings [item %d]", itemIndex))
				}
			}

			extractionValuesRaw, err := core.GetNodeParameter(Name, "extractionValues", nil, input, itemIndex)
			if err != nil {
				return core.GenerateFailedResponse(Name, err)
			}
			extractionValues, ok := extractionValuesRaw.(map[string]interface{})
			if !ok {
				return core.GenerateFailedResponse(Name, fmt.Errorf("extractionValues is invalid"))
			}

			options, err := core.GetNodeParameterAsType(Name, "options", ParameterOptions{
				TrimValues: true,
			}, input, itemIndex)
			if err != nil {
				return core.GenerateFailedResponse(Name, err)
			}

			var htmlArray []string
			if sourceData == "json" {

				json := item["json"].(map[string]interface{})
				value, ok := core.GetMapValueByPath(json, dataPropertyPath)
				if !ok || value == nil {
					return core.GenerateFailedResponse(Name, fmt.Errorf("no property named \"%s\" exists [item %d]", dataPropertyName, itemIndex))
				}
				html, ok := value.(string)
				if ok {
					// TODO may be json string of array
					htmlArray = []string{html}
				} else if core.IsArray(value) {
					htmlArray, ok = value.([]string)
				}
				if !ok {
					return core.GenerateFailedResponse(Name, fmt.Errorf("property \"%s\" is not a string or array of strings [item %d]", dataPropertyName, itemIndex))
				}

			} else {
				// TODO binary
				// helpers.AssertBinaryData(itemIndex, dataPropertyName)
				// binaryDataBuffer := helpers.GetBinaryDataBuffer(itemIndex, dataPropertyName)
				// htmlArray = append(htmlArray, string(binaryDataBuffer))
			}

			if len(htmlArray) == 0 {
				continue
			}

			for index, html := range htmlArray {
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
				if err != nil {
					if core.ContinueOnFail(input.Params) {
						returnData = append(returnData, core.NewNodeSingleDataError(err, itemIndex, map[string]interface{}{
							"htmlIndex": index,
						}))
						continue ItemLoop
					}
					return core.GenerateFailedResponse(Name, err)
				}

				newItem := structs.NodeSingleData{
					"json": map[string]interface{}{},
				}

				valueDataArray, ok := extractionValues["values"].([]interface{})
				if !ok {
					return core.GenerateFailedResponse(Name, fmt.Errorf("extractionValues.values is not an array [item %d]", itemIndex))
				}

				for _, valueData := range valueDataArray {

					valueDataOptions, err := toExtractionValueOptions(valueData.(map[string]interface{}))
					if err != nil {
						return core.GenerateFailedResponse(Name, err)
					}
					jsonMap := newItem["json"].(map[string]interface{})

					htmlElement := doc.Find(valueDataOptions.CssSelector)
					key := valueDataOptions.Key

					if valueDataOptions.ReturnArray {
						values := make([]string, 0)
						htmlElement.Each(func(_ int, el *goquery.Selection) {
							val, err := GetValue(el, valueDataOptions, options)
							if err != nil {
								values = append(values, fmt.Sprintf("Error: %v", err))
							}
							values = append(values, val)
						})
						jsonMap[key] = values
					} else {
						val, err := GetValue(htmlElement, valueDataOptions, options)
						if err != nil {
							if core.ContinueOnFail(input.Params) {
								jsonMap[key] = fmt.Sprintf("Error: %v", err)
							} else {
								return core.GenerateFailedResponse(Name, err)
							}
						}
						jsonMap[key] = val
					}
				}

				returnData = append(returnData, newItem)
			}
		}

	}

	return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{returnData})

}
