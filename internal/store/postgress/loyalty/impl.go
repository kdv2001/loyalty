package loyalty

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/logger"
	"github.com/kdv2001/loyalty/internal/pkg/serviceErorrs"
)

type Implementation struct {
	c *pgxpool.Pool
}

var accrualTable = `create table if not exists orders (
                                      id           bigint GENERATED ALWAYS AS IDENTITY,
                                      user_id      bigint NOT NULL,
                                      order_id     bigint NOT NULL,                                    
                                      created_at   timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
                                      currency     text,
                                      amount       decimal,
                                      primary key  (user_id, order_id))
`

var jobAccrualTable = `create table if not exists jobs (
     id           bigint GENERATED ALWAYS AS IDENTITY,
     order_id     bigint NOT NULL primary key,
     state        text                  
)`

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
	jobAccrualTable,
}

// NewImplementation ...
func NewImplementation(ctx context.Context, c *pgxpool.Pool) (*Implementation, error) {
	for _, t := range tables {
		_, err := c.Exec(ctx, t)
		if err != nil {
			return nil, err
		}
	}

	return &Implementation{
		c: c,
	}, nil
}

type orderModel struct {
	ID            sql.NullInt64  `db:"id"`
	UserID        sql.NullInt64  `db:"user_id"`
	OrderID       sql.NullInt64  `db:"order_id"`
	Status        sql.NullString `db:"status"`
	CreatedAt     sql.NullTime   `db:"created_at"`
	AccrualAmount sql.NullInt64  `db:"amount"`
	Currency      sql.NullString `db:"currency"`
}

// AddOrder создает заявку на начисление баллов в заказе
func (i *Implementation) AddOrder(ctx context.Context, userID domain.ID, order domain.Order) error {
	tx, err := i.c.Begin(ctx)
	if err != nil {
		return serviceErorrs.NewAppError(err)
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// TODO сделать тут селект
	sqlOrder := orderModel{}
	err = tx.QueryRow(ctx, `select id, user_id, order_id from orders where order_id = $1`, order.ID.ID).Scan(&sqlOrder.ID, &sqlOrder.UserID, &sqlOrder.OrderID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return serviceErorrs.NewAppError(err)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		switch {
		case uint64(sqlOrder.UserID.Int64) != userID.ID:
			return serviceErorrs.NewConflict().Wrap(err, "order belongs other user")
		case uint64(sqlOrder.UserID.Int64) == userID.ID:
			return serviceErorrs.NewBadRequest().Wrap(domain.ErrActionCompletedEarly, "")
		}
	}

	_, err = tx.Exec(ctx, "INSERT INTO orders (user_id, order_id) VALUES ($1, $2)", userID.ID, order.ID.ID)
	if err != nil {
		return serviceErorrs.NewAppError(err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO jobs (order_id, state) VALUES ($1, $2)", order.ID.ID, domain.New)

	if err = tx.Commit(ctx); err != nil {
		return serviceErorrs.NewAppError(err)
	}

	return nil

}

func (i *Implementation) GetOrders(ctx context.Context, userID domain.ID) (domain.Orders, error) {
	iter, err := i.c.Query(ctx, `select user_id, orders.order_id, state, created_at, currency, amount
			from orders join jobs on orders.order_id = jobs.order_id where orders.user_id = $1`, userID.ID)
	if err != nil {
		return nil, serviceErorrs.NewAppError(err)
	}

	defer iter.Close()

	res := make(domain.Orders, 0)
	for iter.Next() {
		order := orderModel{}
		err = iter.Scan(&order.UserID, &order.OrderID, &order.Status,
			&order.CreatedAt, &order.Currency, &order.AccrualAmount)
		if err != nil {
			return nil, serviceErorrs.NewAppError(err)
		}

		state := domain.StateFromString(order.Status.String)
		if state == domain.Invalid {
			logger.Errorf(ctx, " data consistency is broken invalid state: %s, orderID: %d",
				order.Status.String, order.ID.Int64)
		}

		res = append(res, domain.Order{
			ID: domain.ID{
				ID: uint64(order.OrderID.Int64),
			},
			State:     state,
			CreatedAt: order.CreatedAt.Time,
			AccrualAmount: domain.Money{
				Currency: order.Currency.String,
				Amount:   decimal.NewFromInt(order.AccrualAmount.Int64),
			},
		})
	}

	return res, nil
}

func (i *Implementation) AccrualPoints(ctx context.Context) error {
	return nil
}

func (i *Implementation) GetBalance(ctx context.Context, userID domain.ID) domain.Money {
	return domain.Money{}
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
	return nil
}
