package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gitlab.com/fforster/gitlab-mcp/lib/build"
)

// newVersionCommand returns a new Cobra command for displaying the current version of the GitLab MCP server.
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  "Display the current version of the GitLab MCP server",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("GitLab MCP version: %s (%s)\n", build.Version(), build.Date())
			fmt.Println()
			fmt.Println("Copyright (c) 2011-2022 GitLab, Inc.")
			fmt.Println("Released under the MIT License.")
			fmt.Println()
			fmt.Println("https://gitlab.com/fforster/gitlab-mcp")

			return nil
		},
		Args: cobra.NoArgs,
	}
}
