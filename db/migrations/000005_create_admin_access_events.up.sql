CREATE TABLE admin_access_events (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    target_user_id BIGINT REFERENCES users (id) ON DELETE RESTRICT,
    resource_type TEXT NOT NULL,
    resource_id BIGINT,
    action TEXT NOT NULL,
    outcome TEXT NOT NULL,
    request_id TEXT NOT NULL,
    safe_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT admin_access_events_resource_type_not_blank CHECK (
        length(btrim(resource_type)) > 0
    ),
    CONSTRAINT admin_access_events_action_not_blank CHECK (length(btrim(action)) > 0),
    CONSTRAINT admin_access_events_outcome_not_blank CHECK (length(btrim(outcome)) > 0),
    CONSTRAINT admin_access_events_request_id_not_blank CHECK (length(btrim(request_id)) > 0),
    CONSTRAINT admin_access_events_safe_metadata_object CHECK (
        jsonb_typeof(safe_metadata) = 'object'
    )
);

CREATE INDEX admin_access_events_actor_idx
    ON admin_access_events (actor_user_id, created_at DESC);

CREATE INDEX admin_access_events_target_idx
    ON admin_access_events (target_user_id, created_at DESC);

REVOKE ALL ON admin_access_events FROM PUBLIC;
