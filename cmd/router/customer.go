package router

import (
	"api/handler"
	"api/repository"

	"github.com/go-chi/chi/v5"
)

func (a *API) CustomerRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/orders", a.CustomerOnlineOrdersRoutes)
	router.Route("/products", a.CustomerProductsRoutes)

	return router
}

func (a *API) CustomerOnlineOrdersRoutes(router chi.Router) {

	repo := repository.New(a.db)
	handle := handler.NewOrderHandler(a.db, repo)

	router.Group(func(r chi.Router) {

		r.Use(a.auth.AuthJWT)

		r.Post("/", handle.CreateOnlineOrder)
	})
}

func (a *API) CustomerProductsRoutes(router chi.Router) {

	repo := repository.New(a.db)
	handle := handler.NewProductHandler(repo)

	router.Group(func(r chi.Router) {

		r.Use(a.auth.AuthJWT)

		r.Get("/{sku}", handle.CustomerFindOne)
	})
}
