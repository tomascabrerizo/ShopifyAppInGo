package main

import (
	"fmt"
	"bytes"
	"errors"

	"net/http"
	
	"encoding/json"
)

type DimensionCm struct {
	Width  float64
	Height float64
	Length float64
}

type Metafield struct {
	Value string `json:"value"`
}

type UserError struct {
	Field   []string `json:"field"`
	Message string `json:"message"`
}

type CarrierService struct {
	ID                       string `json:"id,omitempty"`
	Name                     string `json:"name"`
	CallbackURL              string `json:"callbackUrl"`
	SupportsServiceDiscovery bool   `json:"supportsServiceDiscovery"`
	Active                   bool   `json:"active"`
}

// TODO: All this functinos ask to database for access token, just receive access token as parameter 

func (app *Application) GetCarrierServices(shop string)  ([]CarrierService, error) {
	token, err := app.db.GetAccessToken(shop)
	if err != nil {
		return nil, err
	}

	type GraphQLPayload struct {
		Query     string           `json:"query"`
	}
	
	payload := GraphQLPayload{
		Query: "query CarrierServiceList { carrierServices(first: 10, query: \"active:true\") { edges { node { id name callbackUrl active supportsServiceDiscovery } } } }",
	}

	body, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}

	url := "https://"+shop+"/admin/api/2025-10/graphql.json" 
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type",  "application/json")
	req.Header.Set("X-Shopify-Access-Token", token.Access)
	
	resp, err := app.client.Do(req)
  if err != nil {
		return nil, err
  }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("shopify get carrier service failed")
	}

	var graphql struct {
		Data struct {
			CarrierServices struct {
				Edges []struct {
						CarrierService CarrierService `json:"node"`
				} `json:"edges"`
			} `json:"carrierServices"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&graphql); err != nil {
		return nil, err
	}
	
	services := []CarrierService{}
	for _, node := range graphql.Data.CarrierServices.Edges {
		services = append(services, node.CarrierService)	
	}

	return services, nil
}

type CarrierServiceCreate struct {
	CarrierService CarrierService `json:"carrierService"`	
	UserErrors     []UserError    `json:"userErrors"`
}

func (app *Application) CarrierServiceCreate(shop, name, callbackUrl string) (*CarrierServiceCreate, error) {
	token, err := app.db.GetAccessToken(shop)
	if err != nil {
		return nil, err
	}

	type GraphQLVariables struct {
		Input CarrierService `json:"input"`
	}

	type GraphQLPayload struct {
		Query     string           `json:"query"`
		Variables GraphQLVariables `json:"variables"`
	}
	
	payload := GraphQLPayload{
		Query: "mutation CarrierServiceCreate($input: DeliveryCarrierServiceCreateInput!) { carrierServiceCreate(input: $input) { carrierService { id name callbackUrl active supportsServiceDiscovery } userErrors { field message } } }",
		Variables:  GraphQLVariables{
			Input: CarrierService{
				Name: name,
				CallbackURL: callbackUrl,
				SupportsServiceDiscovery: true,
				Active: true,
			},
		},
	}

	body, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}
	
	url := "https://"+shop+"/admin/api/2025-10/graphql.json" 
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type",  "application/json")
	req.Header.Set("X-Shopify-Access-Token", token.Access)
	
	resp, err := app.client.Do(req)
  if err != nil {
		return nil, err
  }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("shopify create carrier service failed")
	}

	var graphql struct {
		Data struct {
			CarrierServiceCreate CarrierServiceCreate `json:"carrierServiceCreate"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&graphql); err != nil {
		return nil, err
	}
	
	return &graphql.Data.CarrierServiceCreate, nil

}

type CarrierServiceDelete struct {
	DeletedID  string      `json:"deletedId"`
	UserErrors []UserError `json:"userErrors"`
}

func (app *Application) CarrierServiceDelete(shop, id string) (*CarrierServiceDelete, error) {
	token, err := app.db.GetAccessToken(shop)
	if err != nil {
		return nil, err
	}

	type GraphQLVariables struct {
		ID string `json:"id"`
	}

	type GraphQLPayload struct {
		Query     string           `json:"query"`
		Variables GraphQLVariables `json:"variables"`
	}

	payload := GraphQLPayload{
		Query: "mutation CarrierServiceDelete($id: ID!) { carrierServiceDelete(id: $id) { deletedId userErrors { field message } } }",
		Variables: GraphQLVariables{
			ID: id,
		},
	}

	body, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}

	url := "https://"+shop+"/admin/api/2025-10/graphql.json" 
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type",  "application/json")
	req.Header.Set("X-Shopify-Access-Token", token.Access)
	
	resp, err := app.client.Do(req)
  if err != nil {
		return nil, err
  }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("shopify delete carrier service failed")
	}

	var graphql struct {
		Data struct {
			CarrierServiceDelete CarrierServiceDelete `json:"carrierServiceDelete"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&graphql); err != nil {
		return nil, err
	}
	
	return &graphql.Data.CarrierServiceDelete, nil
}

func parseDimension(m *Metafield) (float64, error) {
	if m == nil || m.Value == "" {
		return 0, fmt.Errorf("invalid metafield\n")
	}

	type DimensionValue struct {
		Value float64 `json:"value"`
		Unit  string  `json:"unit"`
	}

	var dim DimensionValue 
	err := json.Unmarshal([]byte(m.Value), &dim)
	if err != nil {
		return 0, err
	}

	// TODO: Convert to the correct unit
	if dim.Unit != "CENTIMETERS" {
		return 0, fmt.Errorf("invalid dim unit %s\n", dim.Unit)
	}

	return dim.Value, nil

} 

func(app *Application) GetProductDimensions(shop, id string) (*DimensionCm, error) {
	token, err := app.db.GetAccessToken(shop)
	if err != nil {
		return nil, err
	}

	type GraphQLVariables struct {
		OwnerID string `json:"ownerId"`
	}

	type GraphQLPayload struct {
		Query     string           `json:"query"`
		Variables GraphQLVariables `json:"variables"`
	}

	query := `
		query ProductMetafields($ownerId: ID!) { 
		    product(id: $ownerId) {
		        largo: metafield(namespace: "custom", key: "largo") {
		          value
		        }
		        ancho: metafield(namespace: "custom", key: "ancho") {
		          value
		        }
		        alto: metafield(namespace: "custom", key: "alto") {
		          value
		        }
		    }
		}
	`

	payload := GraphQLPayload{
		Query: query, 
		Variables: GraphQLVariables{
			OwnerID: id,
		},
	}

	body, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}

	url := "https://"+shop+"/admin/api/2025-10/graphql.json" 
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type",  "application/json")
	req.Header.Set("X-Shopify-Access-Token", token.Access)
	
	resp, err := app.client.Do(req)
  if err != nil {
		return nil, err
  }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("shopify get dimensions failed")
	}

	var graphql struct {
		Data struct {
			Product struct {
				Largo *Metafield `json:"largo"`
				Ancho *Metafield `json:"ancho"`
				Alto  *Metafield `json:"alto"`
			} `json:"product"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&graphql); err != nil {
		return nil, err
	}
	
	largo, err := parseDimension(graphql.Data.Product.Largo)
	if err != nil {
		return nil, err
	}
	
	ancho, err := parseDimension(graphql.Data.Product.Ancho)
	if err != nil {
		return nil, err
	}
	
	alto, err := parseDimension(graphql.Data.Product.Alto)
	if err != nil {
		return nil, err
	}
	
	dim := &DimensionCm{
		Width: alto,
		Height: ancho,
		Length: largo,
	}

	return dim, nil
}
