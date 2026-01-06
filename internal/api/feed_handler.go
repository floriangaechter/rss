// Package api
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"strconv"

	"github.com/floriangaechter/rss/internal/store"
	"github.com/go-chi/chi/v5"
)

type FeedHandler struct {
	feedStore store.FeedStore
}

func NewFeedHanlder(feedStore store.FeedStore) *FeedHandler {
	return &FeedHandler{
		feedStore: feedStore,
	}
}

func (fh *FeedHandler) HandleGetFeedByID(w http.ResponseWriter, r *http.Request) {
	paramsFeedID := chi.URLParam(r, "id")
	if paramsFeedID == "" {
		http.NotFound(w, r)
		return
	}

	feedID, err := strconv.ParseInt(paramsFeedID, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	feed, err := fh.feedStore.GetFeedByID(feedID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to fetch the feed", http.StatusInternalServerError)
		return
	}
	if feed == nil {
		fmt.Println("feed not found")
		http.Error(w, "feed not found", http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(feed)
}

func (fh *FeedHandler) HandleCreateFeed(w http.ResponseWriter, r *http.Request) {
	var feed store.Feed
	err := json.NewDecoder(r.Body).Decode(&feed)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to create feed", http.StatusInternalServerError)
		return
	}

	createdFeed, err := fh.feedStore.CreateFeed(&feed)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to create feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createdFeed)
}

func (fh *FeedHandler) HandleUpdateFeedByID(w http.ResponseWriter, r *http.Request) {
	paramsFeedID := chi.URLParam(r, "id")
	if paramsFeedID == "" {
		http.NotFound(w, r)
		return
	}

	feedID, err := strconv.ParseInt(paramsFeedID, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	feed, err := fh.feedStore.GetFeedByID(feedID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to fetch the feed", http.StatusInternalServerError)
		return
	}
	if feed == nil {
		fmt.Println("feed not found")
		http.Error(w, "feed not found", http.StatusNotFound)
	}

	var updateFeedRequest struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Link        *string `json:"link"`
	}
	err = json.NewDecoder(r.Body).Decode(&updateFeedRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		fmt.Println("update error", err)
		http.Error(w, "failed to update feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(feed)
}
