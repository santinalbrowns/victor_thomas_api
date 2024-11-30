// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: product.sql

package repository

import (
	"context"
	"database/sql"
)

const countProducts = `-- name: CountProducts :one
SELECT COUNT(*) AS count
FROM products
`

func (q *Queries) CountProducts(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countProducts)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const deleteProduct = `-- name: DeleteProduct :exec
DELETE FROM products WHERE id = ?
`

func (q *Queries) DeleteProduct(ctx context.Context, id uint64) error {
	_, err := q.db.ExecContext(ctx, deleteProduct, id)
	return err
}

const findProduct = `-- name: FindProduct :one
SELECT id, slug, name, description, sku, category_id, status, visibility, created_at FROM products WHERE id = ?
`

func (q *Queries) FindProduct(ctx context.Context, id uint64) (Product, error) {
	row := q.db.QueryRowContext(ctx, findProduct, id)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Description,
		&i.Sku,
		&i.CategoryID,
		&i.Status,
		&i.Visibility,
		&i.CreatedAt,
	)
	return i, err
}

const findProductBySKU = `-- name: FindProductBySKU :one
SELECT id, slug, name, description, sku, category_id, status, visibility, created_at FROM products
WHERE sku = ?
`

func (q *Queries) FindProductBySKU(ctx context.Context, sku string) (Product, error) {
	row := q.db.QueryRowContext(ctx, findProductBySKU, sku)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Description,
		&i.Sku,
		&i.CategoryID,
		&i.Status,
		&i.Visibility,
		&i.CreatedAt,
	)
	return i, err
}

const findProductBySlug = `-- name: FindProductBySlug :one
SELECT id, slug, name, description, sku, category_id, status, visibility, created_at FROM products
WHERE slug = ?
`

func (q *Queries) FindProductBySlug(ctx context.Context, slug string) (Product, error) {
	row := q.db.QueryRowContext(ctx, findProductBySlug, slug)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Description,
		&i.Sku,
		&i.CategoryID,
		&i.Status,
		&i.Visibility,
		&i.CreatedAt,
	)
	return i, err
}

const findProductStock = `-- name: FindProductStock :one
SELECT 
    COALESCE(purchased.total_purchased, 0) AS purchased,
    COALESCE(sold.total_sold, 0) AS sold,
    COALESCE(purchased.total_purchased, 0) - COALESCE(sold.total_sold, 0) AS remaining
FROM products p
LEFT JOIN 
    (SELECT 
        product_id, 
        SUM(quantity) AS total_purchased 
    FROM purchases 
    GROUP BY product_id) purchased 
ON p.id = purchased.product_id
LEFT JOIN 
    (SELECT 
        product_id, 
        SUM(quantity) AS total_sold 
    FROM order_items 
    GROUP BY product_id) sold 
ON p.id = sold.product_id
WHERE p.id = ?
`

type FindProductStockRow struct {
	Purchased interface{} `json:"purchased"`
	Sold      interface{} `json:"sold"`
	Remaining int32       `json:"remaining"`
}

func (q *Queries) FindProductStock(ctx context.Context, id uint64) (FindProductStockRow, error) {
	row := q.db.QueryRowContext(ctx, findProductStock, id)
	var i FindProductStockRow
	err := row.Scan(&i.Purchased, &i.Sold, &i.Remaining)
	return i, err
}

const findProducts = `-- name: FindProducts :many
SELECT id, slug, name, description, sku, category_id, status, visibility, created_at FROM products
ORDER BY id DESC
LIMIT ? OFFSET ?
`

type FindProductsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) FindProducts(ctx context.Context, arg FindProductsParams) ([]Product, error) {
	rows, err := q.db.QueryContext(ctx, findProducts, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Product
	for rows.Next() {
		var i Product
		if err := rows.Scan(
			&i.ID,
			&i.Slug,
			&i.Name,
			&i.Description,
			&i.Sku,
			&i.CategoryID,
			&i.Status,
			&i.Visibility,
			&i.CreatedAt,
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

const insertProduct = `-- name: InsertProduct :execlastid
INSERT INTO products (slug, name, description, sku, category_id, status, visibility)
VALUES (?, ?, ?, ?, ?, ?, ?)
`

type InsertProductParams struct {
	Slug        string         `json:"slug"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	Sku         string         `json:"sku"`
	CategoryID  sql.NullInt64  `json:"category_id"`
	Status      bool           `json:"status"`
	Visibility  bool           `json:"visibility"`
}

func (q *Queries) InsertProduct(ctx context.Context, arg InsertProductParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertProduct,
		arg.Slug,
		arg.Name,
		arg.Description,
		arg.Sku,
		arg.CategoryID,
		arg.Status,
		arg.Visibility,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const searchProducts = `-- name: SearchProducts :many
SELECT id, slug, name, description, sku, category_id, status, visibility, created_at FROM products
WHERE name LIKE ?
ORDER BY id DESC
LIMIT ? OFFSET ?
`

type SearchProductsParams struct {
	Name   string `json:"name"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

func (q *Queries) SearchProducts(ctx context.Context, arg SearchProductsParams) ([]Product, error) {
	rows, err := q.db.QueryContext(ctx, searchProducts, arg.Name, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Product
	for rows.Next() {
		var i Product
		if err := rows.Scan(
			&i.ID,
			&i.Slug,
			&i.Name,
			&i.Description,
			&i.Sku,
			&i.CategoryID,
			&i.Status,
			&i.Visibility,
			&i.CreatedAt,
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

const updateProduct = `-- name: UpdateProduct :exec
UPDATE products
SET name = COALESCE(?, name),
    slug = COALESCE(?, slug),
    description = COALESCE(?, description),
    sku = COALESCE(?, sku),
    category_id = ?,
    status = COALESCE(?, status),
    visibility = COALESCE(?, visibility)
WHERE id = ?
`

type UpdateProductParams struct {
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description sql.NullString `json:"description"`
	Sku         string         `json:"sku"`
	CategoryID  sql.NullInt64  `json:"category_id"`
	Status      bool           `json:"status"`
	Visibility  bool           `json:"visibility"`
	ID          uint64         `json:"id"`
}

func (q *Queries) UpdateProduct(ctx context.Context, arg UpdateProductParams) error {
	_, err := q.db.ExecContext(ctx, updateProduct,
		arg.Name,
		arg.Slug,
		arg.Description,
		arg.Sku,
		arg.CategoryID,
		arg.Status,
		arg.Visibility,
		arg.ID,
	)
	return err
}
