CREATE TABLE refresh_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    family_id UUID NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    replaced_by UUID REFERENCES refresh_sessions (id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT refresh_sessions_token_hash_not_blank CHECK (length(btrim(token_hash)) > 0),
    CONSTRAINT refresh_sessions_expiry_valid CHECK (expires_at > created_at),
    CONSTRAINT refresh_sessions_revocation_valid CHECK (
        revoked_at IS NULL OR revoked_at >= created_at
    ),
    CONSTRAINT refresh_sessions_replacement_valid CHECK (
        replaced_by IS NULL OR replaced_by <> id
    )
);

CREATE INDEX refresh_sessions_user_family_idx
    ON refresh_sessions (user_id, family_id);

CREATE INDEX refresh_sessions_expiry_idx
    ON refresh_sessions (expires_at);

REVOKE ALL ON refresh_sessions FROM PUBLIC;
