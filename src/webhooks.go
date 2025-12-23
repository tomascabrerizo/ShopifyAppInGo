package main 

import (
	"io"
	"log"
	
	"net/http"
	"encoding/json"

	"tomi/src/database"
	"tomi/src/shopify"
)

func (app *Application) AppUninstalledWebHook(w http.ResponseWriter, r *http.Request) {
	// TODO: Verify shopify webhook
	// TODO: Remove shop from the database
	log.Println("App uninstall webhook reached")
}

func (app *Application) OrdersWebhook(w http.ResponseWriter, r *http.Request) {
	// TODO: Verify shopify webhook

	topic := r.Header.Get("X-Shopify-Topic")
	shop := r.Header.Get("X-Shopify-Shop-Domain")

	// TODO: Check to webhook duplication
	// TODO: Search for a better way to handle webhooks updatetimes

	switch topic {
		case "orders/create":
			log.Println("OrdersCreateHandler:")
			app.OrdersCreateHandler(shop, w, r)
		case "orders/delete":
			log.Println("OrdersDeleteHandler:")
			app.OrdersDeleteHandler(w, r)
		case "orders/updated":
			log.Println("OrdersUpdatedHandler:")
			app.OrdersUpdatedHandler(shop, w, r)
		case "orders/fulfilled":
			log.Println("OrdersFulfilledHandler:")
			app.OrdersFulfilledHandler(shop, w, r)
		case "orders/paid":
			log.Println("OrdersPaidHandler:")
			app.OrdersPaidHandler(shop, w, r)
		case "orders/cancelled":
			log.Println("OrdersCanceledHandler:")
			app.OrdersCanceledHandler(shop, w, r)
		default:
			w.WriteHeader(http.StatusNoContent)
	}
}

func shopifyToDatabaseOrder(shop string, order shopify.Order) database.Order {
	
	subtotalPrice := shopify.GetShopMoney(order.CurrentSubtotalPriceSet)
	shippingPrice := shopify.GetShopMoney(order.CurrentShippingPriceSet)
	discount := shopify.GetShopMoney(order.CurrentTotalDiscountsSet)
	totalPrice := shopify.GetShopMoney(order.CurrentTotalPriceSet)

	var carrierName  *string = nil
	var carrierCode  *string = nil
	var carrierPrice int64  = 0 
	if(len(order.ShippingLines) > 0) {
		shippingLine := order.ShippingLines[0] 
		carrierName = &shippingLine.Title
		carrierCode = shippingLine.Code
		carrierPrice = shopify.GetShopMoney(shippingLine.PriceSet)
	}
	
	var address *database.Address = nil
	if order.ShippingAddress != nil {
		address = &database.Address{
			OrderID: &order.ID,
			Email: order.ContactEmail,
			Phone: order.ShippingAddress.Phone,
			Name: order.ShippingAddress.Name,
			LastName: order.ShippingAddress.LastName, 
			Address1: order.ShippingAddress.Address1,
			Address2: order.ShippingAddress.Address2,
			Number: nil,
			City: order.ShippingAddress.City,
			Zip: order.ShippingAddress.Zip,
			Province: order.ShippingAddress.Province,
			Country: order.ShippingAddress.Country,
		}
	}

	items := []database.OrderItem{}
	for _, lineItem := range order.LinesItems {
		item := database.OrderItem{
			ItemID: lineItem.ID,
			ItemApiID: lineItem.AdminGraphqlApiID,
			OrderID: order.ID,
			Name: lineItem.Name,
			Grams: lineItem.Grams,
			Quantity: lineItem.CurrentQuantity,
			Price: shopify.GetShopMoney(lineItem.PriceSet),
			ProductID: lineItem.ProductID,
			VariantID: lineItem.VariantID,
			Sku: lineItem.Sku,
		}
		items = append(items, item)
	}

	result := database.Order{
		OrderID: order.ID,
		OrderApiID: order.AdminGraphqlApiID,
		Shop: shop,
		Currency: order.Currency,
		SubtotalPrice: subtotalPrice,
		ShippingPrice: shippingPrice,
		Discount: discount,
		TotalPrice: totalPrice,
		CarrierName: carrierName,
		CarrierCode: carrierCode,
		CarrierPrice: &carrierPrice,
		ShippingAddress: address,
		Items: items,
		UpdatedAt: order.UpdatedAt,
	}

	return result
}

