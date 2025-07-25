package main

import (
	"context"
	"net/http"

	"github.com/Arkitecth/apollo/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) setUserContext(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)

	return r.WithContext(ctx)
}

func (app *application) getUserContext(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
