CREATE TABLE IF NOT EXISTS shops (
	shop TEXT PRIMARY KEY,
	access_token TEXT NOT NULL,
	scopes TEXT NOT NULL,
	installed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orders (
  order_id INTEGER PRIMARY KEY,
  order_api_id TEXT NOT NULL,
  
  shop TEXT NOT NULL,

  currency TEXT NOT NULL,
  subtotal_price INTEGER NOT NULL,
  shipping_price INTEGER NOT NULL,
  discount INTEGER NOT NULL,
  total_price INTEGER NOT NULL,
  
  carrier_name TEXT,
  carrier_code TEXT,
  carrier_price INTEGER,
  
  cancelled BOOLEAN NOT NULL DEFAULT FALSE,
  paid BOOLEAN NOT NULL DEFAULT FALSE,
  fulfilled BOOLEAN NOT NULL DEFAULT FALSE,
  deleted BOOLEAN NOT NULL DEFAULT FALSE,
  
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS addresses (
  address_id INTEGER PRIMARY KEY,
  order_id INTEGER,

  email TEXT,
  phone TEXT,
  name TEXT,
  last_name TEXT,
  
  address1 TEXT,
  address2 TEXT,
  "number" INTEGER,
  city TEXT,
  zip TEXT,
  province TEXT,
  country TEXT,

  UNIQUE(order_id),
  FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE CASCADE
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
  variant_id INTEGER,
  sku TEXT NOT NULL,

  FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE CASCADE
);
