-- name: CreateCategory :one
INSERT INTO categories (id, user_id, type, name)
VALUES (sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(type), btrim(sqlc.arg(name)))
RETURNING *;

-- name: GetActiveCategory :one
SELECT *
FROM categories
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND is_delete = FALSE;

-- name: GetCategoryByID :one
SELECT *
FROM categories
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id);

-- name: GetDeletedCategory :one
SELECT *
FROM categories
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND is_delete = TRUE;

-- name: ListActiveCategories :many
SELECT *
FROM categories
WHERE user_id = sqlc.arg(user_id)
  AND is_delete = FALSE
ORDER BY type, lower(name), id;

-- name: ListDeletedCategories :many
SELECT *
FROM categories
WHERE user_id = sqlc.arg(user_id)
  AND is_delete = TRUE
ORDER BY updated_at DESC, id DESC;

-- name: UpdateCategory :one
UPDATE categories
SET name = btrim(sqlc.arg(name)),
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND version = sqlc.arg(expected_version)
  AND is_delete = FALSE
RETURNING *;

-- name: SoftDeleteCategory :one
UPDATE categories
SET is_delete = TRUE,
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND version = sqlc.arg(expected_version)
  AND is_delete = FALSE
RETURNING *;

-- name: RestoreCategory :one
UPDATE categories
SET is_delete = FALSE,
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND version = sqlc.arg(expected_version)
  AND is_delete = TRUE
RETURNING *;
