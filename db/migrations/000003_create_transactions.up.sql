CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    category_id BIGINT NOT NULL,
    type TEXT NOT NULL,
    amount NUMERIC(14, 2) NOT NULL,
    transaction_date DATE NOT NULL,
    title VARCHAR(200) NOT NULL,
    client_request_id BIGINT NOT NULL,
    request_hash TEXT NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    updated_by BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    is_delete BOOLEAN NOT NULL DEFAULT FALSE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT transactions_category_owner_type_fk
        FOREIGN KEY (category_id, user_id, type)
        REFERENCES categories (id, user_id, type)
        ON DELETE RESTRICT,
    CONSTRAINT transactions_type_valid CHECK (type IN ('income', 'expense')),
    CONSTRAINT transactions_amount_valid CHECK (
        amount >= 0.01 AND amount <= 999999999999.99
    ),
    CONSTRAINT transactions_title_trimmed CHECK (
        title = btrim(title) AND length(title) BETWEEN 1 AND 200
    ),
    CONSTRAINT transactions_request_hash_not_blank CHECK (length(btrim(request_hash)) > 0),
    CONSTRAINT transactions_version_positive CHECK (version > 0),
    CONSTRAINT transactions_owner_request_key UNIQUE (user_id, client_request_id)
);

CREATE INDEX transactions_active_list_idx
    ON transactions (user_id, transaction_date DESC, id DESC)
    WHERE is_delete = FALSE;

CREATE INDEX transactions_active_category_idx
    ON transactions (user_id, category_id, transaction_date)
    WHERE is_delete = FALSE;

CREATE INDEX transactions_trash_idx
    ON transactions (user_id, updated_at DESC, id DESC)
    WHERE is_delete = TRUE;

REVOKE ALL ON transactions FROM PUBLIC;
