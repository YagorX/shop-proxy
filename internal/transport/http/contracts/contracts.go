package contracts

import (
	"context"
	"time"
)

type FaultController interface {
	SetDelay(delay time.Duration)
	Delay() time.Duration
}

type AuthService interface {
	ValidateToken(ctx context.Context, token string, appID int64) (string, error)
	IsAdmin(ctx context.Context, userUUID string) (bool, error)
}
