package repository

import "errors"

var (
	ErrorDatabase       = errors.New("error while interacting with the database")
	ErrorUsernameExists = errors.New("user with username already exists")
	ErrorEmailExists    = errors.New("user with email already exists")
	ErrorPhoneNumberExists = errors.New("user with phone number already exists")
)
