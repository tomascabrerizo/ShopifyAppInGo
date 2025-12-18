package database

import (
	"os"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	handle *sql.DB
}

func NewDatabase(dbPath, schemaPath string) (*Database, error) {
	handle, err := sql.Open("sqlite3", dbPath)
	if err != nil {
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
	query := `SELECT shop, access_token, scopes FROM shops WHERE shop = ?`
	if err := db.handle.QueryRow(query, shop).Scan(
		&token.Shop,
		&token.Access,
		&token.Scopes,
	); err != nil {
 		return nil, err
	}
	return token, nil
}
