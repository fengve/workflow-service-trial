-- name: GetIntegration :one
SELECT * FROM identity.integration WHERE organization_id = $1 AND partner = $2 AND service = $3 LIMIT 1;

-- name: ListIntegrations :many
SELECT * FROM identity.integration WHERE organization_id = $1 ORDER BY creation_time;

-- name: ListIntegrationsByPartnerServiceAndStatus :many
SELECT * FROM identity.integration WHERE partner = $1 AND service = $2 AND status = $3 ORDER BY creation_time;

-- name: ListIntegrationsByServiceAndStatus :many
SELECT * FROM identity.integration WHERE service = $1 AND status = $2 ORDER BY creation_time;

-- name: ListAzureIntegrationsByPublisherId :many
SELECT * FROM identity.integration WHERE partner = 'AZURE' AND info->'azureIntegration'->>'publisherID' = @publisher_id;

-- name: ListGcpIntegrationsByProjectNumber :many
SELECT * FROM identity.integration WHERE partner = 'GCP' AND info->'gcpIntegration'->>'gcpProjectNumber' = @project_number::text;

-- name: CreateIntegration :one
INSERT INTO identity.integration (organization_id, partner, service, status, info, created_by, last_updated_by)
    VALUES ($1, $2, $3, $4, $5, $6, $6) RETURNING *;

-- name: UpdateIntegrationInfo :one
UPDATE identity.integration SET info = $4, last_update_time = CURRENT_TIMESTAMP
    WHERE organization_id = $1 AND partner = $2 AND service = $3 RETURNING *;

-- name: UpdateIntegrationStatus :one
UPDATE identity.integration SET status = $4, last_update_time = CURRENT_TIMESTAMP, last_updated_by = $5
    WHERE organization_id = $1 AND partner = $2 AND service = $3 RETURNING *;

-- name: DeleteIntegration :exec
DELETE FROM identity.integration WHERE organization_id = $1 AND partner = $2 AND service = $3;

-- name: GetHubspotIntegrationByPortalId :many
SELECT * FROM identity.integration WHERE partner = 'HUBSPOT' AND service = 'CRM' AND info->'hubspotCrmIntegration'->>'portalId' = @portal_id;