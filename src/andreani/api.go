package andreani

import (
	"fmt"
	"time"

	"net/http"

	"encoding/json"
)

type Api struct {
	clientCode string 
	token      string
	baseUrl    string
	client     *http.Client
}

func NewApi(clientCode string, token string, baseUrl string) *Api {
	return &Api{
		clientCode: clientCode,
		token: token,
		baseUrl: baseUrl,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type Fee struct {
	SeguroDistribucion string `json:"seguroDistribucion"`
	Distribucion       string `json:"distribucion"`
	Total              string `json:"total"`
}

type Rate struct {
	PesoAforado  string `json:"pesoAforado"`
	TarifaSinIva Fee    `json:"tarifaSinIva"`
	TarifaConIva Fee    `json:"tarifaConIva"`
}

func (api *Api) CalculateShippingRate(contract, zip, volume string) (*Rate, error) {
	url := fmt.Sprintf(
		"%s/v1/tarifas?cpDestino=%s&contrato=%s&cliente=%s&bultos[0][volumen]=%s",
		api.baseUrl,
		zip,
		contract,
		api.clientCode,
		volume,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.client.Do(req)
  if err != nil {
		return nil, err
  }
	defer resp.Body.Close()

	rate := &Rate{}
	if err := json.NewDecoder(resp.Body).Decode(rate); err != nil {
		return nil, err
	}
	
	return rate, nil
}
