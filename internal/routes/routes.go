// Package routes houses all the routes
package routes

import (
	"net/http"

	"github.com/floriangaechter/rss/internal/app"
	"github.com/floriangaechter/rss/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	// Serve static files from /static/ path
	r.Mount("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Public routes
	r.Get("/", app.PageHander.HandleHome)
	r.Get("/health", app.HealthCheck)
	r.Post("/users", app.UserHandler.HandleCreateUser)
	r.Post("/login", app.UserHandler.HandleLogin)

	// Protected routes - require authentication
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(app.SessionStore, app.UserStore, app.Logger))

		r.Get("/dashboard", app.PageHander.HandleDashboard)
		r.Post("/logout", app.UserHandler.HandleLogout)
		r.Get("/feeds/{id}", app.FeedHandler.HandleGetFeedByID)
		r.Post("/feeds", app.FeedHandler.HandleCreateFeed)
		r.Put("/feeds/{id}", app.FeedHandler.HandleUpdateFeedByID)
		r.Delete("/feeds/{id}", app.FeedHandler.HandleDeleteFeedByID)
		r.Post("/feeds/{id}/fetch", app.FeedHandler.HandleFetchFeedItems)
	})

	return r
}
