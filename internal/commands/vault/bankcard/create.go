package bankcard

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cardinmemcache "github.com/andymarkow/gophkeeper/internal/cache/bankcard/inmemory"
	"github.com/andymarkow/gophkeeper/internal/client/apiclient/cardapi"
	"github.com/andymarkow/gophkeeper/internal/commands/cmdhooks"
	"github.com/andymarkow/gophkeeper/internal/commands/customflags"
	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
)

type CLI struct {
	apiclient *cardapi.Client
}

func NewCLI() (*CLI, error) {
	cache := cardinmemcache.NewStorage()

	apiclient := cardapi.NewClient(cache,
		cardapi.WithCacheTTL(viper.GetDuration("cache-ttl")),
		cardapi.WithBaseURL(viper.GetString("address")),
		cardapi.WithAuthToken(viper.GetString("auth-token")),
	)

	return &CLI{
		apiclient: apiclient,
	}, nil
}

func (c *CLI) Register(parent *cobra.Command) error {
	createSecretCmd, err := c.CreateSecretCmd()
	if err != nil {
		return fmt.Errorf("CreateSecretCmd: %w", err)
	}

	parent.AddCommand(createSecretCmd)

	return nil
}

func (c *CLI) CreateSecretCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create bank card secret",
		Args:    cobra.ExactArgs(1),
		PreRunE: cmdhooks.ReadAuthConfigFile,
		RunE:    c.runCreateSecretCmd,
	}

	if err := initCreateSecretCmdFlags(cmd); err != nil {
		return nil, fmt.Errorf("failed to init flags: %w", err)
	}

	return cmd, nil
}

func (c *CLI) runCreateSecretCmd(cmd *cobra.Command, args []string) error {
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

	err = c.apiclient.CreateSecret(cmd.Context(), secret)
	if err != nil {
		return fmt.Errorf("failed to create bank card secret: %w", err)
	}

	return nil
}

func initCreateSecretCmdFlags(cmd *cobra.Command) error {
	cmd.Flags().String("name", "", "Bank card name")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		return fmt.Errorf("failed to mark flag required: %w", err)
	}

	cmd.Flags().String("number", "", "Bank card number")

	if err := cmd.MarkFlagRequired("number"); err != nil {
		return fmt.Errorf("failed to mark flag required: %w", err)
	}

	cmd.Flags().String("cvv", "", "Bank card CVV")

	if err := cmd.MarkFlagRequired("cvv"); err != nil {
		return fmt.Errorf("failed to mark flag required: %w", err)
	}

	cmd.Flags().String("expire-at", "", "Bank card expiration date")

	if err := cmd.MarkFlagRequired("expire-at"); err != nil {
		return fmt.Errorf("failed to mark flag required: %w", err)
	}

	cmd.Flags().Var(&customflags.StringMapFlag{}, "metadata", "Bank card metadata")

	return nil
}
