package gelf

import (
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"testing"
)

func TestNewUdpCore(t *testing.T) {

	// Only pass t into top-level Convey calls
	Convey("Create new udp core for logging", t, func() {

		host, err := os.Hostname()

		So(err, ShouldBeNil)

		core, err := NewUdpCore(
			Address("127.0.0.1:12201"),
			Host(host),
		)

		So(core, ShouldNotBeNil)
		So(err, ShouldBeNil)

		var logger = zap.New(
			core,
			zap.AddCaller(),
			zap.AddStacktrace(zap.LevelEnablerFunc(func(l zapcore.Level) bool {
				return core.Enabled(l)
			})),
		)
		defer logger.Sync()

		So(logger, ShouldNotBeNil)

		logger.With(
			zap.String("test", "TestNewUdpCore"),
		).Error(
			"Test an error was accrued TestNewUdpCore",
			zap.String("token", "a sample token"),
			zap.String("id", "a sample id"),
		)

		logger.Sugar().With(
			zap.String("test", "TestNewUdpCore"),
		).Error("Test an error was accrued from TestNewUdpCore")
	})
}
