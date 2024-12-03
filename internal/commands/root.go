// Package commands provides the CLI commands.
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/andymarkow/gophkeeper/internal/commands/usercli"
	bankcardcmd "github.com/andymarkow/gophkeeper/internal/commands/vault/bankcard"
)

func NewCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "keeperctl <command>",
		Short:   "",
		Long:    ``,
		Example: `keeperctl --help`,
	}

	if err := initFlags(cmd); err != nil {
		return nil, fmt.Errorf("failed to init flags: %w", err)
	}

	signupCmd, err := usercli.NewSignUpCmd()
	if err != nil {
		return nil, fmt.Errorf("failed to init signup command: %w", err)
	}

	signinCmd, err := usercli.NewSignInCmd()
	if err != nil {
		return nil, fmt.Errorf("failed to init signin command: %w", err)
	}

	cmd.AddCommand(signupCmd)
	cmd.AddCommand(signinCmd)

	if err := bankcardcmd.Register(cmd); err != nil {
		return nil, fmt.Errorf("failed to init bankcard command: %w", err)
	}

	return cmd, nil
}

func initFlags(cmd *cobra.Command) error {
	cmd.PersistentFlags().StringP("address", "a", "http://localhost:8080", "Server address")
	err := viper.BindPFlag("address", cmd.PersistentFlags().Lookup("address"))
	if err != nil {
		return fmt.Errorf("viper.BindPFlag: %w", err)
	}

	cmd.PersistentFlags().StringP("data-dir", "d", "/tmp/keeperctl", "Data directory")
	err = viper.BindPFlag("data-dir", cmd.PersistentFlags().Lookup("data-dir"))
	if err != nil {
		return fmt.Errorf("viper.BindPFlag: %w", err)
	}

	cmd.PersistentFlags().String("cache-type", "inmemory", "Local cache type")
	err = viper.BindPFlag("cache-type", cmd.PersistentFlags().Lookup("cache-type"))
	if err != nil {
		return fmt.Errorf("viper.BindPFlag: %w", err)
	}

	return nil
}
