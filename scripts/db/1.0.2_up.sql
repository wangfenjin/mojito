ALTER TABLE users ADD COLUMN phone_number VARCHAR(20);
CREATE UNIQUE INDEX idx_users_phone ON users (phone_number);
