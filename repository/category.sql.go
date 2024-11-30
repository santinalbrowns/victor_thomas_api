// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: category.sql

package repository

import (
	"context"
	"database/sql"
)

const countCategories = `-- name: CountCategories :one
SELECT COUNT(*) AS count
FROM categories
`

func (q *Queries) CountCategories(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countCategories)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const deleteCategory = `-- name: DeleteCategory :exec
DELETE FROM categories
WHERE id = ?
`

func (q *Queries) DeleteCategory(ctx context.Context, id uint64) error {
	_, err := q.db.ExecContext(ctx, deleteCategory, id)
	return err
}

const findCategories = `-- name: FindCategories :many
SELECT id, slug, name, enabled, show_in_menu, show_products, image_id FROM categories
ORDER BY id DESC
LIMIT ? OFFSET ?
`

type FindCategoriesParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) FindCategories(ctx context.Context, arg FindCategoriesParams) ([]Category, error) {
	rows, err := q.db.QueryContext(ctx, findCategories, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Category
	for rows.Next() {
		var i Category
		if err := rows.Scan(
			&i.ID,
			&i.Slug,
			&i.Name,
			&i.Enabled,
			&i.ShowInMenu,
			&i.ShowProducts,
			&i.ImageID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findCategory = `-- name: FindCategory :one
SELECT id, slug, name, enabled, show_in_menu, show_products, image_id FROM categories
WHERE id = ?
`

func (q *Queries) FindCategory(ctx context.Context, id uint64) (Category, error) {
	row := q.db.QueryRowContext(ctx, findCategory, id)
	var i Category
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Enabled,
		&i.ShowInMenu,
		&i.ShowProducts,
		&i.ImageID,
	)
	return i, err
}

const findCategoryBySlug = `-- name: FindCategoryBySlug :one
SELECT id, slug, name, enabled, show_in_menu, show_products, image_id FROM categories
WHERE slug = ?
`

func (q *Queries) FindCategoryBySlug(ctx context.Context, slug string) (Category, error) {
	row := q.db.QueryRowContext(ctx, findCategoryBySlug, slug)
	var i Category
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Enabled,
		&i.ShowInMenu,
		&i.ShowProducts,
		&i.ImageID,
	)
	return i, err
}

const insertCategory = `-- name: InsertCategory :execlastid
INSERT INTO categories (slug, name, enabled, show_in_menu, show_products, image_id)
VALUES (?, ?, ?, ?, ?, ?)
`

type InsertCategoryParams struct {
	Slug         string        `json:"slug"`
	Name         string        `json:"name"`
	Enabled      bool          `json:"enabled"`
	ShowInMenu   bool          `json:"show_in_menu"`
	ShowProducts bool          `json:"show_products"`
	ImageID      sql.NullInt64 `json:"image_id"`
}

func (q *Queries) InsertCategory(ctx context.Context, arg InsertCategoryParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertCategory,
		arg.Slug,
		arg.Name,
		arg.Enabled,
		arg.ShowInMenu,
		arg.ShowProducts,
		arg.ImageID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const searchCategories = `-- name: SearchCategories :many
SELECT id, slug, name, enabled, show_in_menu, show_products, image_id FROM categories
WHERE name LIKE ?
ORDER BY id DESC
LIMIT ? OFFSET ?
`

type SearchCategoriesParams struct {
	Name   string `json:"name"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

func (q *Queries) SearchCategories(ctx context.Context, arg SearchCategoriesParams) ([]Category, error) {
	rows, err := q.db.QueryContext(ctx, searchCategories, arg.Name, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Category
	for rows.Next() {
		var i Category
		if err := rows.Scan(
			&i.ID,
			&i.Slug,
			&i.Name,
			&i.Enabled,
			&i.ShowInMenu,
			&i.ShowProducts,
			&i.ImageID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateCategory = `-- name: UpdateCategory :exec
UPDATE categories
SET 
    name = COALESCE(?, name),
    slug = COALESCE(?, slug),
    enabled = COALESCE(?, enabled),
    show_in_menu = COALESCE(?, show_in_menu),
    show_products = COALESCE(?, show_products),
    image_id = ?
WHERE id = ?
`

type UpdateCategoryParams struct {
	Name         string        `json:"name"`
	Slug         string        `json:"slug"`
	Enabled      bool          `json:"enabled"`
	ShowInMenu   bool          `json:"show_in_menu"`
	ShowProducts bool          `json:"show_products"`
	ImageID      sql.NullInt64 `json:"image_id"`
	ID           uint64        `json:"id"`
}

func (q *Queries) UpdateCategory(ctx context.Context, arg UpdateCategoryParams) error {
	_, err := q.db.ExecContext(ctx, updateCategory,
		arg.Name,
		arg.Slug,
		arg.Enabled,
		arg.ShowInMenu,
		arg.ShowProducts,
		arg.ImageID,
		arg.ID,
	)
	return err
}