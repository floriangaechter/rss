// Package app provides a wrapper for the application
package app

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/floriangaechter/rss/internal/api"
	"github.com/floriangaechter/rss/internal/fetcher"
	"github.com/floriangaechter/rss/internal/store"
	"github.com/floriangaechter/rss/internal/utils"
	"github.com/floriangaechter/rss/migrations"
)

type Application struct {
	Logger       *log.Logger
	FeedHandler  *api.FeedHandler
	UserHandler  *api.UserHandler
	PageHander   *api.PageHandler
	SessionStore store.SessionStore
	UserStore    store.UserStore
	DB           *sql.DB
}

func NewApplication() (*Application, error) {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	sqliteDB, err := store.Open(logger)
	if err != nil {
		return nil, err
	}

	err = store.MigrateFS(sqliteDB, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	feedStore := store.NewSqlite3FeedStore(sqliteDB)
	feedItemStore := store.NewSqlite3FeedItemStore(sqliteDB)
	userStore := store.NewSqlite3UserStore(sqliteDB)
	sessionStore := store.NewSqlite3SessionStore(sqliteDB)

	fetcher := fetcher.NewFetcher(feedStore, feedItemStore, logger)

	feedHandler := api.NewFeedHanlder(feedStore, feedItemStore, fetcher, logger)
	userHandler := api.NewUserHandler(userStore, sessionStore, logger)
	pageHandler := api.NewPageHandler(feedStore, logger)

	app := &Application{
		Logger:       logger,
		FeedHandler:  feedHandler,
		UserHandler:  userHandler,
		PageHander:   pageHandler,
		DB:           sqliteDB,
		SessionStore: sessionStore,
		UserStore:    userStore,
	}

	return app, nil
}

func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	_ = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"ping": "pong"})
}
