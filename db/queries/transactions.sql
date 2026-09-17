-- name: LockScopedCategoryForTransaction :one
SELECT * FROM categories
WHERE id = sqlc.arg(category_id) AND user_id = sqlc.arg(user_id)
FOR UPDATE;

-- name: CreateTransaction :one
INSERT INTO transactions (
    id, user_id, category_id, type, amount, transaction_date, title,
    client_request_id, request_hash, created_by, updated_by
)
SELECT sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(category_id), sqlc.arg(type),
       sqlc.arg(amount), sqlc.arg(transaction_date), btrim(sqlc.arg(title)),
       sqlc.arg(client_request_id), sqlc.arg(request_hash),
       sqlc.arg(actor_user_id), sqlc.arg(actor_user_id)
FROM categories
WHERE categories.id = sqlc.arg(category_id)
  AND categories.user_id = sqlc.arg(user_id)
  AND categories.type = sqlc.arg(type)
  AND categories.is_delete = FALSE
ON CONFLICT (user_id, client_request_id) DO NOTHING
RETURNING transactions.*;

-- name: GetTransactionByRequestIDForUpdate :one
SELECT * FROM transactions
WHERE user_id = sqlc.arg(user_id) AND client_request_id = sqlc.arg(client_request_id)
FOR UPDATE;

-- name: GetTransactionStateForUpdate :one
SELECT * FROM transactions
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
FOR UPDATE;

-- name: GetActiveTransaction :one
SELECT t.*, c.name AS category_name
FROM transactions t
JOIN categories c ON c.id = t.category_id AND c.user_id = t.user_id AND c.type = t.type
WHERE t.id = sqlc.arg(id) AND t.user_id = sqlc.arg(user_id) AND t.is_delete = FALSE;

-- name: GetDeletedTransaction :one
SELECT t.*, c.name AS category_name
FROM transactions t
JOIN categories c ON c.id = t.category_id AND c.user_id = t.user_id AND c.type = t.type
WHERE t.id = sqlc.arg(id) AND t.user_id = sqlc.arg(user_id) AND t.is_delete = TRUE;

-- name: ListActiveTransactions :many
SELECT t.*, c.name AS category_name
FROM transactions t
JOIN categories c ON c.id = t.category_id AND c.user_id = t.user_id AND c.type = t.type
WHERE t.user_id = sqlc.arg(user_id)
  AND t.is_delete = FALSE
  AND t.transaction_date BETWEEN sqlc.arg(start_date) AND sqlc.arg(end_date)
  AND (sqlc.arg(filter_type)::text = '' OR t.type = sqlc.arg(filter_type))
  AND (NOT sqlc.arg(has_category)::boolean OR t.category_id = sqlc.arg(category_id))
  AND (NOT sqlc.arg(has_cursor)::boolean OR (t.transaction_date, t.created_at, t.id) <
      (sqlc.arg(cursor_date)::date, sqlc.arg(cursor_created_at)::timestamptz, sqlc.arg(cursor_id)::uuid))
ORDER BY t.transaction_date DESC, t.created_at DESC, t.id DESC
LIMIT sqlc.arg(page_size);

-- name: ListDeletedTransactions :many
SELECT t.*, c.name AS category_name
FROM transactions t
JOIN categories c ON c.id = t.category_id AND c.user_id = t.user_id AND c.type = t.type
WHERE t.user_id = sqlc.arg(user_id)
  AND t.is_delete = TRUE
  AND t.transaction_date BETWEEN sqlc.arg(start_date) AND sqlc.arg(end_date)
  AND (sqlc.arg(filter_type)::text = '' OR t.type = sqlc.arg(filter_type))
  AND (NOT sqlc.arg(has_category)::boolean OR t.category_id = sqlc.arg(category_id))
  AND (NOT sqlc.arg(has_cursor)::boolean OR (t.updated_at, t.id) <
      (sqlc.arg(cursor_updated_at)::timestamptz, sqlc.arg(cursor_id)::uuid))
ORDER BY t.updated_at DESC, t.id DESC
LIMIT sqlc.arg(page_size);

-- name: UpdateTransaction :one
UPDATE transactions
SET category_id = sqlc.arg(category_id), type = sqlc.arg(type), amount = sqlc.arg(amount),
    transaction_date = sqlc.arg(transaction_date), title = btrim(sqlc.arg(title)),
    updated_by = sqlc.arg(actor_user_id), version = version + 1, updated_at = CURRENT_TIMESTAMP
