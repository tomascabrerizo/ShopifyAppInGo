CREATE TABLE IF NOT EXISTS shops (
	shop TEXT PRIMARY KEY,
	access_token TEXT NOT NULL,
	scopes TEXT NOT NULL,
	installed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS addresses (
  address_id INTEGER PRIMARY KEY,

  email TEXT,
  phone TEXT,
  name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  
  address1 TEXT NOT NULL,
  address2 TEXT,
  "number" INTEGER,
  city TEXT,
  zip TEXT NOT NULL,
  province TEXT,
  country TEXT
);

CREATE TABLE IF NOT EXISTS orders (
  order_id INTEGER PRIMARY KEY,
  order_api_id TEXT NOT NULL,
  
  shop TEXT NOT NULL,

  currency TEXT NOT NULL,
  subtotal_price INTEGER NOT NULL,
  shipping_price INTEGER,
  discount INTEGER,
  total_price INTEGER NOT NULL,
  
  carrier_name TEXT,
  carrier_code TEXT,
  carrier_price INTEGER,
  
  shipping_address_id INTEGER NOT NULL,

	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY (shop) REFERENCES shops(shop),
  FOREIGN KEY (shipping_address_id) REFERENCES addresses(address_id)
);

CREATE TABLE IF NOT EXISTS order_items (
  item_id INTEGER PRIMARY KEY,
  item_api_id TEXT NOT NULL,

  order_id INTEGER NOT NULL,
  
  name TEXT NOT NULL,
  grams INTEGER NOT NULL,
  quantity INTEGER NOT NULL,

  currency TEXT NOT NULL,
  price INTEGER NOT NULL,

  product_id INTEGER NOT NULL,
  variant_id INTEGER NOT NULL,
  sku TEXT NOT NULL,

  FOREIGN KEY (order_id) REFERENCES orders(order_id)
);
