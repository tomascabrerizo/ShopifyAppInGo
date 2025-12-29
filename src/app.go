package main

import (
	"os"
	"log"
	"time"
	
	"encoding/json"

	"net/url"
	"net/http"
	"net/http/httputil"

	"tomi/src/database"
	"tomi/src/shopify"
)

type Application struct {
	db *database.Database
	proxy *httputil.ReverseProxy
	shopify *shopify.Shopify
	client *http.Client
	
	events chan Event
	lastEventIds *EventIdSB
}

func NewAppication() (*Application, error){
	db, err := database.NewDatabase("./database/schema.sql")
	if(err != nil) {
		return nil, err
	}

	target, err := url.Parse("http://localhost:5173")
	if(err != nil) {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	shopify := shopify.NewShop(
		os.Getenv("SHOPIFY_CLIENT_ID"),
		os.Getenv("SHOPIFY_CLIENT_SECRET"), 
	)

	events := make(chan Event,  512)
	
	app := &Application{
		db: db,
		proxy: proxy,
		shopify: shopify,
		events: events,
		lastEventIds: NewEventIdSB(),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	go app.ProcessEvents()

	return app, nil
}

func (app *Application) Shutdown() {
	app.db.Close()
}

func (app *Application) MainHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.shopify.Verify(r); err == nil {
		_, err := app.db.GetAccessToken(r.URL.Query().Get("shop"))
		if err != nil {
			http.ServeFile(w, r, "./app_bridge/dist/index.html")
			return
		}
	}
	app.proxy.ServeHTTP(w, r)
}

func (app *Application) AuthHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.shopify.Verify(r); err != nil {
		app.proxy.ServeHTTP(w, r)
		return
	}

	shop := r.URL.Query().Get("shop")
	// TODO: Generate random state and save it to a cookie
	url := app.shopify.OAuthUrl(r.Host, shop, "123")
	http.Redirect(w, r, url, http.StatusFound)
}

func (app *Application) AuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.shopify.Verify(r); err != nil {
		http.Error(w, "unauthorize request", http.StatusUnauthorized)
		return
	}
	
	shop := r.URL.Query().Get("shop")
	code := r.URL.Query().Get("code")
	host := r.URL.Query().Get("host")

	tokenResp, err := app.shopify.OAuthRequestAccessToken(shop, code)
	if err != nil {
		log.Printf("fail to get access token: %s\n", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
	
	token := database.AccessToken{
		Shop: shop,
		Access: tokenResp.AccessToken,
		Scopes: tokenResp.Scope,
	}
	if err := app.db.InsertAccessToken(&token); err != nil {
		log.Printf("fail to save access token: %s\n", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	embeddedUrl, err := app.shopify.EmbeddedUrl(host) 
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, embeddedUrl, http.StatusFound)
}

func (app *Application) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	shop := os.Getenv("SHOPIFY_SHOP_NAME")
	orders, err := app.db.GetUnfulfilledOrders(shop)
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

func (app *Application) CreateCarrierServiceHandler(w http.ResponseWriter, r *http.Request) {
	shop := os.Getenv("SHOPIFY_SHOP_NAME")	
	
	type CreateCarrierServicePayload struct {
		Name        string `json:"name"`
		CallbackURL string `json:"callbackUrl"`
	}
	
	var payload CreateCarrierServicePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	carrierService, err := app.CarrierServiceCreate(
		shop, 
		payload.Name,
		payload.CallbackURL,
	)
	
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	log.Println(payload.Name)
	log.Println(payload.CallbackURL)
	log.Printf("%v", carrierService)

	w.WriteHeader(http.StatusOK)
}

func (app *Application) GetCarrierServicesHandler(w http.ResponseWriter, r *http.Request) {
	shop := os.Getenv("SHOPIFY_SHOP_NAME")
	
	services, err := app.GetCarrierServices(shop)
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
