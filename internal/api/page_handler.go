package api

import (
	"log"
	"net/http"
	"text/template"

	"github.com/floriangaechter/rss/internal/store"
	"github.com/floriangaechter/rss/internal/utils"
)

type PageHandler struct {
	feedStore store.FeedStore
	logger    *log.Logger
}

func NewPageHandler(feedStore store.FeedStore, logger *log.Logger) *PageHandler {
	return &PageHandler{
		feedStore: feedStore,
		logger:    logger,
	}
}

func (h *PageHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/index.html")

	// Get error from query parameter if present
	errorMsg := r.URL.Query().Get("error")

	data := struct {
		Error string
	}{
		Error: errorMsg,
	}

	err := t.Execute(w, data)
	if err != nil {
		h.logger.Printf("ERROR: HandleHome %v", err)
		return
	}
}

func (h *PageHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	feeds, err := h.feedStore.GetFeedsByUserID(int64(user.ID))
	if err != nil {
		h.logger.Printf("ERROR: GetFeedsByUserID: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("templates/dashboard.html")
	if err != nil {
		h.logger.Printf("ERROR: ParseFiles: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	data := struct {
		Feeds []*store.Feed
	}{
		Feeds: feeds,
	}
	err = t.Execute(w, data)
	if err != nil {
		h.logger.Printf("ERROR: HandleDashboard %v", err)
		return
	}
}
