package loyalty

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/serviceErorrs"
)

type Implementation struct {
	c *pgxpool.Pool
}

var accrualTable = `create table if not exists orders (
                                      id           bigint GENERATED ALWAYS AS IDENTITY,
                                      user_id      bigint NOT NULL,
                                      order_id     bigint NOT NULL,
                                      state        text,
                                      amount       decimal,
                                      currency     text,
                                      primary key (user_id, order_id))
`

var operationTable = `create table if not exists operation (
                                         id            bigint GENERATED ALWAYS AS IDENTITY primary key,
                                         order_id      bigint not null,
                                         user_id       bigint not null,
                                         amount        decimal,
                                         currency      text,
                                         operation     text,
                                         created_at    timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'));
`

var tables = []string{
	accrualTable,
	operationTable,
}

// NewImplementation ...
func NewImplementation(ctx context.Context, c *pgxpool.Pool) (*Implementation, error) {
	for _, t := range tables {
		_, err := c.Exec(ctx, t)
		if err != nil {
			return nil, nil
		}
	}

	return &Implementation{
		c: c,
	}, nil
}

func (i *Implementation) AccrualPoints(ctx context.Context) error {

}

func (i *Implementation) WithdrawPoints(ctx context.Context, operation domain.Operation) error {
	tx, err := i.c.Begin(ctx)
	if err != nil {
		return serviceErorrs.AppErrorFromError(err)
	}
	defer func() {
		if err != nil {
			if err = tx.Rollback(ctx); err != nil {

			}
		}
	}()

	tx.QueryRow(ctx, ``)
}

func (i *Implementation) GetBalance(ctx context.Context, userID domain.ID) domain.Money {
	i.
}