package repositories

import (
	"os"
	"testing"

	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileWriter(t *testing.T) {
	t.Run("writes to empty file", func(t *testing.T) {
		fn := "file_writer.json"
		fw, err := NewFileWriter(fn)
		require.NoError(t, err)

		metrics := []models.Metrics{
			{ID: "PollCount", MType: models.Counter, Delta: new(int64(42))},
			{ID: "Alloc", MType: models.Gauge, Value: new(17.03)},
		}

		err = fw.WriteMetrics(metrics)
		require.NoError(t, err)

		err = os.Remove(fn)
		assert.NoError(t, err)
	})

	t.Run("rewrites file content", func(t *testing.T) {
		fn := "file_writer.json"
		fw, err := NewFileWriter(fn)
		require.NoError(t, err)

		metrics := []models.Metrics{
			{ID: "PollCount", MType: models.Counter, Delta: new(int64(42))},
			{ID: "Alloc", MType: models.Gauge, Value: new(17.03)},
		}

		err = fw.WriteMetrics(metrics)
		require.NoError(t, err)

		data1, err := os.ReadFile(fn)
		require.NoError(t, err)

		metrics[0].Delta = new(int64(67))
		metrics[1].Value = new(18.11)

		err = fw.WriteMetrics(metrics)
		require.NoError(t, err)

		data2, err := os.ReadFile(fn)
		require.NoError(t, err)

		err = fw.Close()
		require.NoError(t, err)

		assert.NotEqual(t, data1, data2)
		assert.NotContains(t, data2, data1)

		err = os.Remove(fn)
		assert.NoError(t, err)
	})
}
