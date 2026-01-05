package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"strings"

	"net/url"
	"net/http"
		
	"github.com/joho/godotenv"
	"github.com/golang-jwt/jwt/v5"
	
	// TODO: Remove from main.go
	"tomi/src/andreani"
)

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	  w.Header().Set("Access-Control-Allow-Origin", "*")
	  w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	  w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	  if r.Method == "OPTIONS" {
	    w.WriteHeader(http.StatusOK)
	    return
	  }
	  next.ServeHTTP(w, r)
	})
}

func unauthorizedResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error": unauthorized request}`))
}

func matchShops(iss string, dest string) bool {
  issURL, err := url.Parse(iss)
  if err != nil {
		return false
  }
  
	destURL, err := url.Parse(dest)
	if err != nil {
		return false
  }
	
	baseISS := fmt.Sprintf("%s://%s", issURL.Scheme, issURL.Host)
	baseDest := fmt.Sprintf("%s://%s", destURL.Scheme, destURL.Host)
	
	if !strings.EqualFold(baseISS, baseDest) {
		return false;
	}
  
	if !strings.EqualFold(issURL.Host, os.Getenv("SHOPIFY_SHOP_NAME")) {
		return false;
	}

	return true
}

func shopifyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")	
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			unauthorizedResponse(w)
			return
		}

		token, err := jwt.Parse(
			parts[1], 
			func(token *jwt.Token) (any, error) {
				return []byte(os.Getenv("SHOPIFY_CLIENT_SECRET")), nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			jwt.WithExpirationRequired(),
			jwt.WithLeeway(30*time.Second),
		)
		if err != nil {
			log.Println("jwt parser fails!")
			log.Println(err.Error())
			unauthorizedResponse(w)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims); 
		if !ok {
			unauthorizedResponse(w)
			return
		}

		issVal, ok := claims["iss"].(string)
		if !ok {
		  unauthorizedResponse(w)
		  return
		}
		
		destVal, ok := claims["dest"].(string)
		if !ok {
		  unauthorizedResponse(w)
		  return
		}
		
		if !matchShops(issVal, destVal) {
		  unauthorizedResponse(w)
		  return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
  
	if err := godotenv.Load(); err != nil {
    log.Fatal("Error loading .env file")
  }
	
	app, err := NewAppication()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer app.Shutdown()
	
	
	/////////////////////////////////////////////////////////
	// Localidad test 
	/////////////////////////////////////////////////////////
	
	locations, err := app.andApi.GetLocations(andreani.LocationQuery{
		Province: "mendoza",
		Location: "guaymallen",
		Zips: []string{"5521"},
	})
	
	if err != nil {
		log.Fatal(err.Error())
	}
	
	for _, location := range locations {
		log.Printf("%v\n", location)
	}

	/////////////////////////////////////////////////////////

	http.HandleFunc("/webhooks/app-uninstalled", app.AppUninstalledWebHook)
	http.HandleFunc("/webhooks/orders", app.OrdersWebhook)

	fs := http.FileServer(http.Dir("./app_bridge/dist"))
	http.Handle("/app_bridge/assets/", http.StripPrefix("/app_bridge/", fs))

	http.HandleFunc("/", app.MainHandler)
	http.HandleFunc("/api/auth", app.AuthHandler)
	http.HandleFunc("/api/auth/callback", app.AuthCallbackHandler)

	http.Handle(
		"GET /api/orders",
		shopifyAuth(http.HandlerFunc(app.GetOrdersHandler)),
	)

	http.Handle(
		"GET /api/orders/{orderID}/fulfillments",
		shopifyAuth(http.HandlerFunc(app.GetOrderFulfillmentsHandler)),
	)

	http.Handle(
		"POST /api/carrier-service",
		shopifyAuth(http.HandlerFunc(app.CreateCarrierServiceHandler)),
	)

	http.Handle(
		"GET /api/carrier-service",
		shopifyAuth(http.HandlerFunc(app.GetCarrierServicesHandler)),
	)

	http.Handle(
		"DELETE /api/carrier-service/{serviceID}",
		shopifyAuth(http.HandlerFunc(app.DeleteCarrierServicesHandler)),
	)

	http.HandleFunc("POST /api/carrier-service/callback", app.CarrierServiceCallbackHandler)

	log.Print("Listening...")
	http.ListenAndServe("0.0.0.0:3000", cors(http.DefaultServeMux))

}
