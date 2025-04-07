CREATE TABLE items (
  id UUID PRIMARY KEY,
  title VARCHAR(200) NOT NULL,
  description TEXT,
  owner_id UUID NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT fk_items_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX idx_items_owner ON items (owner_id);
CREATE INDEX idx_items_created ON items (created_at);
CREATE INDEX idx_items_updated ON items (updated_at);
CREATE INDEX idx_items_deleted_at ON items (deleted_at);
