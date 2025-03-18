package http_request

// TODO: Validate return values from GetValueFromMap or GetNodeParameter
// and confirm that all ignored errors in this file are safe before removing this TODO.

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (
	Category = structs.CategoryExecutor
	Name     = "n8n-nodes-base.httpRequest"
)

var (
	//go:embed node.json
	rawJson []byte
	//go:embed httprequest.svg
	icon               []byte
	binaryContentTypes = []string{
		"image/",
		"audio/",
		"video/",
		"application/octet-stream",
		"application/gzip",
		"application/zip",
		"application/vnd.rar",
		"application/epub+zip",
		"application/x-bzip",
		"application/x-bzip2",
		"application/x-cdf",
		"application/vnd.amazon.ebook",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-fontobject",
		"application/vnd.oasis.opendocument.presentation",
		"application/pdf",
		"application/x-tar",
		"application/vnd.visio",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/x-7z-compressed",
	}

	fileTypes = map[string]string{
		"image/":                 "image",
		"audio/":                 "audio",
		"video/":                 "video",
		"text/html":              "html",
		"text/":                  "text",
		"application/json":       "json",
		"application/javascript": "text",
		"application/pdf":        "pdf",
	}

	extensions = map[string]string{
		"application/pdf":             ".pdf",
		"application/zip":             ".zip",
		"application/gzip":            ".gz",
		"application/x-gzip":          ".gz",
		"application/x-tar":           ".tar",
		"application/x-7z-compressed": ".7z",
		"audio/mpeg":                  ".mp3",
		"audio/wav":                   ".wav",
		"audio/ogg":                   ".ogg",
		"image/jpeg":                  ".jpeg",
		"image/png":                   ".png",
		"image/gif":                   ".gif",
		"image/bmp":                   ".bmp",
		"image/webp":                  ".webp",
		"video/mp4":                   ".mp4",
		"video/mpeg":                  ".mpeg",
		"video/quicktime":             ".mov",
		"video/x-msvideo":             ".avi",
		"text/plain":                  ".txt",
		"text/html":                   ".html",
		"text/css":                    ".css",
		"application/json":            ".json",
		"application/xml":             ".xml",
	}
)

