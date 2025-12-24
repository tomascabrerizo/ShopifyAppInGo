package main 

import (
	"io"
	"log"
	"time"
	
	"net/http"
	
	"encoding/json"
	"tomi/src/shopify"
)

func (app *Application) AppUninstalledWebHook(w http.ResponseWriter, r *http.Request) {
	// TODO: Verify shopify webhook
	// TODO: Remove shop from the database
	log.Println("App uninstall webhook reached")
}

type EventIdSB struct {
	ids   [256]string
	index int
}

func NewEventIdSB() *EventIdSB {
	b := &EventIdSB{
		index: 0,
	}
	return b
}

func (b *EventIdSB) Add(id string) {
	b.ids[b.index] = id;
	b.index = (b.index + 1) & (len(b.ids)-1)
}

func (b *EventIdSB) Contains(id string) bool {
	if(id == "") {
		return false
	}
	for _, other := range b.ids {
		if id == other {
			return true
		}
	}
	return false
}

type Event struct {
	Shop      string
	Topic     string
	EventID   string
	TriggerAt time.Time
	ReceiveAt time.Time
	Body      []byte
}

func (app *Application) OrdersWebhook(w http.ResponseWriter, r *http.Request) {
	// TODO: Verify shopify webhook
	_ = r.Header.Get("X-Shopify-Hmac-Sha256")

  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }
	

	triggeredAt, err := time.Parse(time.RFC3339, r.Header.Get("X-Shopify-Triggered-At"))
	if err != nil {
		log.Printf("invalid Triggered-At header: %v", err)
		triggeredAt = time.Now().UTC()
	}
	
	event := Event{
		Shop:    	 r.Header.Get("X-Shopify-Shop-Domain"), 
		Topic:   	 r.Header.Get("X-Shopify-Topic"),
		EventID: 	 r.Header.Get("X-Shopify-Event-Id"),
		TriggerAt: triggeredAt, 
		ReceiveAt: time.Now(),
		Body:      body,
	}

	app.events <- event
	
	w.WriteHeader(http.StatusOK)
}

func (app *Application) ProcessEvents() {
	log.Println("Waiting for shopify events...")
	for event := range app.events {
		
		if app.lastEventIds.Contains(event.EventID) {
			log.Println("duplicate event received")
			continue
		}
		app.lastEventIds.Add(event.EventID)

		switch event.Topic {
			case "orders/create":
				payload := shopify.Order{}
				if err := json.Unmarshal(event.Body, &payload); err != nil {
					continue
				}
				order := payload.ToDatabaseOrder(event.Shop)
				app.OnCreateOrderEvent(&order)
			case "orders/delete":
				payload := struct { 
					ID string `json:"id"` 
				}{} 
				if err := json.Unmarshal(event.Body, &payload); err != nil {
					continue
				}
				app.OnDeleteOrderEvent(payload.ID)
			case "orders/updated":
				payload := shopify.Order{}
				if err := json.Unmarshal(event.Body, &payload); err != nil {
					continue
				}
				order := payload.ToDatabaseOrder(event.Shop)
				app.OnUpdateOrderEvent(&order)
			case "orders/fulfilled":
				payload := shopify.Order{}
				if err := json.Unmarshal(event.Body, &payload); err != nil {
					continue
				}
				order := payload.ToDatabaseOrder(event.Shop)
				app.OnFulfilledOrderEvent(&order)
			case "orders/paid":
				payload := shopify.Order{}
				if err := json.Unmarshal(event.Body, &payload); err != nil {
					continue
				}
				order := payload.ToDatabaseOrder(event.Shop)
				app.OnPaidOrderEvent(&order)
			case "orders/cancelled":
				payload := shopify.Order{}
				if err := json.Unmarshal(event.Body, &payload); err != nil {
					continue
				}
				order := payload.ToDatabaseOrder(event.Shop)
				app.OnCancelledOrderEvent(&order)
		}
	}
}

