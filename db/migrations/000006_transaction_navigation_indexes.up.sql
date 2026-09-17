DROP INDEX transactions_active_list_idx;

CREATE INDEX transactions_active_list_idx
    ON transactions (user_id, transaction_date DESC, created_at DESC, id DESC)
    WHERE is_delete = FALSE;
