package router

import (
	"api/handler"
	"api/repository"

	"github.com/go-chi/chi/v5"
)

func (a *API) CashierRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/orders", a.CashierInStoreOrdersRoutes)
	router.Route("/profile", a.CashierProfileRoutes)

	return router
}

func (a *API) CashierInStoreOrdersRoutes(router chi.Router) {

	repo := repository.New(a.db)
	handle := handler.NewOrderHandler(a.db, repo)

	router.Group(func(r chi.Router) {

		r.Use(a.auth.AuthJWT)

		r.Post("/", handle.CreateInStoreOrder)
		r.Get("/", handle.CashierFindInStoreOrders)
		r.Get("/{orderID}/store/{storeID}", handle.CashierFindInStoreOrder)
	})
}

func (a *API) CashierProfileRoutes(router chi.Router) {

	repo := repository.New(a.db)
	handle := handler.NewProfileHandler(repo)

	router.Group(func(r chi.Router) {

		r.Use(a.auth.AuthJWT)

		r.Get("/", handle.CashierProfile)
	})
}
