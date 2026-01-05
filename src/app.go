package main

import (
	"strconv"
	"fmt"
	"log"
	"os"

	"encoding/json"

	"net/http"
	"net/http/httputil"
	"net/url"

	"tomi/src/andreani"
	"tomi/src/database"
	"tomi/src/shopify"
)

type Application struct {
	db    *database.Database
	proxy *httputil.ReverseProxy

	shop string

	shopApi *shopify.Api
	andApi  *andreani.Api

	events       chan Event
	lastEventIds *EventIdSB
}

func NewAppication() (*Application, error) {
	db, err := database.NewDatabase("./database/schema.sql")
	if err != nil {
		return nil, err
	}

	target, err := url.Parse("http://localhost:5173")
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	shopApi := shopify.NewApi(
		os.Getenv("SHOPIFY_CLIENT_ID"),
		os.Getenv("SHOPIFY_CLIENT_SECRET"),
	)

	andApi := andreani.NewApi(
		os.Getenv("ANDREANI_CLIENT_CODE"),
		os.Getenv("ANDREANI_ACCESS_TOKEN"),
		os.Getenv("ANDREANI_BASE_URL"),
	)

	events := make(chan Event, 512)

	app := &Application{
		db:           db,
		proxy:        proxy,
		shop:         os.Getenv("SHOPIFY_SHOP_NAME"),
		shopApi:      shopApi,
		andApi:       andApi,
		events:       events,
		lastEventIds: NewEventIdSB(),
	}

	go app.ProcessEvents()

	return app, nil
}

func (app *Application) Shutdown() {
	app.db.Close()
}

func (app *Application) MainHandler(w http.ResponseWriter, r *http.Request) {
	if err := shopify.Verify(app.shopApi, r); err == nil {
		_, err := app.db.GetAccessToken(r.URL.Query().Get("shop"))
		if err != nil {
			http.ServeFile(w, r, "./app_bridge/dist/index.html")
			return
		}
	}
	app.proxy.ServeHTTP(w, r)
}

func (app *Application) AuthHandler(w http.ResponseWriter, r *http.Request) {
	if err := shopify.Verify(app.shopApi, r); err != nil {
		app.proxy.ServeHTTP(w, r)
		return
	}

	shop := r.URL.Query().Get("shop")
	// TODO: Generate random state and save it to a cookie
	url := shopify.OAuthUrl(app.shopApi, r.Host, shop, "123")
	http.Redirect(w, r, url, http.StatusFound)
}

func (app *Application) AuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if err := shopify.Verify(app.shopApi, r); err != nil {
		http.Error(w, "unauthorize request", http.StatusUnauthorized)
		return
	}

	shop := r.URL.Query().Get("shop")
	code := r.URL.Query().Get("code")
	host := r.URL.Query().Get("host")

	tokenResp, err := app.shopApi.OAuthRequestAccessToken(shop, code)
	if err != nil {
		log.Printf("fail to get access token: %s\n", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}

	token := database.AccessToken{
		Shop:   shop,
		Access: tokenResp.AccessToken,
		Scopes: tokenResp.Scope,
	}
	if err := app.db.InsertAccessToken(&token); err != nil {
		log.Printf("fail to save access token: %s\n", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	embeddedUrl, err := app.shopApi.EmbeddedUrl(host)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, embeddedUrl, http.StatusFound)
}

func (app *Application) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orders, err := app.db.GetUnfulfilledOrders(app.shop)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		log.Println("json encode error:", err.Error())
	}
}

func (app *Application) GetOrderFulfillmentsHandler(w http.ResponseWriter, r *http.Request) {
	token, err := app.db.GetAccessToken(app.shop)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	
	id := r.PathValue("orderID")
	if id == "" {
		http.Error(w, "missing orderID", http.StatusBadRequest)
		return
	}

	unscaped, err := url.PathUnescape(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	fulfillments, err := app.shopApi.GetFulfillments(app.shop, token.Access, unscaped)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(fulfillments); err != nil {
		log.Println("json encode error:", err.Error())
	}
}

func (app *Application) CreateCarrierServiceHandler(w http.ResponseWriter, r *http.Request) {
	token, err := app.db.GetAccessToken(app.shop)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Payload struct {
		Name        string `json:"name"`
		CallbackURL string `json:"callbackUrl"`
	}

	var payload Payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	carrierService, err := app.shopApi.CarrierServiceCreate(
		app.shop,
		token.Access,
		payload.Name,
		payload.CallbackURL,
	)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(carrierService); err != nil {
		log.Println("json encode error:", err.Error())
	}
}

func (app *Application) GetCarrierServicesHandler(w http.ResponseWriter, r *http.Request) {
	token, err := app.db.GetAccessToken(app.shop)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	services, err := app.shopApi.GetCarrierServices(app.shop, token.Access)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(services); err != nil {
		log.Println("json encode error:", err.Error())
	}
}

func (app *Application) DeleteCarrierServicesHandler(w http.ResponseWriter, r *http.Request) {
	token, err := app.db.GetAccessToken(app.shop)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	id := r.PathValue("serviceID")
	if id == "" {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	unscaped, err := url.PathUnescape(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	log.Println(unscaped)

	carrierService, err := app.shopApi.CarrierServiceDelete(
		app.shop,
		token.Access,
		unscaped,
	)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(carrierService); err != nil {
		log.Println("json encode error:", err.Error())
	}

}

func (app *Application) CarrierServiceCallbackHandler(w http.ResponseWriter, r *http.Request) {
	token, err := app.db.GetAccessToken(app.shop)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Address struct {
		Country    string `json:"country"`
		Province   string `json:"province"`
		PostalCode string `json:"postal_code"`
	}

	type Item struct {
		Quantity  int   `json:"quantity"`
		Grams     int64 `json:"grams"`
		ProductID int64 `json:"product_id"`
		VariantID int64 `json:"variant_id"`
	}

	var payload struct {
		Rate struct {
			Origin      Address `json:"origin"`
			Destination Address `json:"destination"`
			Items       []Item  `json:"items"`
		} `json:"rate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	items := make([]PackageItem, 0, len(payload.Rate.Items))
	for _, it := range payload.Rate.Items {
		item := PackageItem{
			ProductID: fmt.Sprintf("gid://shopify/Product/%d", it.ProductID),
			Quantity:  it.Quantity,
		}
		items = append(items, item)
	}

	volumen, err := calculatePackageVolumen(app.shopApi, token.Access, app.shop, items)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: These contracts needs to be loaded from database
	contratoEntrega := "400017493"
	zip := onlyDigits(payload.Rate.Destination.PostalCode)
	volumenStr := strconv.FormatFloat(volumen, 'f', 2, 64)
	
	rate, err := app.andApi.CalculateShippingRate(contratoEntrega, zip, volumenStr)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	
	totalPrice, err := strconv.ParseFloat(rate.TarifaConIva.Total, 64)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Rate struct {
		ServiceName string `json:"service_name"`
		ServiceCode string `json:"service_code"`
		TotalPrice  string `json:"total_price"`
		Description string `json:"description"`
		Currency    string `json:"currency"`
	}
		
	var result struct {
		Rates []Rate `json:"rates"`
	}

	result.Rates = append(result.Rates, Rate{
		ServiceName: "Andreani a domicilio",
		ServiceCode: contratoEntrega,
		TotalPrice: fmt.Sprintf("%d", int64(totalPrice * 100)),
		Description: "envio directo a tu domicilio",
		Currency: "ARS",
	})
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Println("json encode error:", err.Error())
	}
}
