package users

// SignUpUserRequest represents a request to create a new user.
type SignUpUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// SignUpUserResponse represents a response to create a new user.
type SignUpUserResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

// SignInUserRequest represents a request to login a user.
type SignInUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// SignInUserResponse represents a response to login a user.
type SignInUserResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}
