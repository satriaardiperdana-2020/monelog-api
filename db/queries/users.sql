-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, timezone, currency)
VALUES (
    sqlc.arg(id),
    lower(btrim(sqlc.arg(email))),
    sqlc.arg(password_hash),
    sqlc.arg(timezone),
    sqlc.arg(currency)
)
RETURNING *;

-- name: GetActiveUserByID :one
SELECT *
FROM users
WHERE id = sqlc.arg(id)
  AND is_delete = FALSE;

-- name: GetActiveUserByEmail :one
SELECT *
FROM users
WHERE email = lower(btrim(sqlc.arg(email)))
  AND is_delete = FALSE;

-- name: LockUsersForUpdate :many
SELECT *
FROM users
WHERE id = ANY(sqlc.arg(ids)::uuid[])
ORDER BY id
FOR UPDATE;

-- name: UpdateUserProfile :one
UPDATE users
SET email = lower(btrim(sqlc.arg(email))),
    timezone = sqlc.arg(timezone),
    currency = sqlc.arg(currency),
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
  AND version = sqlc.arg(expected_version)
  AND is_delete = FALSE
RETURNING *;

-- name: SoftDeleteUser :one
UPDATE users
SET is_delete = TRUE,
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
  AND version = sqlc.arg(expected_version)
  AND is_delete = FALSE
RETURNING *;

-- name: RestoreUser :one
UPDATE users
SET is_delete = FALSE,
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
  AND version = sqlc.arg(expected_version)
  AND is_delete = TRUE
RETURNING *;
