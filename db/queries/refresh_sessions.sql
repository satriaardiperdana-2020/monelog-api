-- Create a refresh session after successful login or refresh rotation.
-- name: CreateRefreshSession :one
INSERT INTO refresh_sessions (
    user_id,
    family_id,
    token_hash,
    expires_at
)
VALUES (
    sqlc.arg(user_id),
    COALESCE(sqlc.narg(family_id)::bigint, nextval('refresh_session_families_seq')),
    sqlc.arg(token_hash),
    sqlc.arg(expires_at)
)
RETURNING *;

-- Find a refresh session for validation without taking a write lock.
-- name: GetRefreshSessionByTokenHash :one
SELECT *
FROM refresh_sessions
WHERE token_hash = sqlc.arg(token_hash);

-- Find and lock a refresh session before rotating or revoking it.
-- name: GetRefreshSessionByTokenHashForUpdate :one
SELECT *
FROM refresh_sessions
WHERE token_hash = sqlc.arg(token_hash)
FOR UPDATE;

-- Revoke one active refresh session and optionally link its replacement.
-- name: RevokeRefreshSession :one
UPDATE refresh_sessions
SET revoked_at = COALESCE(revoked_at, CURRENT_TIMESTAMP),
    replaced_by = sqlc.narg(replaced_by)
WHERE id = sqlc.arg(id)
  AND revoked_at IS NULL
RETURNING *;

-- Revoke every active session in one refresh-token family during logout or reuse detection.
-- name: RevokeRefreshSessionFamily :execrows
UPDATE refresh_sessions
SET revoked_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)
  AND family_id = sqlc.arg(family_id)
  AND revoked_at IS NULL;

-- Revoke all active refresh sessions for a user during account deletion or security action.
-- name: RevokeAllUserRefreshSessions :execrows
UPDATE refresh_sessions
SET revoked_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL;
