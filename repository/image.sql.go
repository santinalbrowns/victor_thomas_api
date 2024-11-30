// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: image.sql

package repository

import (
	"context"
)

const deleteImage = `-- name: DeleteImage :exec
DELETE FROM images WHERE id = ?
`

func (q *Queries) DeleteImage(ctx context.Context, id uint64) error {
	_, err := q.db.ExecContext(ctx, deleteImage, id)
	return err
}

const findImage = `-- name: FindImage :one
SELECT id, name, created_at FROM images WHERE id = ?
`

func (q *Queries) FindImage(ctx context.Context, id uint64) (Image, error) {
	row := q.db.QueryRowContext(ctx, findImage, id)
	var i Image
	err := row.Scan(&i.ID, &i.Name, &i.CreatedAt)
	return i, err
}

const insertImage = `-- name: InsertImage :execlastid
INSERT INTO images (name) VALUES (?)
`

func (q *Queries) InsertImage(ctx context.Context, name string) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertImage, name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}