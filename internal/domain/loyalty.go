package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type AccrualState string

const (
	// Registered заказ зарегистрирован, но начисление не рассчитано
	Registered AccrualState = "PROCESSED"
	// Invalid заказ не принят к расчёту, и вознаграждение не будет начислено
	Invalid AccrualState = "INVALID"
	// Processing расчёт начисления в процессе
	Processing AccrualState = "PROCESSING"
	// Processed расчёт начисления окончен
	Processed AccrualState = "PROCESSED"
)

type OperationType string

const (
	// Accrual начисление баллов
	Accrual OperationType = "ACCRUAL"
	// Withdraw списание баллов
	Withdraw OperationType = "WITHDRAW"
)

type Order struct {
	ID uint64
}

type Money struct {
	Currency string
	Amount   decimal.Decimal
}

type Operation struct {
	ID       ID
	OrderID  ID
	Type     OperationType
	Amount   Money
	CratedAt time.Time
}
