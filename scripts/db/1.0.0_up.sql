CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255) NOT NULL,
  password VARCHAR(255) NOT NULL,
  full_name VARCHAR(100) NOT NULL,
  is_active BOOLEAN DEFAULT true,
  is_superuser BOOLEAN DEFAULT false,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT idx_users_email UNIQUE (email)
);

CREATE INDEX idx_users_created ON users (created_at);
CREATE INDEX idx_users_updated ON users (updated_at);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);
