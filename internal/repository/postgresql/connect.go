package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"launlog-be/repository/sqlc"
)

type Repository struct {
	Queries *sqlc.Queries
	DB      *pgxpool.Pool
}

func NewRepository(dbURL string) (*Repository, error) {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, err
	}
	queries := sqlc.New(pool)
	return &Repository{
		Queries: queries,
		DB:      pool,
	}, nil
}

func (r *Repository) Close() {
	r.DB.Close()
}
