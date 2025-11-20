package repository

import "errors"

var (
    // Token related errors
    ErrorRefreshTokenNotFound    = errors.New("refresh token not found")
    ErrorRefreshTokenExpired     = errors.New("refresh token expired")
    ErrorRefreshTokenStoreFailed = errors.New("failed to store refresh token")
    ErrorRefreshTokenDeleteFailed = errors.New("failed to delete refresh token")

    // Wallet related errors
    ErrorWalletAddressAlreadyExists = errors.New("wallet address already exists")

    // Generic repository errors
    ErrorDatabaseServiceNotSet = errors.New("database service not set in repository")
    ErrorUserNotModified       = errors.New("user not found or not modified")
    ErrorUnhandledUniqueConstraint = errors.New("unhandled unique constraint")
    ErrorUpdatePasswordFailed  = errors.New("failed to update password")
    ErrorUpdateEmailFailed     = errors.New("failed to update email")
)
