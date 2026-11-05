package logger

import "go.uber.org/zap"

var Log *zap.Logger = zap.NewNop()

func Initialize(levelRaw string) error {
	level, err := zap.ParseAtomicLevel(levelRaw)
	if err != nil {
		return err
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = level

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}
