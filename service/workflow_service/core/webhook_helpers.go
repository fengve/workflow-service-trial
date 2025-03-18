package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/errors"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type (
	// Webhook node options
	WebhookNodeOptions struct {
		NoResponseBody     bool                  `json:"noResponseBody,omitempty"`
		RawBody            bool                  `json:"rawBody,omitempty"`
		ResponseData       string                `json:"responseData,omitempty"`
		BinaryPropertyName string                `json:"binaryPropertyName,omitempty"`
		ResponseHeaders    ResponseHeadersOption `json:"responseHeaders,omitempty"`
		// for responseMode:lastNode
		ResponseContentType  string `json:"responseContentType,omitempty"`
		ResponsePropertyName string `json:"responsePropertyName,omitempty"`
	}

	ResponseHeadersOption struct {
		Entries []HeaderEntry `json:"entries,omitempty"`
	}

	HeaderEntry struct {
		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
	}
)

// n8n: WebhookHelpers.getWorkflowWebhooks
// Get all webhooks in the structs.
func GetWorkflowWebhooks(workflowEntity *structs.WorkflowEntity, isTest bool) []structs.WebhookData {
	results := []structs.WebhookData{}
	allNodeObjects := GetAllNodeObjects()
	for index := range workflowEntity.Nodes {
		node := workflowEntity.Nodes[index]
		// If the node is disabled, skip it.
		if node.Disabled {
			continue
		}
		if node.WebhookId == "" {
			// If the node does not have a webhookId, skip it.
			continue
		}

		// If the node is not a webhook, skip it.
		if nodeObject, ok := allNodeObjects[node.Type]; ok {
			webhooks := nodeObject.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.Webhooks
			// Skip if the node does not have webhooks.
			if len(webhooks) == 0 {
				continue
			}

			// If the node is a webhook, add it to the result.
			results = append(
				results,
				structs.WebhookData{
					HttpMethod: getWebhookMethod(&node),
					Node:       node.Name,
					NodeType:   node.Type,
					NodeId:     node.ID,
					// The webhook path is the webhookId by default
					Path:                            getWebhookPath(node.WebhookId, isTest),
					WorkflowId:                      workflowEntity.ID,
					WebhookId:                       node.WebhookId,
					WorkflowExecutionAdditionalData: structs.WorkflowExecuteAdditionalData{},
					IsTest:                          isTest,
				})
		}
	}
	return results
}

// Create the webhook if it does not exist.
// Call the webhook checkExists and create method. It varies depending on the implementation of the node.
func CallWebhookCreateMethod(
	ctx context.Context,
	webhookData *structs.WebhookData,
	workflowEntity *structs.WorkflowEntity,
) {
	allNodes := GetAllNodeObjects()
	// Call webhook checkExists and create method
	nodeObject := allNodes[webhookData.NodeType]
	// Call webhook delete method
	var nodeMethods NodeWebhookMethods
	if newObject, ok := nodeObject.(NodeWebhookMethods); ok {
		nodeMethods = newObject
		webhookCheckExistsFunc := nodeMethods.WebhookMethods().CheckExists
		node := findNodeById(workflowEntity, webhookData.NodeId)
		webhookExists, _ := webhookCheckExistsFunc(ctx, workflowEntity, node, webhookData)
		if !webhookExists {
			webhookCreateFunc := nodeMethods.WebhookMethods().Create
			webhookCreateFunc(ctx, workflowEntity, node, webhookData)
		}
	}
}

func CallWebhookDeleteMethod(
	ctx context.Context,
	webhookData *structs.WebhookData,
	workflowEntity *structs.WorkflowEntity,
) {
	allNodes := GetAllNodeObjects()
	// Call webhook delete method
	nodeObject := allNodes[webhookData.NodeType]
	var nodeMethods NodeWebhookMethods
	if newObject, ok := nodeObject.(NodeWebhookMethods); ok {
		nodeMethods = newObject
		webhookDeleteFunc := nodeMethods.WebhookMethods().Delete
		node := findNodeById(workflowEntity, webhookData.NodeId)
		webhookDeleteFunc(ctx, workflowEntity, node, webhookData)
	}
}

