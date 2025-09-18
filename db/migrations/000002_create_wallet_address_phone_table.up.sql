CREATE TABLE wallet_address_phone (
  id SERIAL PRIMARY KEY,
  address TEXT NOT NULL,
  phone TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
