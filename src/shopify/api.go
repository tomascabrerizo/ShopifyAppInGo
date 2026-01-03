package shopify 

import (
	"fmt"
	"time"
	"bytes"
	"errors"

	"net/http"

	"encoding/json"
)

type Api struct {
	ID     string
	Secret string
	client *http.Client
}

func NewApi(clientId, clientSecret string) *Api {
	return &Api{
		ID: clientId,
		Secret: clientSecret,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type DimensionCm struct {
	Width  float64
	Height float64
	Length float64
}

type Metafield struct {
	Value string `json:"value"`
}

type Metric struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type SupportedAction struct {
	Action string `json:"action"`
}

type Product struct {
	Title *string    `json:"tilte"`
	Largo *Metafield `json:"largo"`
	Ancho *Metafield `json:"ancho"`
	Alto  *Metafield `json:"alto"`
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

func (api *Api) GetCarrierServices(shop, token string)  ([]CarrierService, error) {
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
	req.Header.Set("X-Shopify-Access-Token", token)
	
	resp, err := api.client.Do(req)
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

func (api *Api) CarrierServiceCreate(shop, token, name, callbackUrl string) (*CarrierServiceCreate, error) {
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
	req.Header.Set("X-Shopify-Access-Token", token)
	
	resp, err := api.client.Do(req)
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

func (api *Api) CarrierServiceDelete(shop, token, id string) (*CarrierServiceDelete, error) {
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
	req.Header.Set("X-Shopify-Access-Token", token)
	
	resp, err := api.client.Do(req)
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

	var dim Metric 
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

func(api *Api) GetProductDimensions(shop, token, id string) (*DimensionCm, error) {
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
	req.Header.Set("X-Shopify-Access-Token", token)
	
	resp, err := api.client.Do(req)
  if err != nil {
		return nil, err
  }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("shopify get dimensions failed")
	}

	var graphql struct {
		Data struct {
		Product Product `json:"product"`
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

type Address struct {
	Address1     *string  `json:"address1"`
	Address2     *string  `json:"address2"`
	Phone        *string  `json:"phone"`
	City         *string  `json:"city"`
	Zip          *string  `json:"zip"`
	Province     *string  `json:"province"`
	Country      *string  `json:"country"`
	ContryCode   *string  `json:"country_code"`
	ProvinceCode *string  `json:"province_code"`
}

type LineItems struct {
	Nodes []struct {
		ID                string   `json:"id"`
		TotalQuantity     int      `json:"totalQuantity"`
		RemainingQuantity int      `json:"remainingQuantity"`
		Weigth           	Metric   `json:"weight"` 
		LineItem          struct {
			Product struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"product"`
		}                          `json:"lineItem"`
	} `json:"nodes"`
}

type AssignedLocation struct {
	Location struct {
		ID      string  `json:"id"`
		Name    string  `json:"name"`
		Address Address `json:"address"`
	} `json:"location"`
}

type FulfillmentOrders struct {
	Nodes []struct {
		ID               string   				 `json:"id"`
		Status           string   				 `json:"status"`
		SupportedActions []SupportedAction `json:"supportedActions"`
		AssignedLocation AssignedLocation  `json:"assignedLocation"` 
		LineItems LineItems 							 `json:"lineItems"`
	} `json:"nodes"`
}

func (api *Api) GetFulfillments(shop, token, orderID string) (*FulfillmentOrders, error) {

	type GraphQLVariables struct {
		OrderID string `json:"orderID"`
	}

	type GraphQLPayload struct {
		Query     string           `json:"query"`
		Variables GraphQLVariables `json:"variables"`
	}

	query := `
		query ($orderID: ID!) {
		  order(id: $orderID) {
		    fulfillmentOrders(first: 64, query: "status:OPEN") {
		      nodes {
		        id
		        status
		        supportedActions {
		          action
		        }
        		assignedLocation {
        		  location {
        		    id
        		    name
        		    address {
        		      address1
        		      address2
        		      city
        		      province
        		      provinceCode
        		      zip
        		      country
        		      countryCode
        		    }
        		  }
        		}
		        lineItems(first: 64) {
		          nodes {
		            id
            		totalQuantity
            		remainingQuantity
                weight {
                	unit
                	value
                }
								lineItem {
									product {
										id
										title
									}
								}
		          }
		        }
		      }
		    }
		  }
		}
	`

	payload := GraphQLPayload{
		Query: query, 
		Variables: GraphQLVariables{
			OrderID: orderID,
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
	req.Header.Set("X-Shopify-Access-Token", token)
	
	resp, err := api.client.Do(req)
  if err != nil {
		return nil, err
  }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("shopify get dimensions failed")
	}

	var graphql struct {
		Data struct {
			Order struct {
				FulfillmentOrders FulfillmentOrders `json:"fulfillmentOrders"`
			} `json:"order"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&graphql); err != nil {
		return nil, err
	}

	return &graphql.Data.Order.FulfillmentOrders, nil
}

