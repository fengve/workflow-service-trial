package structs

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/hashicorp/go-multierror"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
)

func (workflowEntity *WorkflowEntity) GetNodeById(nodeId string) *WorkflowNode {
	if workflowEntity == nil {
		return nil
	}

	for _, node := range workflowEntity.Nodes {
		if node.ID == nodeId {
			return &node
		}
	}
	return nil
}

func (node *WorkflowNode) GetWebhookResponseMode() (WebhookResponseMode, error) {
	if node == nil {
		return "", errors.New("node is nil")
	}
	value := node.Parameters["responseMode"]
	if value == nil {
		// If not specified, default to onReceived
		return WebhookResponseMode_OnReceived, nil
	} else if str, ok := value.(string); ok {
		return WebhookResponseMode(str), nil
	} else {
		return "", errors.New("responseMode must be a string")
	}
}

func (node *WorkflowNode) GetWebhookResponseData() (WebhookResponseData, error) {
	if node == nil {
		return "", errors.New("node is nil")
	}
	value := node.Parameters["responseData"]
	if value == nil {
		// If not specified, default to firstEntryJson
		return WebhookResponseData_FirstEntryJson, nil
	} else if str, ok := value.(string); ok {
		return WebhookResponseData(str), nil
	} else {
		return "", errors.New("responseData must be a string")
	}
}

// ToWorkflowEntity converts a rdsDbLib.WorkflowWorkflowEntity to a WorkflowEntity.
func ToWorkflowEntity(entity rdsDbLib.WorkflowWorkflowEntity) (WorkflowEntity, error) {
	workflowEntity := WorkflowEntity{
		ID:           entity.ID,
		Name:         entity.Name,
		Active:       entity.Active,
		TriggerCount: int(entity.TriggerCount),
		VersionId:    entity.VersionId.String,
		CreatedAt:    &entity.CreatedAt,
		UpdatedAt:    &entity.UpdatedAt,
		SugerOrgId:   entity.SugerOrgId,
	}

	var combinedErr error
	if err := UnmarshalOmitEmpty(entity.Nodes, &workflowEntity.Nodes); err != nil {
		combinedErr = multierror.Append(combinedErr, err)
	}
	if err := UnmarshalOmitEmpty(entity.Connections, &workflowEntity.Connections); err != nil {
		combinedErr = multierror.Append(combinedErr, err)
	}
	if err := UnmarshalOmitEmpty(entity.Settings.RawMessage, &workflowEntity.Settings); err != nil {
		combinedErr = multierror.Append(combinedErr, err)
	}
	if err := UnmarshalOmitEmpty(entity.StaticData.RawMessage, &workflowEntity.StaticData); err != nil {
		combinedErr = multierror.Append(combinedErr, err)
	}
	if err := UnmarshalOmitEmpty(entity.PinData.RawMessage, &workflowEntity.PinData); err != nil {
		combinedErr = multierror.Append(combinedErr, err)
	}
	if err := UnmarshalOmitEmpty(entity.Meta.RawMessage, &workflowEntity.Meta); err != nil {
		combinedErr = multierror.Append(combinedErr, err)
	}
	return workflowEntity, combinedErr
}

func UnmarshalOmitEmpty(from []byte, to interface{}) error {
	if from == nil || len(from) == 0 {
		return nil
	}
	return json.Unmarshal(from, to)
}

// Parse the workflow webhook publicUrl to extract the workflowId, nodeId, and webhookId.
func ParseWorkflowWebhookUrl(name, publicUrl string) (WorkflowWebhook, error) {
	// publicUrl pattern /public/webhook/workflow/{flowId}/node/{nodeId}?webhookId={webhookId}
	pattern := `/public/webhook/workflow/([^/]+)/node/([^?]+)\?webhookId=([^&]+)`
	webhook := WorkflowWebhook{}

	// We could utilize either url.Parse or regex for the task at hand.
	// However, regex might offer a more straightforward and simpler solution.
	regex := regexp.MustCompile(pattern)

	// Find matches
	matches := regex.FindStringSubmatch(publicUrl)

	if len(matches) < 4 {
		return webhook, fmt.Errorf("invalid webhook publicUrl: %v", publicUrl)
	}

	// Extract parameters from matches
	workflowId := matches[1]
	nodeId := matches[2]
	webhookId := matches[3]

	webhook.HookName = name
	webhook.PublicUrl = publicUrl
	webhook.WebhookId = webhookId
	webhook.WorkflowId = workflowId
	webhook.NodeId = nodeId

	return webhook, nil
}
