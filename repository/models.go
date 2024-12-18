// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package repository

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

type OrdersChannel string

const (
	OrdersChannelOnline  OrdersChannel = "online"
	OrdersChannelInStore OrdersChannel = "in-store"
)

func (e *OrdersChannel) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = OrdersChannel(s)
	case string:
		*e = OrdersChannel(s)
	default:
		return fmt.Errorf("unsupported scan type for OrdersChannel: %T", src)
	}
	return nil
}

type NullOrdersChannel struct {
	OrdersChannel OrdersChannel `json:"orders_channel"`
	Valid         bool          `json:"valid"` // Valid is true if OrdersChannel is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullOrdersChannel) Scan(value interface{}) error {
	if value == nil {
		ns.OrdersChannel, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.OrdersChannel.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullOrdersChannel) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.OrdersChannel), nil
}

type OrdersStatus string

const (
	OrdersStatusPending   OrdersStatus = "pending"
	OrdersStatusCompleted OrdersStatus = "completed"
	OrdersStatusCanceled  OrdersStatus = "canceled"
)

func (e *OrdersStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = OrdersStatus(s)
	case string:
		*e = OrdersStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for OrdersStatus: %T", src)
	}
	return nil
}

type NullOrdersStatus struct {
	OrdersStatus OrdersStatus `json:"orders_status"`
	Valid        bool         `json:"valid"` // Valid is true if OrdersStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullOrdersStatus) Scan(value interface{}) error {
	if value == nil {
		ns.OrdersStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.OrdersStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullOrdersStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.OrdersStatus), nil
}

type Category struct {
	ID           uint64        `json:"id"`
	Slug         string        `json:"slug"`
	Name         string        `json:"name"`
	Enabled      bool          `json:"enabled"`
	ShowInMenu   bool          `json:"show_in_menu"`
	ShowProducts bool          `json:"show_products"`
	ImageID      sql.NullInt64 `json:"image_id"`
}

type Image struct {
	ID        uint64       `json:"id"`
	Name      string       `json:"name"`
	CreatedAt sql.NullTime `json:"created_at"`
}

type InStoreOrderDetail struct {
	ID        uint64        `json:"id"`
	OrderID   uint64        `json:"order_id"`
	CashierID sql.NullInt64 `json:"cashier_id"`
	StoreID   uint64        `json:"store_id"`
}

type OnlineOrderDetail struct {
	ID         uint64        `json:"id"`
	OrderID    uint64        `json:"order_id"`
	CustomerID sql.NullInt64 `json:"customer_id"`
}

type Order struct {
	ID        uint64        `json:"id"`
	Number    string        `json:"number"`
	Channel   OrdersChannel `json:"channel"`
	Status    OrdersStatus  `json:"status"`
	Total     float64       `json:"total"`
	CreatedAt sql.NullTime  `json:"created_at"`
	UpdatedAt sql.NullTime  `json:"updated_at"`
}

type OrderItem struct {
	ID        uint64          `json:"id"`
	OrderID   uint64          `json:"order_id"`
	ProductID uint64          `json:"product_id"`
	Quantity  int32           `json:"quantity"`
	Price     float64         `json:"price"`
	Total     sql.NullFloat64 `json:"total"`
}

type Product struct {
	ID          uint64         `json:"id"`
	Slug        string         `json:"slug"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	Sku         string         `json:"sku"`
	CategoryID  sql.NullInt64  `json:"category_id"`
	Status      bool           `json:"status"`
	Visibility  bool           `json:"visibility"`
	CreatedAt   sql.NullTime   `json:"created_at"`
}

type ProductImage struct {
	ProductID uint64 `json:"product_id"`
	ImageID   uint64 `json:"image_id"`
}

type Purchase struct {
	ID           uint64        `json:"id"`
	ProductID    uint64        `json:"product_id"`
	Date         time.Time     `json:"date"`
	Quantity     int32         `json:"quantity"`
	OrderPrice   float64       `json:"order_price"`
	SellingPrice float64       `json:"selling_price"`
	StoreID      sql.NullInt64 `json:"store_id"`
	UserID       sql.NullInt64 `json:"user_id"`
	CreatedAt    sql.NullTime  `json:"created_at"`
	UpdatedAt    sql.NullTime  `json:"updated_at"`
}

type Role struct {
	ID        uint64       `json:"id"`
	Name      string       `json:"name"`
	CreatedAt sql.NullTime `json:"created_at"`
}

type Store struct {
	ID     uint64 `json:"id"`
	Slug   string `json:"slug"`
	Name   string `json:"name"`
	Status bool   `json:"status"`
}

type StoreUser struct {
	UserID  uint64 `json:"user_id"`
	StoreID uint64 `json:"store_id"`
}

type User struct {
	ID        uint64         `json:"id"`
	Firstname string         `json:"firstname"`
	Lastname  string         `json:"lastname"`
	Email     string         `json:"email"`
	Phone     sql.NullString `json:"phone"`
	Password  string         `json:"password"`
}

type UserRole struct {
	UserID uint64 `json:"user_id"`
	RoleID uint64 `json:"role_id"`
}
