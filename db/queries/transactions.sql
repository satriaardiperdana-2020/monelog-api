-- name: CreateTransaction :one
INSERT INTO transactions (
    id,
    user_id,
    category_id,
    type,
    amount,
    transaction_date,
    title,
    client_request_id,
    request_hash,
    created_by,
    updated_by
)
SELECT
    sqlc.arg(id),
    sqlc.arg(user_id),
    sqlc.arg(category_id),
    sqlc.arg(type),
    sqlc.arg(amount),
    sqlc.arg(transaction_date),
    btrim(sqlc.arg(title)),
    sqlc.arg(client_request_id),
    sqlc.arg(request_hash),
    sqlc.arg(actor_user_id),
    sqlc.arg(actor_user_id)
FROM categories
WHERE categories.id = sqlc.arg(category_id)
  AND categories.user_id = sqlc.arg(user_id)
  AND categories.type = sqlc.arg(type)
  AND categories.is_delete = FALSE
RETURNING transactions.*;

-- name: GetTransactionByRequestID :one
SELECT *
FROM transactions
WHERE user_id = sqlc.arg(user_id)
  AND client_request_id = sqlc.arg(client_request_id);

-- name: GetActiveTransaction :one
SELECT *
FROM transactions
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND is_delete = FALSE;

-- name: ListActiveTransactions :many
SELECT *
FROM transactions
WHERE user_id = sqlc.arg(user_id)
  AND is_delete = FALSE
ORDER BY transaction_date DESC, id DESC
LIMIT sqlc.arg(page_size);

-- name: ListDeletedTransactions :many
SELECT *
FROM transactions
WHERE user_id = sqlc.arg(user_id)
  AND is_delete = TRUE
ORDER BY updated_at DESC, id DESC
LIMIT sqlc.arg(page_size);

-- name: UpdateTransaction :one
UPDATE transactions
SET category_id = sqlc.arg(category_id),
    type = sqlc.arg(type),
    amount = sqlc.arg(amount),
    transaction_date = sqlc.arg(transaction_date),
    title = btrim(sqlc.arg(title)),
    updated_by = sqlc.arg(actor_user_id),
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE transactions.id = sqlc.arg(id)
  AND transactions.user_id = sqlc.arg(user_id)
  AND transactions.version = sqlc.arg(expected_version)
  AND transactions.is_delete = FALSE
  AND EXISTS (
      SELECT 1
      FROM categories
      WHERE categories.id = sqlc.arg(category_id)
        AND categories.user_id = sqlc.arg(user_id)
        AND categories.type = sqlc.arg(type)
        AND categories.is_delete = FALSE
  )
RETURNING *;

-- name: SoftDeleteTransaction :one
UPDATE transactions
SET is_delete = TRUE,
    updated_by = sqlc.arg(actor_user_id),
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND version = sqlc.arg(expected_version)
  AND is_delete = FALSE
RETURNING *;

-- name: RestoreTransaction :one
UPDATE transactions
SET is_delete = FALSE,
    updated_by = sqlc.arg(actor_user_id),
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE transactions.id = sqlc.arg(id)
  AND transactions.user_id = sqlc.arg(user_id)
  AND transactions.version = sqlc.arg(expected_version)
  AND transactions.is_delete = TRUE
  AND EXISTS (
      SELECT 1
      FROM categories
      WHERE categories.id = transactions.category_id
        AND categories.user_id = transactions.user_id
        AND categories.type = transactions.type
        AND categories.is_delete = FALSE
  )
RETURNING *;
