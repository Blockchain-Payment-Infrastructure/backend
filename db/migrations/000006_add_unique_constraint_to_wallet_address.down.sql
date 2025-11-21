-- Remove unique constraint
ALTER TABLE wallet_address_phone 
DROP CONSTRAINT IF EXISTS unique_wallet_address;

