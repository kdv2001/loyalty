package mock

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/kdv2001/loyalty/internal/domain"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) GetAccruals(_ context.Context, orderID domain.ID) (domain.Order, error) {
	return domain.Order{
		ID:    orderID,
		State: domain.Processed,
		AccrualAmount: domain.Money{
			Currency: "GopherMarketBonuses",
			Amount:   decimal.NewFromInt(150),
		},
	}, nil
}
