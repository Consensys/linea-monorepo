-- Coordinator postgres bootstrap.
-- Runs on first boot of `coordinator-pg`. POSTGRES_DB is `linea_coordinator`
-- (matches the coordinator config's [database].schema field, which the app
-- uses as the JDBC database name).
--
-- The coordinator app handles its own Flyway migrations against the public
-- schema of this database; no schema-level setup needed here.
SELECT 1;
