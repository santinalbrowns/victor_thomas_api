-- name: InsertPurchase :execlastid
INSERT INTO purchases (product_id, date, quantity, order_price, selling_price, store_id, user_id)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: FindPurchases :many
SELECT * FROM purchases
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: FindPurchase :one
SELECT * FROM purchases WHERE id  = ?;

-- name: FindPurchasesByProductSKU :many
SELECT s.* FROM purchases s
JOIN products p ON p.id = s.product_id
WHERE p.sku = ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: FindPurchaseByProductSKU :one
SELECT s.* FROM purchases s
JOIN products p ON p.id = s.product_id
WHERE p.sku = ?
LIMIT 1;

-- name: DeletePurchase :exec
DELETE FROM purchases
WHERE id = ?;

-- name: CountPurchases :one
SELECT COUNT(*) AS count
FROM purchases;