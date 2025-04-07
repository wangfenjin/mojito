-- Migration: 1.0.2
BEGIN;
ALTER TABLE "users" ADD "phone_number" varchar(20);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_phone" ON "users" ("phone_number");
COMMIT;
