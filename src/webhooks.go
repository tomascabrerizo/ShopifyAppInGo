package main 

import (
	"io"
	"log"
	"fmt"
	"bytes"
	"strings"
	"strconv"
	
	"net/http"
	"encoding/json"

	"tomi/src/database"
)

func getString(m map[string]any, key string) (string, error) {
	if s, ok := m[key].(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("failed to get string for key %s\n", key)
}

func getInt(m map[string]any, key string) (int, error) {
	if n, ok := m[key].(json.Number); ok {
		i, _ := n.Int64()
		return int(i), nil
	}
	return 0, fmt.Errorf("failed to get int for key %s\n", key)
}

func getInt64(m map[string]any, key string) (int64, error) {
	if n, ok := m[key].(json.Number); ok {
		i, _ := n.Int64()
		return i, nil
	}
	return 0, fmt.Errorf("failed to get int64 for key %s\n", key)
}

func parseAmount(amount string) (int64, error) {
	parts := strings.Split(amount, ".")
	if len(parts) == 0 || len(parts) > 2 {
		return 0, fmt.Errorf("failed to parse amount: %s\n", amount)
	}
	
	if len(parts) == 1 {
		i, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, err
		}
		return i, nil
	}
	
	whole := parts[0]
	frac := parts[1]
	frac = frac[:2]

	i, err := strconv.ParseInt(whole+frac, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func getPrice(m map[string]any, key string) (int64, error) {
	v, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("invalid key: %s\n", key)
	}

	priceSet, ok := v.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("fail to get price set for key: %s\n", key)
	}

	shopMoney, ok := priceSet["shop_money"].(map[string]any)
	if !ok {
		return 0, fmt.Errorf("fail to get shop money for key: %s\n", key)
	}

	amount, ok := shopMoney["amount"].(string)
	if !ok {
		return 0, fmt.Errorf("fail to get amount for key: %s\n", key)
	}

	res, err := parseAmount(amount)
	if err != nil {
		return 0, err
	}

	return res, nil
}

func parseNumberFromAddressStr(address string) (int, error) {
	// TODO: implement this function
	return 0, nil
}

func payloadToAddress(email string, payload map[string]any) (*database.Address, error)  {
	var err error
	address := &database.Address{}
	
	address.Email = email
	
	if address.Phone, err = getString(payload, "phone"); err != nil {
		return nil, err
	}

	if address.Name , err = getString(payload, "first_name"); err != nil {
		return nil, err
	}

	if address.LastName, err = getString(payload, "last_name"); err != nil {
		return nil, err
	}

	if address.Address1 , err = getString(payload, "address1"); err != nil {
		return nil, err
	}

	if address.Address2 , err = getString(payload, "address2"); err != nil {
		address.Address2 = ""
	}

	if address.Number, err = parseNumberFromAddressStr(address.Address1); err != nil {
		return nil, err
	}

	if address.City, err = getString(payload, "city"); err != nil {
		return nil, err
	}

	if address.Zip, err = getString(payload, "zip"); err != nil {
		return nil, err
	}

	if address.Province, err = getString(payload, "province"); err != nil {
		return nil, err
	}

	if address.Country, err = getString(payload, "country"); err != nil {
		return nil, err
	}

	return address, nil
}

func payloadToOrderItems(payload []any) ([]database.OrderItem, error)  {
	return nil, nil
}

