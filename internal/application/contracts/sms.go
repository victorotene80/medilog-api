package contracts

import "context"

type SMSSender interface {
	Send(
		ctx context.Context,
		recipient string,
		message string,
	) error
}