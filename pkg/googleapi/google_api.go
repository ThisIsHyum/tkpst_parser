package googleapi

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func getClient(ctx context.Context, credentialsPath string) (*http.Client, error) {
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read credentials: %v", err)
	}
	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse credentials: %v", err)
	}
	return config.Client(ctx), nil
}

func NewSheetsService(ctx context.Context, credentialsPath string) (*sheets.Service, error) {
	client, err := getClient(ctx, credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to get client: %w", err)
	}
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Sheets client: %v", err)
	}
	return srv, nil
}
