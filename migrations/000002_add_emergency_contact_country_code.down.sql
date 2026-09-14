-- Intentionally a no-op.
--
-- country_code is created by 000001_init.up.sql, not by this migration, so
-- dropping it here would tear down a column this migration does not own and
-- would leave the schema inconsistent with 000001 after a `down 1`.
SELECT 1;
