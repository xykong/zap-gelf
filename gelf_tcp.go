package gelf

import (
	"compress/gzip"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
)

type (
	// implement io.Writer
	tcpWriter struct {
		conn             net.Conn
		compressionType  int
		compressionLevel int
	}
)

// NewCore zap core constructor.
func NewTcpCore(options ...Option) (_ zapcore.Core, err error) {
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

	var w = &tcpWriter{
		compressionType:  conf.compressionType,
		compressionLevel: conf.compressionLevel,
	}

	if w.conn, err = net.Dial("tcp", conf.address); err != nil {
		return nil, err
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
func (w *tcpWriter) Write(buf []byte) (n int, err error) {

	if n, err = w.conn.Write(buf); err != nil {
		return n, err
	}

	if n, err = w.conn.Write([]byte{0}); err != nil {
		return n, err
	}

	return n, nil
}
