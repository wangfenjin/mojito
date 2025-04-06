-- Migration: 1.0.0
BEGIN;
CREATE TABLE "users" ("id" uuid,"email" text NOT NULL,"password" text NOT NULL,"full_name" text NOT NULL,"is_active" boolean DEFAULT true,"is_superuser" boolean DEFAULT false,"created_at" timestamptz,"updated_at" timestamptz,"deleted_at" timestamptz,PRIMARY KEY ("id"));
CREATE INDEX IF NOT EXISTS "idx_users_deleted_at" ON "users" ("deleted_at");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_email" ON "users" ("email");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_email" ON "users" ("email");
CREATE INDEX IF NOT EXISTS "idx_users_deleted_at" ON "users" ("deleted_at");
COMMIT;
