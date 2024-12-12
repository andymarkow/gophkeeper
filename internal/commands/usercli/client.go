package usercli

import (
	"github.com/andymarkow/gophkeeper/internal/client/apiclient/userapi"
)

func initClient(addr string) (*userapi.Client, error) {
	client := userapi.NewClient(userapi.WithBaseURL(addr))

	return client, nil
}
