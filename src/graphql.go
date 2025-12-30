package main

import (
	"bytes"
	"errors"

	"net/http"
	
	"encoding/json"
)


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
