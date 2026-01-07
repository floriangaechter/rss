// Package routes houses all the routes
package routes

import (
	"github.com/floriangaechter/rss/internal/app"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", app.HealthCheck)
	r.Get("/feeds/{id}", app.FeedHandler.HandleGetFeedByID)
	r.Post("/feeds", app.FeedHandler.HandleCreateFeed)
	r.Put("/feeds/{id}", app.FeedHandler.HandleUpdateFeedByID)
	r.Delete("/feeds/{id}", app.FeedHandler.HandleDeleteFeedByID)

	return r
}
