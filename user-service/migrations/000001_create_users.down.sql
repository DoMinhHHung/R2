DROP TRIGGER IF EXISTS set_users_updated_at ON users;
DROP FUNCTION IF EXISTS trigger_set_updated_at();
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_gender;
DROP TYPE IF EXISTS user_status;