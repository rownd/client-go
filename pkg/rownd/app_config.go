package rownd

import (
	"context"
	"net/http"
)

type AppConfig struct {
	ID string `json:"app.config.id"`
}

type appConfigClient struct {
	*Client
}

func (c *appConfigClient) LoadAppConfig(ctx context.Context) {
	config, err := c.FetchAppConfig(ctx)
	if err != nil {
		panic(err)
	}

	c.appID = config.ID
}

func (c *appConfigClient) FetchAppConfig(ctx context.Context) (*AppConfig, error) {
	endpoint, err := c.rowndURL("hub", "app-config")
	if err != nil {
		return nil, err
	}

	var response *AppConfig
	if err := c.request(ctx, http.MethodGet, endpoint.String(), nil, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	return response, nil
}
