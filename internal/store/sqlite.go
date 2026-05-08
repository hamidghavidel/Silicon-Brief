package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"

	"github.com/hamidghavidel/silicon-brief/internal/fetcher"
	_ "github.com/mattn/go-sqlite3"
)

// Store handles persistence of published posts using SQLite.
type Store struct {
	db *sql.DB
}

// New creates a new SQLite store.
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS published_posts (
			url_hash TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			title TEXT NOT NULL,
			source TEXT NOT NULL,
			score REAL,
			published_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("sqlite create table: %w", err)
	}

	return &Store{db: db}, nil
}

// IsPublished checks if a story has already been published.
func (s *Store) IsPublished(ctx context.Context, story fetcher.Story) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM published_posts WHERE url_hash = ?)",
		hashURL(story.URL),
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("sqlite query: %w", err)
	}
	return exists, nil
}

// MarkPublished records a story as published.
func (s *Store) MarkPublished(ctx context.Context, story fetcher.Story, finalScore float64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO published_posts
		(url_hash, url, title, source, score, published_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		hashURL(story.URL), story.URL, story.Title, story.Source, finalScore, story.PublishedAt,
	)
	if err != nil {
		return fmt.Errorf("sqlite insert: %w", err)
	}
	return nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

func hashURL(url string) string {
	h := sha256.New()
	h.Write([]byte(url))
	return fmt.Sprintf("%x", h.Sum(nil))
}
