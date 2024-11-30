-- name: InsertProduct :execlastid
INSERT INTO products (slug, name, description, sku, category_id, status, visibility)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: FindProduct :one
SELECT * FROM products WHERE id = ?;

-- name: FindProductBySlug :one
SELECT * FROM products
WHERE slug = ?;

-- name: FindProductBySKU :one
SELECT * FROM products
WHERE sku = ?;

-- name: FindProducts :many
SELECT * FROM products
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: UpdateProduct :exec
UPDATE products
SET name = COALESCE(?, name),
    slug = COALESCE(?, slug),
    description = COALESCE(?, description),
    sku = COALESCE(?, sku),
    category_id = ?,
    status = COALESCE(?, status),
    visibility = COALESCE(?, visibility)
WHERE id = ?;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = ?;

-- name: CountProducts :one
SELECT COUNT(*) AS count
FROM products;

-- name: SearchProducts :many
SELECT * FROM products
WHERE name LIKE ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: FindProductStock :one
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
WHERE p.id = ?;
