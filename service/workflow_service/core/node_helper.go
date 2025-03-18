package core

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sugerio/workflow-service-trial/shared/structs"
)

// Some func from n8n NodeHelpers.ts and NodeExecuteFunctions.ts

// NodeExecuteFunctions.getNodeWebhookUrl
// https://github.com/sugerio/workflow-service/blob/c1b5d949658247b19abfdb598cf4b427089cb099/packages/core/src/NodeExecuteFunctions.ts#L3818
// https://github.com/sugerio/workflow-service/blob/c1b5d949658247b19abfdb598cf4b427089cb099/packages/core/src/NodeExecuteFunctions.ts#L2322
func GetNodeWebhookUrl(
	name string,
	workflowId string,
	node *structs.WorkflowNode,
	additionalData structs.WorkflowExecuteAdditionalData,
	mode structs.WorkflowActivateMode,
	isTest bool) string {
	baseUrl := additionalData.WebhookBaseUrl
	if isTest {
		baseUrl = additionalData.WebhookTestBaseUrl
	}
	webhookDescription := GetWebhookDescription(name, node)
	if webhookDescription == nil {
		return ""
	}

	// TODO parse path,isFullPath if they contains expression
	path := webhookDescription.Path
	isFullPath := webhookDescription.IsFullPath

	return GetNodeWebhookUrlByBaseUrl(baseUrl, workflowId, node, path, isFullPath)
}

// Get WebhookDescription by name(webhook.name) and workflowNode
// NodeExecuteFunctions.getWebhookDescription
// https://github.com/sugerio/workflow-service/blob/c1b5d949658247b19abfdb598cf4b427089cb099/packages/core/src/NodeExecuteFunctions.ts#L2374
func GetWebhookDescription(name string, node *structs.WorkflowNode) *structs.WebhookDescription {
	allNodeObjects := GetAllNodeObjects()
	nodeName := node.Type
	if nodeObject, ok := allNodeObjects[nodeName]; ok {
		webhooks := nodeObject.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.Webhooks
		for _, webhook := range webhooks {
			if webhook.Name == name {
				return &webhook
			}
		}
	}
	return nil
}

// NodeHelpers.getNodeWebhookUrl
// https://github.com/sugerio/workflow-service/blob/c1b5d949658247b19abfdb598cf4b427089cb099/packages/workflow/src/NodeHelpers.ts#L1019
func GetNodeWebhookUrlByBaseUrl(
	baseUrl string,
	workflowId string,
	node *structs.WorkflowNode,
	path string,
	isFullPath bool,
) string {
	if (strings.HasPrefix(path, ":") || strings.Contains(path, "/:")) && node.WebhookId != "" {
		isFullPath = false
	}
	path = strings.TrimPrefix(path, "/")
	webhookPath := GetNodeWebhookPath(workflowId, node, path, isFullPath, false)
	return fmt.Sprintf("%s/%s", baseUrl, webhookPath)
}

func GetNodeWebhookPath(
	workflowId string,
	node *structs.WorkflowNode,
	path string,
	isFullPath bool,
	restartWebhook bool) string {
	if restartWebhook {
		return path
	}
	if node.WebhookId == "" {
		encodedStr := url.QueryEscape(strings.ToLower(node.Name))
		return fmt.Sprintf("%s/%s/%s", workflowId, encodedStr, path)
	} else {
		if isFullPath {
			return path
		} else {
			return fmt.Sprintf("%s/%s", node.WebhookId, path)
		}
	}
}
