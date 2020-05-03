package logger

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestWithZapContext(t *testing.T) {
	ctx := context.Background()

	if log := ctx.Value(zapKey); log != nil {
		t.Fatalf("zapKey isn't empty. %+v", log)
	}

	ctx = WithZapContext(ctx, zap.NewNop())

	logger := ctx.Value(zapKey)
	if logger == nil {
		t.Fatalf("zapKey doesn't include anything")
	}

	if _, ok := logger.(*zap.Logger); !ok {
		t.Fatalf("zapKey doesn't zap logger")
	}
}

func TestZapFromContext(t *testing.T) {
	ctx := context.Background()

	if log := ZapFromContext(ctx); log != nil {
		t.Fatalf("zapKey isn't empty. %+v", log)
	}

	ctx = context.WithValue(ctx, zapKey, zap.NewNop())

	logger := ZapFromContext(ctx)
	if logger == nil {
		t.Fatalf("zapKey doesn't include anything")
	}
}
