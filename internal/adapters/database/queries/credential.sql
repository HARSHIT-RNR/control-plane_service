-- --- Credential Queries ---

-- name: CreateCredential :one
INSERT INTO credentials (user_id, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetCredential :one
SELECT * FROM credentials
WHERE user_id = $1;

-- name: UpdateCredential :exec
UPDATE credentials
SET password_hash = $2, updated_at = NOW()
WHERE user_id = $1;

-- name: DeleteCredential :exec
DELETE FROM credentials
WHERE user_id = $1;

-- name: GetUserAndCredentialByEmail :one
SELECT u.id as user_id, u.status, c.password_hash
FROM users u
JOIN credentials c ON u.id = c.user_id
WHERE u.email = $1 AND u.tenant_id = $2;


-- --- Token Queries (for password reset, email verification etc.) ---

-- name: CreateToken :exec
INSERT INTO tokens (hash, user_id, expiry, scope)
VALUES ($1, $2, $3, $4);

-- name: GetToken :one
SELECT * FROM tokens
WHERE hash = $1;

-- name: DeleteToken :exec
DELETE FROM tokens
WHERE hash = $1;

-- name: DeleteExpiredTokens :exec
DELETE FROM tokens
WHERE expiry < NOW();

-- name: DeleteTokensByUserIDAndScope :exec
DELETE FROM tokens
WHERE user_id = $1 AND scope = $2;
