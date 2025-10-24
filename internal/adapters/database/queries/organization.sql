-- --- Department Management Queries ---

-- name: CreateDepartment :one
INSERT INTO departments (id, name, description, tenant_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetDepartment :one
SELECT * FROM departments
WHERE id = $1;

-- name: ListDepartments :many
SELECT * FROM departments
WHERE tenant_id = $1
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: CountDepartments :one
SELECT COUNT(*) FROM departments
WHERE tenant_id = $1;

-- name: UpdateDepartment :one
UPDATE departments
SET name = $2, description = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteDepartment :exec
DELETE FROM departments
WHERE id = $1;


-- --- Designation Management Queries ---

-- name: CreateDesignation :one
INSERT INTO designations (id, name, description, tenant_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetDesignation :one
SELECT * FROM designations
WHERE id = $1;

-- name: ListDesignations :many
SELECT * FROM designations
WHERE tenant_id = $1
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: CountDesignations :one
SELECT COUNT(*) FROM designations
WHERE tenant_id = $1;

-- name: UpdateDesignation :one
UPDATE designations
SET name = $2, description = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteDesignation :exec
DELETE FROM designations
WHERE id = $1;

