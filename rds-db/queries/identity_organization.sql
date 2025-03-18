-- name: GetOrganizationById :one
SELECT * FROM identity.organization WHERE id = $1 AND status != 'DELETED' LIMIT 1;

-- name: GetOrganizationByClientSignupId :one
SELECT * FROM identity.organization WHERE info->'clientSignupPageConfigInfo'->>'signupId' = @signup_id AND status != 'DELETED' LIMIT 1;

-- name: CreateOrganization :one
INSERT INTO identity.organization (id, name, email_domain, website, description, allowed_auth_methods, created_by, auth_id, status)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;

-- name: CreateOrganizationWithInfo :one
INSERT INTO identity.organization (id, name, email_domain, website, description, allowed_auth_methods, created_by, auth_id, status, info)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: UpdateOrganization :one
UPDATE identity.organization SET name = $2, website = $3, description = $4, allowed_auth_methods = $5, last_update_time = CURRENT_TIMESTAMP WHERE id = $1 AND status != 'DELETED' RETURNING *;

-- name: GetOrganizationInfo :one
SELECT info FROM identity.organization WHERE id = $1 AND status != 'DELETED' LIMIT 1;

-- name: UpdateOrganizationInfo :one
UPDATE identity.organization SET info = $2 WHERE id = $1 AND status != 'DELETED' RETURNING *;

-- name: UpdateOrganizationConfigInfo :one
UPDATE identity.organization SET info = jsonb_set(info, '{organizationConfigInfo}', @organization_config_info::jsonb), last_update_time = CURRENT_TIMESTAMP
    WHERE id = $1 AND status != 'DELETED' RETURNING *;

-- name: UpdateCosellConfigInfo :one
UPDATE identity.organization SET info = jsonb_set(info, '{cosellConfigInfoV2}', @cosell_config_info::jsonb), last_update_time = CURRENT_TIMESTAMP
    WHERE id = $1 AND status != 'DELETED' RETURNING *;

-- name: UpdateCosellFillerConfigInfo :one
UPDATE identity.organization SET info = jsonb_set(info, '{cosellFillerConfigInfo}', @cosell_filler_config_info::jsonb), last_update_time = CURRENT_TIMESTAMP
    WHERE id = $1 AND status != 'DELETE' RETURNING *;

-- name: UpdateUsageMeteringConfigInfo :one
UPDATE identity.organization SET info = jsonb_set(info, '{usageMeteringConfigInfo}', @usage_metering_config_info::jsonb), last_update_time = CURRENT_TIMESTAMP
    WHERE id = $1 AND status != 'DELETED' RETURNING *;

-- name: UpdateClientSignupPageConfigInfo :one
UPDATE identity.organization SET info = jsonb_set(info, '{clientSignupPageConfigInfo}', @client_signup_page_config_info::jsonb), last_update_time = CURRENT_TIMESTAMP
    WHERE id = $1 AND status != 'DELETED' RETURNING *;

-- name: UpdateNotificationConfigInfo :one
UPDATE identity.organization SET info = jsonb_set(info, '{notificationConfigInfo}', @notification_config_info::jsonb), last_update_time = CURRENT_TIMESTAMP
    WHERE id = $1 AND status != 'DELETED' RETURNING *;

-- name: UpdateOfferConfigInfos :one
UPDATE identity.organization SET info = jsonb_set(info, '{offerConfigInfos}', @offer_config_infos::jsonb), last_update_time = CURRENT_TIMESTAMP
    WHERE id = $1 AND status != 'DELETED' RETURNING *;

-- name: SoftDeleteOrganization :one
UPDATE identity.organization SET status = 'DELETED' WHERE id = $1 RETURNING *;

-- name: ListAllOrganizations :many
SELECT * FROM identity.organization WHERE status != 'DELETED';

-- name: ListAllActiveOrganizations :many
SELECT * FROM identity.organization WHERE status = 'ACTIVE';

-- name: UpdateOrganizationStatus :one
UPDATE identity.organization SET status = $2 WHERE id = $1 AND status != 'DELETED' RETURNING *;
