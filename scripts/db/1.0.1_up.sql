-- Migration: 1.0.1
BEGIN;
CREATE TABLE "items" ("id" uuid,"title" varchar(200) NOT NULL,"description" text,"owner_id" uuid NOT NULL,"created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,"updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,"deleted_at" timestamptz,PRIMARY KEY ("id"),CONSTRAINT "fk_items_owner" FOREIGN KEY ("owner_id") REFERENCES "users"("id"));
CREATE INDEX IF NOT EXISTS "idx_items_owner" ON "items" ("owner_id");
CREATE INDEX IF NOT EXISTS "idx_items_deleted_at" ON "items" ("deleted_at");
CREATE INDEX IF NOT EXISTS "idx_items_updated" ON "items" ("updated_at");
CREATE INDEX IF NOT EXISTS "idx_items_created" ON "items" ("created_at");
CREATE INDEX IF NOT EXISTS "idx_items_owner" ON "items" ("owner_id");
CREATE INDEX IF NOT EXISTS "idx_items_created" ON "items" ("created_at");
CREATE INDEX IF NOT EXISTS "idx_items_updated" ON "items" ("updated_at");
CREATE INDEX IF NOT EXISTS "idx_items_deleted_at" ON "items" ("deleted_at");
COMMIT;
