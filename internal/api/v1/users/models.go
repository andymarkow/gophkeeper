package users

// CreateUserRequest represents a request to create a new user.
type CreateUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// CreateUserResponse represents a response to create a new user.
type CreateUserResponse struct {
	Token string `json:"token"`
}

// LoginUserRequest represents a request to login a user.
type LoginUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginUserResponse represents a response to login a user.
type LoginUserResponse struct {
	Token string `json:"token"`
}
