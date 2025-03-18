-- name: GetUserById :one
SELECT * FROM identity.user WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM identity.user WHERE LOWER(email) = LOWER(@email) LIMIT 1;

-- name: CreateUser :one
INSERT INTO identity.user AS u(id, first_name, last_name, email)
    VALUES ($1, $2, $3, $4) 
    ON CONFLICT (email) 
    DO UPDATE 
        SET last_update_time = CURRENT_TIMESTAMP WHERE u.email = $4
    RETURNING *;

-- name: UpdateUser :one
UPDATE identity.user SET first_name = $2, last_name = $3, last_update_time = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;