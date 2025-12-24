package main

import (
	"os"
	"log"
	
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
