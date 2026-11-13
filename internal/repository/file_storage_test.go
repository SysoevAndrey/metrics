package repositories

import (
	"os"
	"testing"

	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage(t *testing.T) {
	fn := "file_storage.json"
	fs, err := NewFileStorage(fn)
	require.NoError(t, err)

	metrics := []models.Metrics{
		{ID: "PollCount", MType: models.Counter, Delta: new(int64(42))},
		{ID: "Alloc", MType: models.Gauge, Value: new(17.03)},
	}

	err = fs.WriteMetrics(metrics)
	require.NoError(t, err)

	result, err := fs.ReadMetrics()
	require.NoError(t, err)

	assert.EqualValues(t, metrics, result)

	err = os.Remove(fn)
	assert.NoError(t, err)
}
