DROP TRIGGER IF EXISTS set_auth_users_updated_at ON auth_users;
DROP FUNCTION IF EXISTS trigger_set_updated_at();
DROP TABLE IF EXISTS auth_users;
DROP TYPE IF EXISTS user_role;