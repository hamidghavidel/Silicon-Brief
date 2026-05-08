package store

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/hamidghavidel/silicon-brief/internal/fetcher"
	"google.golang.org/api/option"
)

// Store handles persistence of published posts.
type Store struct {
	client *firestore.Client
}

// PostRecord represents a stored post in Firestore.
type PostRecord struct {
	URL         string    `firestore:"url"`
	Title       string    `firestore:"title"`
	Source      string    `firestore:"source"`
	Score       float64   `firestore:"score"`
	PublishedAt time.Time `firestore:"publishedAt"`
	CreatedAt   time.Time `firestore:"createdAt"`
}

// New creates a new Firestore store.
func New(ctx context.Context, projectID, serviceAccountJSON string) (*Store, error) {
	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsJSON([]byte(serviceAccountJSON)))
	if err != nil {
		return nil, fmt.Errorf("firestore new client: %w", err)
	}

	return &Store{client: client}, nil
}

// IsPublished checks if a story has already been published.
func (s *Store) IsPublished(ctx context.Context, story fetcher.Story) (bool, error) {
	docID := hashURL(story.URL)
	doc, err := s.client.Collection("published_posts").Doc(docID).Get(ctx)
	if err != nil {
		// Document not found means not published
		return false, nil
	}
	return doc.Exists(), nil
}

// MarkPublished records a story as published.
func (s *Store) MarkPublished(ctx context.Context, story fetcher.Story, finalScore float64) error {
	docID := hashURL(story.URL)
	record := PostRecord{
		URL:         story.URL,
		Title:       story.Title,
		Source:      story.Source,
		Score:       finalScore,
		PublishedAt: story.PublishedAt,
		CreatedAt:   time.Now(),
	}

	_, err := s.client.Collection("published_posts").Doc(docID).Set(ctx, record)
	if err != nil {
		return fmt.Errorf("firestore set: %w", err)
	}
	return nil
}

// Close closes the Firestore client.
func (s *Store) Close() error {
	return s.client.Close()
}

func hashURL(url string) string {
	h := sha256.New()
	h.Write([]byte(url))
	return fmt.Sprintf("%x", h.Sum(nil))
}
