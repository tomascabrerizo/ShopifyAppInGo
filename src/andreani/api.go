package andreani

import (
	"time"

	"net/http"
)

type Api struct {
	clientCode string 
	token      string
	client     *http.Client
}

func NewApi(clientCode string, token string) *Api {
	return &Api{
		clientCode: clientCode,
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
