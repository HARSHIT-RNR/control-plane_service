-- --- Role Management Queries ---

-- name: CreateRole :one
INSERT INTO roles (id, tenant_id, name, description, permissions)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM roles
WHERE id = $1;

-- name: ListRoles :many
SELECT * FROM roles
WHERE tenant_id = $1
ORDER BY name;

-- name: UpdateRole :one
UPDATE roles
SET name = $2, description = $3, permissions = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1;


-- --- RBAC & Authorization Queries ---

-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, tenant_id, role_id)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: RevokeRoleFromUser :exec
DELETE FROM user_roles
WHERE user_id = $1 AND role_id = $2;

-- name: GetUserRoles :many
SELECT r.* FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1 AND ur.tenant_id = $2;

