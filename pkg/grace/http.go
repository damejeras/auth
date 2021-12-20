package grace

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

func Serve(ctx context.Context, server *http.Server, listener net.Listener) error {
	errChan := make(chan error)

	go func() {
		log.Printf("listen %q", listener.Addr().String())
		if err := server.Serve(listener); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Printf("shutdown %q", listener.Addr().String())
		defer log.Printf("%q shutdown complete", listener.Addr().String())
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
