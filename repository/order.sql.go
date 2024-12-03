// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: order.sql

package repository

import (
	"context"
	"database/sql"
)

const countOnlineOrders = `-- name: CountOnlineOrders :one
SELECT COUNT(o.id) AS count FROM orders o
JOIN online_order_details i ON o.id = i.order_id
ORDER BY o.id DESC
`

func (q *Queries) CountOnlineOrders(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countOnlineOrders)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countStoreOrders = `-- name: CountStoreOrders :one
SELECT COUNT(o.id) AS count FROM orders o
JOIN in_store_order_details i ON o.id = i.order_id
JOIN stores s ON s.id = i.store_id
WHERE s.id = ?
ORDER BY o.id DESC
`

func (q *Queries) CountStoreOrders(ctx context.Context, id uint64) (int64, error) {
	row := q.db.QueryRowContext(ctx, countStoreOrders, id)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const findLastCreatedOrder = `-- name: FindLastCreatedOrder :one
SELECT id, number, channel, status, total, created_at, updated_at 
FROM orders 
ORDER BY created_at DESC 
LIMIT 1
`

func (q *Queries) FindLastCreatedOrder(ctx context.Context) (Order, error) {
	row := q.db.QueryRowContext(ctx, findLastCreatedOrder)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.Number,
		&i.Channel,
		&i.Status,
		&i.Total,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findOnlineOrder = `-- name: FindOnlineOrder :one
SELECT o.id, o.number, o.channel, o.status, o.total, o.created_at, o.updated_at FROM orders o
JOIN online_order_details i ON o.id = i.order_id
WHERE o.id = ?
LIMIT 1
`

func (q *Queries) FindOnlineOrder(ctx context.Context, id uint64) (Order, error) {
	row := q.db.QueryRowContext(ctx, findOnlineOrder, id)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.Number,
		&i.Channel,
		&i.Status,
		&i.Total,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findOnlineOrderDetails = `-- name: FindOnlineOrderDetails :one
SELECT id, order_id, customer_id FROM online_order_details WHERE order_id = ?
`

func (q *Queries) FindOnlineOrderDetails(ctx context.Context, orderID uint64) (OnlineOrderDetail, error) {
	row := q.db.QueryRowContext(ctx, findOnlineOrderDetails, orderID)
	var i OnlineOrderDetail
	err := row.Scan(&i.ID, &i.OrderID, &i.CustomerID)
	return i, err
}

const findOnlineOrders = `-- name: FindOnlineOrders :many
SELECT o.id, o.number, o.channel, o.status, o.total, o.created_at, o.updated_at FROM orders o
JOIN online_order_details i ON o.id = i.order_id
ORDER BY o.id DESC
LIMIT ? OFFSET ?
`

type FindOnlineOrdersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) FindOnlineOrders(ctx context.Context, arg FindOnlineOrdersParams) ([]Order, error) {
	rows, err := q.db.QueryContext(ctx, findOnlineOrders, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Order
	for rows.Next() {
		var i Order
		if err := rows.Scan(
			&i.ID,
			&i.Number,
			&i.Channel,
			&i.Status,
			&i.Total,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const findOrder = `-- name: FindOrder :one
SELECT id, number, channel, status, total, created_at, updated_at FROM orders WHERE id = ?
`

func (q *Queries) FindOrder(ctx context.Context, id uint64) (Order, error) {
	row := q.db.QueryRowContext(ctx, findOrder, id)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.Number,
		&i.Channel,
		&i.Status,
		&i.Total,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findOrderItemByProductSKU = `-- name: FindOrderItemByProductSKU :one
SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.total
FROM order_items oi
JOIN products p ON p.id = oi.product_id
WHERE p.sku = ?
`

func (q *Queries) FindOrderItemByProductSKU(ctx context.Context, sku string) (OrderItem, error) {
	row := q.db.QueryRowContext(ctx, findOrderItemByProductSKU, sku)
	var i OrderItem
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.ProductID,
		&i.Quantity,
		&i.Price,
		&i.Total,
	)
	return i, err
}

const findOrderItems = `-- name: FindOrderItems :many
SELECT id, order_id, product_id, quantity, price, total FROM order_items WHERE order_id = ?
`

func (q *Queries) FindOrderItems(ctx context.Context, orderID uint64) ([]OrderItem, error) {
	rows, err := q.db.QueryContext(ctx, findOrderItems, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []OrderItem
	for rows.Next() {
		var i OrderItem
		if err := rows.Scan(
			&i.ID,
			&i.OrderID,
			&i.ProductID,
			&i.Quantity,
			&i.Price,
			&i.Total,
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

const findOrderWithChannel = `-- name: FindOrderWithChannel :one
SELECT id, number, channel, status, total, created_at, updated_at FROM orders WHERE id = ? AND channel = ?
`

type FindOrderWithChannelParams struct {
	ID      uint64        `json:"id"`
	Channel OrdersChannel `json:"channel"`
}

func (q *Queries) FindOrderWithChannel(ctx context.Context, arg FindOrderWithChannelParams) (Order, error) {
	row := q.db.QueryRowContext(ctx, findOrderWithChannel, arg.ID, arg.Channel)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.Number,
		&i.Channel,
		&i.Status,
		&i.Total,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findOrders = `-- name: FindOrders :many
SELECT id, number, channel, status, total, created_at, updated_at FROM orders
ORDER BY id DESC
LIMIT ? OFFSET ?
`

type FindOrdersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) FindOrders(ctx context.Context, arg FindOrdersParams) ([]Order, error) {
	rows, err := q.db.QueryContext(ctx, findOrders, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Order
	for rows.Next() {
		var i Order
		if err := rows.Scan(
			&i.ID,
			&i.Number,
			&i.Channel,
			&i.Status,
			&i.Total,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const findStoreOrder = `-- name: FindStoreOrder :one
SELECT o.id, o.number, o.channel, o.status, o.total, o.created_at, o.updated_at FROM orders o
JOIN in_store_order_details i ON o.id = i.order_id
JOIN stores s ON s.id = i.store_id
WHERE o.id = ? AND i.store_id = ?
LIMIT 1
`

type FindStoreOrderParams struct {
	ID      uint64 `json:"id"`
	StoreID uint64 `json:"store_id"`
}

func (q *Queries) FindStoreOrder(ctx context.Context, arg FindStoreOrderParams) (Order, error) {
	row := q.db.QueryRowContext(ctx, findStoreOrder, arg.ID, arg.StoreID)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.Number,
		&i.Channel,
		&i.Status,
		&i.Total,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findStoreOrderDetails = `-- name: FindStoreOrderDetails :one
SELECT id, order_id, cashier_id, store_id FROM in_store_order_details WHERE order_id = ?
`

func (q *Queries) FindStoreOrderDetails(ctx context.Context, orderID uint64) (InStoreOrderDetail, error) {
	row := q.db.QueryRowContext(ctx, findStoreOrderDetails, orderID)
	var i InStoreOrderDetail
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.CashierID,
		&i.StoreID,
	)
	return i, err
}

const findStoreOrders = `-- name: FindStoreOrders :many
SELECT o.id, o.number, o.channel, o.status, o.total, o.created_at, o.updated_at FROM orders o
JOIN in_store_order_details i ON o.id = i.order_id
JOIN stores s ON s.id = i.store_id
WHERE s.id = ?
ORDER BY o.id DESC
LIMIT ? OFFSET ?
`

type FindStoreOrdersParams struct {
	ID     uint64 `json:"id"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

func (q *Queries) FindStoreOrders(ctx context.Context, arg FindStoreOrdersParams) ([]Order, error) {
	rows, err := q.db.QueryContext(ctx, findStoreOrders, arg.ID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Order
	for rows.Next() {
		var i Order
		if err := rows.Scan(
			&i.ID,
			&i.Number,
			&i.Channel,
			&i.Status,
			&i.Total,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const insertInStoreOrderDetails = `-- name: InsertInStoreOrderDetails :exec
INSERT INTO in_store_order_details (order_id, cashier_id, store_id)
VALUES (?, ?, ?)
`

type InsertInStoreOrderDetailsParams struct {
	OrderID   uint64        `json:"order_id"`
	CashierID sql.NullInt64 `json:"cashier_id"`
	StoreID   uint64        `json:"store_id"`
}

func (q *Queries) InsertInStoreOrderDetails(ctx context.Context, arg InsertInStoreOrderDetailsParams) error {
	_, err := q.db.ExecContext(ctx, insertInStoreOrderDetails, arg.OrderID, arg.CashierID, arg.StoreID)
	return err
}

const insertOnlineOrderDetails = `-- name: InsertOnlineOrderDetails :exec
INSERT INTO online_order_details (order_id, customer_id)
VALUES (?, ?)
`

type InsertOnlineOrderDetailsParams struct {
	OrderID    uint64        `json:"order_id"`
	CustomerID sql.NullInt64 `json:"customer_id"`
}

func (q *Queries) InsertOnlineOrderDetails(ctx context.Context, arg InsertOnlineOrderDetailsParams) error {
	_, err := q.db.ExecContext(ctx, insertOnlineOrderDetails, arg.OrderID, arg.CustomerID)
	return err
}

const insertOrder = `-- name: InsertOrder :execlastid
INSERT INTO orders (number, channel, status, total)
VALUES(?, ?, ?, ?)
`

type InsertOrderParams struct {
	Number  string        `json:"number"`
	Channel OrdersChannel `json:"channel"`
	Status  OrdersStatus  `json:"status"`
	Total   float64       `json:"total"`
}

func (q *Queries) InsertOrder(ctx context.Context, arg InsertOrderParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertOrder,
		arg.Number,
		arg.Channel,
		arg.Status,
		arg.Total,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const insertOrderItem = `-- name: InsertOrderItem :exec
INSERT INTO order_items (order_id, product_id, quantity, price)
VALUES (?, ?, ?, ?)
`

type InsertOrderItemParams struct {
	OrderID   uint64  `json:"order_id"`
	ProductID uint64  `json:"product_id"`
	Quantity  int32   `json:"quantity"`
	Price     float64 `json:"price"`
}

func (q *Queries) InsertOrderItem(ctx context.Context, arg InsertOrderItemParams) error {
	_, err := q.db.ExecContext(ctx, insertOrderItem,
		arg.OrderID,
		arg.ProductID,
		arg.Quantity,
		arg.Price,
	)
	return err
}

const updateOrderStatus = `-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`

type UpdateOrderStatusParams struct {
	Status OrdersStatus `json:"status"`
	ID     uint64       `json:"id"`
}

func (q *Queries) UpdateOrderStatus(ctx context.Context, arg UpdateOrderStatusParams) error {
	_, err := q.db.ExecContext(ctx, updateOrderStatus, arg.Status, arg.ID)
	return err
}
