package storage

import (
	"fmt"
	"onlineSubscription/internal/config"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	DB *sqlx.DB
}

func NewPostgresStorage(cfg *config.Config) (*PostgresStorage, error) {
	host := cfg.Database.Host
	port := cfg.Database.Port

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host,
		port,
		cfg.Database.User,
		cfg.Database.Pass,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	return &PostgresStorage{DB: db}, nil
}

func (s *PostgresStorage) Close() error {
	return s.DB.Close()
}
