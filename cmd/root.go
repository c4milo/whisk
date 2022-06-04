package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"slack/whisk"
	"slack/whisk/chef"

	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

var rootCmd = &cobra.Command{
	Use:           "whisk [flags] <role_path>",
	Short:         "Whisk helps you mix and match Chef cookbooks without creating cycles",
	Long:          `More info at https://slack-github.com/slack/goslackgo/tree/master/whisk`,
	SilenceErrors: true,
	SilenceUsage:  true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			if err := cmd.Help(); err != nil {
				return fmt.Errorf("failed displaying usage: %w", err)
			}

			return errors.New("a role file path is required")
		}

		return nil
	},
	RunE: root,
}

var (
	cookbookPath string
	outputFormat string
)

// Execute parses CLI flags and arguments and runs the CLI command.
func Execute() error {
	rootCmd.PersistentFlags().StringVarP(&cookbookPath, "cookbook-path", "c", "./cookbooks", "Comma-separated cookbook paths")
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "ascii", "Output format, either ascii, json or dot")

	// Add subcommands to the root command here
	rootCmd.AddCommand(lintCmd)

	return rootCmd.Execute()
}

func root(cmd *cobra.Command, args []string) error {
	rolePath := args[0]

	tree := treeprint.New()
	handler := whisk.NewHandler(strings.Split(cookbookPath, ","), filepath.Dir(rolePath))

	role, err := chef.NewRole(rolePath)
	if err != nil {
		return fmt.Errorf("failed loading role: %w", err)
	}

	if err := handler.WalkRole(role.Name, tree); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := handler.FindSCCs(); err != nil {
		return fmt.Errorf("failed to find strongly connected components: %w", err)
	}

	if err := handler.FindCycles(); err != nil {
		return fmt.Errorf("failed to enumerate distinct cyles: %w", err)
	}

	switch outputFormat {
	case "json":
		if err := handler.JSON(os.Stdout); err != nil {
			return fmt.Errorf("failed to encode graph to JSON: %w", err)
		}
	case "dot":
		if err := handler.DOT(os.Stdout); err != nil {
			return fmt.Errorf("failed to encode graph to DOT: %w", err)
		}
	default:
		handler.ASCII(tree, os.Stdout)
	}

	return nil
}
