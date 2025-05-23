// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: item_query.sql

package gen

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createItem = `-- name: CreateItem :one
INSERT INTO public.item (
    id,
    owner_id,
    title,
    description
) VALUES (
    $1, $2, $3, $4
) RETURNING id, owner_id, title, description, created_at, updated_at
`

type CreateItemParams struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Title       string
	Description pgtype.Text
}

func (q *Queries) CreateItem(ctx context.Context, arg CreateItemParams) (Item, error) {
	row := q.db.QueryRow(ctx, createItem,
		arg.ID,
		arg.OwnerID,
		arg.Title,
		arg.Description,
	)
	var i Item
	err := row.Scan(
		&i.ID,
		&i.OwnerID,
		&i.Title,
		&i.Description,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteItem = `-- name: DeleteItem :exec
DELETE FROM public.item WHERE id = $1
`

func (q *Queries) DeleteItem(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteItem, id)
	return err
}

const getItemByID = `-- name: GetItemByID :one
SELECT id, owner_id, title, description, created_at, updated_at FROM public.item WHERE id = $1 LIMIT 1
`

func (q *Queries) GetItemByID(ctx context.Context, id uuid.UUID) (Item, error) {
	row := q.db.QueryRow(ctx, getItemByID, id)
	var i Item
	err := row.Scan(
		&i.ID,
		&i.OwnerID,
		&i.Title,
		&i.Description,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listItemsByOwner = `-- name: ListItemsByOwner :many
SELECT id, owner_id, title, description, created_at, updated_at FROM public.item 
WHERE owner_id = $1 
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`

type ListItemsByOwnerParams struct {
	OwnerID uuid.UUID
	Limit   int64
	Offset  int64
}

func (q *Queries) ListItemsByOwner(ctx context.Context, arg ListItemsByOwnerParams) ([]Item, error) {
	rows, err := q.db.Query(ctx, listItemsByOwner, arg.OwnerID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Item
	for rows.Next() {
		var i Item
		if err := rows.Scan(
			&i.ID,
			&i.OwnerID,
			&i.Title,
			&i.Description,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateItem = `-- name: UpdateItem :one
UPDATE public.item SET
    title = COALESCE($2, title),
    description = COALESCE($3, description)
WHERE id = $1
RETURNING id, owner_id, title, description, created_at, updated_at
`

type UpdateItemParams struct {
	ID          uuid.UUID
	Title       string
	Description pgtype.Text
}

func (q *Queries) UpdateItem(ctx context.Context, arg UpdateItemParams) (Item, error) {
	row := q.db.QueryRow(ctx, updateItem, arg.ID, arg.Title, arg.Description)
	var i Item
	err := row.Scan(
		&i.ID,
		&i.OwnerID,
		&i.Title,
		&i.Description,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
