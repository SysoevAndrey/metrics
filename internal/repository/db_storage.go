package repositories

import (
	"context"
	"database/sql"
	"errors"

	models "github.com/SysoevAndrey/metrics/internal/model"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{db: db}
}

func (s *DBStorage) SetGauge(ctx context.Context, name string, value float64) error {
	_, err := s.db.ExecContext(
		ctx,
		`
		INSERT INTO metrics (id, mtype, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (id, mtype) DO UPDATE
		SET value = $3
		`,
		name, models.Gauge, value,
	)

	return err
}

func (s *DBStorage) AddCounter(ctx context.Context, name string, value int64) error {
	_, err := s.db.ExecContext(
		ctx,
		`
		INSERT INTO metrics (id, mtype, delta)
		VALUES ($1, $2, $3)
		ON CONFLICT (id, mtype) DO UPDATE
		SET delta = metrics.delta + EXCLUDED.delta
		`,
		name, models.Counter, value,
	)

	return err
}

func (s *DBStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	row := s.db.QueryRowContext(
		ctx,
		`
		SELECT value 
		FROM metrics
		WHERE id = $1 AND mtype = $2
		`,
		name, models.Gauge,
	)

	var gauge sql.NullFloat64

	err := row.Scan(&gauge)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, models.ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	if gauge.Valid {
		return gauge.Float64, nil
	}
	return 0, models.ErrNotFound
}

func (s *DBStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	row := s.db.QueryRowContext(
		ctx,
		`
		SELECT delta 
		FROM metrics
		WHERE id = $1 AND mtype = $2
		`,
		name, models.Counter,
	)

	var delta sql.NullInt64

	err := row.Scan(&delta)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, models.ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	if delta.Valid {
		return delta.Int64, nil
	}
	return 0, models.ErrNotFound
}

func (s *DBStorage) GetAll(ctx context.Context) (models.Snapshot, error) {
	emptySnapshot := models.Snapshot{}

	rows, err := s.db.QueryContext(
		ctx,
		`
		SELECT id, mtype, value, delta
		FROM metrics
		`,
	)
	if err != nil {
		return emptySnapshot, err
	}
	defer rows.Close()

	gauges := map[string]float64{}
	counters := map[string]int64{}

	for rows.Next() {
		var (
			id    string
			mType string
			value sql.NullFloat64
			delta sql.NullInt64
		)

		err = rows.Scan(&id, &mType, &value, &delta)
		if err != nil {
			return emptySnapshot, err
		}

		switch mType {
		case models.Gauge:
			if !value.Valid {
				return emptySnapshot, models.ErrMissingValue
			}
			gauges[id] = value.Float64
		case models.Counter:
			if !delta.Valid {
				return emptySnapshot, models.ErrMissingValue
			}
			counters[id] = delta.Int64
		default:
			return emptySnapshot, models.ErrUnknownType
		}
	}

	if err = rows.Err(); err != nil {
		return emptySnapshot, err
	}

	return models.Snapshot{
		Gauges:   gauges,
		Counters: counters,
	}, nil
}
