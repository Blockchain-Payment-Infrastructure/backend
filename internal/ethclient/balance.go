package ethclient

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// GetETHBalance gets the ETH balance of an address in Wei
func (c *Client) GetETHBalance(ctx context.Context, address string) (*big.Int, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("invalid Ethereum address: %s", address)
	}

	account := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, account, nil)
	if err != nil {
		slog.Error("Failed to get ETH balance", slog.Any("error", err), slog.String("address", address))
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// GetETHBalanceAtBlock gets the ETH balance of an address at a specific block
func (c *Client) GetETHBalanceAtBlock(ctx context.Context, address string, blockNumber *big.Int) (*big.Int, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("invalid Ethereum address: %s", address)
	}

	account := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, account, blockNumber)
	if err != nil {
		slog.Error("Failed to get ETH balance at block",
			slog.Any("error", err),
			slog.String("address", address),
			slog.Any("blockNumber", blockNumber))
		return nil, fmt.Errorf("failed to get balance at block: %w", err)
	}

	return balance, nil
}

// GetETHBalanceInEther gets the ETH balance of an address in Ether (as float64)
func (c *Client) GetETHBalanceInEther(ctx context.Context, address string) (float64, error) {
	balanceWei, err := c.GetETHBalance(ctx, address)
	if err != nil {
		return 0, err
	}

	balanceEther := WeiToEther(balanceWei)
	result, _ := balanceEther.Float64()
	return result, nil
}

// GetMultipleBalances gets ETH balances for multiple addresses
func (c *Client) GetMultipleBalances(ctx context.Context, addresses []string) (map[string]*big.Int, error) {
	balances := make(map[string]*big.Int)

	for _, address := range addresses {
		if !common.IsHexAddress(address) {
			slog.Warn("Skipping invalid address", slog.String("address", address))
			continue
		}

		balance, err := c.GetETHBalance(ctx, address)
		if err != nil {
			slog.Error("Failed to get balance for address",
				slog.Any("error", err),
				slog.String("address", address))
			// Continue with other addresses instead of failing completely
			balances[address] = big.NewInt(0)
			continue
		}

		balances[address] = balance
	}

	return balances, nil
}

// HasSufficientBalance checks if an address has sufficient balance for a transaction
func (c *Client) HasSufficientBalance(ctx context.Context, address string, requiredAmount *big.Int, includeGas bool) (bool, error) {
	balance, err := c.GetETHBalance(ctx, address)
	if err != nil {
		return false, err
	}

	if !includeGas {
		return balance.Cmp(requiredAmount) >= 0, nil
	}

	// Estimate gas cost (simplified approach)
	gasPrice, err := c.GetGasPrice(ctx)
	if err != nil {
		slog.Warn("Failed to get gas price for balance check", slog.Any("error", err))
		// Use a default gas price estimate if we can't get current price
		gasPrice = big.NewInt(20000000000) // 20 Gwei
	}

	// Estimate gas limit (21000 for simple ETH transfer)
	gasLimit := big.NewInt(21000)
	gasCost := new(big.Int).Mul(gasPrice, gasLimit)

	// Total required = amount + gas cost
	totalRequired := new(big.Int).Add(requiredAmount, gasCost)

	return balance.Cmp(totalRequired) >= 0, nil
}

// ValidateAddress checks if an address is a valid Ethereum address
func ValidateAddress(address string) bool {
	return common.IsHexAddress(address)
}

// NormalizeAddress normalizes an Ethereum address to proper checksum format
func NormalizeAddress(address string) string {
	if !common.IsHexAddress(address) {
		return ""
	}
	return common.HexToAddress(address).Hex()
}

// FormatBalanceForDisplay formats a Wei balance for display purposes
func FormatBalanceForDisplay(balanceWei *big.Int, decimals int) string {
	if balanceWei == nil {
		return "0"
	}

	balanceEther := WeiToEther(balanceWei)
	format := fmt.Sprintf("%%.%df ETH", decimals)
	result, _ := balanceEther.Float64()

	return fmt.Sprintf(format, result)
}

// GetBalanceChange calculates the balance change between two blocks
func (c *Client) GetBalanceChange(ctx context.Context, address string, fromBlock, toBlock *big.Int) (*big.Int, error) {
	fromBalance, err := c.GetETHBalanceAtBlock(ctx, address, fromBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance at from block: %w", err)
	}

	toBalance, err := c.GetETHBalanceAtBlock(ctx, address, toBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance at to block: %w", err)
	}

	// Calculate difference (to - from)
	change := new(big.Int).Sub(toBalance, fromBalance)
	return change, nil
}
