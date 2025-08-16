// Package cmd implements the root command for Cobra.
package cmd

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/build"
	"gitlab.com/fforster/gitlab-mcp/lib/tools"
)

// New creates a new command hierarchy for Cobra with the provided GitLab client.
func New(client *gitlab.Client) *cobra.Command {
	cmd := newRootCommand(client)

	cmd.AddCommand(newVersionCommand())

	return cmd
}

type rootCommand struct {
	client *gitlab.Client
}

// newRootCommand returns the root command for the CLI.
func newRootCommand(client *gitlab.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "gitlab-mcp",
		Short: "GitLab MCP server",
		Long:  "A command-line tool that provides an MCP server for interacting with GitLab.",
		RunE:  (&rootCommand{client}).run,
		Args:  cobra.NoArgs,
	}
}

func (c *rootCommand) run(_ *cobra.Command, _ []string) error {
	// Create a new MCP server
	s := server.NewMCPServer(
		"GitLab",
		build.Version(),
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	user, _, err := c.client.Users.CurrentUser()
	if err != nil {
		return fmt.Errorf("CurrentUser: %w", err)
	}

	tools := tools.New(c.client, user.Username)

	tools.AddTo(s)

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		return fmt.Errorf("ServeStdio: %w", err)
	}

	return nil
}
