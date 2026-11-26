package repositories

import (
	"testing"
	"time"

	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

var (
	ce = &pgconn.PgError{Code: pgerrcode.ConnectionException}
	cf = &pgconn.PgError{Code: pgerrcode.ConnectionFailure}
	uv = &pgconn.PgError{Code: pgerrcode.UniqueViolation}
	mv = models.ErrMissingValue
)

func Test_isRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "connection exception",
			err:  ce,
			want: true,
		},
		{
			name: "wrapped connection failure",
			err:  cf,
			want: true,
		},
		{
			name: "unique violation",
			err:  uv,
			want: false,
		},
		{
			name: "missing value error",
			err:  mv,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryable(tt.err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDBStorage_withRetry(t *testing.T) {

	tests := []struct {
		name      string
		errs      []error
		wantCalls int
		wantErr   error
	}{
		{
			name:      "retries and returns last error",
			errs:      []error{cf, ce, cf, ce},
			wantCalls: 4,
			wantErr:   ce,
		},
		{
			name:      "retries and returns non retryable error",
			errs:      []error{ce, uv},
			wantCalls: 2,
			wantErr:   uv,
		},
		{
			name:      "retries and returns nil",
			errs:      []error{ce, cf, nil},
			wantCalls: 3,
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDBStorage(nil, &RetryConfig{RetryCount: 3, RetryAfter: func(attempt int) time.Duration {
				return 10 * time.Millisecond
			}})

			calls := 0
			err := s.withRetry(t.Context(), func() error {
				calls++
				return tt.errs[calls-1]
			})

			assert.Equal(t, tt.wantCalls, calls)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
