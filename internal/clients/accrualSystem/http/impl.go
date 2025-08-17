package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/serviceErorrs"
)

const retryNums = 3

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	client    httpClient
	serverURL url.URL
}

func NewClient(client httpClient, serverURL url.URL) *Client {
	return &Client{
		client:    client,
		serverURL: serverURL,
	}
}

type accrualsResponse struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual int64  `json:"accrual"`
}

func (c *Client) GetAccruals(ctx context.Context, orderID domain.ID) (domain.Order, error) {
	getAccrualsURL := c.serverURL.JoinPath("api/orders", strconv.FormatUint(orderID.ID, 10))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getAccrualsURL.String(), nil)
	if err != nil {
		return domain.Order{}, serviceErorrs.AppErrorFromError(err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.Order{}, serviceErorrs.AppErrorFromError(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNoContent:
			return domain.Order{}, serviceErorrs.NewNoContent().Wrap(domain.ErrNoContent, "accrual status")
		case http.StatusInternalServerError:
			return domain.Order{}, serviceErorrs.NewAppError(nil).Wrap(nil, "accrual status")
		case http.StatusTooManyRequests:
			return domain.Order{}, serviceErorrs.NewTooManyRequests().Wrap(nil, "accrual status")
		}
	}

	ar := new(accrualsResponse)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.Order{}, serviceErorrs.NewAppError(nil)
	}

	err = json.Unmarshal(body, ar)
	if err != nil {
		return domain.Order{}, serviceErorrs.NewAppError(nil)
	}

	return domain.Order{
		ID:    orderID,
		State: domain.StateFromString(ar.Status),
		AccrualAmount: domain.Money{
			Currency: "GopherMarketBonuses",
			Amount:   decimal.NewFromInt(ar.Accrual),
		},
	}, nil
}
