DO $migration$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'users'
          AND column_name = 'id'
          AND data_type = 'uuid'
    ) THEN
        CREATE TEMP TABLE identifier_map (
            entity TEXT NOT NULL,
            old_id UUID NOT NULL,
            new_id BIGINT NOT NULL,
            PRIMARY KEY (entity, old_id),
            UNIQUE (entity, new_id)
        ) ON COMMIT DROP;

        INSERT INTO identifier_map (entity, old_id, new_id)
        SELECT 'users', id, row_number() OVER (ORDER BY created_at, id) FROM users;
        INSERT INTO identifier_map (entity, old_id, new_id)
        SELECT 'categories', id, row_number() OVER (ORDER BY created_at, id) FROM categories;
        INSERT INTO identifier_map (entity, old_id, new_id)
        SELECT 'transactions', id, row_number() OVER (ORDER BY created_at, id) FROM transactions;
        INSERT INTO identifier_map (entity, old_id, new_id)
        SELECT 'refresh_sessions', id, row_number() OVER (ORDER BY created_at, id) FROM refresh_sessions;
        INSERT INTO identifier_map (entity, old_id, new_id)
        SELECT 'refresh_families', family_id, row_number() OVER (ORDER BY family_id)
        FROM (SELECT DISTINCT family_id FROM refresh_sessions) families;
        INSERT INTO identifier_map (entity, old_id, new_id)
        SELECT 'client_requests', client_request_id, row_number() OVER (ORDER BY client_request_id)
        FROM (SELECT DISTINCT client_request_id FROM transactions) requests;
        INSERT INTO identifier_map (entity, old_id, new_id)
        SELECT 'admin_access_events', id, row_number() OVER (ORDER BY created_at, id)
        FROM admin_access_events;
        INSERT INTO identifier_map (entity, old_id, new_id)
        SELECT 'admin_resources', resource_id, row_number() OVER (ORDER BY resource_id)
        FROM (SELECT DISTINCT resource_id FROM admin_access_events WHERE resource_id IS NOT NULL) resources;

        CREATE FUNCTION pg_temp.identifier_bigint(entity_name TEXT, value UUID)
        RETURNS BIGINT
        LANGUAGE SQL
        STABLE
        AS $function$
            SELECT new_id
            FROM identifier_map
            WHERE entity = entity_name AND old_id = value
        $function$;

        ALTER TABLE transactions
            DROP CONSTRAINT transactions_category_owner_type_fk,
            DROP CONSTRAINT transactions_created_by_fkey,
            DROP CONSTRAINT transactions_updated_by_fkey,
            DROP CONSTRAINT transactions_user_id_fkey;
        ALTER TABLE categories DROP CONSTRAINT categories_user_id_fkey;
        ALTER TABLE refresh_sessions
            DROP CONSTRAINT refresh_sessions_replaced_by_fkey,
            DROP CONSTRAINT refresh_sessions_user_id_fkey;
        ALTER TABLE admin_access_events
            DROP CONSTRAINT admin_access_events_actor_user_id_fkey,
            DROP CONSTRAINT admin_access_events_target_user_id_fkey;

        ALTER TABLE users
            ALTER COLUMN id TYPE BIGINT USING pg_temp.identifier_bigint('users', id);
        ALTER TABLE categories
            ALTER COLUMN id TYPE BIGINT USING pg_temp.identifier_bigint('categories', id),
            ALTER COLUMN user_id TYPE BIGINT USING pg_temp.identifier_bigint('users', user_id);
        ALTER TABLE transactions
            ALTER COLUMN id TYPE BIGINT USING pg_temp.identifier_bigint('transactions', id),
            ALTER COLUMN user_id TYPE BIGINT USING pg_temp.identifier_bigint('users', user_id),
            ALTER COLUMN category_id TYPE BIGINT USING pg_temp.identifier_bigint('categories', category_id),
            ALTER COLUMN client_request_id TYPE BIGINT USING pg_temp.identifier_bigint('client_requests', client_request_id),
            ALTER COLUMN created_by TYPE BIGINT USING pg_temp.identifier_bigint('users', created_by),
            ALTER COLUMN updated_by TYPE BIGINT USING pg_temp.identifier_bigint('users', updated_by);
        ALTER TABLE refresh_sessions
            ALTER COLUMN id TYPE BIGINT USING pg_temp.identifier_bigint('refresh_sessions', id),
            ALTER COLUMN user_id TYPE BIGINT USING pg_temp.identifier_bigint('users', user_id),
            ALTER COLUMN family_id TYPE BIGINT USING pg_temp.identifier_bigint('refresh_families', family_id),
            ALTER COLUMN replaced_by TYPE BIGINT USING pg_temp.identifier_bigint('refresh_sessions', replaced_by);
        ALTER TABLE admin_access_events
            ALTER COLUMN id TYPE BIGINT USING pg_temp.identifier_bigint('admin_access_events', id),
            ALTER COLUMN actor_user_id TYPE BIGINT USING pg_temp.identifier_bigint('users', actor_user_id),
            ALTER COLUMN target_user_id TYPE BIGINT USING pg_temp.identifier_bigint('users', target_user_id),
            ALTER COLUMN resource_id TYPE BIGINT USING COALESCE(
                CASE resource_type
                    WHEN 'user' THEN pg_temp.identifier_bigint('users', resource_id)
                    WHEN 'category' THEN pg_temp.identifier_bigint('categories', resource_id)
                    WHEN 'transaction' THEN pg_temp.identifier_bigint('transactions', resource_id)
                END,
                pg_temp.identifier_bigint('admin_resources', resource_id)
            );

        CREATE SEQUENCE users_id_seq AS BIGINT OWNED BY users.id;
        CREATE SEQUENCE categories_id_seq AS BIGINT OWNED BY categories.id;
        CREATE SEQUENCE transactions_id_seq AS BIGINT OWNED BY transactions.id;
        CREATE SEQUENCE refresh_sessions_id_seq AS BIGINT OWNED BY refresh_sessions.id;
        CREATE SEQUENCE refresh_session_families_seq AS BIGINT;
        CREATE SEQUENCE admin_access_events_id_seq AS BIGINT OWNED BY admin_access_events.id;

        ALTER TABLE users ALTER COLUMN id SET DEFAULT nextval('users_id_seq');
        ALTER TABLE categories ALTER COLUMN id SET DEFAULT nextval('categories_id_seq');
        ALTER TABLE transactions ALTER COLUMN id SET DEFAULT nextval('transactions_id_seq');
        ALTER TABLE refresh_sessions
            ALTER COLUMN id SET DEFAULT nextval('refresh_sessions_id_seq'),
            ALTER COLUMN family_id SET DEFAULT nextval('refresh_session_families_seq');
        ALTER TABLE admin_access_events ALTER COLUMN id SET DEFAULT nextval('admin_access_events_id_seq');

        PERFORM setval('users_id_seq', COALESCE((SELECT max(id) FROM users), 0) + 1, false);
        PERFORM setval('categories_id_seq', COALESCE((SELECT max(id) FROM categories), 0) + 1, false);
        PERFORM setval('transactions_id_seq', COALESCE((SELECT max(id) FROM transactions), 0) + 1, false);
        PERFORM setval('refresh_sessions_id_seq', COALESCE((SELECT max(id) FROM refresh_sessions), 0) + 1, false);
        PERFORM setval('refresh_session_families_seq', COALESCE((SELECT max(family_id) FROM refresh_sessions), 0) + 1, false);
        PERFORM setval('admin_access_events_id_seq', COALESCE((SELECT max(id) FROM admin_access_events), 0) + 1, false);

        ALTER TABLE categories
            ADD CONSTRAINT categories_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
        ALTER TABLE transactions
            ADD CONSTRAINT transactions_category_owner_type_fk FOREIGN KEY (category_id, user_id, type) REFERENCES categories(id, user_id, type) ON DELETE RESTRICT,
            ADD CONSTRAINT transactions_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,
            ADD CONSTRAINT transactions_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE RESTRICT,
            ADD CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
        ALTER TABLE refresh_sessions
            ADD CONSTRAINT refresh_sessions_replaced_by_fkey FOREIGN KEY (replaced_by) REFERENCES refresh_sessions(id) ON DELETE RESTRICT,
            ADD CONSTRAINT refresh_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
        ALTER TABLE admin_access_events
            ADD CONSTRAINT admin_access_events_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES users(id) ON DELETE RESTRICT,
            ADD CONSTRAINT admin_access_events_target_user_id_fkey FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE RESTRICT;
    END IF;
END
$migration$;
