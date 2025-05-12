-- name: CreateAdAccount :one
INSERT INTO ad_accounts (
    platform_account_id,
    name,
    platform_type,
    credentials,
    status,
    owner_user_id,
    additional_config
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetAdAccountByID :one
SELECT * FROM ad_accounts
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetAdAccountByPlatformAccountID :one
SELECT * FROM ad_accounts
WHERE platform_account_id = $1 AND platform_type = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: ListAdAccounts :many
SELECT * FROM ad_accounts
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateAdAccount :one
UPDATE ad_accounts
SET
    name = COALESCE(sqlc.narg(name), name),
    platform_type = COALESCE(sqlc.narg(platform_type), platform_type),
    credentials = COALESCE(sqlc.narg(credentials), credentials),
    status = COALESCE(sqlc.narg(status), status),
    owner_user_id = COALESCE(sqlc.narg(owner_user_id), owner_user_id),
    additional_config = COALESCE(sqlc.narg(additional_config), additional_config),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteAdAccount :exec
UPDATE ad_accounts
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListAdAccountsIncludingDeleted :many
SELECT * FROM ad_accounts
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetAdAccountByIDIncludingDeleted :one
SELECT * FROM ad_accounts
WHERE id = $1
LIMIT 1;