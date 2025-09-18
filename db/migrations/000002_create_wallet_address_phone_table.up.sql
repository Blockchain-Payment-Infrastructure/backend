CREATE TABLE wallet_address_phone (
  id SERIAL PRIMARY KEY,
  address TEXT UNIQUE NOT NULL,
  phone_number TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
