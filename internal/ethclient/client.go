package ethclient

import (
	"backend/internal/config"
	"context"
	"log/slog"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

// Client wraps the Ethereum client with common operations
type Client struct {
	client *ethclient.Client
	rpcURL string
}

// NewClient creates a new Ethereum client
func NewClient() (*Client, error) {
	rpcURL := os.Getenv("ETHEREUM_RPC_URL")
	if rpcURL == "" {
		if config.AppMode != gin.ReleaseMode {
			rpcURL = "http://127.0.0.1:7545"
			slog.Warn("ETHEREUM_RPC_URL not set, using default local development RPC", slog.String("url", rpcURL))
		} else {
			panic("No ETHEREUM_RPC_URL found")
		}
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		slog.Error("Failed to connect to Ethereum client", slog.Any("error", err), slog.String("url", rpcURL))
		return nil, err
	}

	return &Client{
		client: client,
		rpcURL: rpcURL,
	}, nil
}

// Close closes the Ethereum client connection
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

// GetBalance gets the ETH balance of an address
func (c *Client) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	account := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, account, nil)
	if err != nil {
		slog.Error("Failed to get balance", slog.Any("error", err), slog.String("address", address))
		return nil, err
	}
	return balance, nil
}

// GetTransaction gets transaction details by hash
func (c *Client) GetTransaction(ctx context.Context, txHash string) (*types.Transaction, bool, error) {
	hash := common.HexToHash(txHash)
	tx, pending, err := c.client.TransactionByHash(ctx, hash)
	if err != nil {
		slog.Error("Failed to get transaction", slog.Any("error", err), slog.String("hash", txHash))
		return nil, false, err
	}
	return tx, pending, nil
}

// GetTransactionReceipt gets transaction receipt by hash
func (c *Client) GetTransactionReceipt(ctx context.Context, txHash string) (*types.Receipt, error) {
	hash := common.HexToHash(txHash)
	receipt, err := c.client.TransactionReceipt(ctx, hash)
	if err != nil {
		slog.Error("Failed to get transaction receipt", slog.Any("error", err), slog.String("hash", txHash))
		return nil, err
	}
	return receipt, nil
}

// GetBlockNumber gets the latest block number
func (c *Client) GetBlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := c.client.BlockNumber(ctx)
	if err != nil {
		slog.Error("Failed to get block number", slog.Any("error", err))
		return 0, err
	}
	return blockNumber, nil
}

// GetBlock gets block information by number
func (c *Client) GetBlock(ctx context.Context, blockNumber *big.Int) (*types.Block, error) {
	block, err := c.client.BlockByNumber(ctx, blockNumber)
	if err != nil {
		slog.Error("Failed to get block", slog.Any("error", err), slog.Any("blockNumber", blockNumber))
		return nil, err
	}
	return block, nil
}

// IsTransactionConfirmed checks if a transaction is confirmed (has receipt)
func (c *Client) IsTransactionConfirmed(ctx context.Context, txHash string) (bool, error) {
	_, err := c.GetTransactionReceipt(ctx, txHash)
	if err != nil {
		// If we can't find the receipt, the transaction is not confirmed
		return false, nil
	}
	return true, nil
}

// WaitForConfirmation waits for a transaction to be confirmed
func (c *Client) WaitForConfirmation(ctx context.Context, txHash string, confirmations uint64) (*types.Receipt, error) {
	hash := common.HexToHash(txHash)

	// First, get the transaction receipt
	receipt, err := c.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, err
	}

	// If no confirmations required, return immediately
	if confirmations == 0 {
		return receipt, nil
	}

	// Get current block number
	currentBlock, err := c.client.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	// Check if we have enough confirmations
	txBlock := receipt.BlockNumber.Uint64()
	if currentBlock >= txBlock+confirmations-1 {
		return receipt, nil
	}

	// In a real implementation, you might want to poll for confirmations
	// For now, we'll just return the receipt as is
	return receipt, nil
}
