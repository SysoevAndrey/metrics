package repositories

import (
	"sync"

	models "github.com/SysoevAndrey/metrics/internal/model"
)

type FileStorage struct {
	writer *FileWriter
	reader *FileReader

	mu sync.Mutex
}

func NewFileStorage(filename string) (*FileStorage, error) {
	writer, err := NewFileWriter(filename)
	if err != nil {
		return nil, err
	}

	reader, err := NewFileReader(filename)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		writer: writer,
		reader: reader,
	}, nil
}

func (fs *FileStorage) WriteMetrics(metrics []models.Metrics) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.writer.WriteMetrics(metrics)
}

func (fs *FileStorage) ReadMetrics() ([]models.Metrics, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.reader.ReadMetrics()
}
