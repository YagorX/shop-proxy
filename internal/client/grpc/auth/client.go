package authclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	authv1 "github.com/YagorX/shop-contracts/gen/go/proto/auth/v1"
	"github.com/YagorX/shop-proxy/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	log     *slog.Logger
	addr    string
	timeout time.Duration

	conn   *grpc.ClientConn
	client authv1.AuthServiceClient
}

func NewClient(log *slog.Logger, addr string, timeout time.Duration, tlsCfg config.TLSConfig) (*Client, error) {
	if log == nil {
		log = slog.Default()
	}
	if addr == "" {
		return nil, errors.New("auth grpc addr is required")
	}
	if timeout <= 0 {
		return nil, errors.New("auth grpc timeout must be > 0")
	}

	transportCreds, err := buildTransportCredentials(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("build auth transport credentials: %w", err)
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(
		dialCtx,
		addr,
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("dial auth grpc: %w", err)
	}

	return &Client{
		log:     log,
		addr:    addr,
		timeout: timeout,
		conn:    conn,
		client:  authv1.NewAuthServiceClient(conn),
	}, nil
}

func buildTransportCredentials(tlsCfg config.TLSConfig) (credentials.TransportCredentials, error) {
	if !tlsCfg.Enabled {
		return insecure.NewCredentials(), nil
	}

	caPEM, err := os.ReadFile(tlsCfg.CAFile)
	if err != nil {
		return nil, fmt.Errorf("read ca file: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, errors.New("append ca cert to pool")
	}

	clientCert, err := tls.LoadX509KeyPair(tlsCfg.ClientCertFile, tlsCfg.ClientKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load client certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		RootCAs:      pool,
		ServerName:   tlsCfg.ServerName,
		Certificates: []tls.Certificate{clientCert},
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(tlsConfig), nil
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) ValidateToken(ctx context.Context, token string, appID int64) (string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.ValidateToken(reqCtx, &authv1.ValidateTokenRequest{
		Token: token,
		AppId: appID,
	})
	if err != nil {
		return "", err
	}

	return resp.GetUserUuid(), nil
}

func (c *Client) IsAdmin(ctx context.Context, userUUID string) (bool, error) {
	reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.IsAdmin(reqCtx, &authv1.IsAdminRequest{
		UserUuid: userUUID,
	})
	if err != nil {
		return false, err
	}

	return resp.GetIsAdmin(), nil
}
