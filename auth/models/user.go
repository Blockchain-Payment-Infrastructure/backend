package models

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type User struct {
	ID             string `db:"id"`
	Username       string `db:"username"`
	Email          string `db:"email"`
	Phone          string `db:"phone"`
	HashedPassword string `db:"hashed_password"`
}
