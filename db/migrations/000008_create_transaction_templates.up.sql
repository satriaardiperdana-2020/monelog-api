CREATE TABLE transaction_templates (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    category_id BIGINT NOT NULL,
    type TEXT NOT NULL,
    name VARCHAR(80) NOT NULL,
    amount NUMERIC(14, 2) NOT NULL,
    title VARCHAR(200) NOT NULL,
    is_delete BOOLEAN NOT NULL DEFAULT FALSE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT transaction_templates_category_owner_type_fk
        FOREIGN KEY (category_id, user_id, type)
        REFERENCES categories (id, user_id, type)
        ON DELETE RESTRICT,
    CONSTRAINT transaction_templates_type_valid CHECK (type IN ('income', 'expense')),
    CONSTRAINT transaction_templates_name_trimmed CHECK (
        name = btrim(name) AND length(name) BETWEEN 1 AND 80
    ),
    CONSTRAINT transaction_templates_amount_valid CHECK (
        amount >= 0.01 AND amount <= 999999999999.99
    ),
    CONSTRAINT transaction_templates_title_trimmed CHECK (
        title = btrim(title) AND length(title) BETWEEN 1 AND 200
    ),
    CONSTRAINT transaction_templates_version_positive CHECK (version > 0)
);

CREATE INDEX transaction_templates_active_list_idx
    ON transaction_templates (user_id, lower(name), id)
    WHERE is_delete = FALSE;

CREATE INDEX transaction_templates_trash_idx
    ON transaction_templates (user_id, updated_at DESC, id DESC)
    WHERE is_delete = TRUE;

REVOKE ALL ON transaction_templates FROM PUBLIC;
