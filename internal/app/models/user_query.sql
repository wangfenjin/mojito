-- name: CreateUser :one
INSERT INTO public."user" (
    id,
    email,
    hashed_password,
    is_active,
    is_superuser,
    full_name
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM public."user" WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM public."user" WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM public."user" 
ORDER BY email
LIMIT $1 
OFFSET $2;

-- name: UpdateUser :one
UPDATE public."user" SET
    email = COALESCE(sqlc.narg(email), email),
    full_name = COALESCE(sqlc.narg(full_name), full_name),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    is_superuser = COALESCE(sqlc.narg(is_superuser), is_superuser),
    hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password)
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
UPDATE public."user" SET
    is_active = false
WHERE id = $1
RETURNING *;

-- name: IsUserEmailExists :one
SELECT EXISTS (
    SELECT 1 
    FROM public."user" 
    WHERE email = $1
) AS email_exists;
