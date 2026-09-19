CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    timezone TEXT NOT NULL DEFAULT 'Asia/Jakarta',
    currency TEXT NOT NULL DEFAULT 'IDR',
    is_delete BOOLEAN NOT NULL DEFAULT FALSE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT users_email_normalized CHECK (
        email = lower(btrim(email)) AND length(email) > 0
    ),
    CONSTRAINT users_password_hash_not_blank CHECK (length(btrim(password_hash)) > 0),
    CONSTRAINT users_role_valid CHECK (role IN ('user', 'admin')),
    CONSTRAINT users_timezone_not_blank CHECK (length(btrim(timezone)) > 0),
    CONSTRAINT users_currency_idr CHECK (currency = 'IDR'),
    CONSTRAINT users_version_positive CHECK (version > 0)
);

REVOKE ALL ON users FROM PUBLIC;
