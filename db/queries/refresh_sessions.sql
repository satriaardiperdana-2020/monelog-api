-- name: CreateRefreshSession :one
INSERT INTO refresh_sessions (
    id,
    user_id,
    family_id,
    token_hash,
    expires_at
)
VALUES (
    sqlc.arg(id),
    sqlc.arg(user_id),
    sqlc.arg(family_id),
    sqlc.arg(token_hash),
    sqlc.arg(expires_at)
)
RETURNING *;

-- name: GetRefreshSessionByTokenHash :one
SELECT *
FROM refresh_sessions
WHERE token_hash = sqlc.arg(token_hash);

-- name: GetRefreshSessionByTokenHashForUpdate :one
SELECT *
FROM refresh_sessions
WHERE token_hash = sqlc.arg(token_hash)
FOR UPDATE;

-- name: RevokeRefreshSession :one
UPDATE refresh_sessions
SET revoked_at = COALESCE(revoked_at, CURRENT_TIMESTAMP),
    replaced_by = sqlc.narg(replaced_by)
WHERE id = sqlc.arg(id)
  AND revoked_at IS NULL
RETURNING *;

-- name: RevokeRefreshSessionFamily :execrows
UPDATE refresh_sessions
SET revoked_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)
  AND family_id = sqlc.arg(family_id)
  AND revoked_at IS NULL;

-- name: RevokeAllUserRefreshSessions :execrows
UPDATE refresh_sessions
SET revoked_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL;
