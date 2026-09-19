DO $migration$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'users'
          AND column_name = 'id'
          AND data_type = 'bigint'
    ) THEN
        RAISE EXCEPTION 'migration 000007 is irreversible because external UUID values cannot be reconstructed';
    END IF;
END
$migration$;