type (
	HttpRequestExecutor struct {
		spec *structs.WorkflowNodeSpec
	}

	ParameterOptions struct {
		Batching               BatchOption    `json:"batching,omitempty"`
		AllowUnauthorizedCerts bool           `json:"allowUnauthorizedCerts,omitempty"`
		QueryParameterArrays   string         `json:"queryParameterArrays,omitempty"`
		Redirect               RedirectOption `json:"redirect,omitempty"`
		Response               ResponseOption `json:"response,omitempty"`
		Proxy                  string         `json:"proxy,omitempty"`
		Timeout                int            `json:"timeout,omitempty"`
	}

	BatchOption struct {
		Batch BatchOptionValue `json:"batch,omitempty"`
	}

	BatchOptionValue struct {
		BatchSize     int `json:"batchSize,omitempty"`
		BatchInterval int `json:"batchInterval,omitempty"`
	}

	RedirectOption struct {
		Redirect RedirectOptionValue `json:"redirect,omitempty"`
	}

	RedirectOptionValue struct {
		FollowRedirects bool `json:"followRedirects,omitempty"`
		MaxRedirects    int  `json:"maxRedirects,omitempty"`
	}

	ResponseOption struct {
		Response ResponseOptionValue `json:"response,omitempty"`
	}

	ResponseOptionValue struct {
		FullResponse       bool   `json:"fullResponse,omitempty"`
		NeverError         bool   `json:"neverError,omitempty"`
		ResponseFormat     string `json:"responseFormat,omitempty"`
		OutputPropertyName string `json:"outputPropertyName,omitempty"`
	}

	Pagination struct {
		PaginationMode          string               `json:"paginationMode,omitempty"`
		NextURL                 string               `json:"nextURL,omitempty"`
		Parameters              PaginationParameters `json:"parameters,omitempty"`
		PaginationCompleteWhen  string               `json:"paginationCompleteWhen,omitempty"`
		StatusCodesWhenComplete string               `json:"statusCodesWhenComplete,omitempty"`
		LimitPagesFetched       bool                 `json:"limitPagesFetched,omitempty"`
		MaxRequests             int                  `json:"maxRequests,omitempty"`
	}

	PaginationParameters struct {
		Parameters []PaginationParameterItem `json:"parameters,omitempty"`
	}

	PaginationParameterItem struct {
		Type  string `json:"type,omitempty"`
		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
	}

	NormalParameter struct {
		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
		// for BodyParameter
		ParameterType      string `json:"parameterType,omitempty"`
		InputDataFieldName string `json:"inputDataFieldName,omitempty"`
	}

	ResponseWithIndex struct {
		Index    int
		Response *http.Response
	}

	// RequestBuildOptions, RequestOptionsHeader are used for http request contruct, not parsed from node input.
	RequestBuildOptions struct {
		Headers                 RequestBuildHeader
		Method                  string
		URI                     string
		Gzip                    bool
		RejectUnauthorized      bool
		FollowRedirect          bool // not implemented yet.
		FollowAllRedirect       bool // not implemented yet.
		ResolveWithFullResponse bool
		MaxRedirects            int    // not implemented yet.
		Simple                  bool   // not implemented yet.
		Proxy                   string // not implemented yet.
		Timeout                 int
		Qs                      map[string]string
		QsStringifyOptions      map[string]string // not implemented yet.
		ContentType             string
		BodyFormUrlencoded      map[string]string
		BodyFormData            map[string]interface{}
		BodyString              string
		BodyBytes               []byte
		Encoding                string // not implemented yet.
		Json                    bool   // not implemented yet.
		UseStream               bool   // not implemented yet.
	}

	RequestBuildHeader map[string]string
)

func init() {
	hre := &HttpRequestExecutor{
		spec: &structs.WorkflowNodeSpec{},
	}
	hre.spec.JsonConfig = rawJson
	hre.spec.GenerateSpec()

	core.Register(hre)
	core.RegisterEmbedIcons(Name, icon)
}

func (hre *HttpRequestExecutor) Category() structs.NodeObjectCategory {
	return Category
}

func (hre *HttpRequestExecutor) Name() string {
	return Name
}

func (hre *HttpRequestExecutor) DefaultSpec() interface{} {
	return hre.spec
}

