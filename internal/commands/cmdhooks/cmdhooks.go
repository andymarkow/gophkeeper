// Package cmdhooks provides the command hooks.
package cmdhooks

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/andymarkow/gophkeeper/internal/client/authconfig"
)

func ReadAuthConfigFile(cmd *cobra.Command, _ []string) error {
	dataDir := viper.GetString("data-dir")

	authCfg, err := authconfig.ReadFile(dataDir)
	if err != nil {
		return fmt.Errorf("failed to read auth config: %w", err)
	}

	viper.Set("user-id", authCfg.UserID())
	viper.Set("username", authCfg.Username())
	viper.Set("auth-token", authCfg.Token())

	return nil
}
