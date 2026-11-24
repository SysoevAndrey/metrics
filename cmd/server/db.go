package main

import (
	"database/sql"
	"errors"
	"io"
	"strings"

	handlers "github.com/SysoevAndrey/metrics/internal/handler"
	repositories "github.com/SysoevAndrey/metrics/internal/repository"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type store struct {
	metrics handlers.MetricStorage
	pinger  handlers.Pinger
	closer  io.Closer
}

func newStore() (*store, error) {
	if !useDB() {
		return &store{
			metrics: repositories.NewMemStorage(),
		}, nil
	}

	db, err := sql.Open("pgx", config.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	if err = runMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	return &store{
		metrics: repositories.NewDBStorage(db),
		pinger:  db,
		closer:  db,
	}, nil
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	return err
}

func useDB() bool {
	return len(strings.TrimSpace(config.DatabaseDSN)) > 0
}
