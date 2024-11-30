-- name: InsertCategory :execlastid
INSERT INTO categories (slug, name, enabled, show_in_menu, show_products, image_id)
VALUES (?, ?, ?, ?, ?, ?);

-- name: FindCategory :one
SELECT * FROM categories
WHERE id = ?;

-- name: FindCategoryBySlug :one
SELECT * FROM categories
WHERE slug = ?;

-- name: FindCategories :many
SELECT * FROM categories
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchCategories :many
SELECT * FROM categories
WHERE name LIKE ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: CountCategories :one
SELECT COUNT(*) AS count
FROM categories;

-- name: UpdateCategory :exec
UPDATE categories
SET 
    name = COALESCE(?, name),
    slug = COALESCE(?, slug),
    enabled = COALESCE(?, enabled),
    show_in_menu = COALESCE(?, show_in_menu),
    show_products = COALESCE(?, show_products),
    image_id = ?
WHERE id = ?;

-- name: DeleteCategory :exec
DELETE FROM categories
WHERE id = ?;

