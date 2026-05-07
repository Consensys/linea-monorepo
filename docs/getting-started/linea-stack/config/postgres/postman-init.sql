-- Postman postgres bootstrap.
-- Runs on first boot of `postman-pg` (POSTGRES_DB=postman already created).
-- The postman app handles its own migrations.

-- No schema overrides needed; postman uses the public schema.
SELECT 1;
