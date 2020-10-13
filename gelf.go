package gelf

import (
	"bytes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// CompressionNone don't use compression.
	CompressionNone = 0

	// CompressionGzip use gzip compression.
	CompressionGzip = 1

	// CompressionZlib use zlib compression.
	CompressionZlib = 2
)

type (
	// Option interface.
	Option interface {
		apply(conf *optionConf) error
	}

	// coreConf core.
	optionConf struct {
		address          string
		host             string
		version          string
		enabler          zap.AtomicLevel
		encoder          zapcore.EncoderConfig
		chunkSize        int
		writeSyncers     []zapcore.WriteSyncer
		compressionType  int
		compressionLevel int
	}

	// optionFunc wraps a func so it satisfies the Option interface.
	optionFunc func(conf *optionConf) error

	// implement io.WriteCloser.
	writeCloser struct {
		bytes.Buffer
	}

	// implement zapcore.Core.
	wrappedCore struct {
		core zapcore.Core
	}
)