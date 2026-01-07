// Package routes houses all the routes
package routes

import (
	"net/http"

	"github.com/floriangaechter/rss/internal/app"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	// Serve static files from /static/ path
	r.Mount("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	r.Get("/", app.PageHander.HandleHome)

	r.Get("/health", app.HealthCheck)

	r.Get("/feeds/{id}", app.FeedHandler.HandleGetFeedByID)
	r.Post("/feeds", app.FeedHandler.HandleCreateFeed)
	r.Put("/feeds/{id}", app.FeedHandler.HandleUpdateFeedByID)
	r.Delete("/feeds/{id}", app.FeedHandler.HandleDeleteFeedByID)

	r.Post("/users", app.UserHandler.HandleCreateUser)

	return r
}
