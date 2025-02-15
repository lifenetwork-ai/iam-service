package clients

import (
	"net/http"
)

type lifeAIClient struct {
	endpoint   string
	httpClient *http.Client
}

func NewLifeAIClient(endpoint string) *lifeAIClient {
	return &lifeAIClient{
		endpoint:   endpoint,
		httpClient: &http.Client{},
	}
}
