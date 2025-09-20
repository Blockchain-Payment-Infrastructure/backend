package model

type WalletAddress struct {
	Address string `json:"address"`
}

type PhoneNumber struct {
	Phone string `json:"phone_number"`
}

type ConnectWalletRequest struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}
