package main

import (
	"tomi/src/database"
)

func (app *Application) OnCreateOrderEvent(order *database.Order) {

}

func (app *Application) OnDeleteOrderEvent(id string) {

}

func (app *Application) OnUpdateOrderEvent(order *database.Order) {

}

func (app *Application) OnFulfilledOrderEvent(order *database.Order) {

}

func (app *Application) OnPaidOrderEvent(order *database.Order) {

}

func (app *Application) OnCancelledOrderEvent(order *database.Order) {

}
