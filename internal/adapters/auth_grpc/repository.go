package authgrpc

import (
	"context"
	"fmt"

	authclient "github.com/YagorX/shop-proxy/internal/client/grpc/auth"
)

type Repository struct {
	client *authclient.Client
}

func NewRepository(client *authclient.Client) (*Repository, error) {
	if client == nil {
		return nil, fmt.Errorf("auth grpc client is nil")
	}

	return &Repository{client: client}, nil
}

func (r *Repository) ValidateToken(ctx context.Context, token string, appID int64) (string, error) {
	return r.client.ValidateToken(ctx, token, appID)
}

func (r *Repository) IsAdmin(ctx context.Context, userUUID string) (bool, error) {
	return r.client.IsAdmin(ctx, userUUID)
}
