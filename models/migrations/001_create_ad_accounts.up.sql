-- Create ENUM types for fixed value sets
CREATE TYPE platform_type_enum AS ENUM ('META', 'GOOGLE_ADS', 'TIKTOK_ADS');
CREATE TYPE account_status_enum AS ENUM ('ACTIVE', 'INACTIVE', 'REQUIRES_REAUTH', 'SUSPENDED');

CREATE TABLE ad_accounts (
    id UUID PRIMARY KEY,
    platform_account_id TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    platform_type platform_type_enum NOT NULL,
    credentials TEXT NOT NULL,
    status account_status_enum NOT NULL,
    owner_user_id UUID,
    additional_config JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_ad_accounts_platform_account_id ON ad_accounts(platform_account_id);
CREATE INDEX idx_ad_accounts_platform_type ON ad_accounts(platform_type);
CREATE INDEX idx_ad_accounts_status ON ad_accounts(status);
CREATE INDEX idx_ad_accounts_deleted_at ON ad_accounts(deleted_at);

CREATE TRIGGER update_ad_accounts_updated_at
BEFORE UPDATE ON ad_accounts
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();