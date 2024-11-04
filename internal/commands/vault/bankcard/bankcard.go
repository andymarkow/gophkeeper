// Package bankcard provides the CLI command.
package bankcard

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Register(parent *cobra.Command) error {
	cmd := &cobra.Command{
		Use:   "bankcard",
		Short: "Bank card secret management",
	}

	cli, err := NewCLI()
	if err != nil {
		return fmt.Errorf("NewCLI: %w", err)
	}

	if err := cli.Register(cmd); err != nil {
		return fmt.Errorf("Register: %w", err)
	}

	parent.AddCommand(cmd)

	return nil
}
