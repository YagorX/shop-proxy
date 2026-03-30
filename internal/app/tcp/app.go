package tcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync/atomic"
	"time"

	"github.com/YagorX/shop-proxy/internal/observability"
)

type Server struct {
	listenAddr   string
	upstreamAddr string
	logger       *slog.Logger
	listener     net.Listener
	metrics      *observability.Metrics
	delayNanos   atomic.Int64
}

func New(listenAddr, upstreamAddr string, logger *slog.Logger, metrics *observability.Metrics) (*Server, error) {
	if listenAddr == "" {
		return nil, errors.New("tcp listen address is required")
	}
	if upstreamAddr == "" {
		return nil, errors.New("tcp upstream address is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	if metrics == nil {
		return nil, errors.New("metrics is required")
	}

	srv := &Server{
		listenAddr:   listenAddr,
		upstreamAddr: upstreamAddr,
		logger:       logger,
		metrics:      metrics,
	}

	srv.SetDelay(0)
	return srv, nil
}

func (s *Server) Run(ctx context.Context) error {
	if s == nil {
		return errors.New("tcp server is nil")
	}

	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("listen tcp %s: %w", s.listenAddr, err)
	}
	s.listener = ln

	s.logger.Info("tcp proxy server started",
		slog.String("listen_addr", s.listenAddr),
		slog.String("upstream_addr", s.upstreamAddr),
	)

	go func() {
		<-ctx.Done()
		_ = s.Shutdown(context.Background())
	}()

	for {
		conn, err := ln.Accept()
		s.metrics.ConnectionsTotal.Inc()

		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				s.logger.Info("tcp proxy server stopped")
				return nil
			}
			s.logger.Error("tcp accept failed",
				slog.String("error", err.Error()),
			)
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil {
		return nil
	}
	if s.listener == nil {
		return nil
	}

	s.logger.Info("shutting down tcp proxy server")

	if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("close tcp listener: %w", err)
	}

	return nil
}

func (s *Server) Addr() string {
	if s == nil {
		return ""
	}
	return s.listenAddr
}

func (s *Server) handleConn(clientConn net.Conn) {
	s.metrics.ConnectionsActive.Inc()
	defer s.metrics.ConnectionsActive.Dec()

	if clientConn == nil {
		return
	}
	defer clientConn.Close()

	upstreamConn, err := net.Dial("tcp", s.upstreamAddr)
	if err != nil {
		s.logger.Error("tcp upstream dial failed",
			slog.String("upstream_addr", s.upstreamAddr),
			slog.String("error", err.Error()),
		)
		return
	}
	defer upstreamConn.Close()

	s.logger.Debug("tcp proxy connection opened",
		slog.String("client_remote_addr", clientConn.RemoteAddr().String()),
		slog.String("upstream_addr", s.upstreamAddr),
	)

	errCh := make(chan error, 2)

	go func() {
		errCh <- s.proxyCopy(upstreamConn, clientConn, "upstream")
	}()

	go func() {
		errCh <- s.proxyCopy(clientConn, upstreamConn, "downstream")
	}()

	err = <-errCh
	if err != nil && !errors.Is(err, io.EOF) {
		s.logger.Warn("tcp proxy copy finished with error",
			slog.String("error", err.Error()),
		)
	}

	s.logger.Debug("tcp proxy connection closed",
		slog.String("client_remote_addr", clientConn.RemoteAddr().String()),
	)
}

func (s *Server) proxyCopy(dst net.Conn, src net.Conn, direction string) error {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {

			delay := s.Delay()
			if delay > 0 {
				time.Sleep(delay)
			}
			written, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				if s.metrics != nil {
					s.metrics.CopyErrorsTotal.WithLabelValues(direction).Inc()
				}
				return writeErr
			}

			if written > 0 && s.metrics != nil {
				switch direction {
				case "upstream":
					s.metrics.BytesUpstreamTotal.Add(float64(written))
				case "downstream":
					s.metrics.BytesDownstreamTotal.Add(float64(written))
				}
			}

			if written != n {
				if s.metrics != nil {
					s.metrics.CopyErrorsTotal.WithLabelValues(direction).Inc()
				}
				return io.ErrShortWrite
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			if s.metrics != nil {
				s.metrics.CopyErrorsTotal.WithLabelValues(direction).Inc()
			}
			return err
		}
	}
}

func (s *Server) SetDelay(delay time.Duration) {
	if s == nil {
		return
	}

	s.delayNanos.Store(int64(delay))

	if s.metrics != nil {
		s.metrics.ConfiguredDelayMillis.Set(float64(delay.Milliseconds()))
	}
}

func (s *Server) Delay() time.Duration {
	if s == nil {
		return 0
	}

	return time.Duration(s.delayNanos.Load())
}
