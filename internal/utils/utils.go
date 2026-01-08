// Package utils
package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/floriangaechter/rss/internal/middleware"
	"github.com/floriangaechter/rss/internal/store"
	"github.com/go-chi/chi/v5"
)

type Envelope map[string]any

func WriteJSON(w http.ResponseWriter, status int, data Envelope) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(js); err != nil {
		return err
	}

	return nil
}

func ReadIDParam(r *http.Request) (int64, error) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		return 0, errors.New("invalid id")
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return 0, errors.New("invalid id type")
	}

	return id, nil
}

func ParseFeedDate(dateStr string) string {
	if dateStr == "" {
		return time.Now().Format("2006-01-02 00:00:00")
	}

	// Parse RFC 1123Z format (RSS spec)
	t, err := time.Parse(time.RFC1123Z, dateStr)
	if err != nil {
		// If parsing fails, return current date
		return time.Now().Format("2006-01-02 00:00:00")
	}

	return t.Format("2006-01-02 00:00:00")
}

// GetUserFromContext retrieves the authenticated user from the request context
func GetUserFromContext(r *http.Request) *store.User {
	user, ok := r.Context().Value(middleware.UserContextKey).(*store.User)
	if !ok {
		return nil
	}
	return user
}
