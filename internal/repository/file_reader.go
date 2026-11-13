package repositories

import (
	"encoding/json"
	"io"
	"os"

	models "github.com/SysoevAndrey/metrics/internal/model"
)

type FileReader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewFileReader(filename string) (*FileReader, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &FileReader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (fr *FileReader) ReadMetrics() ([]models.Metrics, error) {
	metrics := []models.Metrics{}
	err := fr.decoder.Decode(&metrics)
	if err == io.EOF {
		return []models.Metrics{}, nil
	}
	if err != nil {
		return []models.Metrics{}, err
	}
	return metrics, nil
}

func (fr *FileReader) Close() error {
	return fr.file.Close()
}
