-- --- User Management Queries ---

-- name: CreateUser :one
INSERT INTO users (
    full_name, email, tenant_id, employee_id, department_id, designation_id, job_title, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: CreateInitialAdmin :one
-- Creates the first admin user for a tenant, setting their status to PENDING_SETUP.
INSERT INTO users (
    full_name, email, tenant_id, status
) VALUES (
    $1, $2, $3, 'PENDING_SETUP'
) RETURNING *;


-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmailAndTenant :one
-- Used to find a user during login or before inviting a new one.
SELECT * FROM users
WHERE email = $1 AND tenant_id = $2;

-- name: ListUsers :many
SELECT * FROM users
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: UpdateUser :one
UPDATE users
SET
    full_name = $2,
    employee_id = $3,
    department_id = $4,
    designation_id = $5,
    job_title = $6
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: ActivateInvitedUser :exec
-- Updates a user's status from PENDING_INVITE to ACTIVE and sets their name.
UPDATE users
SET status = 'ACTIVE', full_name = $2
WHERE id = $1;

-- name: UpdateUserStatus :exec
UPDATE users
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateLastLogin :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND tenant_id = $2;
