-- Add unique constraint to prevent same wallet address from being linked to multiple accounts
ALTER TABLE wallet_address_phone 
ADD CONSTRAINT unique_wallet_address UNIQUE (address);

