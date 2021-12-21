package grace

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

func Serve(ctx context.Context, server *http.Server, listener net.Listener) error {
	errChan := make(chan error)

	go func() {
		if err := server.Serve(listener); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errChan:
		return err
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		return errors.Wrap(err, "server shutdown with context")
	}

	return nil
}
