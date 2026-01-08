// Package fetcher
package fetcher

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/floriangaechter/rss/internal/store"
	"github.com/floriangaechter/rss/internal/utils"
	"github.com/mmcdole/gofeed"
)

type Fetcher struct {
	feedStore     store.FeedStore
	feedItemStore store.FeedItemStore
	logger        *log.Logger
}

func NewFetcher(feedStore store.FeedStore, feedItemStore store.FeedItemStore, logger *log.Logger) *Fetcher {
	return &Fetcher{
		feedStore:     feedStore,
		feedItemStore: feedItemStore,
		logger:        logger,
	}
}

func (f *Fetcher) FetchFeedItems(feedID int64) error {
	feed, err := f.feedStore.GetFeedByID(feedID)
	if err != nil {
		return err
	}
	if feed == nil {
		return errors.New("feed not found")
	}

	fp := gofeed.NewParser()
	fp.UserAgent = "RSS/1.0"

	parsedFeed, err := fp.ParseURL(feed.Link)
	if err != nil {
		return err
	}

	var newItemsCount int
	for _, item := range parsedFeed.Items {
		feedItem := &store.FeedItem{
			FeedID:      feed.ID,
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
		}

		if item.PublishedParsed != nil {
			feedItem.PublishedAt = item.PublishedParsed.Format(time.RFC3339)
		} else if item.Published != "" {
			feedItem.PublishedAt = utils.ParseFeedDate(item.Published)
		} else {
			feedItem.PublishedAt = time.Now().Format(time.RFC3339)
		}

		_, err := f.feedItemStore.CreateFeedItem(feedItem)
		if err != nil {
			// Check if it's a unique constraint violation (duplicate)
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				// Duplicate item - skip it (this is expected)
				continue
			}
			// Some other error - log it but continue with other items
			f.logger.Printf("ERROR: Failed to create feed item %s: %v", item.Link, err)
			continue
		}

		newItemsCount++
	}

	f.logger.Printf("Fetched %d new items for feed %d", newItemsCount, feedID)
	return nil
}
