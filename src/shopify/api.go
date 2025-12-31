package shopify 

import (
	"fmt"
	"time"
	"sort"
	"bytes"
	"errors"

	"crypto/hmac"
	"crypto/sha256"

	"net/url"
	"net/http"
	
	"encoding/hex"
	"encoding/json"
	"encoding/base64"
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

func parseHmacAndMessage(r *http.Request) (string, string, error) {
	hmac := r.URL.Query().Get("hmac")
	if hmac == "" {
		return "", "", errors.New("missing hmac parameter") 
	}
	
	values := r.URL.Query()
	values.Del("hmac")
	
	params := []string{}
	for key, value := range values {
		paramName := key
		for _ , v := range value {
			paramValue := v 
			params = append(params, paramName+"="+paramValue)
		}
	}
	
	var message string
	sort.Strings(params)
	for i := 0; i < len(params); i++ {
		if(i > 0) {
			message += "&"
		}
		message += params[i]
	}

	return hmac, message, nil
}

func hmacSHA256(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Api) Verify(r *http.Request) error {
	received, message, err := parseHmacAndMessage(r)		
	if(err != nil) {
		return err
	}
	computed := hmacSHA256(message, s.Secret)
	if !hmac.Equal([]byte(computed), []byte(received)) {
		return errors.New("failed to verify shopify request")
	}
	return nil
}

func (s *Api) OAuthUrl(host, shop, state string) string {
	redirectUri := "https://" + host + "/api/auth/callback";
	u := url.URL{
		Scheme: "https",
		Host:   shop,
		Path:   "/admin/oauth/authorize",
	}
	q := u.Query()
	q.Set("client_id", s.ID)
	q.Set("redirect_uri", redirectUri)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	authUrl := u.String()
	return authUrl
}

func (s *Api) EmbeddedUrl(host string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(host)
	if err != nil {
		return "", err 
	}
	embeddedUrl := "https://"+string(decoded)+"/apps/"+s.ID+"/"
	return embeddedUrl, nil
}

type AccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
}

func (s *Api) OAuthRequestAccessToken(shop, code string) (*AccessTokenResponse, error) {
	type Payload struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
		Expiring     int    `json:"expiring"`
	}

	body := Payload{
		ClientID: s.ID,
		ClientSecret: s.Secret,
		Code: code,
		Expiring: 0,
	}

	jsonBody, err := json.Marshal(body)
  if err != nil {
		return nil, err
  }
	
	url := "https://"+shop+"/admin/oauth/access_token"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
  if err != nil {
		return nil, err
  }
	req.Header.Set("Content-Type", "application/json")


	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
  if err != nil {
		return nil, err
  }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("shopify token exchange failed")
	}

	tokenResp := &AccessTokenResponse{} 
	if err := json.NewDecoder(resp.Body).Decode(tokenResp); err != nil {
		return nil, err
	}

	return tokenResp, nil
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