WHERE transactions.id = sqlc.arg(id) AND transactions.user_id = sqlc.arg(user_id)
  AND transactions.version = sqlc.arg(expected_version) AND transactions.is_delete = FALSE
  AND EXISTS (SELECT 1 FROM categories WHERE categories.id = sqlc.arg(category_id)
      AND categories.user_id = sqlc.arg(user_id) AND categories.type = sqlc.arg(type)
      AND categories.is_delete = FALSE)
RETURNING *;

-- name: SoftDeleteTransaction :one
UPDATE transactions
SET is_delete = TRUE, updated_by = sqlc.arg(actor_user_id),
    version = version + 1, updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
  AND version = sqlc.arg(expected_version) AND is_delete = FALSE
RETURNING *;

-- name: RestoreTransaction :one
UPDATE transactions
SET is_delete = FALSE, updated_by = sqlc.arg(actor_user_id),
    version = version + 1, updated_at = CURRENT_TIMESTAMP
WHERE transactions.id = sqlc.arg(id) AND transactions.user_id = sqlc.arg(user_id)
  AND transactions.version = sqlc.arg(expected_version) AND transactions.is_delete = TRUE
  AND EXISTS (SELECT 1 FROM categories WHERE categories.id = transactions.category_id
      AND categories.user_id = transactions.user_id AND categories.type = transactions.type
      AND categories.is_delete = FALSE)
RETURNING *;

-- name: ListDailySummaries :many
SELECT transaction_date,
       COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0)::numeric(14,2) AS income,
       COALESCE(SUM(amount) FILTER (WHERE type = 'expense'), 0)::numeric(14,2) AS expense
FROM transactions
WHERE user_id = sqlc.arg(user_id) AND is_delete = FALSE
  AND transaction_date BETWEEN sqlc.arg(start_date) AND sqlc.arg(end_date)
  AND (NOT sqlc.arg(has_cursor)::boolean OR transaction_date < sqlc.arg(cursor_date))
GROUP BY transaction_date
ORDER BY transaction_date DESC
LIMIT sqlc.arg(page_size);

-- name: GetReportSummary :one
SELECT
    COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0)::numeric(14,2) AS income,
    COALESCE(SUM(amount) FILTER (WHERE type = 'expense'), 0)::numeric(14,2) AS expense
FROM transactions
WHERE user_id = sqlc.arg(user_id)
  AND is_delete = FALSE
  AND transaction_date BETWEEN sqlc.arg(start_date) AND sqlc.arg(end_date);

-- name: ListTopReportCategories :many
WITH totals AS (
    SELECT t.type, t.category_id, c.name,
           SUM(t.amount)::numeric(14,2) AS amount,
           ROW_NUMBER() OVER (PARTITION BY t.type ORDER BY SUM(t.amount) DESC, t.category_id ASC) AS rank
    FROM transactions t
    JOIN categories c ON c.id = t.category_id AND c.user_id = t.user_id AND c.type = t.type
    WHERE t.user_id = sqlc.arg(user_id)
      AND t.is_delete = FALSE
      AND t.transaction_date BETWEEN sqlc.arg(start_date) AND sqlc.arg(end_date)
    GROUP BY t.type, t.category_id, c.name
)
SELECT type, category_id, name, amount
FROM totals
WHERE rank <= 5
ORDER BY type ASC, amount DESC, category_id ASC;

-- name: ListReportPeriodBreakdown :many
SELECT date_trunc(sqlc.arg(group_by)::text, transaction_date)::date AS period_start,
       COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0)::numeric(14,2) AS income,
       COALESCE(SUM(amount) FILTER (WHERE type = 'expense'), 0)::numeric(14,2) AS expense
FROM transactions
WHERE user_id = sqlc.arg(user_id)
  AND is_delete = FALSE
  AND transaction_date BETWEEN sqlc.arg(start_date) AND sqlc.arg(end_date)
GROUP BY date_trunc(sqlc.arg(group_by)::text, transaction_date)::date
ORDER BY period_start ASC;

-- name: ListReportCategoryBreakdown :many
SELECT t.type, t.category_id, c.name,
       SUM(t.amount)::numeric(14,2) AS amount
FROM transactions t
JOIN categories c ON c.id = t.category_id AND c.user_id = t.user_id AND c.type = t.type
WHERE t.user_id = sqlc.arg(user_id)
  AND t.is_delete = FALSE
  AND t.transaction_date BETWEEN sqlc.arg(start_date) AND sqlc.arg(end_date)
GROUP BY t.type, t.category_id, c.name
ORDER BY t.type ASC, amount DESC, t.category_id ASC;
