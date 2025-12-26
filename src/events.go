package main

import (
	"fmt"
	"time"
	"errors"

	"database/sql"
	
	"tomi/src/database"
)

func (app *Application) validateOrder(order *database.Order) (bool, time.Time, error) {
	last, err := app.db.GetLastUpdatedFromOrder(order.OrderID)

	if errors.Is(err, sql.ErrNoRows) {
		if app.db.OrderWasDeleted(order.OrderID) {
			return false, time.Time{}, fmt.Errorf("order: %d, was deleted", order.OrderID)
		}
		return false, time.Time{}, nil
	}

	if err != nil {
		return false, time.Time{}, err
	}
	
	return true, last, nil
}

func (app *Application) upsertOrden(order *database.Order) error {
	exist, last, err := app.validateOrder(order) 
	if err != nil {
		return err
	}

	if !exist {
		return app.db.InsertOrder(order)
	}

	if order.UpdatedAt.After(last) {
		return app.db.UpdateOrder(order)
	}
	
	return nil
}

func (app *Application) OnCreateOrderEvent(order *database.Order) error {
	if err := app.upsertOrden(order); err != nil {
		return err
	}
	return nil
}

func (app *Application) OnDeleteOrderEvent(id int64)  error {
	if err := app.db.DeleteOrder(id); err != nil {
		return err
	}
	return nil
}

func (app *Application) OnUpdateOrderEvent(order *database.Order) error {
	if err := app.upsertOrden(order); err != nil {
		return err
	}
	return nil
}

func (app *Application) OnFulfilledOrderEvent(order *database.Order) error {
	if err := app.upsertOrden(order); err != nil {
		return err
	}
	if err := app.db.FulfillOrder(order); err != nil {
		return err
	}
	return nil
}

func (app *Application) OnPaidOrderEvent(order *database.Order) error {
	if err := app.upsertOrden(order); err != nil {
		return err
	}
	if err := app.db.PayOrder(order); err != nil {
		return err
	}
	return nil
}

func (app *Application) OnCancelledOrderEvent(order *database.Order) error {
	if err := app.upsertOrden(order); err != nil {
		return err
	}
	if err := app.db.CancelOrder(order); err != nil {
		return err
	}
	return nil
}
