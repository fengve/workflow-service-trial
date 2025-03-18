-- name: CreateRole :one
INSERT INTO identity.role (organization_id, id, name, description, permissions) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetRole :one
SELECT * FROM identity.role WHERE organization_id = $1 and id = $2 LIMIT 1;

-- name: UpdateRole :one
UPDATE identity.role SET name = $1, description = $2, permissions = $3 WHERE organization_id = $4 and id = $5 RETURNING *;

-- name: UpdateRoleNameAndDescription :one
UPDATE identity.role SET name = $1, description = $2 WHERE organization_id = $3 and id = $4 RETURNING *;

-- name: UpdateRolePermissions :one
UPDATE identity.role SET permissions = $1 WHERE organization_id = $2 and id = $3 RETURNING *;

-- name: ListRoles :many
SELECT * FROM identity.role WHERE organization_id = $1;