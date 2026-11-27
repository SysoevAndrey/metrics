package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	setGaugeUpsert = `
		INSERT INTO metrics (id, mtype, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (id, mtype) DO UPDATE
		SET value = $3
	`
	addCounterUpsert = `
		INSERT INTO metrics (id, mtype, delta)
		VALUES ($1, $2, $3)
		ON CONFLICT (id, mtype) DO UPDATE
		SET delta = metrics.delta + EXCLUDED.delta
	`
)

type DBStorage struct {
	db          *sql.DB
	retryConfig *RetryConfig
}

type RetryConfig struct {
	RetryCount int
	RetryAfter RetryAfterFunc
}

type RetryAfterFunc func(attempt int) time.Duration

func NewDBStorage(db *sql.DB, retryConfig *RetryConfig) *DBStorage {
	config := newDefaultRetryConfig()

	if retryConfig != nil && retryConfig.RetryCount != 0 {
		config.RetryCount = retryConfig.RetryCount
	}
	if retryConfig != nil && retryConfig.RetryAfter != nil {
		config.RetryAfter = retryConfig.RetryAfter
	}

	return &DBStorage{
		db:          db,
		retryConfig: config,
	}
}

func newDefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		RetryCount: 3,
		RetryAfter: func(attempt int) time.Duration {
			return 2*time.Second*time.Duration(attempt) - time.Second
		},
	}
}

func (s *DBStorage) SetGauge(ctx context.Context, name string, value float64) error {
	return s.withRetry(ctx, func() error {
		return s.setGauge(ctx, name, value)
	})
}

func (s *DBStorage) AddCounter(ctx context.Context, name string, value int64) error {
	return s.withRetry(ctx, func() error {
		return s.addCounter(ctx, name, value)
	})
}

func (s *DBStorage) ApplyBatch(ctx context.Context, metrics []models.Metrics) error {
	return s.withRetry(ctx, func() error {
		return s.applyBatch(ctx, metrics)
	})
}

func (s *DBStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	var gauge float64
	err := s.withRetry(ctx, func() error {
		var err error
		gauge, err = s.getGauge(ctx, name)
		return err
	})
	return gauge, err
}

func (s *DBStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	var counter int64
	err := s.withRetry(ctx, func() error {
		var err error
		counter, err = s.getCounter(ctx, name)
		return err
	})
	return counter, err
}

func (s *DBStorage) GetAll(ctx context.Context) (models.Snapshot, error) {
	var sh models.Snapshot
	err := s.withRetry(ctx, func() error {
		var err error
		sh, err = s.getAll(ctx)
		return err
	})
	return sh, err
}

func (s *DBStorage) setGauge(ctx context.Context, name string, value float64) error {
	_, err := s.db.ExecContext(
		ctx,
		setGaugeUpsert,
		name, models.Gauge, value,
	)

	return err
}

func (s *DBStorage) addCounter(ctx context.Context, name string, value int64) error {
	_, err := s.db.ExecContext(
		ctx,
		addCounterUpsert,
		name, models.Counter, value,
	)

	return err
}

func (s *DBStorage) applyBatch(ctx context.Context, metrics []models.Metrics) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	setGaugeStmt, err := tx.PrepareContext(ctx, setGaugeUpsert)
	if err != nil {
		return err
	}
	defer setGaugeStmt.Close()

	addCounterStmt, err := tx.PrepareContext(ctx, addCounterUpsert)
	if err != nil {
		return err
	}
	defer addCounterStmt.Close()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				return models.ErrMissingValue
			}
			if _, err := setGaugeStmt.ExecContext(ctx, metric.ID, models.Gauge, *metric.Value); err != nil {
				return err
			}
		case models.Counter:
			if metric.Delta == nil {
				return models.ErrMissingValue
			}
			if _, err := addCounterStmt.ExecContext(ctx, metric.ID, models.Counter, *metric.Delta); err != nil {
				return err
			}
		default:
			return models.ErrUnknownType
		}
	}

	return tx.Commit()
}

func (s *DBStorage) getGauge(ctx context.Context, name string) (float64, error) {
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

func (s *DBStorage) getCounter(ctx context.Context, name string) (int64, error) {
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

func (s *DBStorage) getAll(ctx context.Context) (models.Snapshot, error) {
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

func (s *DBStorage) withRetry(ctx context.Context, fn func() error) error {
	if s.retryConfig.RetryCount == 0 || s.retryConfig.RetryAfter == nil {
		return fn()
	}

	var err error

	for attempt := 0; attempt <= s.retryConfig.RetryCount; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return errors.Join(err, ctx.Err())
			case <-time.After(s.retryConfig.RetryAfter(attempt)):
			}
		}

		err = fn()
		if err == nil {
			return nil
		}
		if !isRetryable(err) {
			return err
		}
	}

	return err
}

func isRetryable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)
}
