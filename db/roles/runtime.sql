\if :{?runtime_role}
\else
\echo 'runtime_role must be provided, for example: psql -v runtime_role=monelog_runtime -f db/roles/runtime.sql'
\quit
\endif

BEGIN;

GRANT USAGE ON SCHEMA public TO :"runtime_role";

GRANT SELECT, INSERT, UPDATE ON users TO :"runtime_role";
GRANT SELECT, INSERT, UPDATE ON categories TO :"runtime_role";
GRANT SELECT, INSERT, UPDATE ON transactions TO :"runtime_role";
GRANT SELECT, INSERT, UPDATE, DELETE ON refresh_sessions TO :"runtime_role";
GRANT SELECT, INSERT ON admin_access_events TO :"runtime_role";

GRANT USAGE, SELECT ON SEQUENCE
    users_id_seq,
    categories_id_seq,
    transactions_id_seq,
    refresh_sessions_id_seq,
    refresh_session_families_seq,
    admin_access_events_id_seq
TO :"runtime_role";

REVOKE DELETE, TRUNCATE ON users, categories, transactions FROM :"runtime_role";
REVOKE UPDATE, DELETE, TRUNCATE ON admin_access_events FROM :"runtime_role";

COMMIT;
