package zerologgorm_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/rs/zerolog"
	zerologgorm "github.com/skynet2/zerolog-gorm"
	"gorm.io/gorm"
	"testing"
	"time"
)

func newTestLogger(buf *bytes.Buffer, opts ...zerologgorm.OptionFn) (zerologgorm.Logger, context.Context) {
	logger := zerologgorm.NewLogger(opts...)
	ctx := zerolog.New(buf).With().Logger().WithContext(context.Background())
	return logger, ctx
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !bytes.Contains([]byte(s), []byte(substr)) {
		t.Errorf("Expected substring %q in %q", substr, s)
	}
}

func TestLogger_InfoWarnError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, ctx := newTestLogger(buf)

	logger.Info(ctx, "info %s", "test")
	logger.Warn(ctx, "warn %s", "test")
	logger.Error(ctx, "error %s", "test")

	logs := buf.String()
	assertContains(t, logs, "info test")
	assertContains(t, logs, "warn test")
	assertContains(t, logs, "error test")
}

func TestLogger_Trace_SlowQuery(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, ctx := newTestLogger(buf, zerologgorm.WithSlowThreshold(1*time.Nanosecond))

	start := time.Now().Add(-2 * time.Millisecond)
	logger.Trace(ctx, start, func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	logs := buf.String()
	assertContains(t, logs, "warn")
	assertContains(t, logs, "SELECT 1")
}

func TestLogger_Trace_ErrorAndNotFound(t *testing.T) {
	buf := &bytes.Buffer{}
	commonOpts := []zerologgorm.OptionFn{
		zerologgorm.WithSkipFrames(1),
		zerologgorm.WithSlowThreshold(1 * time.Nanosecond),
		zerologgorm.WithDefaultLogLevel(zerolog.DebugLevel),
		zerologgorm.WithSqlFieldName("test"),
	}

	logger, ctx := newTestLogger(buf, commonOpts...)

	start := time.Now()
	logger.Trace(ctx, start, func() (string, int64) {
		return "SELECT 1", 1
	}, errors.New("some error"))

	assertContains(t, buf.String(), "error")

	// Test ignore not found
	buf.Reset()
	logger, _ = newTestLogger(buf, zerologgorm.WithIgnoreNotFoundError())
	logger.Trace(ctx, start, func() (string, int64) {
		return "SELECT 1", 1
	}, gorm.ErrRecordNotFound)

	if buf.Len() != 0 {
		t.Errorf("Expected no log for not found error, got: %s", buf.String())
	}
}

func TestLogger_ParamsFilter(t *testing.T) {
	logger := zerologgorm.NewLogger()
	_, params := logger.ParamsFilter(context.Background(), "SELECT ?", 1)
	if params != nil {
		t.Errorf("Expected params to be nil when logParams is false")
	}

	logger = zerologgorm.NewLogger(zerologgorm.WithLogParams())
	_, params = logger.ParamsFilter(context.Background(), "SELECT ?", 1)
	if len(params) != 1 || params[0] != 1 {
		t.Errorf("Expected params to be returned when logParams is true")
	}
}
