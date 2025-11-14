package exporter

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Server wraps the HTTP server with graceful shutdown support.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

// NewServer constructs a Server.
func NewServer(addr string, handler http.Handler, logger *slog.Logger) *Server {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return &Server{httpServer: srv, logger: logger}
}

// Run starts the HTTP server and blocks until the context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("starting HTTP server", slog.String("addr", s.httpServer.Addr))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		return err
	}
}
