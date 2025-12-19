package main

import (
	"log"
	"net/http"
	
	"github.com/joho/godotenv"
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

func main() {
  
	if err := godotenv.Load(); err != nil {
    log.Fatal("Error loading .env file")
  }
	
	app, err := NewAppication()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer app.Shutdown()

	fs := http.FileServer(http.Dir("./app_bridge/dist"))
	http.Handle("/app_bridge/assets/", http.StripPrefix("/app_bridge/", fs))

	http.HandleFunc("/", app.MainHandler)
	http.HandleFunc("/api/auth", app.AuthHandler)
	http.HandleFunc("/api/auth/callback", app.AuthCallbackHandler)

	log.Print("Listening...")
	http.ListenAndServe("0.0.0.0:3000", cors(http.DefaultServeMux))

}
