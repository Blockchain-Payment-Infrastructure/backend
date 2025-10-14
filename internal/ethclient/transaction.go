package ethclient

import (
	"backend/internal/model"
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// VerifyTransaction verifies a transaction and returns detailed information
func (c *Client) VerifyTransaction(ctx context.Context, txHash string) (*model.TransactionDetails, error) {
	hash := common.HexToHash(txHash)

	// Get transaction details
	tx, pending, err := c.client.TransactionByHash(ctx, hash)
	if err != nil {
		slog.Error("Failed to get transaction", slog.Any("error", err), slog.String("hash", txHash))
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if pending {
		slog.Info("Transaction is still pending", slog.String("hash", txHash))
	}

	// Get transaction receipt for status and block information
	receipt, err := c.client.TransactionReceipt(ctx, hash)
	if err != nil {
		// Transaction might be pending, return basic info
		return &model.TransactionDetails{
			Hash:     tx.Hash().Hex(),
			From:     getTransactionSender(tx),
			To:       getTransactionRecipient(tx),
			Value:    tx.Value().String(),
			Gas:      tx.Gas(),
			GasPrice: tx.GasPrice().String(),
		}, nil
	}

	// Convert block number to int64 for our model
	var blockNumber *int64
	if receipt.BlockNumber != nil {
		bn := receipt.BlockNumber.Int64()
		blockNumber = &bn
	}

	return &model.TransactionDetails{
		Hash:        tx.Hash().Hex(),
		From:        getTransactionSender(tx),
		To:          getTransactionRecipient(tx),
		Value:       tx.Value().String(),
		Gas:         tx.Gas(),
		GasPrice:    tx.GasPrice().String(),
		BlockNumber: blockNumber,
		Status:      receipt.Status,
	}, nil
}

// ValidateTransactionForPayment validates if a transaction matches the payment request
func (c *Client) ValidateTransactionForPayment(ctx context.Context, txDetails *model.TransactionDetails, req *model.CreatePaymentRequest, fromAddress string) error {
	// Validate addresses
	if !common.IsHexAddress(txDetails.From) {
		return fmt.Errorf("invalid from address: %s", txDetails.From)
	}

	if !common.IsHexAddress(txDetails.To) {
		return fmt.Errorf("invalid to address: %s", txDetails.To)
	}

	// Check if the from address matches the user's wallet
	if !equalAddresses(txDetails.From, fromAddress) {
		return fmt.Errorf("transaction from address %s does not match user's wallet %s", txDetails.From, fromAddress)
	}

	// Check if the to address matches the request
	if !equalAddresses(txDetails.To, req.ToAddress) {
		return fmt.Errorf("transaction to address %s does not match requested address %s", txDetails.To, req.ToAddress)
	}

	// Validate amount (convert both to big.Int for comparison)
	txValue, ok := new(big.Int).SetString(txDetails.Value, 10)
	if !ok {
		return fmt.Errorf("invalid transaction value: %s", txDetails.Value)
	}

	reqValue, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		return fmt.Errorf("invalid requested amount: %s", req.Amount)
	}

	if txValue.Cmp(reqValue) != 0 {
		return fmt.Errorf("transaction amount %s does not match requested amount %s", txDetails.Value, req.Amount)
	}

	// Check transaction status (1 = success, 0 = failed)
	if txDetails.Status == 0 {
		return fmt.Errorf("transaction failed on blockchain")
	}

	return nil
}

// GetTransactionConfirmations returns the number of confirmations for a transaction
func (c *Client) GetTransactionConfirmations(ctx context.Context, txHash string) (uint64, error) {
	receipt, err := c.GetTransactionReceipt(ctx, txHash)
	if err != nil {
		return 0, fmt.Errorf("transaction not confirmed yet")
	}

	currentBlock, err := c.GetBlockNumber(ctx)
	if err != nil {
		return 0, err
	}

	confirmations := currentBlock - receipt.BlockNumber.Uint64() + 1
	return confirmations, nil
}

// EstimateGas estimates gas for a transaction
func (c *Client) EstimateGas(ctx context.Context, from, to common.Address, value *big.Int) (uint64, error) {
	msg := ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: value,
	}

	gasEstimate, err := c.client.EstimateGas(ctx, msg)
	if err != nil {
		slog.Error("Failed to estimate gas", slog.Any("error", err))
		return 0, err
	}

	return gasEstimate, nil
}

// GetGasPrice gets the current gas price
func (c *Client) GetGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		slog.Error("Failed to get gas price", slog.Any("error", err))
		return nil, err
	}
	return gasPrice, nil
}

// WeiToEther converts Wei to Ether
func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
}

// EtherToWei converts Ether to Wei
func EtherToWei(eth *big.Float) *big.Int {
	wei := new(big.Float).Mul(eth, big.NewFloat(1e18))
	result, _ := wei.Int(nil)
	return result
}

// ParseEtherAmount parses an ether amount string to Wei
func ParseEtherAmount(amount string) (*big.Int, error) {
	// Try to parse as float first
	ethAmount, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		// If that fails, try to parse as big.Int (assuming it's already in Wei)
		wei, ok := new(big.Int).SetString(amount, 10)
		if !ok {
			return nil, fmt.Errorf("invalid amount format: %s", amount)
		}
		return wei, nil
	}

	// Convert ether to wei
	ethBig := big.NewFloat(ethAmount)
	wei := EtherToWei(ethBig)
	return wei, nil
}

// Helper functions
func getTransactionSender(tx *types.Transaction) string {
	// Use types.Sender for side-chain IDs. This will work for all transaction types.
	signer := types.LatestSignerForChainID(tx.ChainId())
	sender, err := types.Sender(signer, tx)
	if err != nil {
		slog.Error("Failed to get transaction sender", slog.Any("error", err))
		// Fallback for older transaction types if needed, though LatestSigner should handle it.
		// For example, you might try another signer if the first fails.
		// For this application, logging the error and returning empty is sufficient.
		return ""
	}
	return sender.Hex()
}

func getTransactionRecipient(tx *types.Transaction) string {
	if tx.To() == nil {
		return "" // Contract creation
	}
	return tx.To().Hex()
}

func equalAddresses(addr1, addr2 string) bool {
	return common.HexToAddress(addr1) == common.HexToAddress(addr2)
}
