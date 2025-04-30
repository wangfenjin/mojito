-- name: CreateItem :one
INSERT INTO public.item (
    id,
    owner_id,
    title,
    description
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetItemByID :one
SELECT * FROM public.item WHERE id = $1 LIMIT 1;

-- name: ListItems :many
SELECT * FROM public.item 
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListItemsByOwner :many
SELECT * FROM public.item 
WHERE owner_id = $1 
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateItem :one
UPDATE public.item SET
    title = COALESCE($2, title),
    description = COALESCE($3, description)
WHERE id = $1
RETURNING *;

-- name: DeleteItem :exec
DELETE FROM public.item WHERE id = $1;

-- name: CountItemsByOwner :one
SELECT COUNT(*) FROM public.item WHERE owner_id = $1;