-- name: GetWorkflowExecutionData :one
SELECT * FROM workflow.execution_data WHERE "executionId" = $1;

-- name: CreateWorkflowExecutionData :one
INSERT INTO workflow.execution_data("executionId", "workflowData", data)
    VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteWorkflowExecutionData :exec
DELETE FROM workflow.execution_data WHERE "executionId" = $1 AND  "workflowData"->>'id'::text = @workflow_id::text;

-- name: BatchDeleteWorkflowExecutionData :exec
DELETE FROM workflow.execution_data WHERE "workflowData"->>'id'::text = @workflow_id::text AND "executionId" = ANY(@execution_ids::integer[]);

-- name: UpdateWorkflowExecutionData :one
UPDATE workflow.execution_data SET "workflowData" = $2, data = $3
    WHERE "executionId" = $1 RETURNING *;