func payloadToOrder(shop string, payload map[string]any) (*database.Order, error) {
	var err error
	
	mail, err := getString(payload, "email")
	if err != nil {
		return nil, err
	}

	shippingAddressPayload, ok := payload["shipping_address"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid key shipping_address")
	}
	shippingAddress, err := payloadToAddress(mail, shippingAddressPayload)
	if err != nil {
		return nil, err
	}
	
	itemsPayload, ok := payload["line_items"].([]any)
	if !ok {
		return nil, fmt.Errorf("invalid key line_items")
	}
	items, err := payloadToOrderItems(itemsPayload)
	if err != nil {
		return nil, err
	}

	var order *database.Order = &database.Order{}
	order.Shop = shop
	order.ShippingAddress = *shippingAddress
	order.Items = items
	
	if order.OrderID, err = getInt64(payload, "id"); err != nil {
		return nil, err
	}
	
	if order.OrderApiID, err = getString(payload, "admin_graphql_api_id"); err != nil {
		return nil, err
	}
	
	if order.Currency, err = getString(payload, "currency"); err != nil {
		return nil, err
	}
	
	if order.SubtotalPrice, err = getPrice(payload, "current_subtotal_price_set"); err != nil {
		return nil, err
	}
	
	if order.ShippingPrice, err = getPrice(payload, "current_shipping_price_set"); err != nil {
		return nil, err
	}
	
	if order.Discount, err = getPrice(payload, "current_total_discounts_set"); err != nil {
		return nil, err
	}
	
	if order.TotalPrice, err = getPrice(payload, "current_total_price_set"); err != nil {
		return nil, err
	}
	
	// TODO: Find andreani shipping line
	shoppingLines, ok := payload["shipping_lines"].([]any)
	if !ok {
		return nil, fmt.Errorf("invalid key shipping_lines")
	}
	
	if len(shoppingLines) == 0 {
		return nil, fmt.Errorf("no carrier for the order")
	}
	
	shoppingLine, ok := shoppingLines[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid shoping line")
	}

	if order.CarrierName, err = getString(shoppingLine, "carrier_identifier"); err != nil {
		order.CarrierName = ""
	}

	if order.CarrierCode, err = getString(shoppingLine, "code"); err != nil {
		order.CarrierCode = ""
	}

	if order.CarrierPrice, err = getPrice(shoppingLine, "price_set"); err != nil {
		return nil, err
	}

	return order, nil
}

func (app *Application) AppUninstalledWebHook(w http.ResponseWriter, r *http.Request) {
	// TODO: Verify shopify webhook
	// TODO: Remove shop from the database
	log.Println("App uninstall webhook reached")
}

func (app *Application) OrdersWebhook(w http.ResponseWriter, r *http.Request) {
	// TODO: Verify shopify webhook

	topic := r.Header.Get("X-Shopify-Topic")
	switch topic {
		case "orders/create":
			log.Println("OrdersCreateHandler:")
			app.OrdersCreateHandler(w, r)
		case "orders/delete":
			log.Println("OrdersDeleteHandler:")
			app.OrdersDeleteHandler(w, r)
		case "orders/updated":
			log.Println("OrdersUpdatedHandler:")
			app.OrdersUpdatedHandler(w, r)
		case "orders/fulfilled":
			log.Println("OrdersFulfilledHandler:")
			app.OrdersFulfilledHandler(w, r)
		case "orders/paid":
			log.Println("OrdersPaidHandler:")
			app.OrdersPaidHandler(w, r)
		default:
			w.WriteHeader(http.StatusNoContent)
	}
}

func (app *Application) OrdersCreateHandler(w http.ResponseWriter, r *http.Request) {
  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()

	var payload map[string]any
	if err := dec.Decode(&payload); err != nil {
	  log.Printf("failed to parse orders webhook: %s\n", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	order, err := payloadToOrder("flichman", payload)
	if err != nil {
	  log.Printf("failed to parse order: %s\n", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	
	if err := app.db.InsertOrder(order); err != nil {
	  log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Order: %v\n", order)
  w.WriteHeader(http.StatusOK)
}

func (app *Application) OrdersDeleteHandler(w http.ResponseWriter, r *http.Request) {
  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }
  log.Println(string(body))
	w.WriteHeader(http.StatusOK)
}

func (app *Application) OrdersUpdatedHandler(w http.ResponseWriter, r *http.Request) {
  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }
  log.Println(string(body))
	w.WriteHeader(http.StatusOK)
}

func (app *Application) OrdersFulfilledHandler(w http.ResponseWriter, r *http.Request) {
  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }
  log.Println(string(body))
	w.WriteHeader(http.StatusOK)
}

func (app *Application) OrdersPaidHandler(w http.ResponseWriter, r *http.Request) {
  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }
  log.Println(string(body))
	w.WriteHeader(http.StatusOK)
}
