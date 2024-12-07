package slog

import (
	"github.com/sirupsen/logrus"
	acerrors "github.com/zhughes3/go-accelerate/pkg/errors"
	"io"
	"time"
)

type LoggerBuilder struct {
	writer     io.Writer
	format     string
	tsFormat   string
	level      string
	labels     fieldLabels
	extractors []contextExtractor
}

func NewLoggerBuilder() *LoggerBuilder {
	return &LoggerBuilder{
		format:   "json",
		tsFormat: time.RFC3339Nano,
		level:    "info",
		labels:   defaultLabels,
	}
}

func (b *LoggerBuilder) WithWriter(writer io.Writer) *LoggerBuilder {
	b.writer = writer
	return b
}

func (b *LoggerBuilder) WithJSONFormatting() *LoggerBuilder {
	b.format = "json"
	return b
}

func (b *LoggerBuilder) WithTextFormatting() *LoggerBuilder {
	b.format = "text"
	return b
}

func (b *LoggerBuilder) WithTimestampFormat(tsFormat string) *LoggerBuilder {
	b.tsFormat = tsFormat
	return b
}

func (b *LoggerBuilder) WithLevel(level string) *LoggerBuilder {
	b.level = level
	return b
}

func (b *LoggerBuilder) WithLowercaseLabels() *LoggerBuilder {
	b.labels = lowercaseLabels
	return b
}

func (b *LoggerBuilder) WithContextExtractors(extractors []contextExtractor) *LoggerBuilder {
	b.extractors = append(b.extractors, extractors...)
	return b
}

func (b *LoggerBuilder) Build() (Logger, error) {
	ll := logrus.New()

	if b.writer != nil {
		ll.SetOutput(b.writer)
	}

	ll.SetFormatter(b.determineFormat())

	level, err := logrus.ParseLevel(b.level)
	if err != nil {
		return nil, acerrors.Wrapf(err, "problem parsing level '%s'", b.level)
	}
	ll.SetLevel(level)

	return newLogger(logrus.NewEntry(ll), b.labels, b.extractors), nil
}

func (b *LoggerBuilder) determineFormat() logrus.Formatter {
	if b.format == "json" {
		return &logrus.JSONFormatter{TimestampFormat: b.tsFormat}
	}

	return &logrus.TextFormatter{TimestampFormat: b.tsFormat}
}
