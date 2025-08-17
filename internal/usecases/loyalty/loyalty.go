package loyalty

import (
	"context"
	"errors"

	"github.com/kdv2001/loyalty/internal/domain"
)

type loyaltyClient interface {
	GetAccruals(ctx context.Context, orderID domain.ID) error
}

type loyaltyStore interface {
	AddOrder(ctx context.Context, userID domain.ID, order domain.Order) error
	GetOrders(ctx context.Context, userID domain.ID) (domain.Orders, error)
}

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

// AddOrder ...
func (i *Implementation) AddOrder(ctx context.Context, userID domain.ID, order domain.Order) error {
	err := i.store.AddOrder(ctx, userID, order)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrActionCompletedEarly):
			// TODO это не ошибка, можно подвязаться на статус

			return nil

		}
		return err
	}

	return nil
}

func (i *Implementation) GetOrders(ctx context.Context, userID domain.ID) (domain.Orders, error) {
	return i.store.GetOrders(ctx, userID)
}
