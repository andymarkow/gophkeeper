// Package usercli provides the CLI command.
package usercli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/andymarkow/gophkeeper/internal/client/authconfig"
)

func initFlags(cmd *cobra.Command) error {
	cmd.Flags().StringP("username", "u", "", "Username")
	cmd.Flags().StringP("password", "p", "", "Password")

	if err := cmd.MarkFlagRequired("username"); err != nil {
		return fmt.Errorf("failed to mark flag required: %w", err)
	}

	if err := cmd.MarkFlagRequired("password"); err != nil {
		return fmt.Errorf("failed to mark flag required: %w", err)
	}

	return nil
}

func NewSignUpCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "signup",
		Short: "Sign up user",
		RunE:  runSignUpCmd,
	}

	if err := initFlags(cmd); err != nil {
		return nil, fmt.Errorf("failed to init flags: %w", err)
	}

	return cmd, nil
}

func runSignUpCmd(cmd *cobra.Command, args []string) error {
	addr := viper.GetString("address")
	username := cmd.Flags().Lookup("username").Value.String()
	password := cmd.Flags().Lookup("password").Value.String()

	client, err := initClient(addr)
	if err != nil {
		return fmt.Errorf("failed to init client: %w", err)
	}

	result, err := client.DoSignUpUser(cmd.Context(), username, password)
	if err != nil {
		return fmt.Errorf("failed to sign up user: %w", err)
	}

	authcfg := authconfig.NewAuthConfig(result.ID, username, result.Token)

	dataDir := viper.GetString("data-dir")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	if err := authcfg.WriteFile(dataDir); err != nil {
		return fmt.Errorf("failed to save auth config: %w", err)
	}

	fmt.Fprintf(os.Stdout, "User %s successfully signed up\n", username)

	return nil
}

func NewSignInCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "signin",
		Short: "Sign in user",
		RunE:  runSignInCmd,
	}

	if err := initFlags(cmd); err != nil {
		return nil, fmt.Errorf("failed to init flags: %w", err)
	}

	return cmd, nil
}

func runSignInCmd(cmd *cobra.Command, args []string) error {
	addr := viper.GetString("address")
	username := cmd.Flags().Lookup("username").Value.String()
	password := cmd.Flags().Lookup("password").Value.String()

	client, err := initClient(addr)
	if err != nil {
		return fmt.Errorf("failed to init client: %w", err)
	}

	result, err := client.DoSignInUser(cmd.Context(), username, password)
	if err != nil {
		return fmt.Errorf("failed to sign up user: %w", err)
	}

	authcfg := authconfig.NewAuthConfig(result.ID, username, result.Token)

	dataDir := viper.GetString("data-dir")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	if err := authcfg.WriteFile(dataDir); err != nil {
		return fmt.Errorf("failed to save auth config: %w", err)
	}

	fmt.Fprintf(os.Stdout, "User %s successfully signed in\n", username)

	return nil
}
