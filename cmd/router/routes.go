package router

import (
	"api/handler"
	"api/repository"

	"github.com/go-chi/chi/v5"
)

func (a *API) Routes() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/categories", a.CategoriesRoutes)
	router.Route("/products", a.ProductsRoutes)
	router.Route("/stores", a.StoresRoutes)
	router.Route("/purchases", a.PurchasesRoutes)
	router.Route("/orders", a.InStoreOrdersRoutes)
	router.Route("/cashiers", a.StoreUsersRoutes)

	return router
}

func (a *API) AuthRoutes() *chi.Mux {
	router := chi.NewRouter()

	repo := repository.New(a.db)

	handle := handler.NewAuthHandler(repo, a.issuer)

	router.Post("/register", handle.Register)
	router.Post("/login", handle.Login)

	return router
}

func (a *API) CategoriesRoutes(router chi.Router) {

	repo := repository.New(a.db)
	handle := handler.NewCategoryHandler(repo)

	router.Group(func(r chi.Router) {

		r.Use(a.auth.AuthJWT)

		r.Post("/", handle.Create)
		r.Get("/", handle.AdminFindAll)
		r.Get("/{id}", handle.AdminFindOne)
		r.Put("/{id}", handle.AdminUpdate)
		r.Delete("/{id}", handle.AdminDelete)
	})
}

func (a *API) ProductsRoutes(router chi.Router) {

	repo := repository.New(a.db)
	handle := handler.NewProductHandler(repo)

	router.Group(func(r chi.Router) {

		r.Use(a.auth.AuthJWT)

		r.Post("/", handle.Create)
		r.Get("/", handle.AdminFindAll)
		r.Get("/{id}", handle.AdminFindOne)
		r.Put("/{id}", handle.AdminUpdate)
		r.Delete("/{id}", handle.AdminDelete)
	})
}

func (a *API) StoresRoutes(router chi.Router) {

	repo := repository.New(a.db)
	handle := handler.NewStoreHandler(repo)

	router.Group(func(r chi.Router) {

		r.Use(a.auth.AuthJWT)

		r.Post("/", handle.Create)
		r.Get("/", handle.AdminFindAll)
		r.Get("/{id}", handle.AdminFindOne)
		r.Put("/{id}", handle.AdminUpdate)
		r.Delete("/{id}", handle.AdminDelete)
	})
}

func (a *API) PurchasesRoutes(router chi.Router) {

	repo := repository.New(a.db)
	handle := handler.NewPurchaseHandler(repo)

	router.Group(func(r chi.Router) {

		r.Use(a.auth.AuthJWT)

		r.Post("/", handle.Create)
		r.Get("/", handle.AdminFindAll)
		r.Delete("/{id}", handle.AdminDelete)
	})
}

func (a *API) InStoreOrdersRoutes(router chi.Router) {

	repo := repository.New(a.db)
	handle := handler.NewOrderHandler(a.db, repo)

	router.Group(func(r chi.Router) {

		r.Use(a.auth.AuthJWT)

		r.Post("/", handle.CreateInStoreOrder)
		r.Get("/", handle.AdminFindInStoreOrders)
		r.Get("/{id}", handle.AdminFindInStoreOrder)
	})
}

func (a *API) StoreUsersRoutes(router chi.Router) {
	repo := repository.New(a.db)
	handle := handler.NewUserHandler(repo)

	router.Group(func(r chi.Router) {

		r.Use(a.auth.AuthJWT)

		r.Post("/", handle.AdminAssignStoreUser)
		r.Get("/stores/{id}", handle.AdminFindStoreUsers)
		r.Delete("/{userID}/stores/{storeID}", handle.AdminDeleteStoreUser)
	})
}