func (hre *HttpRequestExecutor) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	items := core.GetInputData(input.Data)

	// Get Pagination
	paginationRaw, err := core.GetNodeParameter(Name, "options.pagination.pagination",
		map[string]interface{}{},
		input,
		0,
		core.GetNodeParameterOptions{},
	)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}
	pagination, err := toPagination(paginationRaw)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}

	// Response option
	responseOption, err := core.GetNodeParameterAsType(Name, "options.response.response",
		ResponseOptionValue{}, input, 0)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}

	if responseOption.ResponseFormat == "" {
		responseOption.ResponseFormat = "autodetect"
	}
	if responseOption.OutputPropertyName == "" {
		responseOption.OutputPropertyName = "data"
	}
	autoDetectResponseFormat := responseOption.ResponseFormat == "autodetect"

	sendBodyHttpMethods := []string{"PATCH", "POST", "PUT", "GET"}
	responseChannel := make(chan ResponseWithIndex)
	client := &http.Client{}
	// TODO: set total timeout here
	parentCtx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	for itemIndex := range items {
		item := items[itemIndex]
		requestMethod, err := core.GetNodeParameterAsBasicType(Name, "method", "GET",
			input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		// Send Query Parameters
		sendQuery, err := core.GetNodeParameterAsBasicType(Name, "sendQuery", false,
			input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		queryParametersRaw, err := core.GetNodeParameter(Name, "queryParameters.parameters",
			[]interface{}{}, input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		queryParameters, err := toNormalParameters(queryParametersRaw)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		specifyQuery, err := core.GetNodeParameterAsBasicType(Name, "specifyQuery", "keypair",
			input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		jsonQueryParameter, err := core.GetNodeParameterAsBasicType(Name, "jsonQuery", "",
			input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}

		// Send Body

		sendBody, err := core.GetNodeParameterAsBasicType(Name, "sendBody", false,
			input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		bodyContentType, err := core.GetNodeParameterAsBasicType(Name, "contentType", "",
			input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		specifyBody, err := core.GetNodeParameterAsBasicType(Name, "specifyBody", "",
			input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		bodyParametersRaw, err := core.GetNodeParameter(Name, "bodyParameters.parameters",
			[]interface{}{}, input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		bodyParameters, err := toNormalParameters(bodyParametersRaw)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		jsonBody, err := core.GetNodeParameterAsBasicType(Name, "jsonBody", "", input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		body, err := core.GetNodeParameterAsBasicType(Name, "body", "", input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}

		// Send Headers
		sendHeaders, err := core.GetNodeParameterAsBasicType(Name, "sendHeaders", false, input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		headerParametersRaw, err := core.GetNodeParameter(Name, "headerParameters.parameters",
			[]interface{}{}, input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		headerParameters, err := toNormalParameters(headerParametersRaw)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		specifyHeaders, err := core.GetNodeParameterAsBasicType(Name, "specifyHeaders", "keypair", input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		jsonHeadersParameter, err := core.GetNodeParameterAsBasicType(Name, "jsonHeaders", "", input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}

		// Options
		optionsRaw, err := core.GetNodeParameter(Name, "options", map[string]interface{}{}, input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		options, err := toOptions(optionsRaw)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}

		// Url
		url, err := core.GetNodeParameterAsBasicType(Name, "url", "", input, itemIndex)
		if err != nil || url == "" {
			return core.GenerateFailedResponse(Name, err)
		}
		err = validateUrl(url)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}

		// Default batch size adjust
		batchSize := options.Batching.Batch.BatchSize
		if batchSize <= 0 {
			batchSize = 1
		}

		batchInterval := options.Batching.Batch.BatchInterval
		if itemIndex > 0 && batchSize > 0 && batchInterval > 0 {
			if itemIndex%batchSize == 0 {
				time.Sleep(time.Duration(batchInterval) * time.Millisecond)
			}
		}

		requestBuildOptions := RequestBuildOptions{
			Headers:                 RequestBuildHeader{},
			Method:                  requestMethod,
			URI:                     url,
			Gzip:                    true,
			RejectUnauthorized:      !options.AllowUnauthorizedCerts || false,
			ResolveWithFullResponse: true,
		}

		if requestBuildOptions.Method != "GET" {
			requestBuildOptions.FollowAllRedirect = false
		}

		if options.Redirect.Redirect.FollowRedirects {
			requestBuildOptions.FollowRedirect = true
			requestBuildOptions.FollowAllRedirect = true
		}

		requestBuildOptions.MaxRedirects = options.Redirect.Redirect.MaxRedirects

		if options.Response.Response.NeverError {
			requestBuildOptions.Simple = false
		}
		if options.Proxy != "" {
			requestBuildOptions.Proxy = options.Proxy
		}
		if options.Timeout > 0 {
			requestBuildOptions.Timeout = options.Timeout
		} else {
			requestBuildOptions.Timeout = 300_000
		}
		if sendQuery && options.QueryParameterArrays != "" {
			requestBuildOptions.QsStringifyOptions = map[string]string{
				"arrayFormat": options.QueryParameterArrays,
			}
		}

		// Change the way data get send in case a different content-type than JSON got selected
		if sendBody && checkArrayContains(sendBodyHttpMethods, requestMethod) {
			requestBuildOptions.ContentType = bodyContentType
			if bodyContentType == "form-urlencoded" {
				// keypair is default
				if specifyBody == "string" {
					requestBuildOptions.BodyFormUrlencoded = urlParamsToMap(body)
				} else {
					requestBuildOptions.BodyFormUrlencoded = normalParametersToMap(bodyParameters)
				}
				requestBuildOptions.Headers["content-type"] = "application/x-www-form-urlencoded"
			} else if bodyContentType == "multipart-form-data" {
				requestBuildOptions.BodyFormData = parametersToMap(bodyParameters, item)
				requestBuildOptions.Headers["content-type"] = "multipart/form-data"
			} else if bodyContentType == "json" {
				// keypair is default
				if specifyBody == "json" {
					requestBuildOptions.BodyString = jsonBody
				} else {
					bodyParameterMap := parametersToMap(bodyParameters, item)
					jsonData, err := json.Marshal(bodyParameterMap)
					if err != nil {
						return core.GenerateFailedResponse(Name, fmt.Errorf("contentType json specifyBody keypair parse failed"))
					}
					requestBuildOptions.BodyString = string(jsonData)
				}
				requestBuildOptions.Headers["content-type"] = "application/json"
			} else if bodyContentType == "binaryData" {
				inputDataFieldName, err := core.GetNodeParameterAsBasicType(Name, "inputDataFieldName", "", input, itemIndex)
				if err != nil || inputDataFieldName == "" {
					return core.GenerateFailedResponse(Name, fmt.Errorf("binaryData inputDataFieldName empty"))
				}
				// get binary data from item.binary
				binary, err := core.ConvertInterfaceToType[map[string]structs.WorkflowBinaryData](item["binary"])
				if err != nil {
					return core.GenerateFailedResponse(Name, fmt.Errorf("the item.binary is null"))
				}
				binaryData, ok := (*binary)[inputDataFieldName]
				if !ok {
					return core.GenerateFailedResponse(Name,
						fmt.Errorf("the item binary didn't contains inputDataFieldName of %s", inputDataFieldName))
				}

				// Get binary data
				content, err := core.GetDataFromBinaryData(&binaryData)
				if err != nil {
					return core.GenerateFailedResponse(Name, err)
				}
				// Data as bytes
				decodedBytes := []byte(content)

				contentLength := len(decodedBytes)
				contentType := "application/octet-stream"
				if binaryData.MimeType != "" {
					contentType = binaryData.MimeType
				}
				requestBuildOptions.BodyBytes = decodedBytes
				requestBuildOptions.Headers["content-length"] = strconv.Itoa(contentLength)
				requestBuildOptions.Headers["content-type"] = contentType
			} else if bodyContentType == "raw" {
				requestBuildOptions.BodyString = body
				rawContentType, err := core.GetNodeParameterAsBasicType(
					Name, "rawContentType", "", input, itemIndex)
				if err != nil {
					return core.GenerateFailedResponse(Name, err)
				}
				requestBuildOptions.Headers["content-type"] = rawContentType
			}
		}

		if sendQuery && queryParameters != nil && len(*queryParameters) > 0 {
			// keypair is default
			if specifyQuery == "json" {
				queryParameter := map[string]string{}
				err := json.Unmarshal([]byte(jsonQueryParameter), &queryParameter)
				if err != nil {
					return core.GenerateFailedResponse(Name, fmt.Errorf("json parameter need to be an valid json"))
				}
				requestBuildOptions.Qs = queryParameter
			} else {
				requestBuildOptions.Qs = normalParametersToMap(queryParameters)
			}
		}

		if sendHeaders && len(*headerParameters) > 0 {
			additionalHeaders := map[string]string{}
			// keypair is default
			if specifyHeaders == "json" {
				err := json.Unmarshal([]byte(jsonHeadersParameter), &additionalHeaders)
				if err != nil {
					return core.GenerateFailedResponse(Name, fmt.Errorf("json parameter need to be an valid json"))
				}
			} else {
				additionalHeaders = normalParametersToMap(headerParameters)
			}
			for k, v := range additionalHeaders {
				requestBuildOptions.Headers[strings.ToLower(k)] = v
			}
		}

		// Options
		if autoDetectResponseFormat || responseOption.ResponseFormat == "file" {
			requestBuildOptions.Encoding = ""
			requestBuildOptions.Json = false
			requestBuildOptions.UseStream = true
		} else if bodyContentType == "raw" {
			requestBuildOptions.Json = false
			requestBuildOptions.UseStream = true
		} else {
			requestBuildOptions.Json = true
		}

		if _, ok := requestBuildOptions.Headers["accept"]; !ok {
			if responseOption.ResponseFormat == "json" {
				requestBuildOptions.Headers["accept"] = "application/json,text/*;q=0.99"
			} else if responseOption.ResponseFormat == "text" {
				requestBuildOptions.Headers["accept"] = "application/json,text/html,application/xhtml+xml,application/xml,text/*;q=0.9, */*;q=0.1"
			} else {
				requestBuildOptions.Headers["accept"] = "application/json,text/html,application/xhtml+xml,application/xml,text/*;q=0.9, image/*;q=0.8, */*;q=0.7"
			}
		}

		// Pagination request
		if pagination.PaginationMode != "" && pagination.PaginationMode != "off" {
			// TODO handle pagination request
		} else {
			go sendRequest(parentCtx, itemIndex, client, requestBuildOptions, responseChannel)
		}
	}

	result := structs.NodeData{}
	for i := 0; i < len(items); i++ {
		select {
		case responseWithIndex := <-responseChannel:
			requestResult := handleResponse(responseWithIndex, responseOption.ResponseFormat == "autodetect",
				responseOption.ResponseFormat, responseOption.FullResponse, responseOption.OutputPropertyName)
			result = append(result, requestResult...)
		case <-parentCtx.Done():
			fmt.Println("parentCtx timeout or canceled")
		}
	}

	close(responseChannel)

	return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
}

func handleResponse(
	responseWithIndex ResponseWithIndex,
	autoDetectResponseFormat bool,
	responseFormat string,
	fullResponse bool,
	outputPropertyName string) []map[string]interface{} {
	result := []map[string]interface{}{}

	index := responseWithIndex.Index
	response := responseWithIndex.Response

	if response == nil {
		result = append(result, map[string]interface{}{
			"json": map[string]interface{}{
				"error": "request error or timeout",
			},
			"pairedItem": map[string]interface{}{
				"item": index,
			},
		})
		return result
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		result = append(result, map[string]interface{}{
			"json": map[string]interface{}{
				"error": "read response body error",
			},
			"pairedItem": map[string]interface{}{
				"item": index,
			},
		})
		return result
	}

	if autoDetectResponseFormat {
		responseContentType := response.Header.Get("content-type")
		if strings.Contains(responseContentType, "application/json") {
			responseFormat = "json"
		} else if checkBinaryContentType(responseContentType) {
			responseFormat = "file"
		} else {
			responseFormat = "text"
		}
	}

	if responseFormat == "file" {
		itemResult := map[string]interface{}{
			"json":   map[string]interface{}{},
			"binary": map[string]structs.WorkflowBinaryData{},
			"pairedItem": map[string]interface{}{
				"item": index,
			},
		}
		itemResultJson := map[string]interface{}{}
		if fullResponse {
			itemResultJson["headers"] = response.Header
			itemResultJson["statusCode"] = response.StatusCode
			itemResultJson["statusMessage"] = response.Status
			itemResult["json"] = itemResultJson
		}

		binaryData := prepareBinaryData(response, body)
		result = append(result, map[string]interface{}{
			"json": itemResultJson,
			"binary": map[string]structs.WorkflowBinaryData{
				outputPropertyName: binaryData,
			},
			"pairedItem": map[string]interface{}{
				"item": index,
			},
		})

	} else if responseFormat == "text" {
		itemResultJson := map[string]interface{}{}
		if fullResponse {
			itemResultJson[outputPropertyName] = string(body)
			itemResultJson["headers"] = response.Header
			itemResultJson["statusCode"] = response.StatusCode
			itemResultJson["statusMessage"] = response.Status
		}
		itemResultJson[outputPropertyName] = string(body)
		result = append(result, map[string]interface{}{
			"json": itemResultJson,
			"pairedItem": map[string]interface{}{
				"item": index,
			},
		})
	} else {
		// responseFormat is json
		var jsonObject interface{}
		err := json.Unmarshal(body, &jsonObject)
		if err != nil {
			result = append(result, map[string]interface{}{
				"json": map[string]interface{}{
					"error": "parse response json error",
				},
				"pairedItem": map[string]interface{}{
					"item": index,
				},
			})
			return result
		}

		itemResultJson := map[string]interface{}{}
		if fullResponse {
			itemResultJson["headers"] = response.Header
			itemResultJson["statusCode"] = response.StatusCode
			itemResultJson["statusMessage"] = response.Status
			itemResultJson["body"] = jsonObject
			result = append(result, map[string]interface{}{
				"json": itemResultJson,
				"pairedItem": map[string]interface{}{
					"item": index,
				},
			})
		} else {
			if isArray(jsonObject) {
				if jsonArray, ok := jsonObject.([]interface{}); ok {
					for _, jsonSubItem := range jsonArray {
						result = append(result, map[string]interface{}{
							"json": jsonSubItem,
							"pairedItem": map[string]interface{}{
								"item": index,
							},
						})
					}
				}
			} else {
				result = append(result, map[string]interface{}{
					"json": jsonObject,
					"pairedItem": map[string]interface{}{
						"item": index,
					},
				})
			}
		}
	}
	return result
}

// Construct a binary data from response body. The content is base64 encoded as a string.
func prepareBinaryData(response *http.Response, body []byte) structs.WorkflowBinaryData {
	contentType := response.Header.Get("Content-Type")
	filePath := response.Request.URL.Path

	if contentType == "" {
		contentType = "text/plain"
	}

	data := structs.WorkflowBinaryData{
		MimeType:      contentType,
		FileType:      structs.WorkflowBinaryFileType(fileTypeFromMimeType(contentType)),
		FileExtension: fileExtensionFromMimeType(contentType),
	}

	if filePath != "" {
		if strings.Contains(filePath, "?") {
			filePath = strings.Split(filePath, "?")[0]
		}
		filePathParts := path.Base(filePath)
		dir := path.Dir(filePath)
		if dir != "." {
			data.Directory = dir
		}
		data.FileName = filePathParts
		fileExtension := strings.TrimPrefix(path.Ext(filePath), ".")
		if fileExtension != "" {
			data.FileExtension = fileExtension
		}
	}

	base64String := base64.StdEncoding.EncodeToString(body)
	data.Base64Encoded = true
	data.Data = base64String
	data.FileSize = prettyBytes(float64(len(body)))
	return data
}

func fileTypeFromMimeType(mimeType string) string {
	for ct, fileType := range fileTypes {
		if strings.HasPrefix(mimeType, ct) {
			return fileType
		}
	}
	return ""
}

func fileExtensionFromMimeType(mimeType string) string {
	for ct, ext := range extensions {
		if strings.HasPrefix(mimeType, ct) {
			return ext
		}
	}
	return ""
}

func toPagination(raw interface{}) (*Pagination, error) {
	pagination := &Pagination{}
	if raw == nil {
		return pagination, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, pagination)
	if err != nil {
		return nil, err
	}
	return pagination, nil
}

func toResponseOption(raw interface{}) (*ResponseOptionValue, error) {
	responseOption := &ResponseOptionValue{}
	if raw == nil {
		return responseOption, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, responseOption)
	if err != nil {
		return nil, err
	}
	return responseOption, nil
}

func toNormalParameters(raw interface{}) (*[]NormalParameter, error) {
	parameters := &[]NormalParameter{}
	if raw == nil {
		return parameters, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, parameters)
	if err != nil {
		return nil, err
	}
	return parameters, nil
}

func toOptions(raw interface{}) (*ParameterOptions, error) {
	options := &ParameterOptions{}
	if raw == nil {
		return options, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, options)
	if err != nil {
		return nil, err
	}
	return options, nil
}

func normalParametersToMap(parameters *[]NormalParameter) map[string]string {
	result := make(map[string]string)
	for index := range *parameters {
		entry := (*parameters)[index]
		if entry.ParameterType == "formBinaryData" {
			fmt.Println("parameters which contains formBinaryData shouldn't use the method")
		} else {
			result[entry.Name] = entry.Value
		}
	}
	return result
}

func parametersToMap(parameters *[]NormalParameter, item map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for index := range *parameters {
		entry := (*parameters)[index]
		if entry.ParameterType == "formBinaryData" {
			if entry.InputDataFieldName == "" {
				continue
			}
			// get binary data from item.binary
			binary, err := core.ConvertInterfaceToType[map[string]structs.WorkflowBinaryData](item["binary"])
			if err != nil {
				fmt.Println("the item didn't contains binary")
				continue
			}

			binaryData, ok := (*binary)[entry.InputDataFieldName]
			if !ok {
				fmt.Println("the item binary didn't contains inputDataFieldName of", entry.InputDataFieldName)
				continue
			}

			// Get binary data
			data, err := core.GetDataFromBinaryData(&binaryData)
			if err != nil {
				fmt.Println("get data from binary data failed")
				continue
			}

			v := map[string]interface{}{
				"value": []byte(data),
			}
			options := map[string]string{}
			if binaryData.FileName != "" {
				options["filename"] = binaryData.FileName
			}
			if binaryData.MimeType != "" {
				options["contentType"] = binaryData.MimeType
			}
			v["options"] = options
			result[entry.Name] = v
		} else {
			result[entry.Name] = entry.Value
		}
	}
	return result
}

func urlParamsToMap(str string) map[string]string {
	urlValues := parseURLSearchParams(str)
	return convertUrlValuesToMap(urlValues)
}

func parseURLSearchParams(body string) url.Values {
	values := make(url.Values)
	params := strings.Split(body, "&")
	for _, param := range params {
		pair := strings.Split(param, "=")
		if len(pair) == 2 {
			key := pair[0]
			value := pair[1]
			values.Add(key, value)
		}
	}
	return values
}

func convertUrlValuesToMap(values url.Values) map[string]string {
	result := make(map[string]string)
	for key, value := range values {
		if len(value) > 0 {
			result[key] = value[0]
		}
	}
	return result
}

func checkArrayContains(arr []string, target string) bool {
	for _, item := range arr {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func sendRequest(
	ctx context.Context,
	index int,
	client *http.Client,
	requestOptions RequestBuildOptions,
	responseChannel chan<- ResponseWithIndex) {
	request, err := buildRequest(requestOptions)
	if err != nil {
		fmt.Println("build request error:", err)
		responseChannel <- ResponseWithIndex{Index: index, Response: nil}
		return
	}
	currentCtx, cancel := context.WithTimeout(ctx, time.Duration(requestOptions.Timeout)*time.Millisecond)
	defer cancel()

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(request.URL.String(), "request error:", err)
		responseChannel <- ResponseWithIndex{Index: index, Response: nil}
		return
	}
	select {
	case <-currentCtx.Done():
		fmt.Println(request.URL.String(), "request timeout", index)
		responseChannel <- ResponseWithIndex{Index: index, Response: nil}
		return
	default:
		responseChannel <- ResponseWithIndex{Index: index, Response: response}
	}
}

func buildRequest(requestOptions RequestBuildOptions) (*http.Request, error) {
	method := requestOptions.Method
	baseUrl := requestOptions.URI
	url, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	// QueryParameters
	if qs := requestOptions.Qs; len(qs) > 0 {
		params := convertParameterMapToUrlValues(qs)
		url.RawQuery = params.Encode()
	}

	// Body
	var body io.Reader
	bodyContentType := requestOptions.ContentType
	if bodyContentType == "form-urlencoded" {
		formData := convertParameterMapToUrlValues(requestOptions.BodyFormUrlencoded)
		body = strings.NewReader(formData.Encode())
	} else if bodyContentType == "multipart-form-data" {
		bodyBuf, contentType, err := convertFormDataToBuffer(requestOptions.BodyFormData)
		if err != nil {
			return nil, err
		}
		body = bodyBuf
		requestOptions.Headers["content-type"] = contentType
	} else if bodyContentType == "json" {
		body = bytes.NewBuffer([]byte(requestOptions.BodyString))
	} else if bodyContentType == "binaryData" {
		body = bytes.NewBuffer(requestOptions.BodyBytes)
	} else if bodyContentType == "raw" {
		body = strings.NewReader(requestOptions.BodyString)
	}

	// Request
	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return nil, err
	}

	// Header
	for k, v := range requestOptions.Headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

func convertFormDataToBuffer(formData map[string]interface{}) (*bytes.Buffer, string, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	for k, v := range formData {
		if value, ok := v.(string); ok {
			textFieldWriter, err := bodyWriter.CreateFormField(k)
			if err != nil {
				return nil, "", err
			}
			_, err = textFieldWriter.Write([]byte(value))
			if err != nil {
				return nil, "", err
			}
		} else {
			vmap, ok := v.(map[string]interface{})
			if !ok {
				fmt.Printf("the formdata is invalid because the field of %s is not a map\n", k)
				continue
			}
			bytes, ok := vmap["value"].([]byte)
			if !ok {
				fmt.Println("the formdata is invalid because the field value format is not []byte")
				continue
			}
			valueOptions, ok := vmap["options"].(map[string]string)
			if !ok {
				fmt.Println("the formdata is invalid because options of the value is not a map")
				continue
			}
			fileName, ok := valueOptions["filename"]
			if !ok || len(fileName) == 0 {
				return nil, "", fmt.Errorf("filename in formdata field options is invalid")
			}
			fileWriter, err := bodyWriter.CreateFormFile(k, fileName)
			if err != nil {
				return nil, "", err
			}
			_, err = fileWriter.Write(bytes)
			if err != nil {
				return nil, "", err
			}
		}
	}
	bodyWriter.Close()
	contentType := bodyWriter.FormDataContentType()
	return bodyBuf, contentType, nil
}

func convertParameterMapToUrlValues(qs map[string]string) url.Values {
	params := url.Values{}
	for k, v := range qs {
		params.Set(k, v)
	}
	return params
}

func isArray(obj interface{}) bool {
	if obj == nil {
		return false
	}
	value := reflect.TypeOf(obj)
	return value.Kind() == reflect.Array || value.Kind() == reflect.Slice
}

func checkBinaryContentType(contentType string) bool {
	for _, item := range binaryContentTypes {
		if strings.Contains(contentType, item) {
			return true
		}
	}
	return false
}

// Validate the URL. It should not be an IP address or a local hostname.
// Return an error if the URL is invalid. Otherwise, return nil.
func validateUrl(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	hostname := parsedURL.Hostname()
	ip := net.ParseIP(hostname)
	if ip != nil {
		if strings.HasPrefix(hostname, "10.") || strings.HasPrefix(hostname, "192.168") || hostname == "127.0.0.1" {
			return fmt.Errorf("%s forbidden ip address", urlStr)
		}
	} else {
		if strings.HasSuffix(hostname, ".svc") || strings.Contains(hostname, "localhost") {
			return fmt.Errorf("%s forbidden hostname", urlStr)
		}
	}

	return nil
}

func prettyBytes(size float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}

	unitIndex := 0
	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}
