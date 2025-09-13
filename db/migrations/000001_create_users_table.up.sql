CREATE TABLE users (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    phone_number TEXT UNIQUE NOT NULL,
    dob DATE NOT NULL DEFAULT now(),
    created_at TIMESTAMP DEFAULT now(),
    phone_verified BOOLEAN DEFAULT FALSE,
    account_complete BOOLEAN DEFAULT FALSE
);
