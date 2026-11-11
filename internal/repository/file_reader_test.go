package repositories

import (
	"os"
	"testing"

	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileReader(t *testing.T) {
	t.Run("reads from empty file", func(t *testing.T) {
		fn := "file_reader.json"
		fr, err := NewFileReader(fn)
		require.NoError(t, err)

		metrics, err := fr.ReadMetrics()
		require.NoError(t, err)
		assert.Empty(t, metrics)

		err = os.Remove(fn)
		assert.NoError(t, err)
	})

	t.Run("reads metrics from file", func(t *testing.T) {
		fn := "file_reader.json"
		fw, err := NewFileWriter(fn)
		require.NoError(t, err)

		metrics := []models.Metrics{
			{ID: "PollCount", MType: models.Counter, Delta: new(int64(42))},
			{ID: "Alloc", MType: models.Gauge, Value: new(17.03)},
		}

		err = fw.WriteMetrics(metrics)
		require.NoError(t, err)
		err = fw.Close()
		require.NoError(t, err)

		fr, err := NewFileReader(fn)
		require.NoError(t, err)

		result, err := fr.ReadMetrics()
		require.NoError(t, err)
		assert.EqualValues(t, metrics, result)
		err = fr.Close()
		assert.NoError(t, err)

		err = os.Remove(fn)
		assert.NoError(t, err)
	})
}
