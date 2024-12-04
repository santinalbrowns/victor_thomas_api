package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"api/cmd/middleware"
	"api/handler/dto"
	"api/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	"github.com/santinalbrowns/paychangu"
)

type orderHandler struct {
	repo *repository.Queries
	db   *sql.DB
}

func NewOrderHandler(db *sql.DB, repo *repository.Queries) *orderHandler {
	return &orderHandler{db: db, repo: repo}
}

func (h *orderHandler) CreateInStoreOrder(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	cashierID, err := middleware.GuardCashier(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var form dto.CreateStoreOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(form); err != nil {
		var msg string

		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				msg = fmt.Sprintf("%s is a required field", err.Field())
			case "gte":
				msg = fmt.Sprintf("%s should at least be greater than %s", err.Field(), err.Param())
			case "email":
				msg = fmt.Sprintf("%s provided is invalid", err.Field())
			}
		}

		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	total := calculateTotal(form.Items)
	var number string

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	repo := h.repo.WithTx(tx)

	if len(form.Items) < 1 {
		http.Error(w, "Please add items", http.StatusBadRequest)
		return
	}

	u, err := repo.FindUserByID(ctx, cashierID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Cashier not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	hasRole, err := repo.CheckStoreUser(ctx, repository.CheckStoreUserParams{
		StoreID: form.StoreID,
		UserID:  u.ID,
	})
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if !hasRole {
		http.Error(w, "Not allowed to perform this task", http.StatusForbidden)
		return
	}

	o, err := repo.FindLastCreatedOrder(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			number, err = incrementNumber("00000")
			if err != nil {
				http.Error(w, "Failed to create order number", http.StatusInternalServerError)
				return
			}

		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	} else {
		number, err = incrementNumber(o.Number)
		if err != nil {
			http.Error(w, "Failed to create order number", http.StatusInternalServerError)
			return
		}
	}

	s, err := repo.FindStore(ctx, form.StoreID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Store not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	orderID, err := repo.InsertOrder(ctx, repository.InsertOrderParams{
		Number:  number,
		Total:   total,
		Channel: repository.OrdersChannelInStore,
		Status:  repository.OrdersStatusPending,
	})
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	for _, item := range form.Items {
		p, err := repo.FindProductBySKU(ctx, item.SKU)
		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		if !p.Status {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Sorry, you cannot order item SKU: %s", p.Sku), http.StatusBadRequest)
			return
		}

		stock, err := repo.FindProductStock(ctx, p.ID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		if item.Quantity > stock.Remaining {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Sorry, insufficient stock for the requested quantity of item SKU: %s", p.Sku), http.StatusBadRequest)
			return
		}

		if item.Quantity < 1 {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("The minimum order quantity for item SKU: %s is 1", p.Sku), http.StatusBadRequest)
			return
		}

		err = repo.InsertOrderItem(ctx, repository.InsertOrderItemParams{
			OrderID:   uint64(orderID),
			ProductID: p.ID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to add order item", http.StatusInternalServerError)
			return
		}
	}

	err = repo.InsertInStoreOrderDetails(ctx, repository.InsertInStoreOrderDetailsParams{
		OrderID:   uint64(orderID),
		CashierID: sql.NullInt64{Int64: int64(cashierID), Valid: true},
		StoreID:   s.ID,
	})
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to add order details", http.StatusInternalServerError)
		return
	}

	orderResult, err := repo.FindOrder(ctx, uint64(orderID))
	if err != nil {
		tx.Rollback()
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	orderItems, err := repo.FindOrderItems(ctx, orderResult.ID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var items = []dto.ItemResponse{}

	for _, item := range orderItems {

		p, err := repo.FindProduct(ctx, item.ProductID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		imagesResults, err := h.repo.FindProductImages(ctx, p.ID)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		var images = []dto.ImageResponse{}

		for _, i := range imagesResults {
			images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
		}

		items = append(items, dto.ItemResponse{
			ID:         p.ID,
			Slug:       p.Slug,
			Name:       p.Name,
			SKU:        p.Sku,
			Status:     p.Status,
			Visibility: p.Visibility,
			Images:     images,
			Quantity:   item.Quantity,
			Price:      item.Price,
		})
	}

	store := dto.StoreResponse{
		ID:     s.ID,
		Slug:   s.Slug,
		Name:   s.Name,
		Status: s.Status,
	}

	cashier := dto.UserResponse{
		ID:        u.ID,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
	}

	if u.Phone.Valid {
		cashier.Phone = u.Phone.String
	}

	order := dto.StoreOrderResponse{
		ID:      orderResult.ID,
		Number:  orderResult.Number,
		Channel: string(orderResult.Channel),
		Status:  string(orderResult.Status),
		Total:   orderResult.Total,
		Items:   items,
		Details: dto.StoreOrderDetails{
			Store:   store,
			Cashier: cashier,
		},
		CreatedAt: orderResult.CreatedAt.Time.UTC().Format(time.RFC3339),
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *orderHandler) CreateOnlineOrder(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	customerID, err := middleware.GuardCustomer(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var form dto.CreateOnlineOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(form); err != nil {
		var msg string

		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				msg = fmt.Sprintf("%s is a required field", err.Field())
			case "gte":
				msg = fmt.Sprintf("%s should at least be greater than %s", err.Field(), err.Param())
			case "email":
				msg = fmt.Sprintf("%s provided is invalid", err.Field())
			}
		}

		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	total := calculateTotal(form.Items)
	var number string

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	repo := h.repo.WithTx(tx)

	if len(form.Items) < 1 {
		http.Error(w, "Please add items", http.StatusBadRequest)
		return
	}

	u, err := repo.FindUserByID(ctx, customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	o, err := repo.FindLastCreatedOrder(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			number, err = incrementNumber("00000")
			if err != nil {
				http.Error(w, "Failed to create order number", http.StatusInternalServerError)
				return
			}

		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	} else {
		number, err = incrementNumber(o.Number)
		if err != nil {
			http.Error(w, "Failed to create order number", http.StatusInternalServerError)
			return
		}
	}

	orderID, err := repo.InsertOrder(ctx, repository.InsertOrderParams{
		Number:  number,
		Total:   total,
		Channel: repository.OrdersChannelOnline,
		Status:  repository.OrdersStatusPending,
	})
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	for _, item := range form.Items {
		p, err := repo.FindProductBySKU(ctx, item.SKU)
		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		if !p.Status {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Sorry, you cannot order item SKU: %s", p.Sku), http.StatusBadRequest)
			return
		}

		if !p.Visibility {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Sorry, you cannot order item SKU: %s", p.Sku), http.StatusBadRequest)
			return
		}

		stock, err := repo.FindProductStock(ctx, p.ID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		if item.Quantity > stock.Remaining {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Sorry, insufficient stock for the requested quantity of item SKU: %s", p.Sku), http.StatusBadRequest)
			return
		}

		if item.Quantity < 1 {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("The minimum order quantity for item SKU: %s is 1", p.Sku), http.StatusBadRequest)
			return
		}

		err = repo.InsertOrderItem(ctx, repository.InsertOrderItemParams{
			OrderID:   uint64(orderID),
			ProductID: p.ID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to add order item", http.StatusInternalServerError)
			return
		}
	}

	err = repo.InsertOnlineOrderDetails(ctx, repository.InsertOnlineOrderDetailsParams{
		OrderID:    uint64(orderID),
		CustomerID: sql.NullInt64{Int64: int64(customerID), Valid: true},
	})
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to add order details", http.StatusInternalServerError)
		return
	}

	orderResult, err := repo.FindOrder(ctx, uint64(orderID))
	if err != nil {
		tx.Rollback()
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	orderItems, err := repo.FindOrderItems(ctx, orderResult.ID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var items = []dto.ItemResponse{}

	for _, item := range orderItems {

		p, err := repo.FindProduct(ctx, item.ProductID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		imagesResults, err := h.repo.FindProductImages(ctx, p.ID)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		var images = []dto.ImageResponse{}

		for _, i := range imagesResults {
			images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
		}

		items = append(items, dto.ItemResponse{
			ID:         p.ID,
			Slug:       p.Slug,
			Name:       p.Name,
			SKU:        p.Sku,
			Status:     p.Status,
			Visibility: p.Visibility,
			Images:     images,
			Quantity:   item.Quantity,
			Price:      item.Price,
		})
	}

	customer := dto.UserResponse{
		ID:        u.ID,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
	}

	if u.Phone.Valid {
		customer.Phone = u.Phone.String
	}

	client := paychangu.New("SEC-W3BIN3UUUUq9tACnLOjeCYL5tFCg1YdS")

	//TODO change the IP Address
	request := paychangu.Request{
		Amount:      10500,
		Currency:    "MWK",
		FirstName:   customer.Firstname,
		LastName:    customer.Lastname,
		Email:       customer.Email,
		CallbackURL: "http://192.168.0.101:5000/success",
		ReturnURL:   "http://192.168.0.101:5000/cancel",
		TxRef:       strconv.Itoa(int(orderID)),
		Customization: struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}{
			Title:       "Service Payment",
			Description: "Payment for services rendered",
		},
		Meta: struct {
			UUID     string `json:"uuid"`
			Response string `json:"response"`
		}{
			UUID:     "unique_user_identifier",
			Response: "custom_response_data",
		},
	}

	response, err := client.InitiatePayment(request)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	order := dto.OnlineOrderResponse{
		ID:      orderResult.ID,
		Number:  orderResult.Number,
		Channel: response.Data.CheckoutURL,
		Status:  string(orderResult.Status),
		Total:   orderResult.Total,
		Items:   items,
		Details: dto.OnlineOrderDetails{
			Customer: customer,
		},
		CreatedAt: orderResult.CreatedAt.Time.UTC().Format(time.RFC3339),
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *orderHandler) AdminFindInStoreOrders(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters for pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	storeIDStr := r.URL.Query().Get("store")
	if storeIDStr == "" {
		http.Error(w, "Provide store paraman from URL query", http.StatusBadRequest)
		return
	}

	// Convert limit and offset to integers
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20 // Default limit
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0 // Default offset
	}

	if storeIDStr == string(repository.OrdersChannelOnline) {

		count, err := h.repo.CountOnlineOrders(ctx)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		results, err := h.repo.FindOnlineOrders(ctx, repository.FindOnlineOrdersParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		var orders = []dto.OnlineOrderResponse{}

		for _, or := range results {

			orderItems, err := h.repo.FindOrderItems(ctx, or.ID)
			if err != nil {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
				return
			}

			var items = []dto.ItemResponse{}

			for _, item := range orderItems {

				p, err := h.repo.FindProduct(ctx, item.ProductID)
				if err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "Product not found", http.StatusNotFound)
					} else {
						http.Error(w, "Something went wrong", http.StatusInternalServerError)
					}
					return
				}

				imagesResults, err := h.repo.FindProductImages(ctx, p.ID)
				if err != nil {
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
					return
				}

				var images = []dto.ImageResponse{}

				for _, i := range imagesResults {
					images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
				}

				items = append(items, dto.ItemResponse{
					ID:         p.ID,
					Slug:       p.Slug,
					Name:       p.Name,
					SKU:        p.Sku,
					Status:     p.Status,
					Visibility: p.Visibility,
					Images:     images,
					Quantity:   item.Quantity,
					Price:      item.Price,
				})
			}

			od, err := h.repo.FindOnlineOrderDetails(ctx, or.ID)
			if err != nil {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
				return
			}

			//TODO Cashier can be removed
			u, err := h.repo.FindUserByID(ctx, uint64(od.CustomerID.Int64))
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Customer not found", http.StatusNotFound)
				} else {
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
				}
				return
			}

			customer := dto.UserResponse{
				ID:        u.ID,
				Firstname: u.Firstname,
				Lastname:  u.Lastname,
				Email:     u.Email,
			}

			if u.Phone.Valid {
				customer.Phone = u.Phone.String
			}

			order := dto.OnlineOrderResponse{
				ID:      or.ID,
				Number:  or.Number,
				Channel: string(or.Channel),
				Status:  string(or.Status),
				Total:   or.Total,
				Items:   items,
				Details: dto.OnlineOrderDetails{
					Customer: customer,
				},
				CreatedAt: or.CreatedAt.Time.UTC().Format(time.RFC3339),
			}

			orders = append(orders, order)
		}

		response := map[string]interface{}{
			"total":  count,
			"limit":  limit,
			"offset": offset,
			"data":   orders,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

	} else {
		// Get the category ID from URL query
		storeID, err := strconv.ParseUint(storeIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid store ID", http.StatusBadRequest)
			return
		}

		s, err := h.repo.FindStore(ctx, storeID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Store not found", http.StatusNotFound)
			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		count, err := h.repo.CountStoreOrders(ctx, s.ID)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		results, err := h.repo.FindStoreOrders(ctx, repository.FindStoreOrdersParams{
			ID:     s.ID,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		var orders = []dto.StoreOrderResponse{}

		for _, or := range results {

			orderItems, err := h.repo.FindOrderItems(ctx, or.ID)
			if err != nil {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
				return
			}

			var items = []dto.ItemResponse{}

			for _, item := range orderItems {

				p, err := h.repo.FindProduct(ctx, item.ProductID)
				if err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "Product not found", http.StatusNotFound)
					} else {
						http.Error(w, "Something went wrong", http.StatusInternalServerError)
					}
					return
				}

				imagesResults, err := h.repo.FindProductImages(ctx, p.ID)
				if err != nil {
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
					return
				}

				var images = []dto.ImageResponse{}

				for _, i := range imagesResults {
					images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
				}

				items = append(items, dto.ItemResponse{
					ID:         p.ID,
					Slug:       p.Slug,
					Name:       p.Name,
					SKU:        p.Sku,
					Status:     p.Status,
					Visibility: p.Visibility,
					Images:     images,
					Quantity:   item.Quantity,
					Price:      item.Price,
				})
			}

			od, err := h.repo.FindStoreOrderDetails(ctx, or.ID)
			if err != nil {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
				return
			}

			store := dto.StoreResponse{
				ID:     s.ID,
				Slug:   s.Slug,
				Name:   s.Name,
				Status: s.Status,
			}

			//TODO Cashier can be removed
			u, err := h.repo.FindUserByID(ctx, uint64(od.CashierID.Int64))
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Cashier not found", http.StatusNotFound)
				} else {
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
				}
				return
			}

			cashier := dto.UserResponse{
				ID:        u.ID,
				Firstname: u.Firstname,
				Lastname:  u.Lastname,
				Email:     u.Email,
			}

			if u.Phone.Valid {
				cashier.Phone = u.Phone.String
			}

			order := dto.StoreOrderResponse{
				ID:      or.ID,
				Number:  or.Number,
				Channel: string(or.Channel),
				Status:  string(or.Status),
				Total:   or.Total,
				Items:   items,
				Details: dto.StoreOrderDetails{
					Store:   store,
					Cashier: cashier,
				},
				CreatedAt: or.CreatedAt.Time.UTC().Format(time.RFC3339),
			}

			orders = append(orders, order)
		}

		response := map[string]interface{}{
			"total":  count,
			"limit":  limit,
			"offset": offset,
			"data":   orders,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func (h *orderHandler) AdminFindInStoreOrder(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the order ID from URL params
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid store ID", http.StatusBadRequest)
		return
	}

	or, err := h.repo.FindOrderWithChannel(ctx, repository.FindOrderWithChannelParams{
		ID:      orderID,
		Channel: repository.OrdersChannelInStore,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	orderItems, err := h.repo.FindOrderItems(ctx, or.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order items not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	var items = []dto.ItemResponse{}

	for _, item := range orderItems {

		p, err := h.repo.FindProduct(ctx, item.ProductID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		imagesResults, err := h.repo.FindProductImages(ctx, p.ID)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		var images = []dto.ImageResponse{}

		for _, i := range imagesResults {
			images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
		}

		items = append(items, dto.ItemResponse{
			ID:         p.ID,
			Slug:       p.Slug,
			Name:       p.Name,
			SKU:        p.Sku,
			Status:     p.Status,
			Visibility: p.Visibility,
			Images:     images,
			Quantity:   item.Quantity,
			Price:      item.Price,
		})
	}

	od, err := h.repo.FindStoreOrderDetails(ctx, or.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order details not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	s, err := h.repo.FindStore(ctx, od.StoreID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Store not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	store := dto.StoreResponse{
		ID:     s.ID,
		Slug:   s.Slug,
		Name:   s.Name,
		Status: s.Status,
	}

	//TODO Cashier can be removed
	u, err := h.repo.FindUserByID(ctx, uint64(od.CashierID.Int64))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Cashier not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	cashier := dto.UserResponse{
		ID:        u.ID,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
	}

	if u.Phone.Valid {
		cashier.Phone = u.Phone.String
	}

	order := dto.StoreOrderResponse{
		ID:      or.ID,
		Number:  or.Number,
		Channel: string(or.Channel),
		Status:  string(or.Status),
		Total:   or.Total,
		Items:   items,
		Details: dto.StoreOrderDetails{
			Store:   store,
			Cashier: cashier,
		},
		CreatedAt: or.CreatedAt.Time.UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func (h *orderHandler) AdminReport(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	countCategories, err := h.repo.CountCategories(ctx)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	countProducts, err := h.repo.CountProducts(ctx)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	countPurchases, err := h.repo.CountPurchases(ctx)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	countSales, err := h.repo.CountOnlineOrders(ctx)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	response := dto.Report{
		Categories: countCategories,
		Products:   countProducts,
		Purchases:  countPurchases,
		Sales:      countSales,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func (h *orderHandler) CashierFindInStoreOrders(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	cashierID, err := middleware.GuardCashier(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters for pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	storeIDStr := r.URL.Query().Get("store")
	if storeIDStr == "" {
		http.Error(w, "Provide store paraman from URL query", http.StatusBadRequest)
		return
	}

	// Convert limit and offset to integers
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20 // Default limit
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0 // Default offset
	}

	// Get the category ID from URL query
	storeID, err := strconv.ParseUint(storeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid store ID", http.StatusBadRequest)
		return
	}

	s, err := h.repo.FindStore(ctx, storeID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Store not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	exits, err := h.repo.CheckStoreUser(ctx, repository.CheckStoreUserParams{
		StoreID: s.ID,
		UserID:  cashierID,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}

	if !exits {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	count, err := h.repo.CountStoreOrders(ctx, s.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	results, err := h.repo.FindStoreOrders(ctx, repository.FindStoreOrdersParams{
		ID:     s.ID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var orders = []dto.StoreOrderResponse{}

	for _, or := range results {

		orderItems, err := h.repo.FindOrderItems(ctx, or.ID)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		var items = []dto.ItemResponse{}

		for _, item := range orderItems {

			p, err := h.repo.FindProduct(ctx, item.ProductID)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Product not found", http.StatusNotFound)
				} else {
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
				}
				return
			}

			imagesResults, err := h.repo.FindProductImages(ctx, p.ID)
			if err != nil {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
				return
			}

			var images = []dto.ImageResponse{}

			for _, i := range imagesResults {
				images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
			}

			items = append(items, dto.ItemResponse{
				ID:         p.ID,
				Slug:       p.Slug,
				Name:       p.Name,
				SKU:        p.Sku,
				Status:     p.Status,
				Visibility: p.Visibility,
				Images:     images,
				Quantity:   item.Quantity,
				Price:      item.Price,
			})
		}

		od, err := h.repo.FindStoreOrderDetails(ctx, or.ID)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		store := dto.StoreResponse{
			ID:     s.ID,
			Slug:   s.Slug,
			Name:   s.Name,
			Status: s.Status,
		}

		//TODO Cashier can be removed
		u, err := h.repo.FindUserByID(ctx, uint64(od.CashierID.Int64))
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Cashier not found", http.StatusNotFound)
			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		cashier := dto.UserResponse{
			ID:        u.ID,
			Firstname: u.Firstname,
			Lastname:  u.Lastname,
			Email:     u.Email,
		}

		if u.Phone.Valid {
			cashier.Phone = u.Phone.String
		}

		order := dto.StoreOrderResponse{
			ID:      or.ID,
			Number:  or.Number,
			Channel: string(or.Channel),
			Status:  string(or.Status),
			Total:   or.Total,
			Items:   items,
			Details: dto.StoreOrderDetails{
				Store:   store,
				Cashier: cashier,
			},
			CreatedAt: or.CreatedAt.Time.UTC().Format(time.RFC3339),
		}

		orders = append(orders, order)
	}

	response := map[string]interface{}{
		"total":  count,
		"limit":  limit,
		"offset": offset,
		"data":   orders,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *orderHandler) CashierFindInStoreOrder(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	cashierID, err := middleware.GuardCashier(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the order ID from URL params
	orderIDStr := chi.URLParam(r, "orderID")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Get the order ID from URL params
	storeIDStr := chi.URLParam(r, "storeID")
	storeID, err := strconv.ParseUint(storeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid store ID", http.StatusBadRequest)
		return
	}

	s, err := h.repo.FindStore(ctx, storeID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Store not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	exits, err := h.repo.CheckStoreUser(ctx, repository.CheckStoreUserParams{
		StoreID: s.ID,
		UserID:  cashierID,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}

	if !exits {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	or, err := h.repo.FindOrderWithChannel(ctx, repository.FindOrderWithChannelParams{
		ID:      orderID,
		Channel: repository.OrdersChannelInStore,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	orderItems, err := h.repo.FindOrderItems(ctx, or.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order items not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	var items = []dto.ItemResponse{}

	for _, item := range orderItems {

		p, err := h.repo.FindProduct(ctx, item.ProductID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		imagesResults, err := h.repo.FindProductImages(ctx, p.ID)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		var images = []dto.ImageResponse{}

		for _, i := range imagesResults {
			images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
		}

		items = append(items, dto.ItemResponse{
			ID:         p.ID,
			Slug:       p.Slug,
			Name:       p.Name,
			SKU:        p.Sku,
			Status:     p.Status,
			Visibility: p.Visibility,
			Images:     images,
			Quantity:   item.Quantity,
			Price:      item.Price,
		})
	}

	od, err := h.repo.FindStoreOrderDetails(ctx, or.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order details not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	store := dto.StoreResponse{
		ID:     s.ID,
		Slug:   s.Slug,
		Name:   s.Name,
		Status: s.Status,
	}

	//TODO Cashier can be removed
	u, err := h.repo.FindUserByID(ctx, uint64(od.CashierID.Int64))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Cashier not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	cashier := dto.UserResponse{
		ID:        u.ID,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
	}

	if u.Phone.Valid {
		cashier.Phone = u.Phone.String
	}

	order := dto.StoreOrderResponse{
		ID:      or.ID,
		Number:  or.Number,
		Channel: string(or.Channel),
		Status:  string(or.Status),
		Total:   or.Total,
		Items:   items,
		Details: dto.StoreOrderDetails{
			Store:   store,
			Cashier: cashier,
		},
		CreatedAt: or.CreatedAt.Time.UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func (h *orderHandler) CustomerFindOnlineOrder(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get the order ID from URL params
	sku := chi.URLParam(r, "sku")
	io, err := h.repo.FindOrderItemByProductSKU(ctx, sku)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order Item not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	od, err := h.repo.FindOnlineOrderDetails(ctx, io.OrderID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order details not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	or, err := h.repo.FindOrderWithChannel(ctx, repository.FindOrderWithChannelParams{
		ID:      io.OrderID,
		Channel: repository.OrdersChannelOnline,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	orderItems, err := h.repo.FindOrderItems(ctx, or.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order items not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	var items = []dto.ItemResponse{}

	for _, item := range orderItems {

		p, err := h.repo.FindProduct(ctx, item.ProductID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		imagesResults, err := h.repo.FindProductImages(ctx, p.ID)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		var images = []dto.ImageResponse{}

		for _, i := range imagesResults {
			images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
		}

		items = append(items, dto.ItemResponse{
			ID:         p.ID,
			Slug:       p.Slug,
			Name:       p.Name,
			SKU:        p.Sku,
			Status:     p.Status,
			Visibility: p.Visibility,
			Images:     images,
			Quantity:   item.Quantity,
			Price:      item.Price,
		})
	}

	od, err = h.repo.FindOnlineOrderDetails(ctx, or.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order details not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	u, err := h.repo.FindUserByID(ctx, uint64(od.CustomerID.Int64))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	customer := dto.UserResponse{
		ID:        u.ID,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
	}

	if u.Phone.Valid {
		customer.Phone = u.Phone.String
	}

	order := dto.OnlineOrderResponse{
		ID:      or.ID,
		Number:  or.Number,
		Channel: string(or.Channel),
		Status:  string(or.Status),
		Total:   or.Total,
		Items:   items,
		Details: dto.OnlineOrderDetails{
			Customer: customer,
		},
		CreatedAt: or.CreatedAt.Time.UTC().Format(time.RFC3339),
	}

	if or.Status == repository.OrdersStatusCompleted {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(order)
	} else {
		http.Error(w, "Payment not clear", http.StatusForbidden)
		return
	}
}

func (h *orderHandler) CustomerOrderPaid(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get the order ID from URL params
	orderIDStr := chi.URLParam(r, "orderID")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	o, err := h.repo.FindOrder(ctx, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	err = h.repo.UpdateOrderStatus(ctx, repository.UpdateOrderStatusParams{
		Status: repository.OrdersStatusCompleted,
		ID:     o.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *orderHandler) CustomerFindOnlineOrders(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, err := middleware.GuardCustomer(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters for pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Convert limit and offset to integers
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20 // Default limit
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0 // Default offset
	}

	count, err := h.repo.CountOnlineOrders(ctx)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	results, err := h.repo.FindOnlineOrders(ctx, repository.FindOnlineOrdersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var orders = []dto.OnlineOrderResponse{}

	for _, or := range results {

		orderItems, err := h.repo.FindOrderItems(ctx, or.ID)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		var items = []dto.ItemResponse{}

		for _, item := range orderItems {

			p, err := h.repo.FindProduct(ctx, item.ProductID)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Product not found", http.StatusNotFound)
				} else {
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
				}
				return
			}

			imagesResults, err := h.repo.FindProductImages(ctx, p.ID)
			if err != nil {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
				return
			}

			var images = []dto.ImageResponse{}

			for _, i := range imagesResults {
				images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
			}

			items = append(items, dto.ItemResponse{
				ID:         p.ID,
				Slug:       p.Slug,
				Name:       p.Name,
				SKU:        p.Sku,
				Status:     p.Status,
				Visibility: p.Visibility,
				Images:     images,
				Quantity:   item.Quantity,
				Price:      item.Price,
			})
		}

		od, err := h.repo.FindOnlineOrderDetails(ctx, or.ID)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		//TODO Cashier can be removed
		u, err := h.repo.FindUserByID(ctx, uint64(od.CustomerID.Int64))
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Customer not found", http.StatusNotFound)
			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		customer := dto.UserResponse{
			ID:        u.ID,
			Firstname: u.Firstname,
			Lastname:  u.Lastname,
			Email:     u.Email,
		}

		if u.Phone.Valid {
			customer.Phone = u.Phone.String
		}

		order := dto.OnlineOrderResponse{
			ID:      or.ID,
			Number:  or.Number,
			Channel: string(or.Channel),
			Status:  string(or.Status),
			Total:   or.Total,
			Items:   items,
			Details: dto.OnlineOrderDetails{
				Customer: customer,
			},
			CreatedAt: or.CreatedAt.Time.UTC().Format(time.RFC3339),
		}

		orders = append(orders, order)
	}

	response := map[string]interface{}{
		"total":  count,
		"limit":  limit,
		"offset": offset,
		"data":   orders,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func (h *orderHandler) CustomerGetOnlineOrder(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, err := middleware.GuardCustomer(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the order ID from URL params
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid store ID", http.StatusBadRequest)
		return
	}

	or, err := h.repo.FindOrderWithChannel(ctx, repository.FindOrderWithChannelParams{
		ID:      orderID,
		Channel: repository.OrdersChannelOnline,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	orderItems, err := h.repo.FindOrderItems(ctx, or.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order items not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	var items = []dto.ItemResponse{}

	for _, item := range orderItems {

		p, err := h.repo.FindProduct(ctx, item.ProductID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		imagesResults, err := h.repo.FindProductImages(ctx, p.ID)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		var images = []dto.ImageResponse{}

		for _, i := range imagesResults {
			images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
		}

		items = append(items, dto.ItemResponse{
			ID:         p.ID,
			Slug:       p.Slug,
			Name:       p.Name,
			SKU:        p.Sku,
			Status:     p.Status,
			Visibility: p.Visibility,
			Images:     images,
			Quantity:   item.Quantity,
			Price:      item.Price,
		})
	}

	od, err := h.repo.FindOnlineOrderDetails(ctx, or.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order details not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	//TODO Cashier can be removed
	u, err := h.repo.FindUserByID(ctx, uint64(od.CustomerID.Int64))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Cashier not found", http.StatusNotFound)
		} else {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	customer := dto.UserResponse{
		ID:        u.ID,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
	}

	if u.Phone.Valid {
		customer.Phone = u.Phone.String
	}

	order := dto.OnlineOrderResponse{
		ID:      or.ID,
		Number:  or.Number,
		Channel: string(or.Channel),
		Status:  string(or.Status),
		Total:   or.Total,
		Items:   items,
		Details: dto.OnlineOrderDetails{
			Customer: customer,
		},
		CreatedAt: or.CreatedAt.Time.UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func calculateTotal(items []dto.OrderItem) float64 {
	var total float64
	for _, item := range items {
		total += float64(item.Quantity) * item.Price
	}
	return total
}

func incrementNumber(num string) (string, error) {
	// Convert the input string to an integer
	number, err := strconv.Atoi(num)
	if err != nil {
		return "", err
	}

	// Increment the number
	number++

	// Format the result with leading zeroes, matching the length of the original input
	result := fmt.Sprintf("%0*d", len(num), number)
	return result, nil
}
