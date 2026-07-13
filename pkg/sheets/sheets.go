package sheets

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Sheets struct {
	client     http.Client
	maxRetries int
	retryDelay time.Duration
}

func New(maxRetries int, retryDelay time.Duration) Sheets {
	if maxRetries < 1 {
		maxRetries = 3
	}
	if retryDelay < 0 {
		retryDelay = 2 * time.Second
	}

	return Sheets{
		maxRetries: maxRetries,
		retryDelay: retryDelay,
		client: http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (s Sheets) Values(id string) ([][]string, error) {
	url := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv", id)

	var lastErr error
	for attempt := range s.maxRetries {
		resp, err := s.client.Get(url)
		if err != nil {
			lastErr = err
			if attempt < s.maxRetries-1 {
				time.Sleep(s.retryDelay)
			}
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("bad status: %s", resp.Status)
		}
		return parseCSVRows(resp.Body)
	}
	return nil, fmt.Errorf("failed after %d attempts: %w", s.maxRetries, lastErr)
}

func parseCSVRows(r io.Reader) ([][]string, error) {
	reader := csv.NewReader(r)
	var rows [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, record)
	}
	return rows, nil
}
