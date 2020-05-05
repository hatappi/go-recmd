// Package zap manages logger
package zap

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapKey = struct{}{}

// WithContext set zap logger to context
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, zapKey, logger)
}

// FromContext get zap logger from context
func FromContext(ctx context.Context) *zap.Logger {
	logger := ctx.Value(zapKey)
	if logger == nil {
		return nil
	}
	return logger.(*zap.Logger)
}

// NewZap initialize zap logger
func NewZap(level zap.AtomicLevel) (*zap.Logger, error) {
	logConfig := zap.Config{
		OutputPaths: []string{"stdout"},
		Level:       level,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			LevelKey:    "level",
			MessageKey:  "message",
			EncodeLevel: zapcore.CapitalColorLevelEncoder,
		},
	}

	logger, err := logConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
