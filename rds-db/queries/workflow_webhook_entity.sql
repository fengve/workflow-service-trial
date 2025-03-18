-- name: ListWebhookEntities :many
SELECT * FROM workflow.webhook_entity WHERE "workflowId" = $1 AND "webhookId" = $2;

-- name: ListDistinctWorkflowIdsFromWebhookEntities :many
SELECT DISTINCT workflow.webhook_entity."workflowId" FROM workflow.webhook_entity;

-- name: CreateWebhookEntity :one
INSERT INTO workflow.webhook_entity("webhookPath", method, node, "webhookId", "pathLength", "workflowId")
    VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: DeleteWebhookEntitiesByWorkflowId :exec
DELETE FROM workflow.webhook_entity WHERE "workflowId" = $1;

-- name: DeleteWebhookEntityByWorkflowId_Path_Method :exec
DELETE FROM workflow.webhook_entity WHERE "workflowId" = $1 AND "webhookPath" = $2 AND "method" = $3;

-- name: DeleteAllWebhookEntities :exec
DELETE FROM workflow.webhook_entity;