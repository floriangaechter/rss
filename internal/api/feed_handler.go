// Package api
package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/floriangaechter/rss/internal/fetcher"
	"github.com/floriangaechter/rss/internal/store"
	"github.com/floriangaechter/rss/internal/utils"
)

type FeedHandler struct {
	feedStore     store.FeedStore
	feedItemStore store.FeedItemStore
	fetcher       *fetcher.Fetcher
	logger        *log.Logger
}

func NewFeedHanlder(feedStore store.FeedStore, feedItemStore store.FeedItemStore, fetcher *fetcher.Fetcher, logger *log.Logger) *FeedHandler {
	return &FeedHandler{
		feedStore:     feedStore,
		feedItemStore: feedItemStore,
		fetcher:       fetcher,
		logger:        logger,
	}
}

type CreateFeedInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
}

func (in *CreateFeedInput) ValidateFeed() error {
	if strings.TrimSpace(in.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(in.Link) == "" {
		return errors.New("link is required")
	}
	return nil
}

func (fh *FeedHandler) HandleGetFeedByID(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r)
	if user == nil {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "unauthorized"})
		return
	}

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

	// Check if feed belongs to the authenticated user
	if feed.UserID != user.ID {
		_ = utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "access denied"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"feed": feed})
}

func (fh *FeedHandler) HandleCreateFeed(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r)
	if user == nil {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "unauthorized"})
		return
	}

	var req CreateFeedInput

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fh.logger.Printf("ERROR: decoding HandleCreateFeed: %v", err)
		_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request"})
		return
	}

	if err := req.ValidateFeed(); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}

	feed := store.Feed{
		UserID:      user.ID,
		Title:       req.Title,
		Description: req.Description,
		Link:        req.Link,
	}

	createdFeed, err := fh.feedStore.CreateFeed(&feed)
	if err != nil {
		fh.logger.Printf("ERROR: HandleCreateFeed: %v", err)
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"feed": createdFeed})
}

func (fh *FeedHandler) HandleUpdateFeedByID(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r)
	if user == nil {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "unauthorized"})
		return
	}

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

	// Check if feed belongs to the authenticated user
	if feed.UserID != user.ID {
		_ = utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "access denied"})
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
	user := utils.GetUserFromContext(r)
	if user == nil {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "unauthorized"})
		return
	}

	feedID, err := utils.ReadIDParam(r)
	if err != nil {
		fh.logger.Printf("ERROR: ReadIDParam: %v", err)
		_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid feed id"})
		return
	}

	// First check if feed exists and belongs to user
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

	// Check if feed belongs to the authenticated user
	if feed.UserID != user.ID {
		_ = utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "access denied"})
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

	w.WriteHeader(http.StatusOK)
}

func (fh *FeedHandler) HandleFetchFeedItems(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r)
	if user == nil {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "unauthorized"})
		return
	}

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

	// Check if feed belongs to the authenticated user
	if feed.UserID != user.ID {
		_ = utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "access denied"})
		return
	}

	err = fh.fetcher.FetchFeedItems(feedID)
	if err != nil {
		fh.logger.Printf("ERROR: FetchFeedItems: %v", err)
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "failed to fetch feed items"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "feed items fetched successfully"})
}
