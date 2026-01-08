// Package middleware
package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/floriangaechter/rss/internal/store"
)

type contextKey string

const UserContextKey contextKey = "user"

func RequireAuth(sessionStore store.SessionStore, userStore store.UserStore, logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				handleUnauthorized(w, r, logger)
				return
			}

			session, err := sessionStore.GetSession(cookie.Value)
			if err != nil {
				logger.Printf("ERROR: getting session %v", err)
				handleUnauthorized(w, r, logger)
				return
			}
			if session == nil {
				handleUnauthorized(w, r, logger)
				return
			}

			user, err := userStore.GetUserByID(session.UserID)
			if err != nil {
				logger.Printf("ERROR: getting user %v", err)
				handleUnauthorized(w, r, logger)
				return
			}
			if user == nil {
				handleUnauthorized(w, r, logger)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func handleUnauthorized(w http.ResponseWriter, r *http.Request, logger *log.Logger) {
	// Check if HTMX request
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Regular HTTP request - redirect to login
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
