package bankcard

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/andymarkow/gophkeeper/internal/commands/cmdhooks"
	"github.com/andymarkow/gophkeeper/internal/commands/customflags"
	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
)

func CreateSecretCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create bank card secret",
		Args:    cobra.ExactArgs(1),
		PreRunE: cmdhooks.ReadAuthConfigFile,
		RunE:    RunCreateSecretCmd,
	}

	cmd.Flags().String("name", "", "Bank card name")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		return nil, fmt.Errorf("failed to mark flag required: %w", err)
	}

	cmd.Flags().String("number", "", "Bank card number")

	if err := cmd.MarkFlagRequired("number"); err != nil {
		return nil, fmt.Errorf("failed to mark flag required: %w", err)
	}

	cmd.Flags().String("cvv", "", "Bank card CVV")

	if err := cmd.MarkFlagRequired("cvv"); err != nil {
		return nil, fmt.Errorf("failed to mark flag required: %w", err)
	}

	cmd.Flags().String("expire-at", "", "Bank card expiration date")

	if err := cmd.MarkFlagRequired("expire-at"); err != nil {
		return nil, fmt.Errorf("failed to mark flag required: %w", err)
	}

	cmd.Flags().Var(&customflags.StringMapFlag{}, "metadata", "Bank card metadata")

	return cmd, nil
}

func RunCreateSecretCmd(cmd *cobra.Command, args []string) error {
	secretName := args[0]
	metadataMap := cmd.Flag("metadata").Value.(*customflags.StringMapFlag)

	userID := viper.GetString("user-id")

	data, err := bankcard.NewData(
		cmd.Flag("number").Value.String(),
		cmd.Flag("name").Value.String(),
		cmd.Flag("cvv").Value.String(),
		cmd.Flag("expire-at").Value.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to create bank card data: %w", err)
	}

	secret, err := bankcard.CreateSecret(secretName, userID, metadataMap.GetMap(), data)
	if err != nil {
		return fmt.Errorf("failed to create bank card secret: %w", err)
	}

	apiclient, err := NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create bank card API cleint: %w", err)
	}

	err = apiclient.CreateSecret(cmd.Context(), secret)
	if err != nil {
		return fmt.Errorf("failed to create bank card secret: %w", err)
	}

	return nil
}
