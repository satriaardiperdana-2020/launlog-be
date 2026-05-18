package postgresql

import (
	"context"
	"fmt"
	"github.com/satriaardiperdana-2020/launlog-be/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnection(cfg *config.Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name)
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
