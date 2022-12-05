package shutdown

import (
	"go.uber.org/zap"
	"io"
	"os"
	"os/signal"
)

func Graceful(logger *zap.Logger, signals []os.Signal, closeItems ...io.Closer) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, signals...)
	sig := <-sigc
	logger.Warn("Caught termination signal. Shutting Down.",
		zap.String("signal", sig.String()),
	)

	for _, closer := range closeItems {
		if err := closer.Close(); err != nil {
			logger.Error("failed to close",
				zap.Any("closer", closer),
				zap.Error(err))
		}
	}
}
