package contracts

import "time"

type FaultController interface {
	SetDelay(delay time.Duration)
	Delay() time.Duration
}
