-- Migration: 1.0.2
BEGIN;
ALTER TABLE "users" DROP COLUMN "phone_number";
DROP INDEX "idx_users_phone";
COMMIT;
