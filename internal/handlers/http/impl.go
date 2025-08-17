package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/serviceErorrs"
)

type userClient interface {
	RegisterAndLoginUser(ctx context.Context, reg domain.Login) (domain.SessionToken, error)
	LoginUser(ctx context.Context, reg domain.Login) (domain.SessionToken, error)
}

type loyaltyClient interface {
	AddOrder(ctx context.Context, userID domain.ID, order domain.Order) error
	GetOrders(ctx context.Context, userID domain.ID) (domain.Orders, error)
}

type Implementation struct {
	a userClient
	l loyaltyClient
}

func New(a userClient, l loyaltyClient) *Implementation {
	return &Implementation{
		a: a,
		l: l,
	}
}

type login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register POST /api/user/register — регистрация пользователя;
func (i *Implementation) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		appErr := serviceErorrs.AppErrorFromError(err).LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
		return
	}

	l := login{}
	if err = json.Unmarshal(body, &l); err != nil {
		appErr := serviceErorrs.NewBadRequest().LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
		return
	}

	authInfo, err := i.a.RegisterAndLoginUser(r.Context(), domain.Login{
		Login:    l.Login,
		Password: l.Password,
	})
	if err != nil {
		appErr := serviceErorrs.AppErrorFromError(err).LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    AuthCookiesName,
		Value:   authInfo.Token,
		Expires: time.Now().Add(24 * 180 * time.Hour),
		Secure:  true,
	})

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("success register and authorize")); err != nil {
		appErr := serviceErorrs.AppErrorFromError(err).LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
	}

	return
}

// Login POST /api/user/login — аутентификация пользователя;
func (i *Implementation) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {

		return
	}

	l := login{}
	if err = json.Unmarshal(body, &l); err != nil {
		appErr := serviceErorrs.NewBadRequest().LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
		return
	}

	authInfo, err := i.a.LoginUser(r.Context(), domain.Login{
		Login:    l.Login,
		Password: l.Password,
	})
	if err != nil {
		appErr := serviceErorrs.AppErrorFromError(err).LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    AuthCookiesName,
		Value:   authInfo.Token,
		Expires: time.Now().Add(24 * 180 * time.Hour),
		Secure:  true,
	})

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("success authorize")); err != nil {
		appErr := serviceErorrs.AppErrorFromError(err).LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
	}

	return
}

// AddOrder POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
func (i *Implementation) AddOrder(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(ctx, w, err)
		return
	}

	if isValidID := LuhnAlgorithm(string(body)); !isValidID {
		writeError(ctx, w,
			serviceErorrs.NewBadRequest().Wrap(nil, "invalid order id"))
		return
	}

	userID, isOK := getUserID(ctx)
	if !isOK {
		writeError(ctx, w,
			serviceErorrs.NewAppError(nil).
				Wrap(errors.New("invalid user id type assertion"), ""))
		return
	}

	orderID, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		writeError(ctx, w,
			serviceErorrs.NewAppError(nil).Wrap(err, ""))
		return
	}

	order := domain.Order{
		ID: domain.ID{
			ID: orderID,
		},
	}

	if err = i.l.AddOrder(ctx, userID, order); err != nil {
		writeError(ctx, w, err)
		return
	}

	return
}

type order struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int64  `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}

// GetOrders GET /api/user/orders — получение списка загруженных пользователем номеров заказов,
// статусов их обработки и информации о начислениях;
// TODO пагинация
func (i *Implementation) GetOrders(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()

	userID, isOK := getUserID(ctx)
	if !isOK {
		writeError(ctx, w,
			serviceErorrs.NewAppError(nil).
				Wrap(errors.New("invalid user id type assertion"), ""))
		return
	}

	resOrders, err := i.l.GetOrders(ctx, userID)
	if err != nil {
		writeError(ctx, w, err)
		return
	}

	if resOrders == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	transportOrders := make([]order, 0, len(resOrders))
	for _, ro := range resOrders {
		transportOrders = append(transportOrders, order{
			Number:     strconv.FormatUint(ro.ID.ID, 10),
			Status:     string(ro.State),
			Accrual:    ro.AccrualAmount.Amount.IntPart(),
			UploadedAt: ro.CreatedAt.Format(time.RFC3339),
		})
	}

	jsonBytes, err := json.Marshal(transportOrders)
	if err != nil {
		writeError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBytes)
	if err != nil {
		writeError(ctx, w, err)
		return
	}

	return
}

// GetBalance GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
func (i *Implementation) GetBalance(w http.ResponseWriter, r *http.Request) {
	return
}

// WithdrawalPoints POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
func (i *Implementation) WithdrawalPoints(w http.ResponseWriter, r *http.Request) {
	return
}

// GetWithdrawals GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.
// TODO пагинация
func (i *Implementation) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	return
}

func writeError(ctx context.Context, w http.ResponseWriter, err error) {
	appErr := serviceErorrs.AppErrorFromError(err).LogServerError(ctx)
	http.Error(w, appErr.String(), appErr.Code)
}

func LuhnAlgorithm(number string) bool {
	sum := 0
	alternate := false
	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false // Некорректный символ в номере
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}
	return sum%10 == 0
}
