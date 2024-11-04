package userapi

type SignUpUserResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type SignInUserResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}
