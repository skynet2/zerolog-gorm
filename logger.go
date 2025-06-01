package zerologgorm

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

var _ logger.Interface = (*Logger)(nil)
var _ gorm.ParamsFilter = (*Logger)(nil)

type Logger struct {
	opt *options
}

type options struct {
	logParams         bool
	skipFrames        int
	defaultLogLevel   zerolog.Level
	fieldName         string
	ignoreNotFoundErr bool
	slowThreshold     time.Duration
}

func WithIgnoreNotFoundError() OptionFn {
	return func(o *options) {
		o.ignoreNotFoundErr = true
	}
}

func WithLogParams() OptionFn {
	return func(o *options) {
		o.logParams = true
	}
}

func WithSkipFrames(skip int) OptionFn {
	return func(o *options) {
		o.skipFrames = skip
	}
}

func WithSqlFieldName(name string) OptionFn {
	return func(o *options) {
		o.fieldName = name
	}
}

func WithDefaultLogLevel(level zerolog.Level) OptionFn {
	return func(o *options) {
		o.defaultLogLevel = level
	}
}

func WithSlowThreshold(threshold time.Duration) OptionFn {
	return func(o *options) {
		o.slowThreshold = threshold
	}
}

type OptionFn func(*options)

func NewLogger(
	opts ...OptionFn,
) Logger {
	opt := &options{
		defaultLogLevel: zerolog.DebugLevel,
		fieldName:       "sql",
		slowThreshold:   500 * time.Millisecond,
	}

	for _, fn := range opts {
		fn(opt)
	}

	return Logger{
		opt: opt,
	}
}

func (l Logger) LogMode(logger.LogLevel) logger.Interface {
	return l
}

func (l Logger) Error(ctx context.Context, msg string, opts ...interface{}) {
	l.apply(zerolog.Ctx(ctx).Error()).Msg(fmt.Sprintf(msg, opts...))
}

func (l Logger) Warn(ctx context.Context, msg string, opts ...interface{}) {
	l.apply(zerolog.Ctx(ctx).Warn()).Msg(fmt.Sprintf(msg, opts...))
}

func (l Logger) Info(ctx context.Context, msg string, opts ...interface{}) {
	l.apply(zerolog.Ctx(ctx).Info()).Msg(fmt.Sprintf(msg, opts...))
}

func (l Logger) apply(event *zerolog.Event) *zerolog.Event {
	if l.opt.skipFrames > 0 {
		event.CallerSkipFrame(l.opt.skipFrames)
	}

	return event
}

func (l Logger) Trace(ctx context.Context, begin time.Time, f func() (string, int64), err error) {
	logLevel := l.opt.defaultLogLevel

	if l.opt.slowThreshold > 0 && time.Since(begin) > l.opt.slowThreshold {
		logLevel = zerolog.WarnLevel
	}

	event := zerolog.Ctx(ctx).WithLevel(logLevel)

	if err != nil {
		if l.opt.ignoreNotFoundErr && errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}

		event = zerolog.Ctx(ctx).Error().Err(err)
	}

	if l.opt.skipFrames > 0 {
		event = event.CallerSkipFrame(l.opt.skipFrames)
	}

	event.Int64("elapsed_ms", time.Since(begin).Milliseconds())

	sql, rows := f()

	if sql != "" {
		event.Str(l.opt.fieldName, sql)
	}

	if rows > -1 {
		event.Int64("rows_affected", rows)
	}

	event.Send()
}

func (l Logger) ParamsFilter(_ context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if !l.opt.logParams {
		return sql, nil
	}

	return sql, params
}
