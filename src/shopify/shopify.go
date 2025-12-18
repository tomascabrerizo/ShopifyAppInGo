package shopify

import (
	"sort"
	"errors"
	"time"
	"bytes"

	"net/url"
	"net/http"

	"crypto/hmac"
	"crypto/sha256"

	"encoding/hex"
	"encoding/json"
	"encoding/base64"
)


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

type Shopify struct {
	ID string
	Secret string
}

func NewShop(clientId, clientSecret string) *Shopify {
	return &Shopify{
		ID: clientId,
		Secret: clientSecret,
	}
}

func (s *Shopify) Verify(r *http.Request) error {
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

func (s *Shopify) OAuthUrl(host, shop, state string) string {
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

func (s *Shopify) EmbeddedUrl(host string) (string, error) {
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

func (s *Shopify) OAuthRequestAccessToken(shop, code string) (*AccessTokenResponse, error) {
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
