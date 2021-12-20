package grace

import (
	"context"
	"os"
	"os/signal"
)

func NewAppContext() (context.Context, context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-sigChan
		cancel()
	}()

	return ctx, cancel
}
