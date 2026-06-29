package fetcher

import (
	"context"
	"errors"
	"fmt"

	"github.com/DGreegman/gohunt/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists fetched jobs into the database.
type Store struct {
	queries *db.Queries
}

// NewStore wires a Store to connection pool
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store {
		queries: db.New(pool),
	}
}


// SvaeJobs inserts each job, skipping duplicates
// Returns how many were newly inserted

func (s *Store) SaveJobs(ctx context.Context, jobs [] Job) (int, error) {
	inserted := 0
	for _, j := range jobs {
		_, err := s.queries.CreateJob(ctx, toCreateParams(j))
		if errors.Is(err, pgx.ErrNoRows) {
			// 	NO CONFLICT DO NOTHING returned no row: it was a duplicate.
			continue
		}
		if err != nil {
			return  inserted, fmt.Errorf("Saving Job %q: %w", j.Title, err)
		}

		inserted ++
	}

	return inserted, nil
}

// toCreateParams maps a fetched Job to the generated insert params

func toCreateParams(j Job) db.CreateJobParams {
	return db.CreateJobParams{
		Title: 			j.Title,
		Company: 		j.Company,
		Source: 		j.Source,
		Url: 			j.URL,
		Description: 	j.Description,
		DescriptionHash: "",
		Location: 		j.Location,
		Remote: 		j.Remote,
		PostedAt: 		pgtype.Timestamptz{Time: j.PostedAt, Valid: !j.PostedAt.IsZero()},
		Hash: 			j.Hash,
		RawJson: 		j.RawJSON,
	}
}