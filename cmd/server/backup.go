package main

import (
	"context"
	"strings"
	"time"

	handlers "github.com/SysoevAndrey/metrics/internal/handler"
	"github.com/SysoevAndrey/metrics/internal/logger"
	repositories "github.com/SysoevAndrey/metrics/internal/repository"
	services "github.com/SysoevAndrey/metrics/internal/service"
	"go.uber.org/zap"
)

type hook func(context.Context) error

func setupBackup(ctx context.Context, storage handlers.MetricStorage) (hook, error) {
	if useDB() {
		logger.Log.Info("skipping backup setup", zap.String("reason", "DB configured"))
		return nil, nil
	}

	if !useFileStorage() {
		logger.Log.Info("skipping backup setup", zap.String("reason", "no file"))
		return nil, nil
	}

	fs, err := repositories.NewFileStorage(config.FileStoragePath)
	if err != nil {
		return nil, err
	}

	bc := services.BackupConfig{
		Restore:       config.Restore,
		StoreInterval: time.Duration(config.StoreInterval) * time.Second,
	}
	b := services.NewBackup(storage, fs, bc)
	if err := b.SeedRuntimeStorageIfNeeded(ctx); err != nil {
		logger.Log.Error("failed to seed runtime storage", zap.Error(err))
	}
	go b.SetupPersistentStorageUpdateIfNeeded(ctx)

	return b.CreateSnapshotIfNeeded, nil
}

func useFileStorage() bool {
	return len(strings.TrimSpace(config.FileStoragePath)) > 0
}
