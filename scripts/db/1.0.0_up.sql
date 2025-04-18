-- Migration: 1.0.0
BEGIN;
CREATE TABLE "users" ("id" uuid,"email" varchar(255) NOT NULL,"password" varchar(255) NOT NULL,"full_name" varchar(100) NOT NULL,"is_active" boolean DEFAULT true,"is_superuser" boolean DEFAULT false,"created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,"updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,"deleted_at" timestamptz,PRIMARY KEY ("id"));
CREATE INDEX IF NOT EXISTS "idx_users_deleted_at" ON "users" ("deleted_at");
CREATE INDEX IF NOT EXISTS "idx_users_updated" ON "users" ("updated_at");
CREATE INDEX IF NOT EXISTS "idx_users_created" ON "users" ("created_at");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_email" ON "users" ("email");
COMMIT;
