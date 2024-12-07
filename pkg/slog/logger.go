package slog

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
	"time"
)

var (
	baseLogger = newLogger(logrus.NewEntry(logrus.New()), defaultLabels, nil)
)

func Base() Logger {
	return baseLogger
}

const messageCalled = "Execution complete"

type logger struct {
	entry      *logrus.Entry
	labels     fieldLabels
	extractors []contextExtractor
}

func (l logger) DebugContext(ctx context.Context, args ...any) {
	l.withContext(ctx, l.sourced()).Debug(args...)
}

func (l logger) DebugContextf(ctx context.Context, format string, args ...any) {
	l.withContext(ctx, l.sourced()).Debugf(format, args...)

}

func (l logger) InfoContext(ctx context.Context, args ...any) {
	l.withContext(ctx, l.sourced()).Info(args...)
}

func (l logger) InfoContextf(ctx context.Context, format string, args ...any) {
	l.withContext(ctx, l.sourced()).Infof(format, args...)
}

func (l logger) WarnContext(ctx context.Context, args ...any) {
	l.withContext(ctx, l.sourced()).Warn(args...)
}

func (l logger) WarnContextf(ctx context.Context, format string, args ...any) {
	l.withContext(ctx, l.sourced()).Warnf(format, args...)
}

func (l logger) ErrorContext(ctx context.Context, args ...any) {
	l.withContext(ctx, l.sourced()).Error(args...)
}

func (l logger) ErrorContextf(ctx context.Context, format string, args ...any) {
	l.withContext(ctx, l.sourced()).Errorf(format, args...)
}

func (l logger) FatalContext(ctx context.Context, args ...any) {
	l.withContext(ctx, l.sourced()).Fatal(args...)
}

func (l logger) FatalContextf(ctx context.Context, format string, args ...any) {
	l.withContext(ctx, l.sourced()).Fatalf(format, args...)
}

func (l logger) Called(ctx context.Context, begin time.Time, err error) {

	l.withContext(ctx, l.sourced().WithFields(l.doGetCalled(begin, err))).Info(messageCalled)
}

func (l logger) doGetCalled(begin time.Time, err error) map[string]any {
	return map[string]any{
		l.labels.duration: time.Since(begin).Milliseconds(),
		l.labels.error:    err,
	}
}

func (l logger) With(key string, value any) Logger {
	return l.WithField(key, value)
}

func (l logger) WithField(key string, value any) Logger {
	return l.withEntry(l.entry.WithField(key, value))
}

func (l logger) WithFields(fields map[string]any) Logger {
	return l.withEntry(l.entry.WithFields(fields))
}

func (l logger) WithDur(dur time.Duration) Logger {
	return l.With(l.labels.error, dur)
}

func (l logger) WithError(err error) Logger {
	return l.With(l.labels.error, err)
}

func (l logger) WithContext(ctx context.Context) Logger {
	if len(l.extractors) == 0 {
		return l
	}

	return l.WithFields(l.extractFields(ctx))
}

func (l logger) WithContextExtractor(extractor contextExtractor) Logger {
	return newLogger(l.entry, l.labels, append(l.extractors, extractor))
}

type Logger interface {
	DebugContext(context.Context, ...any)
	DebugContextf(context.Context, string, ...any)
	InfoContext(context.Context, ...any)
	InfoContextf(context.Context, string, ...any)
	WarnContext(context.Context, ...any)
	WarnContextf(context.Context, string, ...any)
	ErrorContext(context.Context, ...any)
	ErrorContextf(context.Context, string, ...any)
	FatalContext(context.Context, ...any)
	FatalContextf(context.Context, string, ...any)

	Called(ctx context.Context, begin time.Time, err error)

	With(key string, value any) Logger
	WithField(key string, value any) Logger
	WithFields(fields map[string]any) Logger
	WithDur(dur time.Duration) Logger
	WithError(err error) Logger
	WithContext(ctx context.Context) Logger
	WithContextExtractor(extractor contextExtractor) Logger
}

func newLogger(entry *logrus.Entry, labels fieldLabels, extractors []contextExtractor) logger {
	return logger{
		entry:      entry,
		labels:     labels,
		extractors: extractors,
	}
}

func (l logger) withEntry(entry *logrus.Entry) Logger {
	return newLogger(entry, l.labels, l.extractors)
}

func (l logger) withContext(ctx context.Context, entry *logrus.Entry) *logrus.Entry {
	if len(l.extractors) == 0 {
		return entry
	}

	return entry.WithFields(l.extractFields(ctx))
}

func (l logger) extractFields(ctx context.Context) map[string]any {
	m := map[string]any{}
	for _, extractor := range l.extractors {
		m := extractor(ctx)
		for k, v := range m {
			m[k] = v
		}
	}

	return m
}

func (l logger) sourced() *logrus.Entry {
	// pc (program counter), where we are in the program
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return l.entry.WithField(l.labels.source, "N/A")
	}

	return l.entry.WithFields(l.doGetSource(pc))
}

func (l logger) doGetSource(pc uintptr) map[string]any {
	frames := runtime.CallersFrames([]uintptr{pc})
	frame, _ := frames.Next()

	return map[string]any{
		l.labels.source:   determineSource(frame),
		l.labels.function: determineFunction(frame),
	}
}

func determineSource(frame runtime.Frame) string {
	fileIndex := strings.LastIndex(frame.File, "/")
	filename := frame.File[fileIndex+1:]
	return fmt.Sprintf("%s:%d", filename, frame.Line)
}

func determineFunction(frame runtime.Frame) string {
	functionIndex := strings.LastIndex(frame.Function, "/")
	return frame.Function[functionIndex+1:]
}
