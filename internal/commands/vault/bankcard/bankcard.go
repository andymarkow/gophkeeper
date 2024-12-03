// Package bankcard provides the CLI command.
package bankcard

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/andymarkow/gophkeeper/internal/client/apiclient/cardapi"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo/cardinmem"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo/cardsqlite"
)

func NewAPIClient() (*cardapi.Client, error) {
	cacheType := viper.GetString("cache-type")

	var store cardrepo.Storage

	store = cardinmem.NewInMemory()

	if cacheType == "sqlite" {
		var err error
		store, err = cardsqlite.NewStorage(viper.GetString("data-dir") + "/sqlite.db")
		if err != nil {
			return nil, fmt.Errorf("cardsqlite.NewStorage: %w", err)
		}
	}

	client := cardapi.NewClient(store,
		cardapi.WithBaseURL(viper.GetString("address")),
		cardapi.WithAuthToken(viper.GetString("auth-token")),
	)

	return client, nil
}

func Register(parent *cobra.Command) error {
	cmd := &cobra.Command{
		Use:   "bankcard",
		Short: "Bank card secret management",
	}

	parent.AddCommand(cmd)

	createSecretCmd, err := CreateSecretCmd()
	if err != nil {
		return fmt.Errorf("failed to init create secret command: %w", err)
	}

	cmd.AddCommand(createSecretCmd)

	return nil
}
