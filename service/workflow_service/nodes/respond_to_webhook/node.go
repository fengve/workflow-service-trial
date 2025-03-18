package respond_to_webhook

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	"github.com/valyala/fasthttp"
)

const (
	// Category is the category of WebhookNode.
	Category = structs.CategoryExecutor

	// Name is the name of RespondToWebhookNode.
	Name = "n8n-nodes-base.respondToWebhook"
)

var (
	//go:embed node.json
	rawJson []byte

	//go:embed webhook.svg
	icon []byte
)

type RespondToWebhook struct {
	spec *structs.WorkflowNodeSpec
}

func init() {
	wh := &RespondToWebhook{
		spec: &structs.WorkflowNodeSpec{},
	}
	wh.spec.JsonConfig = rawJson
	wh.spec.GenerateSpec()

	core.Register(wh)
	core.RegisterEmbedIcons(Name, icon)
}

func (wh *RespondToWebhook) Category() structs.NodeObjectCategory {
	return Category
}

func (wh *RespondToWebhook) Name() string {
	return Name
}

func (wh *RespondToWebhook) DefaultSpec() interface{} {
	return wh.spec
}

func (wh *RespondToWebhook) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	node := input.Params
	if node == nil {
		return core.GenerateFailedResponse(Name, errors.New("node not found"))
	}

	respondWith, err := getRespondWith(node)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}

	responseBytes := make([]byte, 0)
	options, err := getResponseOptions(node)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}

	responseCode := 0
	responseKey := ""
	headers := make(map[string]string)
	if options != nil {
		if options.ResponseCode != 0 {
			responseCode = options.ResponseCode
		}
		if options.ResponseKey != "" {
			responseKey = options.ResponseKey
		}
		if options.ResponseHeaders.Entries != nil {
			for _, entry := range options.ResponseHeaders.Entries {
				headers[entry.Name] = entry.Value
			}
		}
	}

	inputData := core.GetInputData(input.Data)
	switch respondWith {
	case structs.WebhookRespondWith_FirstIncomingItem: // json field of first incoming item
		if len(inputData) == 0 {
			return core.GenerateFailedResponse(Name, errors.New("empty incoming items"))
		}
		jsonField, ok := inputData[0]["json"]
		if !ok {
			return core.GenerateFailedResponse(Name, errors.New("json field of first incoming item not found"))
		}
		responseBytes, err = json.Marshal(wrapResponseData(responseKey, jsonField))
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		setContentType(headers, "application/json; charset=utf-8")
	case structs.WebhookRespondWith_AllIncomingItems:
		responseBytes, err = json.Marshal(wrapResponseData(responseKey, inputData))
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		setContentType(headers, "application/json; charset=utf-8")
	case structs.WebhookRespondWith_Json:
		// In n8n, it may be an object that needs marshaling
		presetBody, err := core.GetNodeParameterAsBasicType(Name, "responseBody", "", input, 0)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		responseBytes = []byte(presetBody)
		setContentType(headers, "application/json; charset=utf-8")
	case structs.WebhookRespondWith_Text:
		presetBody, err := core.GetNodeParameterAsBasicType(Name, "responseBody", "", input, 0)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		responseBytes = []byte(presetBody)
		// n8n did not set content type, and it automatically used text/html
		setContentType(headers, "text/html; charset=utf-8")
	case structs.WebhookRespondWith_Redirect:
		if responseCode == 0 {
			responseCode = 307
		}
		url, err := core.GetNodeParameterAsBasicType(Name, "redirectURL", "", input, 0)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		headers["location"] = url
	case structs.WebhookRespondWith_NoData:
		// do nothing
	// TODO binary type
	default:
		return core.GenerateFailedResponse(Name, errors.New("Unknown value of respondWith parameter: "+string(respondWith)))
	}

	response := &fasthttp.Response{}
	response.SetBody(responseBytes)
	if responseCode == 0 {
		responseCode = 200
	}
	response.SetStatusCode(responseCode)
	for key, value := range headers {
		response.Header.Set(key, value)
	}
	input.AdditionalData.Hooks.ExecutionHookFunctionsSendResponse(ctx, response)
	// Pass through input
	return core.GenerateSuccessResponse(input.Data[0], []structs.NodeData{input.Data[0]})
}

func wrapResponseData(responseKey string, data interface{}) interface{} {
	if len(responseKey) > 0 {
		data = map[string]interface{}{
			responseKey: data,
		}
	}
	return data
}

func getRespondWith(node *structs.WorkflowNode) (structs.WebhookRespondWith, error) {
	if node == nil {
		return "", errors.New("node is nil")
	}
	value := node.Parameters["respondWith"]
	if value == nil {
		// If not specified, default to firstIncomingItem
		return structs.WebhookRespondWith_FirstIncomingItem, nil
	} else if str, ok := value.(string); ok {
		return structs.WebhookRespondWith(str), nil
	} else {
		return "", errors.New("respondWith must be a string")
	}
}

type ResponseOptions struct {
	ResponseCode    int                   `json:"responseCode"`
	ResponseKey     string                `json:"responseKey"`
	ResponseHeaders ResponseHeadersOption `json:"responseHeaders"`
}

type ResponseHeadersOption struct {
	Entries []ResponseHeaderEntry `json:"entries"`
}

type ResponseHeaderEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func getResponseOptions(node *structs.WorkflowNode) (*ResponseOptions, error) {
	if node == nil {
		return nil, errors.New("node is nil")
	}
	value := node.Parameters["options"]
	if value == nil {
		return nil, nil
	}

	options := ResponseOptions{}
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &options)
	if err != nil {
		return nil, err
	}
	return &options, nil
}

func setContentType(headers map[string]string, contentType string) {
	headers["Content-Type"] = contentType
}
