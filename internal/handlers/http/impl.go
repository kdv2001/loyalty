package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Implementation struct{}

func New(mux chi.Mux) *Implementation {
	i := &Implementation{}

	mux.Route("/api/user", func(r chi.Router) {
		r.Post("/register", i.Register)
	})

	return i
}

// Register POST /api/user/register — регистрация пользователя;
func (i *Implementation) Register(w http.ResponseWriter, r *http.Request) {
	return
}

// Login POST /api/user/login — аутентификация пользователя;
func (i *Implementation) Login(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// AddOrder POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
func (i *Implementation) AddOrder(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// GetOrders GET /api/user/orders — получение списка загруженных пользователем номеров заказов,
// статусов их обработки и информации о начислениях;
// TODO пагинация
func (i *Implementation) GetOrders(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// GetBalance GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
func (i *Implementation) GetBalance(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// WithdrawalPoints POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
func (i *Implementation) WithdrawalPoints(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// GetWithdrawals GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.
// TODO пагинация
func (i *Implementation) GetWithdrawals(w http.ResponseWriter, r *http.Request) error {
	return nil
}
