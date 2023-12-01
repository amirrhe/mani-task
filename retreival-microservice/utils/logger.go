package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	cfg.OutputPaths = []string{"stdout"}

	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer l.Sync()
	logger = l
}

func GetLogger() *zap.Logger {
	return logger
}
