package andreani

import (
	"io"
	"fmt"
	"time"
	"bytes"
	"errors"
	"strings"

	"net/http"
	"net/url"

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

type Address struct {
	Calle         string `json:"calle"`
	Numero        string `json:"numero"`
	Provincia     string `json:"provincia"`
	Localidad     string `json:"localidad"`
	Region        string `json:"region"`
	Pais          string `json:"pais"`
	CondigoPostal string `json:"codigoPostal"`
}

type Coordenada struct {
	Latitud  string `json:"latitud"`
	Longitud string `json:"longitud"`
}

type DatosAdicionales struct {
	SeHaceAtencionAlCliente bool   `json:"seHaceAtencionAlCliente"`
	ConBuzonInteligente     bool   `json:"conBuzonInteligente"`
	tipo                    string `json:"tipo"`
}

type Office struct {
	ID                       int64   	        `json:"id"`
	Codigo                   string  	        `json:"codigo"`
	Numero                   string  	        `json:"numero"`
	Descripcion              string  	        `json:"descripcion"`
	Canal                    string  	        `json:"canal"`
	Direccion                Address 	        `json:"direccion"`
	Coordenadas              Coordenada       `json:"coordenada"`
	HorarioDeAtencion        string           `json:"horarioDeAtencion"`
	DatosAdicionales         DatosAdicionales `json:"datosAdicionales"`
	Telefonos                []string         `json:"telefonos"`
	CodigosPostalesAtendidos []string         `json:"codigosPostalesAtendidos"`
}

type Location struct {
	IdDeProvLocalidad string   `json:"idDeProvLocalidad"`
	Localidad         string   `json:"localidad"` 
	Provincia         string   `json:"provincia"`
	CondigosPostales  []string `json:"codigosPostales"`
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



type LocationQuery struct {
	Location string
	Province string
	Zips     []string
}

func (api *Api) GetLocations(query LocationQuery) ([]Location, error) {
	
	baseUrl, err := url.Parse(fmt.Sprintf("%s/v1/localidades", api.baseUrl))
	if err != nil {
		return nil, err
	}
	
	q := baseUrl.Query()

	if query.Location != "" {
		q.Set("localidad", query.Location)
	}

	if query.Province != "" {
		q.Set("provincia", query.Province)
	}
	
	if len(query.Zips) > 0 {
		q.Set("codigosPostales", strings.Join(query.Zips, ","))
	}
	
	baseUrl.RawQuery = q.Encode()

	fmt.Println(baseUrl.String())
	
	req, err := http.NewRequest("GET", baseUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.client.Do(req)
  if err != nil {
		return nil, err
  }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("fail to get location")
	}

	locations := []Location{}
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		return nil, err
	}
	
	return locations, nil
}

type SucursalQuery struct {
	Codigo string
	Sucursal string
	Region string
	Localidad string
	CodigoPostal string
	Canal string
	Numero string
}

func (api *Api) GetOffices(query SucursalQuery) ([]Office, error) {

	baseUrl, err := url.Parse(fmt.Sprintf("%s/v2/localidades", api.baseUrl))
	if err != nil {
		return nil, err
	}
	
	q := baseUrl.Query()

	if query.Codigo != "" {
		q.Set("codigo", query.Codigo)
	}

	if query.Sucursal != "" {
		q.Set("sucursal", query.Sucursal)
	}

	if query.Region != "" {
		q.Set("region", query.Region)
	}

	if query.Localidad != "" {
		q.Set("localidad", query.Localidad)
	}

	if query.CodigoPostal != "" {
		q.Set("codigoPostal", query.CodigoPostal)
	}
	
	if query.Canal != "" {
		q.Set("canal", query.Canal)
	}

	if query.Numero != "" {
		q.Set("numero", query.Numero)
	}
	
	baseUrl.RawQuery = q.Encode()
	
	req, err := http.NewRequest("GET", baseUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.client.Do(req)
  if err != nil {
		return nil, err
  }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("fail to get location")
	}

	offices := []Office{}
	if err := json.NewDecoder(resp.Body).Decode(&offices); err != nil {
		return nil, err
	}
	
	return offices, nil
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
