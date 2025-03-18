-- name: ListWorkflowExecutionEntitiesByWorkflowId :many
SELECT * FROM workflow.execution_entity WHERE "workflowId" = $1 ORDER BY "startedAt" DESC LIMIT $2 OFFSET $3;

-- name: CountWorkflowExecutionEntitiesByWorkflowId :one
SELECT count(*) FROM workflow.execution_entity WHERE "workflowId" = $1;

-- name: GetWorkflowExecutionEntity :one
SELECT * FROM workflow.execution_entity WHERE id = $1;

-- name: CreateWorkflowExecutionEntity :one
INSERT INTO workflow.execution_entity(finished, mode, "retryOf", "retrySuccessId", "startedAt", "stoppedAt", "waitTill", status, "workflowId", "deletedAt")
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: DeleteWorkflowExecutionEntity :exec
DELETE FROM workflow.execution_entity WHERE "workflowId" = $1 AND id = $2;

-- name: BatchDeleteWorkflowExecutionEntities :exec
DELETE FROM workflow.execution_entity WHERE "workflowId" = $1 AND id = ANY(@execution_ids::integer[]);

-- name: UpdateWorkflowExecutionEntity :one
UPDATE workflow.execution_entity SET finished = $2, mode = $3, "retryOf" = $4, "retrySuccessId" = $5, "stoppedAt" = $6, "waitTill" = $7, status = $8
    WHERE id = $1 RETURNING *;