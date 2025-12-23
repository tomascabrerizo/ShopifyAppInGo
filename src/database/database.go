package database

import (
	"os"
	"time"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	handle *sql.DB
}

func NewDatabase(schemaPath string) (*Database, error) {
	handle, err := sql.Open("sqlite3", "file:./database/sqlite.db?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	if err := handle.Ping(); err != nil {
		return nil, err
	}
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		handle.Close()
		return nil, err
	}
	if _, err := handle.Exec(string(schema)); err != nil {
		handle.Close()
		return nil, err
	}
	db := &Database{handle: handle}
	return db, nil
}

func (db *Database) Close() {
	db.handle.Close()
}

type AccessToken struct {
	Shop   string `json:"shop"`
	Access string `json:"access_token"`
	Scopes string `json:"scopes"`
}

func (db *Database) InsertAccessToken(token *AccessToken) error {
	query := `INSERT INTO shops (shop, access_token, scopes) VALUES (?, ?, ?);`
	_, err := db.handle.Exec(query, token.Shop, token.Access, token.Scopes)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) GetAccessToken(shop string) (*AccessToken, error) {
	token := &AccessToken{}
	query := `SELECT shop, access_token, scopes FROM shops WHERE shop = ?;`
	if err := db.handle.QueryRow(query, shop).Scan(
		&token.Shop,
		&token.Access,
		&token.Scopes,
	); err != nil {
 		return nil, err
	}
	return token, nil
}

type Address struct {
	AddressID int64   `json:"address_id"`
	OrderID   *int64  `json:"order_id"`
  Email     *string `json:"email"`
  Phone     *string `json:"phone"`
  Name      *string `json:"name"`
  LastName  *string `json:"last_name"`
  Address1  *string `json:"address1"`
  Address2  *string `json:"address2"`
  Number    *int    `json:"number"`
  City      *string `json:"city"`
  Zip       *string `json:"zip"`
  Province  *string `json:"province"`
  Country   *string `json:"country"`
}

type OrderItem struct {
	ItemID    int64  `json:"item_id"`
	ItemApiID string `json:"item_api_id"`
	OrderID   int64  `json:"order_id"`
	Name      string `json:"name"`
	Grams     int64  `json:"grams"`
	Quantity  int64  `json:"quantity"`
	Currency  string `json:"currency"`
	Price     int64  `json:"price"`
	ProductID int64  `json:"product_id"`
	VariantID *int64 `json:"variant_id"`
	Sku       string `json:"sku"`
}

type Order struct {
	OrderID           int64  			`json:"order_id"`
	OrderApiID        string 			`json:"order_api_id"`
	Shop              string 			`json::"shop"`
	Currency          string 			`json:"currency"`
	SubtotalPrice     int64  			`json:"subtotal_price"`
	ShippingPrice     int64  			`json:"shipping_price"`
	Discount          int64  			`json:"discount"`
	TotalPrice        int64  			`json:"total_price"`
	CarrierName       *string 		`json:"carrier_name"`
	CarrierCode       *string 		`json:"carrier_code"`
	CarrierPrice      *int64  		`json:"carrier_price"`
	CreatedAt         time.Time 	`json:"created_at"`
	
	ShippingAddress   *Address    `json:"shipping_address"`
	Items             []OrderItem `json:"items"`
}

func insertAddressTx(tx *sql.Tx, address *Address) error {
	query := `
		INSERT INTO addresses (
			order_id, email, phone, name, last_name, 
			address1, address2, "number", 
			city, zip, province, country
  	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`

	res, err := tx.Exec(
		query,
		address.OrderID,
		address.Email,
		address.Phone,
		address.Name,
		address.LastName,
		address.Address1,
		address.Address2,
		address.Number,
		address.City,
		address.Zip,
		address.Province,
		address.Country,
	)

	if err != nil {
		return err
	}

	address.AddressID, _ = res.LastInsertId()

	return nil
}

func insertItemTx(tx *sql.Tx, item *OrderItem) error {

	query := `
		INSERT INTO order_items (
			item_id, item_api_id, order_id,
			name, grams, quantity,
			currency, price,
			product_id, variant_id, sku
  	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := tx.Exec(
		query,
		item.ItemID,
		item.ItemApiID,
		item.OrderID,
		item.Name,
		item.Grams,
		item.Quantity,
		item.Currency,
		item.Price,
		item.ProductID,
		item.VariantID,
		item.Sku,
	)

	if err != nil {
		return err
	}

	return nil
}


func (db *Database) InsertOrder(order *Order) error {
	tx, err := db.handle.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	query := `
		INSERT INTO orders (
  	  order_id, order_api_id, shop,
  	  currency, subtotal_price, shipping_price, discount, total_price,
  	  carrier_name, carrier_code, carrier_price
  	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`

	_, err = tx.Exec(
		query,
		order.OrderID,
		order.OrderApiID,
		order.Shop,
		order.Currency,
		order.SubtotalPrice,
		order.ShippingPrice,
		order.Discount,
		order.TotalPrice,
		order.CarrierName,
		order.CarrierCode,
		order.CarrierPrice,
	)

	if err != nil {
		return err
	}

	if order.ShippingAddress != nil {
		if err := insertAddressTx(tx, order.ShippingAddress); err != nil {
			return err
		}
	}

	for i := 0; i < len(order.Items); i++ {
		if err := insertItemTx(tx, &order.Items[i]); err != nil {
			return err
		} 
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (db *Database) DeleteOrder(orderID int64) error {
	query := `DELETE FROM orders WHERE order_id = ?;`
	_, err := db.handle.Exec(query, orderID)
	if err != nil {
		return err
	}
	return nil
}
