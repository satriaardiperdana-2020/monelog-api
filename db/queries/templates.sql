-- name: LockScopedCategoryForTemplate :one
SELECT * FROM categories
WHERE id = sqlc.arg(category_id) AND user_id = sqlc.arg(user_id)
FOR UPDATE;

-- name: CreateTransactionTemplate :one
INSERT INTO transaction_templates (user_id, category_id, type, name, amount, title)
SELECT sqlc.arg(user_id), sqlc.arg(category_id), sqlc.arg(type),
       btrim(sqlc.arg(name)), sqlc.arg(amount), btrim(sqlc.arg(title))
FROM categories
WHERE id = sqlc.arg(category_id)
  AND user_id = sqlc.arg(user_id)
  AND type = sqlc.arg(type)
  AND is_delete = FALSE
RETURNING transaction_templates.*;

-- name: GetActiveTransactionTemplate :one
SELECT * FROM transaction_templates
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND is_delete = FALSE;

-- name: GetDeletedTransactionTemplate :one
SELECT * FROM transaction_templates
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND is_delete = TRUE;

-- name: GetTransactionTemplateStateForUpdate :one
SELECT * FROM transaction_templates
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
FOR UPDATE;

-- name: ListActiveTransactionTemplates :many
SELECT * FROM transaction_templates
WHERE user_id = sqlc.arg(user_id)
  AND is_delete = FALSE
ORDER BY lower(name), id;

-- name: ListDeletedTransactionTemplates :many
SELECT * FROM transaction_templates
WHERE user_id = sqlc.arg(user_id)
  AND is_delete = TRUE
ORDER BY updated_at DESC, id DESC;

-- name: UpdateTransactionTemplate :one
UPDATE transaction_templates
SET category_id = sqlc.arg(category_id),
    type = sqlc.arg(type),
    name = btrim(sqlc.arg(name)),
    amount = sqlc.arg(amount),
    title = btrim(sqlc.arg(title)),
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE transaction_templates.id = sqlc.arg(id)
  AND transaction_templates.user_id = sqlc.arg(user_id)
  AND transaction_templates.version = sqlc.arg(expected_version)
  AND transaction_templates.is_delete = FALSE
  AND EXISTS (
      SELECT 1 FROM categories
      WHERE categories.id = sqlc.arg(category_id)
        AND categories.user_id = sqlc.arg(user_id)
        AND categories.type = sqlc.arg(type)
        AND categories.is_delete = FALSE
  )
RETURNING *;

-- name: SoftDeleteTransactionTemplate :one
UPDATE transaction_templates
SET is_delete = TRUE,
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND version = sqlc.arg(expected_version)
  AND is_delete = FALSE
RETURNING *;

-- name: RestoreTransactionTemplate :one
UPDATE transaction_templates
SET is_delete = FALSE,
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE transaction_templates.id = sqlc.arg(id)
  AND transaction_templates.user_id = sqlc.arg(user_id)
  AND transaction_templates.version = sqlc.arg(expected_version)
  AND transaction_templates.is_delete = TRUE
  AND EXISTS (
      SELECT 1 FROM categories
      WHERE categories.id = transaction_templates.category_id
        AND categories.user_id = transaction_templates.user_id
        AND categories.type = transaction_templates.type
        AND categories.is_delete = FALSE
  )
RETURNING *;
