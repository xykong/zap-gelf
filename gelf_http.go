package gelf

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
)

type (
	// implement io.Writer
	httpWriter struct {
		url              string
		compressionType  int
		compressionLevel int
	}
)

// NewCore zap core constructor.
func NewHttpCore(options ...Option) (_ zapcore.Core, err error) {
	var conf = optionConf{
		address: "http://127.0.0.1:12201/gelf",
		host:    "localhost",
		encoder: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			NameKey:        "_logger",
			LevelKey:       "level",
			CallerKey:      "_caller",
			MessageKey:     "short_message",
			StacktraceKey:  "full_message",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeName:     zapcore.FullNameEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeLevel:    levelEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		},
		version:          "1.1",
		enabler:          zap.NewAtomicLevel(),
		chunkSize:        DefaultChunkSize,
		writeSyncers:     make([]zapcore.WriteSyncer, 0, 8),
		compressionType:  CompressionGzip,
		compressionLevel: gzip.BestCompression,
	}

	for _, option := range options {
		if err = option.apply(&conf); err != nil {
			return nil, err
		}
	}

	var w = &httpWriter{
		url:              conf.address,
		compressionType:  conf.compressionType,
		compressionLevel: conf.compressionLevel,
	}

	var core = zapcore.NewCore(
		zapcore.NewJSONEncoder(conf.encoder),
		zapcore.AddSync(w),
		conf.enabler,
	)

	return &wrappedCore{
		core: core.With([]zapcore.Field{
			zap.String("host", conf.host),
			zap.String("version", conf.version),
		}),
	}, nil
}

// Write implements io.Writer.
func (w *httpWriter) Write(buf []byte) (n int, err error) {
	var (
		cw   io.WriteCloser
		cBuf bytes.Buffer
	)

	switch w.compressionType {
	case CompressionNone:
		cw = &writeCloser{cBuf}
	case CompressionGzip:
		cw, err = gzip.NewWriterLevel(&cBuf, w.compressionLevel)
	case CompressionZlib:
		cw, err = zlib.NewWriterLevel(&cBuf, w.compressionLevel)
	}

	if err != nil {
		return 0, err
	}

	if n, err = cw.Write(buf); err != nil {
		return n, err
	}

	_ = cw.Close()

	// post some data
	res, err := http.Post(
		w.url,
		"application/json; charset=UTF-8",
		bytes.NewReader(cBuf.Bytes()),
	)

	// check for response error
	if err != nil {
		return n, fmt.Errorf("writed %d bytes failed: %v", n, err)
	}

	// close response body
	_ = res.Body.Close()

	return n, nil
}
