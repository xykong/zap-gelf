package gelf

import (
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"testing"
)

func TestSpec(t *testing.T) {

	// Only pass t into top-level Convey calls
	Convey("Given some integer with a starting value", t, func() {
		x := 1

		Convey("When the integer is incremented", func() {
			x++

			Convey("The value should be greater by one", func() {
				So(x, ShouldEqual, 2)
			})
		})
	})
}

func TestNewHttpCore(t *testing.T) {
	Convey("Create new http core for logging", t, func() {

		host, err := os.Hostname()

		So(err, ShouldBeNil)

		core, err := NewHttpCore(
			Address("http://localhost:12201/gelf"),
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
			zap.String("test", "TestNewHttpCore"),
		).Error(
			"Test an error was accrued from TestNewHttpCore",
			zap.String("token", "a sample token"),
			zap.String("id", "a sample id"),
		)

		logger.Sugar().With(
			zap.String("test", "TestNewHttpCore"),
		).Error("Test an error was accrued from TestNewHttpCore")
	})
}
