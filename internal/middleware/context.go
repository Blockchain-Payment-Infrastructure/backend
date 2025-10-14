package middleware

// ContextKey defines a type for context keys to avoid collisions.
type ContextKey string

// UserIDKey is the context key for storing the user ID.
const UserIDKey ContextKey = "userID"