func (app *Application) OrdersCreateHandler(shop string, w http.ResponseWriter, r *http.Request) {
  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }
	
	payload := shopify.Order{}
	if err := json.Unmarshal(body, &payload); err != nil {
    log.Printf("failed to parse order webhook: %s\n", err)
    http.Error(w, "invalid JSON", http.StatusBadRequest)
    return
	}

	order := shopifyToDatabaseOrder(shop, payload)
	if err := app.db.UpsertOrder(&order); err != nil {
    log.Printf("failed to insert order : %s\n", err)
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
	
	}

	log.Println(order.UpdatedAt)
  w.WriteHeader(http.StatusOK)
}

func (app *Application) OrdersDeleteHandler(w http.ResponseWriter, r *http.Request) {
  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }

	payload := struct {
		ID int64 `json:"id"`
	}{} 
	if err := json.Unmarshal(body, &payload); err != nil {
    log.Printf("failed to parse order webhook: %s\n", err)
    http.Error(w, "invalid JSON", http.StatusBadRequest)
    return
	}
	
	if err := app.db.DeleteOrder(payload.ID); err != nil {
    log.Printf("failed to delete order : %s\n", err)
    http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *Application) OrdersUpdatedHandler(shop string, w http.ResponseWriter, r *http.Request) {
  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }

	payload := shopify.Order{}
	if err := json.Unmarshal(body, &payload); err != nil {
    log.Printf("failed to parse order webhook: %s\n", err)
    http.Error(w, "invalid JSON", http.StatusBadRequest)
    return
	}

	order := shopifyToDatabaseOrder(shop, payload)
	if err := app.db.UpsertOrder(&order); err != nil {
    log.Printf("failed to insert order : %s\n", err)
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
	
	}

	log.Println(order.UpdatedAt)
	w.WriteHeader(http.StatusOK)
}

func (app *Application) OrdersFulfilledHandler(shop string, w http.ResponseWriter, r *http.Request) {
  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }

	payload := shopify.Order{}
	if err := json.Unmarshal(body, &payload); err != nil {
    log.Printf("failed to parse order webhook: %s\n", err)
    http.Error(w, "invalid JSON", http.StatusBadRequest)
    return
	}

	order := shopifyToDatabaseOrder(shop, payload)
	order.Fulfilled = true
	if err := app.db.UpsertOrder(&order); err != nil {
    log.Printf("failed to insert order : %s\n", err)
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
	
	}

	log.Println(order.UpdatedAt)
	w.WriteHeader(http.StatusOK)
}

func (app *Application) OrdersPaidHandler(shop string, w http.ResponseWriter, r *http.Request) {
  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }

	payload := shopify.Order{}
	if err := json.Unmarshal(body, &payload); err != nil {
    log.Printf("failed to parse order webhook: %s\n", err)
    http.Error(w, "invalid JSON", http.StatusBadRequest)
    return
	}

	order := shopifyToDatabaseOrder(shop, payload)
	order.Paid = true
	if err := app.db.UpsertOrder(&order); err != nil {
    log.Printf("failed to insert order : %s\n", err)
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
	}

	log.Println(order.UpdatedAt)
	w.WriteHeader(http.StatusOK)
}

func (app *Application) OrdersCanceledHandler(shop string, w http.ResponseWriter, r *http.Request) {
  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }

	payload := shopify.Order{}
	if err := json.Unmarshal(body, &payload); err != nil {
    log.Printf("failed to parse order webhook: %s\n", err)
    http.Error(w, "invalid JSON", http.StatusBadRequest)
    return
	}

	order := shopifyToDatabaseOrder(shop, payload)
	order.Cancelled = true
	if err := app.db.UpsertOrder(&order); err != nil {
    log.Printf("failed to insert order : %s\n", err)
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
	}

	log.Println(order.UpdatedAt)
	w.WriteHeader(http.StatusOK)
}
