package repository

import "errors"

var (
	ErrorDatabase   = errors.New("error while interacting with the database")
	ErrorUserExists = errors.New("user with username or email already exists")
)
