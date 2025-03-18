-- name: AddUserToOrganization :one
INSERT INTO identity.user_organization AS uo(user_id, organization_id, user_role, allowed_auth_methods)
    VALUES ($1, $2, $3, $4) ON CONFLICT (user_id, organization_id) 
    DO UPDATE 
        SET user_role = $3, allowed_auth_methods = $4, last_update_time = CURRENT_TIMESTAMP WHERE uo.user_id = $1 AND uo.organization_id = $2
    RETURNING *;

-- name: UpdateUserInOrganization :one
UPDATE identity.user_organization SET user_role = $3, allowed_auth_methods = $4, last_update_time = CURRENT_TIMESTAMP
    WHERE user_id = $1 AND organization_id = $2 RETURNING *;

-- name: DeleteUserFromOrganization :exec
DELETE FROM identity.user_organization WHERE user_id = $1 AND organization_id = $2;

-- name: DeleteAllUsersFromOrganization :exec
DELETE FROM identity.user_organization WHERE organization_id = $1;

-- name: ListOrganizationsByUser :many
SELECT o.*, uo.user_role, uo.creation_time AS join_organization_time
    FROM identity.user_organization uo INNER JOIN identity.organization o ON uo.organization_id = o.id 
    WHERE uo.user_id = $1 AND o.status = ANY(@status::text[])  ORDER BY uo.creation_time;

-- name: ListUsersByOrganization :many
SELECT u.*, uo.user_role, uo.creation_time AS join_organization_time
    FROM identity.user_organization uo INNER JOIN identity.user u ON uo.user_id = u.id 
    WHERE uo.organization_id = $1 ORDER BY uo.creation_time;

-- name: GetUserRoleByUserAndOrganization :one
SELECT user_role FROM identity.user_organization WHERE user_id = $1 AND organization_id = $2;