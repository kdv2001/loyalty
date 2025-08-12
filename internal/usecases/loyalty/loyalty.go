package loyalty

import (
	"context"

	"github.com/kdv2001/loyalty/internal/domain"
)

type loyaltyClient interface {
	GetAccruals(ctx context.Context, orderID domain.ID) error
}

type loyaltyStore interface{}

type Implementation struct {
	loyaltyClient loyaltyClient
	store         loyaltyStore
}

func NewImplementation(loyaltyClient loyaltyClient, store loyaltyStore) *Implementation {
	return &Implementation{
		loyaltyClient: loyaltyClient,
		store:         store,
	}
}
