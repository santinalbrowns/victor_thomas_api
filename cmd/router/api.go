package router

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	mid "api/cmd/middleware"

	"api/cmd/helper"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type API struct {
	db     *sql.DB
	auth   *mid.Auth
	issuer *helper.Issuer
}

func New(db *sql.DB, issuer *helper.Issuer) *API {
	return &API{db: db, issuer: issuer}
}

func (api *API) Serve(ctx context.Context) error {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://fixchirp.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.Use(middleware.Logger)
	router.Use(middleware.Heartbeat("/"))

	auth, err := mid.NewAuth(os.Getenv("JWT_PUB_CERT_PATH"))
	if err != nil {
		return err
	}

	api.auth = auth

	router.Mount("/admin", api.Routes())
	router.Mount("/auth", api.AuthRoutes())
	router.Mount("/cashier", api.CashierRoutes())
	router.Mount("/customer", api.CustomerRoutes())

	router.Route("/images", api.ImagesRoutes)
	router.Route("/thumbnails", api.ThumbnailRoutes)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", os.Getenv("PORT")),
		Handler: router,
	}

	ch := make(chan error, 1)

	log.Printf("server listening on port: %v", os.Getenv("PORT"))

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}

		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		return server.Shutdown(timeout)
	}

}

func (s *API) ImagesRoutes(router chi.Router) {
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/images/", http.FileServer(http.Dir(os.Getenv("IMAGES_PATH")))).ServeHTTP(w, r)
	})
}

func (s *API) ThumbnailRoutes(router chi.Router) {
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/thumbnails/", http.FileServer(http.Dir(os.Getenv("THUMBNAILS_PATH")))).ServeHTTP(w, r)
	})
}
