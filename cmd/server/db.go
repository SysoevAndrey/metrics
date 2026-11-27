package main

import (
	"database/sql"
	"errors"
	"io"
	"strings"

	handlers "github.com/SysoevAndrey/metrics/internal/handler"
	"github.com/SysoevAndrey/metrics/internal/logger"
	repositories "github.com/SysoevAndrey/metrics/internal/repository"
	"github.com/SysoevAndrey/metrics/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
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
		metrics: repositories.NewDBStorage(db, nil),
		pinger:  db,
		closer:  db,
	}, nil
}

func runMigrations(db *sql.DB) error {
	srcDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", srcDriver, "postgres", dbDriver)
	if err != nil {
		return err
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		logger.Log.Info("skipping migrations", zap.String("reason", "no new migrations found"))
		return nil
	}
	return err
}

func useDB() bool {
	return len(strings.TrimSpace(config.DatabaseDSN)) > 0
}
