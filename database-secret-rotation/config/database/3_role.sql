
CREATE ROLE vault_db_user LOGIN SUPERUSER PASSWORD 'vault_db_password';
CREATE ROLE readonly NOINHERIT;

GRANT SELECT ON ALL TABLES IN SCHEMA public to "readonly";