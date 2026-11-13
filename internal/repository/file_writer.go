package repositories

import (
	"encoding/json"
	"os"

	models "github.com/SysoevAndrey/metrics/internal/model"
)

type FileWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func NewFileWriter(filename string) (*FileWriter, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return &FileWriter{
		file:    file,
		encoder: encoder,
	}, nil
}

func (fw *FileWriter) WriteMetrics(metrics []models.Metrics) error {
	if err := fw.file.Truncate(0); err != nil {
		return err
	}
	if _, err := fw.file.Seek(0, 0); err != nil {
		return err
	}
	return fw.encoder.Encode(metrics)
}

func (fw *FileWriter) Close() error {
	return fw.file.Close()
}
