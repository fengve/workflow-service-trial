-- name: GetWorkflowEntity :one
SELECT * FROM workflow.workflow_entity WHERE "sugerOrgId" = $1 and id = $2;

-- name: GetWorkflowEntityById :one
SELECT * FROM workflow.workflow_entity WHERE id = $1;

-- name: ListWorkflowEntities :many
SELECT * FROM workflow.workflow_entity WHERE "sugerOrgId" = $1;

-- name: ListActiveWorkflowEntities :many
SELECT * FROM workflow.workflow_entity WHERE "sugerOrgId" = $1 AND active = true;

-- name: ListAllActiveWorkflowEntities :many
SELECT * FROM workflow.workflow_entity WHERE active = true;

-- name: CreateWorkflowEntity :one
INSERT INTO workflow.workflow_entity(name, active, nodes, connections, settings, "staticData", "pinData", "versionId", "triggerCount", id, meta, "sugerOrgId")
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING *;

-- name: UpdateWorkflowEntity :one
UPDATE workflow.workflow_entity SET name = $3, active = $4, nodes = $5, connections = $6, settings = $7, "staticData" = $8, "pinData" = $9, "versionId" = $10, "triggerCount" = $11, meta = $12, "updatedAt" = CURRENT_TIMESTAMP
    WHERE "sugerOrgId" = $1 and id = $2 RETURNING *;

-- name: UpdateWorkflowEntityName :one
UPDATE workflow.workflow_entity SET name = $3, "updatedAt" = CURRENT_TIMESTAMP WHERE "sugerOrgId" = $1 and id = $2 RETURNING *;

-- name: UpdateWorkflowEntityActive :one
UPDATE workflow.workflow_entity SET active = $3, "updatedAt" = CURRENT_TIMESTAMP WHERE "sugerOrgId" = $1 and id = $2 RETURNING *;

-- name: UpdateWorkflowEntityNodes :one
UPDATE workflow.workflow_entity SET nodes = $3, "updatedAt" = CURRENT_TIMESTAMP WHERE "sugerOrgId" = $1 and id = $2 RETURNING *;

-- name: UpdateWorkflowEntityConnections :one
UPDATE workflow.workflow_entity SET connections = $3, "updatedAt" = CURRENT_TIMESTAMP WHERE "sugerOrgId" = $1 and id = $2 RETURNING *;

-- name: UpdateWorkflowEntitySettings :one
UPDATE workflow.workflow_entity SET settings = $3, "updatedAt" = CURRENT_TIMESTAMP WHERE "sugerOrgId" = $1 and id = $2 RETURNING *;

-- name: UpdateWorkflowEntityStaticData :one
UPDATE workflow.workflow_entity SET "staticData" = $3, "updatedAt" = CURRENT_TIMESTAMP WHERE "sugerOrgId" = $1 and id = $2 RETURNING *;

-- name: UpdateWorkflowEntityStaticDataByID :one
UPDATE workflow.workflow_entity SET "staticData" = $2, "updatedAt" = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;

-- name: UpdateWorkflowEntityPinData :one
UPDATE workflow.workflow_entity SET "pinData" = $3, "updatedAt" = CURRENT_TIMESTAMP WHERE "sugerOrgId" = $1 and id = $2 RETURNING *;

-- name: UpdateWorkflowEntityVersionId :one
UPDATE workflow.workflow_entity SET "versionId" = $3, "updatedAt" = CURRENT_TIMESTAMP WHERE "sugerOrgId" = $1 and id = $2 RETURNING *;

-- name: UpdateWorkflowEntityTriggerCount :one
UPDATE workflow.workflow_entity SET "triggerCount" = $3, "updatedAt" = CURRENT_TIMESTAMP WHERE "sugerOrgId" = $1 and id = $2 RETURNING *;

-- name: UpdateWorkflowEntityMeta :one
UPDATE workflow.workflow_entity SET meta = $3, "updatedAt" = CURRENT_TIMESTAMP WHERE "sugerOrgId" = $1 and id = $2 RETURNING *;

-- name: DeleteWorkflowEntity :one
DELETE FROM workflow.workflow_entity WHERE "sugerOrgId" = $1 and id = $2 RETURNING *;
