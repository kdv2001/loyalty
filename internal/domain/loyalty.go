package domain

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

type Operation string

const (
	// Accrual начисление баллов
	Accrual Operation = "ACCRUAL"
	// Withdraw списание баллов
	Withdraw Operation = "WITHDRAW"
)

type Order struct {
	ID uint64
}
