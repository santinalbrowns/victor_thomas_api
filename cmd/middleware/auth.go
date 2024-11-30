package middleware

import (
	"api/cmd/helper"
	"api/repository"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type tokenKey string

const tokenContextKey tokenKey = "token"

type Auth struct {
	helper.Validator
}

func NewAuth(pem string) (*Auth, error) {
	validator, err := helper.NewValidator(pem)
	if err != nil {
		return nil, fmt.Errorf("unable to create validator: %w", err)
	}

	return &Auth{
		Validator: *validator,
	}, nil
}

func (a *Auth) HandleHTTP(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.Header.Get("Authorization"), " ")

		if len(parts) < 2 || parts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(
				[]byte("Unauthorised"),
			)
			return
		}

		tokenString := parts[1]

		token, err := a.GetToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorised"))
			return
		}

		// Get a new context with the parsed token
		ctx := ContextWithToken(r.Context(), token)

		// call the next handler with the updated context
		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (a *Auth) AuthJWT(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.Header.Get("Authorization"), " ")

		if len(parts) < 2 || parts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(
				[]byte("Unauthorised"),
			)
			return
		}

		tokenString := parts[1]

		token, err := a.GetToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorised"))
			return
		}

		// Get a new context with the parsed token
		ctx := ContextWithToken(r.Context(), token)

		// call the next handler with the updated context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ContextWithToken(ctx context.Context, token *jwt.Token) context.Context {
	return context.WithValue(ctx, tokenContextKey, token)
}

func ContextGetToken(ctx context.Context) (*jwt.Token, error) {
	val := ctx.Value(tokenContextKey)
	if val == nil {
		return nil, errors.New("Unauthorised")
	}

	t, ok := val.(*jwt.Token)
	if !ok {
		return nil, errors.New("Unauthorised")
	}

	return t, nil
}

func MustContextGetToken(ctx context.Context) (*jwt.Token, error) {
	t, err := ContextGetToken(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func GuardAdmin(ctx context.Context, repo *repository.Queries) (uint64, error) {
	token, err := MustContextGetToken(ctx)
	if err != nil {
		return 0, err
	}

	sub, err := token.Claims.GetSubject()
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseUint(sub, 10, 64)
	if err != nil {
		return 0, err
	}

	isAdmin, err := repo.CheckUserRole(context.Background(), repository.CheckUserRoleParams{UserID: id, RoleID: 1})
	if err != nil {
		return 0, err
	}

	if !isAdmin {
		return 0, fmt.Errorf("user is not an admin")
	}

	return id, nil
}

func GuardCustomer(ctx context.Context, repo *repository.Queries) (uint64, error) {
	token, err := MustContextGetToken(ctx)
	if err != nil {
		return 0, err
	}

	sub, err := token.Claims.GetSubject()
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseUint(sub, 10, 64)
	if err != nil {
		return 0, err
	}

	isAdmin, err := repo.CheckUserRole(context.Background(), repository.CheckUserRoleParams{UserID: id, RoleID: 2})
	if err != nil {
		return 0, err
	}

	if !isAdmin {
		return 0, fmt.Errorf("user is not a customer")
	}

	return id, nil
}

func GuardCashier(ctx context.Context, repo *repository.Queries) (uint64, error) {
	token, err := MustContextGetToken(ctx)
	if err != nil {
		return 0, err
	}

	sub, err := token.Claims.GetSubject()
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseUint(sub, 10, 64)
	if err != nil {
		return 0, err
	}

	isCashier, err := repo.CheckUserRole(context.Background(), repository.CheckUserRoleParams{UserID: id, RoleID: 3})
	if err != nil {
		return 0, err
	}

	if !isCashier {
		return 0, fmt.Errorf("user is not a cashier")
	}

	return id, nil
}
