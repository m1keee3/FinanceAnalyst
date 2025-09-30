package moex

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/m1keee3/FinanceAnalyst/common/models"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/pkg/utils"
	"net/http"
	"net/url"
	"time"
)

const year = 365 * 24 * time.Hour

type httpClient interface {
	Get(string) (*http.Response, error)
}

type Fetcher struct {
	client httpClient
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{},
	}
}

func NewFetcherWithClient(client httpClient) *Fetcher {
	return &Fetcher{
		client: client,
	}
}

func (f *Fetcher) Fetch(ticker string, from, to time.Time) (
	[]models.Candle, error) {

	if !utils.IsLetterOnly(ticker) || !utils.IsAllUpper(ticker) {
		return nil, errors.New("invalid ticker")
	}

	if from.After(to) {
		return nil, errors.New("from date must be before to")
	}

	if to.Sub(from) < year {
		return f.getCandles(ticker, from, to, 24)
	}

	candles := make([]models.Candle, 0, 365*(to.Year()-from.Year()))

	start := from
	end := to
	for !start.After(to) {

		if end.Sub(from) > year {
			end = start.AddDate(1, 0, 0)
		} else {
			end = to
		}

		periodCandles, err := f.getCandles(ticker, start, end, 24)
		if err != nil {
			return nil, err
		}
		candles = append(candles, periodCandles...)

		start = start.AddDate(1, 0, 0)
	}

	return candles, nil
}

// The difference between from and to variables must be less than a year
//
// Supported interval values
//
//	interval = 1 → 1 minute
//	interval = 10 → 10 minute
//	interval = 60 → 1 hour
//	interval = 24 → 1 day
func (f *Fetcher) getCandles(ticker string, from, to time.Time, interval int) (
	[]models.Candle, error) {

	baseURL := fmt.Sprintf(
		"https://iss.moex.com/iss/engines/stock/markets/shares/boards/TQBR/securities/%s/candles.json",
		url.PathEscape(ticker),
	)

	params := url.Values{}
	params.Set("from", from.Format("2006-01-02"))
	params.Set("till", to.Format("2006-01-02"))
	params.Set("interval", fmt.Sprintf("%d", interval))
	params.Set("limit", "1000")

	reqURL := baseURL + "?" + params.Encode()

	resp, err := f.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var result struct {
		Candles struct {
			Data [][]interface{} `json:"data"`
		} `json:"candles"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	var candles []models.Candle
	for _, row := range result.Candles.Data {
		if len(row) < 7 {
			continue
		}

		timestamp, err := time.Parse("2006-01-02 15:04:05", row[6].(string))
		if err != nil {
			return nil, fmt.Errorf("time parse error: %w", err)
		}

		c := models.Candle{
			Date:  timestamp,
			Open:  row[0].(float64),
			Close: row[1].(float64),
			High:  row[2].(float64),
			Low:   row[3].(float64),
		}
		candles = append(candles, c)
	}

	return candles, nil
}
