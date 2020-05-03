// Package logger manages logger
package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapKey = struct{}{}

// WithZapContext set zap logger to context
func WithZapContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, zapKey, logger)
}

// ZapFromContext get zap logger from context
func ZapFromContext(ctx context.Context) *zap.Logger {
	logger := ctx.Value(zapKey)
	if logger == nil {
		return nil
	}
	return logger.(*zap.Logger)
}

// NewZap initialize zap logger
func NewZap(level zapcore.Level) (*zap.Logger, error) {
	logConfig := zap.Config{
		Level:    zap.NewAtomicLevelAt(level),
		Encoding: "json",
	}

	logger, err := logConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
