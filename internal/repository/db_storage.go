package repositories

import (
	"context"
	"database/sql"

	models "github.com/SysoevAndrey/metrics/internal/model"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{db: db}
}

func (s *DBStorage) SetGauge(ctx context.Context, name string, value float64) error {
	panic("not implemented")
}

func (s *DBStorage) AddCounter(ctx context.Context, name string, value int64) error {
	panic("not implemented")
}

func (s *DBStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	panic("not implemented")
}

func (s *DBStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	panic("not implemented")
}

func (s *DBStorage) GetAll(ctx context.Context) (models.Snapshot, error) {
	panic("not implemented")
}
