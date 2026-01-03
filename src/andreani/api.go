package andreani

import (
	"io"
	"fmt"
	"time"
	"bytes"
	"errors"

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

type Postal struct {
	CodigoPostal string `json:"codigoPostal"`
	Calle        string `json:"calle"`
	Numero       string `json:"numero"`
	Localidad    string `json:"localidad"`
}

type Telefono struct {
	Tipo   int    `json:"tipo"`
	Numero string `json:"numero"`
}

type Persona struct {
	NombreCompleto string     `json:"nombreCompleto"`
	Telefonos      []Telefono `json:"telefonos"`
}

type Bulto struct {
	Kilos     float64 `json:"kilos"`
	VolumenCm float64 `json:"volumenCm"`
}

func (api *Api) CreateShipping(
	contrato string,
	origen, destino Postal,
	remitente, destinatario Persona,
	bultos []Bulto) error {
	
	type Origen struct {
		Postal Postal `json:"postal"`
	}
	
	type Destino struct {
		Postal Postal `json:"postal"`
	}

	type Payload struct {
		Contrato 	   string  `json:"contrato"`
		Origen       Origen  `json:"origen"`
		Destino      Destino `json:"destino"`
		Remitente    Persona `json:"remitente"`
		Destinatario []Persona `json:"destinatario"`
		Bultos       []Bulto `json:"bultos"`
	}

	url := api.baseUrl + "/v2/ordenes-de-envio"
	
	payload := Payload {
		Contrato: contrato,
		Origen:  Origen{Postal: origen},
		Destino: Destino{Postal: destino},
		Remitente: remitente,
		Destinatario: []Persona{destinatario},
		Bultos: bultos,
	}

	body, err := json.Marshal(&payload)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type",  "application/json")
	req.Header.Set("x-authorization-token", api.token)
	
	resp, err := api.client.Do(req)
  if err != nil {
		return err
  }
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	fmt.Println(string(bytes))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("shopify create shipping failed")
	}

	return nil
}