func findNodeById(req *structs.WorkflowEntity, nodeId string) *structs.WorkflowNode {
	for _, node := range req.Nodes {
		if node.ID == nodeId {
			return &node
		}
	}
	return nil
}

// Get the http method for the webhook.
func getWebhookMethod(node *structs.WorkflowNode) string {
	if node == nil {
		return ""
	}

	// The http method for sugarNotificationEventTrigger is always POST.
	if strings.Contains(node.Type, "sugerNotificationEventTrigger") {
		return http.MethodPost
	}
	// For other nodes, the http method is specified in the parameters.
	// If not specified, the default is GET.
	methodRaw, ok := node.Parameters["httpMethod"]
	if !ok {
		return http.MethodGet
	}
	method, ok := methodRaw.(string)
	if !ok || method == "" {
		return http.MethodGet
	}
	return method
}

// Save webhook entity to db
func SaveWebhookEntity(
	ctx context.Context, webhookData *structs.WebhookData) (rdsDbLib.WorkflowWebhookEntity, error) {
	if webhookData == nil {
		return rdsDbLib.WorkflowWebhookEntity{}, errors.New("webhookData is nil")
	}

	return GetRdsDbQueries().CreateWebhookEntity(
		ctx,
		rdsDbLib.CreateWebhookEntityParams{
			WebhookPath: webhookData.Path,
			Method:      webhookData.HttpMethod,
			WebhookId:   sql.NullString{String: webhookData.WebhookId, Valid: true},
			Node:        webhookData.Node,
			WorkflowId:  webhookData.WorkflowId,
		})
}

// Delete webhook entity from db.
func DeleteWebhookEntity(ctx context.Context, webhookData *structs.WebhookData) error {
	if webhookData == nil {
		return nil
	}
	return GetRdsDbQueries().DeleteWebhookEntityByWorkflowId_Path_Method(
		ctx,
		rdsDbLib.DeleteWebhookEntityByWorkflowId_Path_MethodParams{
			WorkflowId:  webhookData.WorkflowId,
			WebhookPath: webhookData.Path,
			Method:      webhookData.HttpMethod,
		})
}

// The webhook path in the db table is part of primary key and should be unique.
// Here we use webhookId (and the suffix "/test" for test webhook) to store webhook_entity in the db.
func getWebhookPath(webhookId string, isTest bool) string {
	if isTest {
		return webhookId + "/test"
	} else {
		return webhookId
	}
}

// Get online or test webhook entity by webhookId
func GetWebhookEntity(
	ctx context.Context, workflowId string,
	webhookId string, isTest bool) (*rdsDbLib.WorkflowWebhookEntity, error) {
	// Read from db, it will return online and test webhook if exists
	webhookEitities, err := GetRdsDbQueries().ListWebhookEntities(
		ctx,
		rdsDbLib.ListWebhookEntitiesParams{
			WorkflowId: workflowId,
			WebhookId:  sql.NullString{String: webhookId, Valid: true},
		})
	if err != nil {
		return nil, err
	}
	if len(webhookEitities) == 0 {
		return nil, fmt.Errorf("no such webhook")
	}

	// Filter target online or test webhook
	targetWebhookPath := getWebhookPath(webhookId, isTest)
	for _, webhookEntity := range webhookEitities {
		if targetWebhookPath == webhookEntity.WebhookPath {
			return &webhookEntity, nil
		}
	}
	return nil, fmt.Errorf("no such webhook")
}

// Parse parameters.options of the webhook node
func ParseWebhookNodeOptions(node *structs.WorkflowNode) (*WebhookNodeOptions, error) {
	raw := node.Parameters["options"]

	options := &WebhookNodeOptions{}
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
