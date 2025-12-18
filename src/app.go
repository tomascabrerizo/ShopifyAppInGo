package main

import (
	"os"
	"log"
	"fmt"
	
	"net/url"
	"net/http"
	"net/http/httputil"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"tomi/src/shopify"
)

func dbInit() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./database/sqlite.db")
	if err != nil {
		return nil, err
	}
	schema, err := os.ReadFile("./database/schema.sql")
	if err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
} 

func dbSaveAccessToken(db *sql.DB, shop, accessToken, scopes string) error {
	query := `INSERT INTO shops (shop, access_token, scopes) VALUES (?, ?, ?);`
	_, err := db.Exec(query, shop, accessToken, scopes)
	if err != nil {
		return err
	}
	return nil
}

func dbCheckAuth(db *sql.DB, shop string) (string, error) {
	var token string
	query := `SELECT access_token FROM shops WHERE shop = ?`
  err := db.QueryRow(query, shop).Scan(&token)
  if err != nil {
 		return "", err
	}
	return token, nil
}

type Application struct {
	db *sql.DB
	proxy *httputil.ReverseProxy
	shopify *shopify.Shopify
}

func NewAppication() (*Application, error){
	db, err := dbInit()	
	if(err != nil) {
		return nil, err
	}

	target, err := url.Parse("http://localhost:5173")
	if(err != nil) {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	clientId := os.Getenv("SHOPIFY_CLIENT_ID")
	clientSecret := os.Getenv("SHOPIFY_CLIENT_SECRET") 
	shopify := shopify.NewShop(clientId, clientSecret)
	
	app := &Application{
		db: db,
		proxy: proxy,
		shopify: shopify,
	}

	return app, nil
}

func (app *Application) Shutdown() {
	app.db.Close()
}

func (app *Application) MainHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Handle oauth redirection correctly only escape if app is embedded
	shop := r.URL.Query().Get("shop")
	token, err := dbCheckAuth(app.db, shop)
	if err != nil {
		app.proxy.ServeHTTP(w, r)
	} else {
		w.Write([]byte(fmt.Sprintf("shop: %s has token: %s\n", shop, token)))
	}
}

func (app *Application) AuthHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := dbSaveAccessToken(app.db, shop, tokenResp.AccessToken, tokenResp.Scope); err != nil {
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
