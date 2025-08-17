package loyalty

import (
	"context"
	"errors"
	"time"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/logger"
	"github.com/kdv2001/loyalty/internal/pkg/serviceErorrs"
)

type loyaltyClient interface {
	GetAccruals(ctx context.Context, orderID domain.ID) (domain.Order, error)
}

type loyaltyStore interface {
	AddOrder(ctx context.Context, userID domain.ID, order domain.Order) error
	GetOrders(ctx context.Context, userID domain.ID) (domain.Orders, error)
	GetOrderForAccruals(ctx context.Context) (domain.Orders, error)
	AccrualPoints(ctx context.Context, o domain.Order) error
	UpdateOrderStatus(ctx context.Context, order domain.Order) error
}

type Implementation struct {
	loyaltyClient loyaltyClient
	store         loyaltyStore
}

func NewImplementation(ctx context.Context, loyaltyClient loyaltyClient, store loyaltyStore) *Implementation {
	i := &Implementation{
		loyaltyClient: loyaltyClient,
		store:         store,
	}

	go i.ProcessAccrual(ctx)

	return i
}

// AddOrder ...
func (i *Implementation) AddOrder(ctx context.Context, userID domain.ID, order domain.Order) error {
	err := i.store.AddOrder(ctx, userID, order)
	if err != nil {
		return err
	}

	return nil
}

func (i *Implementation) GetOrders(ctx context.Context, userID domain.ID) (domain.Orders, error) {
	return i.store.GetOrders(ctx, userID)
}

func (i *Implementation) ProcessAccrual(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf(ctx, "recovered from panic: %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := i.processAccrual(ctx)
			if err != nil {
				_ = serviceErorrs.AppErrorFromError(err).LogServerError(ctx)
			}
		}
	}
}

func (i *Implementation) processAccrual(ctx context.Context) error {
	orders, err := i.store.GetOrderForAccruals(ctx)
	if err != nil {
		return err
	}

	for _, order := range orders {
		var res domain.Order
		tCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		res, err = i.loyaltyClient.GetAccruals(tCtx, order.ID)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrNoContent):
				continue
			}

			return err
		}

		if res.State != domain.Processed {
			order.State = res.State
			err = i.store.UpdateOrderStatus(tCtx, order)
			if err != nil {
				return err
			}
			continue
		}

		order.State = res.State
		order.AccrualAmount = res.AccrualAmount

		err = i.store.AccrualPoints(tCtx, order)
		if err != nil {
			return err
		}
	}

	return nil
}
