package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type CurrencyRateFetcher interface {
	GetCourseByDate(time.Time) ([]byte, error)
}

type cbClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) CurrencyRateFetcher {
	return &cbClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *cbClient) GetCourseByDate(date time.Time) ([]byte, error) {

	dateStr := date.Format("02/01/2006")

	fullUrl := fmt.Sprintf("%s?date_req=%s", c.baseURL, dateStr)

	resp, err := c.httpClient.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
