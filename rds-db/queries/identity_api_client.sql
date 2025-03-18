-- name: GetApiClient :one
SELECT * FROM identity.api_client WHERE organization_id = $1 AND id = $2 LIMIT 1;

-- name: GetApiClientById :one
SELECT * FROM identity.api_client WHERE id = $1 LIMIT 1;

-- name: GetApiClientByApiKeyHash :one
SELECT * FROM identity.api_client WHERE api_key_hash = $1 LIMIT 1;

-- name: ListApiClientsByOrganization :many
SELECT * FROM identity.api_client WHERE organization_id = $1;

-- name: CreateApiClient :one
INSERT INTO identity.api_client(id, organization_id, provider, info, role, type, api_key_hash)
    VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: UpdateApiClientApiKeyHash :one
UPDATE identity.api_client SET api_key_hash = $3 WHERE organization_id = $1 AND id = $2 RETURNING *;

-- name: DeleteApiClient :exec
DELETE FROM identity.api_client WHERE organization_id = $1 AND id = $2;

-- name: DeleteAllApiClients :exec
DELETE FROM identity.api_client WHERE organization_id = $1;