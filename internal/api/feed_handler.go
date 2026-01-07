// Package api
package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/floriangaechter/rss/internal/store"
	"github.com/floriangaechter/rss/internal/utils"
)

type FeedHandler struct {
	feedStore store.FeedStore
	logger    *log.Logger
}

func NewFeedHanlder(feedStore store.FeedStore, logger *log.Logger) *FeedHandler {
	return &FeedHandler{
		feedStore: feedStore,
		logger:    logger,
	}
}

func (fh *FeedHandler) HandleGetFeedByID(w http.ResponseWriter, r *http.Request) {
	feedID, err := utils.ReadIDParam(r)
	if err != nil {
		fh.logger.Printf("ERROR: ReadIDParam: %v", err)
		_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid feed id"})
		return
	}

	feed, err := fh.feedStore.GetFeedByID(feedID)
	if err != nil {
		fh.logger.Printf("ERROR: GetFeedByID: %v", err)
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}
	if feed == nil {
		_ = utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "feed not found"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"feed": feed})
}

func (fh *FeedHandler) HandleCreateFeed(w http.ResponseWriter, r *http.Request) {
	var feed store.Feed
	err := json.NewDecoder(r.Body).Decode(&feed)
	if err != nil {
		fh.logger.Printf("ERROR: decoding HandleCreateFeed: %v", err)
		_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request"})
		return
	}

	createdFeed, err := fh.feedStore.CreateFeed(&feed)
	if err != nil {
		fh.logger.Printf("ERROR: HandleCreateFeed: %v", err)
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"feed": createdFeed})
}

// HandleUpdateFeedByID updates a feed based on the ID
func (fh *FeedHandler) HandleUpdateFeedByID(w http.ResponseWriter, r *http.Request) {
	feedID, err := utils.ReadIDParam(r)
	if err != nil {
		fh.logger.Printf("ERROR: ReadIDParam: %v", err)
		_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid feed id"})
		return
	}

	feed, err := fh.feedStore.GetFeedByID(feedID)
	if err != nil {
		fh.logger.Printf("ERROR: GetFeedByID: %v", err)
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}
	if feed == nil {
		_ = utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "feed not found"})
		return
	}

	var updateFeedRequest struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Link        *string `json:"link"`
	}
	err = json.NewDecoder(r.Body).Decode(&updateFeedRequest)
	if err != nil {
		fh.logger.Printf("ERROR: decoding HandleUpdateFeedByID: %v", err)
		_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request"})
		return
	}

	if updateFeedRequest.Title != nil {
		feed.Title = *updateFeedRequest.Title
	}
	if updateFeedRequest.Description != nil {
		feed.Description = *updateFeedRequest.Description
	}
	if updateFeedRequest.Link != nil {
		feed.Link = *updateFeedRequest.Link
	}

	err = fh.feedStore.UpdateFeed(feed)
	if err != nil {
		fh.logger.Printf("ERROR: HandleUpdateFeedByID: %v", err)
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"feed": feed})
}

func (fh *FeedHandler) HandleDeleteFeedByID(w http.ResponseWriter, r *http.Request) {
	feedID, err := utils.ReadIDParam(r)
	if err != nil {
		fh.logger.Printf("ERROR: ReadIDParam: %v", err)
		_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid feed id"})
		return
	}

	err = fh.feedStore.DeleteFeedByID(feedID)
	if err == sql.ErrNoRows {
		http.Error(w, "feed not found", http.StatusNotFound)
		return
	}
	if err != nil {
		fh.logger.Printf("ERROR: DeleteFeedByID: %v", err)
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	// TODO: check if that's the correct header, or if we need "No Content"
	w.WriteHeader(http.StatusOK)
}
