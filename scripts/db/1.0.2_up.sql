-- Migration: 1.0.2
BEGIN;
ALTER TABLE "users" ALTER COLUMN "email" TYPE varchar(255) USING "email"::varchar(255);
ALTER TABLE "users" ADD "phone_number" varchar(20);
ALTER TABLE "users" ALTER COLUMN "password" TYPE varchar(255) USING "password"::varchar(255);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_phone" ON "users" ("phone_number");
COMMIT;
