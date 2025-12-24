package shopify

import (
	"sort"
	"errors"
	"time"
	"bytes"
	"strings"
	"strconv"

	"net/url"
	"net/http"

	"crypto/hmac"
	"crypto/sha256"

	"encoding/hex"
	"encoding/json"
	"encoding/base64"

	"tomi/src/database"
)

type MailingAddress struct {
	FirstName    *string  `json:"first_name"`
	LastName     *string  `json:"last_name"`
	Address1     *string  `json:"address1"`
	Address2     *string  `json:"address2"`
	Phone        *string  `json:"phone"`
	City         *string  `json:"city"`
	Zip          *string  `json:"zip"`
	Province     *string  `json:"province"`
	Country      *string  `json:"country"`
	Company      *string  `json:"company"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	Name         *string  `json:"name"`
	ContryCode   *string  `json:"country_code"`
	ProvinceCode *string  `json:"province_code"`
}

type Money struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}

type MoneyBag struct {
	PresentmentMoney Money `json:"presentment_money"`
	ShopMoney        Money `json:"shop_money"`
}

type ShippingLine struct {
	CarrierIdentifier         *string  `json:"carrier_identifier"`
	Code                      *string  `json:"string"`
	Custom                    bool     `json:"custom"`
	Title                     string   `json:"title"`
	Source                    *string  `json:"source"`
	CurrentDiscountedPriceSet MoneyBag `json:"current_discounted_price_set"`
	DiscountedPriceSet        MoneyBag `json:"dicounted_price_set"`
	PriceSet                  MoneyBag `json:"price_set"`
}

type LineItem struct {
	ID                int64    `json:"id"`
	AdminGraphqlApiID string   `json:"admin_graphql_api_id"` 
	CurrentQuantity   int64    `json:"current_quantity"`
	Grams             int64    `json:"grams"`
	ProductID         int64    `json:"product_id"`
	PriceSet          MoneyBag `json:"price_set"`
	Sku               string   `json:"sku"`
	Name              string   `json:"name"`
	VariantID         *int64   `json:"variant_id"`
}

type Order struct {
	ID                		   int64           `json:"id"`
	AdminGraphqlApiID 		   string          `json:"admin_graphql_api_id"` 
	Currency          		   string          `json:"currency"`
	CurrentShippingPriceSet  MoneyBag        `json:"current_shipping_price_set"`
	CurrentSubtotalPriceSet  MoneyBag        `json:"current_subtotal_price_set"`
	CurrentTotalPriceSet     MoneyBag        `json:"current_total_price_set"`
	CurrentTotalDiscountsSet MoneyBag        `json:"current_total_discounts_set"`
	ContactEmail             *string         `json:"contact_email"`	
	ShippingAddress          *MailingAddress `json:"shipping_address"`
	ShippingLines            []ShippingLine  `json:"shipping_lines"`
	LinesItems               []LineItem      `json:"line_items"`
	UpdatedAt                time.Time       `json:"updated_at"`
}

func getShopMoney(bag MoneyBag) int64 {
	amount := bag.ShopMoney.Amount

	parts := strings.Split(amount, ".")
	if len(parts) == 0 || len(parts) > 2 {
		return 0
	}
	
	if len(parts) == 1 {
		i, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0
		}
		return i
	}
	
	whole := parts[0]
	frac := parts[1]
	frac = frac[:2]

	i, err := strconv.ParseInt(whole+frac, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func (o *Order) ToDatabaseOrder(shop string) database.Order {
	subtotalPrice := getShopMoney(o.CurrentSubtotalPriceSet)
	shippingPrice := getShopMoney(o.CurrentShippingPriceSet)
	discount := getShopMoney(o.CurrentTotalDiscountsSet)
	totalPrice := getShopMoney(o.CurrentTotalPriceSet)

	var carrierName  *string = nil
	var carrierCode  *string = nil
	var carrierPrice int64  = 0 
	if(len(o.ShippingLines) > 0) {
		shippingLine := o.ShippingLines[0] 
		carrierName = &shippingLine.Title
		carrierCode = shippingLine.Code
		carrierPrice = getShopMoney(shippingLine.PriceSet)
	}
	
	var address *database.Address = nil
	if o.ShippingAddress != nil {
		address = &database.Address{
			OrderID: &o.ID,
			Email: o.ContactEmail,
			Phone: o.ShippingAddress.Phone,
			Name: o.ShippingAddress.Name,
			LastName: o.ShippingAddress.LastName, 
			Address1: o.ShippingAddress.Address1,
			Address2: o.ShippingAddress.Address2,
			Number: nil,
			City: o.ShippingAddress.City,
			Zip: o.ShippingAddress.Zip,
			Province: o.ShippingAddress.Province,
			Country: o.ShippingAddress.Country,
		}
	}

	items := []database.OrderItem{}
	for _, lineItem := range o.LinesItems {
		item := database.OrderItem{
			ItemID: lineItem.ID,
			ItemApiID: lineItem.AdminGraphqlApiID,
			OrderID: o.ID,
			Name: lineItem.Name,
			Grams: lineItem.Grams,
			Quantity: lineItem.CurrentQuantity,
			Price: getShopMoney(lineItem.PriceSet),
			ProductID: lineItem.ProductID,
			VariantID: lineItem.VariantID,
			Sku: lineItem.Sku,
		}
		items = append(items, item)
	}

	result := database.Order{
		OrderID: o.ID,
		OrderApiID: o.AdminGraphqlApiID,
		Shop: shop,
		Currency: o.Currency,
		SubtotalPrice: subtotalPrice,
		ShippingPrice: shippingPrice,
		Discount: discount,
		TotalPrice: totalPrice,
		CarrierName: carrierName,
		CarrierCode: carrierCode,
		CarrierPrice: &carrierPrice,
		ShippingAddress: address,
		Items: items,
		UpdatedAt: o.UpdatedAt,
	}

	return result
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
