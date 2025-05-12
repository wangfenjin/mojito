DROP TRIGGER IF EXISTS update_ad_accounts_updated_at ON ad_accounts;
DROP INDEX IF EXISTS idx_ad_accounts_deleted_at;
DROP INDEX IF EXISTS idx_ad_accounts_status;
DROP INDEX IF EXISTS idx_ad_accounts_platform_type;
DROP INDEX IF EXISTS idx_ad_accounts_platform_account_id;
DROP TABLE IF EXISTS ad_accounts;
DROP TYPE IF EXISTS account_status_enum;
DROP TYPE IF EXISTS platform_type_enum;