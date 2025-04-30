-- name: CleanupItems :exec
DELETE FROM public.item;

-- name: CleanupUsers :exec
DELETE FROM public."user";