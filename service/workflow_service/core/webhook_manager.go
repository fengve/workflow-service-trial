package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sugerio/workflow-service-trial/shared/structs"

	"github.com/sqlc-dev/pqtype"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
)

func RegisterWebhook(ctx context.Context, workflowID string, isTest bool) error {
	workflowEntity, err := GetWorkflowEntityById(ctx, workflowID)
	if err != nil {
		Errorf("Failed to get workflow entity %v", err)
		return err
	}

	if workflowEntity == nil {
		return nil
	}

	webhooks := GetWorkflowWebhooks(workflowEntity, isTest)
	if len(webhooks) == 0 {
		return nil
	}
	for _, webhook := range webhooks {
		// Create record in webhook_entity
		_, err := SaveWebhookEntity(ctx, &webhook)
		if err != nil {
			Errorf("webhook %s workflowId %s save failed with err %v", webhook.WebhookId, webhook.WorkflowId, err)
			continue
		}
		CallWebhookCreateMethod(ctx, &webhook, workflowEntity)
	}

	err = updateWorkflowStaticData(ctx, workflowID, workflowEntity.StaticData)
	if err != nil {
		Errorf("Failed to update workflow static data %v", err)
		return err
	}
	return nil
}

func updateWorkflowStaticData(ctx context.Context, workflowId string, staticData map[string]interface{}) error {
	staticDataJson, err := json.Marshal(staticData)
	if err != nil {
		return err
	}
	_, err = GetRdsDbQueries().UpdateWorkflowEntityStaticDataByID(
		ctx,
		rdsDbLib.UpdateWorkflowEntityStaticDataByIDParams{
			ID: workflowId,
			StaticData: pqtype.NullRawMessage{
				RawMessage: staticDataJson,
				Valid:      true,
			},
		})
	if err != nil {
		return err
	}

	return nil
}

// Unregister all webhooks of the given workflow
func UnregisterWebhook(ctx context.Context, workflowID string, isTest bool) error {
	workflowEntity, err := GetWorkflowEntityById(ctx, workflowID)
	if err != nil {
		Errorf("Failed to get workflow entity %v", err)
		return err
	}

	if workflowEntity == nil {
		return nil
	}

	webhooks := GetWorkflowWebhooks(workflowEntity, isTest)
	if len(webhooks) == 0 {
		return nil
	}
	for _, webhook := range webhooks {
		CallWebhookDeleteMethod(ctx, &webhook, workflowEntity)
		err := DeleteWebhookEntity(ctx, &webhook)
		if err != nil {
			Errorf("delete webhook entity error. workflowId:%s webhookPath:%s method:%s",
				workflowID, webhook.Path, webhook.HttpMethod)
		}
	}

	err = updateWorkflowStaticData(ctx, workflowID, workflowEntity.StaticData)
	if err != nil {
		Errorf("Failed to update workflow static data %v", err)
		return err
	}

	return nil
}

// Register all webhooks of all active workflows. Regardless the orgId.
func RegisterAllWebhooks(ctx context.Context) error {
	workflowEntities, err := ListAllActiveWorkflowEntities(ctx)
	if err != nil {
		Errorf(fmt.Sprintf("Failed to list all active workflow entities: %v", err))
		return err
	}
	for _, workflowEntity := range workflowEntities {
		err := RegisterWebhook(ctx, workflowEntity.ID, false)
		if err != nil {
			Errorf(fmt.Sprintf("Failed to register webhook: %v", err))
			// Don't return error here, continue to register webhooks for other workflows.
		}
	}
	return nil
}

// Unregister all webhooks, including test-webhooks, ignoring the workflow status.
func UnregisterAllWebhooks(ctx context.Context) error {
	// Destroy webhooks by calling their delete webhook method, for example the SugerNotificationEventTrigger node
	// should unsubscribe the AWS SNS topic.
	// The workflow_entity.staticData will be updated too because some webhook data kept in it.
	// 1. Get all workflowIds in table webhook_entity
	workflowIds, err := GetRdsDbQueries().ListDistinctWorkflowIdsFromWebhookEntities(ctx)
	if err != nil {
		Errorf(fmt.Sprintf("Failed to list all webhook workflowIds: %v", err))
		return err
	}
	if len(workflowIds) == 0 {
		return nil
	}
	// 2. Unregister webhook and test-webhook of each workflow
	for _, workflowId := range workflowIds {
		err := UnregisterWebhook(ctx, workflowId, true)
		if err != nil {
			Errorf(fmt.Sprintf("Failed to unregister test webhook of %s: %v", workflowId, err))
			// Don't return error here, continue to unregister webhooks for other workflows.
		}
		err = UnregisterWebhook(ctx, workflowId, false)
		if err != nil {
			Errorf(fmt.Sprintf("Failed to unregister webhook of %s: %v", workflowId, err))
			// Don't return error here, continue to unregister webhooks for other workflows.
		}
	}
	// 3. Delete all webhook entities from db. Just to confirm again, because all entities should
	// already been deleted in previous step of UnregisterWebhook.
	return GetRdsDbQueries().DeleteAllWebhookEntities(ctx)
}

// Register test webhooks if a workflow contains webhook when manuallyRun.
// Return true if test webhooks are registered, otherwise return false.
func RegisterTestWebhooksIfAny(ctx context.Context, workflowEntity *structs.WorkflowEntity) bool {
	webhooks := GetWorkflowWebhooks(workflowEntity, true)
	if len(webhooks) == 0 {
		return false
	}

	// If only Wait node, it will return false.
	if !checkStartWebhook(webhooks) {
		return false
	}

	// Register test webhook
	err := RegisterWebhook(ctx, workflowEntity.ID, true)
	if err != nil {
		Errorf("Failed to register test webhook: %v", err)
		return false
	}

	return true
}

func checkStartWebhook(webhooks []structs.WebhookData) bool {
	for _, webhook := range webhooks {
		if !webhook.WebhookDescription.RestartWebhook {
			return true
		}
	}
	return false
}
