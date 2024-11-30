-- name: InsertOrder :execlastid
INSERT INTO orders (number, channel, status, total)
VALUES(?, ?, ?, ?);

-- name: InsertOrderItem :exec
INSERT INTO order_items (order_id, product_id, quantity, price)
VALUES (?, ?, ?, ?);

-- name: InsertInStoreOrderDetails :exec
INSERT INTO in_store_order_details (order_id, cashier_id, store_id)
VALUES (?, ?, ?);

-- name: InsertOnlineOrderDetails :exec
INSERT INTO online_order_details (order_id, customer_id)
VALUES (?, ?);

-- name: FindOrder :one
SELECT * FROM orders WHERE id = ?;

-- name: FindOrderWithChannel :one
SELECT * FROM orders WHERE id = ? AND channel = ?;

-- name: FindLastCreatedOrder :one
SELECT * 
FROM orders 
ORDER BY created_at DESC 
LIMIT 1;


-- name: FindStoreOrder :one
SELECT o.* FROM orders o
JOIN in_store_order_details i ON o.id = i.order_id
JOIN stores s ON s.id = i.store_id
WHERE o.id = ? AND i.store_id = ?
LIMIT 1;

-- name: FindOnlineOrder :one
SELECT o.* FROM orders o
JOIN online_order_details i ON o.id = i.order_id
WHERE o.id = ?
LIMIT 1;

-- name: FindOrders :many
SELECT * FROM orders
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: FindStoreOrders :many
SELECT o.* FROM orders o
JOIN in_store_order_details i ON o.id = i.order_id
JOIN stores s ON s.id = i.store_id
WHERE s.id = ?
ORDER BY o.id DESC
LIMIT ? OFFSET ?;

-- name: FindOnlineOrders :many
SELECT o.* FROM orders o
JOIN online_order_details i ON o.id = i.order_id
ORDER BY o.id DESC
LIMIT ? OFFSET ?;

-- name: FindOrderItems :many
SELECT * FROM order_items WHERE order_id = ?;

-- name: FindOnlineOrderDetails :one
SELECT * FROM online_order_details WHERE order_id = ?;

-- name: FindStoreOrderDetails :one
SELECT * FROM in_store_order_details WHERE order_id = ?;


-- name: CountStoreOrders :one
SELECT COUNT(o.id) AS count FROM orders o
JOIN in_store_order_details i ON o.id = i.order_id
JOIN stores s ON s.id = i.store_id
WHERE s.id = ?
ORDER BY o.id DESC;

-- name: CountOnlineOrders :one
SELECT COUNT(o.id) AS count FROM orders o
JOIN online_order_details i ON o.id = i.order_id
ORDER BY o.id DESC;
