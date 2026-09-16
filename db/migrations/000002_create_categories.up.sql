CREATE TABLE categories (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    type TEXT NOT NULL,
    name VARCHAR(80) NOT NULL,
    is_delete BOOLEAN NOT NULL DEFAULT FALSE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT categories_type_valid CHECK (type IN ('income', 'expense')),
    CONSTRAINT categories_name_trimmed CHECK (
        name = btrim(name) AND length(name) BETWEEN 1 AND 80
    ),
    CONSTRAINT categories_version_positive CHECK (version > 0),
    CONSTRAINT categories_owner_type_key UNIQUE (id, user_id, type)
);

CREATE UNIQUE INDEX categories_owner_type_name_key
    ON categories (user_id, type, lower(name));

REVOKE ALL ON categories FROM PUBLIC;
