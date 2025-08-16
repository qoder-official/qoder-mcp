package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/github/github-mcp-server/internal/ghmcp"
	"github.com/github/github-mcp-server/pkg/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Populated by goreleaser.
	version = "dev"
	commit  = "none"
	date    = "unknown"

	rootCmd = &cobra.Command{
		Use:     "github-mcp-server",
		Short:   "GitHub MCP server",
		Long:    `A server that communicates with GitHub's API and provides MCP functionality.`,
		Version: fmt.Sprintf("Version: %s\nCommit: %s\nBuild Date: %s", version, commit, date),
	}

	stdioCmd = &cobra.Command{
		Use:   "stdio",
		Short: "Start stdio server",
		Long:  `Start a server that communicates via standard input/output streams using JSON-RPC messages.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			// This is the correct place to check for the token
			token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
			if token == "" {
				return fmt.Errorf("GITHUB_PERSONAL_ACCESS_TOKEN not set in environment")
			}

			var enabledToolsets []string
			if err := viper.UnmarshalKey("toolsets", &enabledToolsets); err != nil {
				return fmt.Errorf("failed to unmarshal toolsets: %w", err)
			}

			stdioServerConfig := ghmcp.StdioServerConfig{
				Version:              version,
				Host:                 viper.GetString("host"),
				Token:                token,
				EnabledToolsets:      enabledToolsets,
				DynamicToolsets:      viper.GetBool("dynamic_toolsets"),
				ReadOnly:             viper.GetBool("read-only"),
				ExportTranslations:   viper.GetBool("export-translations"),
				EnableCommandLogging: viper.GetBool("enable-command-logging"),
				LogFilePath:          viper.GetString("log-file"),
			}

			return ghmcp.RunStdioServer(stdioServerConfig)
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.SetVersionTemplate("{{.Short}}\n{{.Version}}\n")
	rootCmd.PersistentFlags().StringSlice("toolsets", github.DefaultTools, "An optional comma separated list of groups of tools to allow, defaults to enabling all")
	rootCmd.PersistentFlags().Bool("dynamic-toolsets", false, "Enable dynamic toolsets")
	rootCmd.PersistentFlags().Bool("read-only", false, "Restrict the server to read-only operations")
	rootCmd.PersistentFlags().String("log-file", "", "Path to log file")
	rootCmd.PersistentFlags().Bool("enable-command-logging", false, "When enabled, the server will log all command requests and responses to the log file")
	rootCmd.PersistentFlags().Bool("export-translations", false, "Save translations to a JSON file")
	rootCmd.PersistentFlags().String("gh-host", "", "Specify the GitHub hostname (for GitHub Enterprise etc.)")

	_ = viper.BindPFlag("toolsets", rootCmd.PersistentFlags().Lookup("toolsets"))
	_ = viper.BindPFlag("dynamic_toolsets", rootCmd.PersistentFlags().Lookup("dynamic-toolsets"))
	_ = viper.BindPFlag("read-only", rootCmd.PersistentFlags().Lookup("read-only"))
	_ = viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
	_ = viper.BindPFlag("enable-command-logging", rootCmd.PersistentFlags().Lookup("enable-command-logging"))
	_ = viper.BindPFlag("export-translations", rootCmd.PersistentFlags().Lookup("export-translations"))
	_ = viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("gh-host"))

	rootCmd.AddCommand(stdioCmd)
}

func initConfig() {
	viper.SetEnvPrefix("github")
	viper.AutomaticEnv()

	if _, err := os.Stat("/.env"); err == nil {
		viper.SetConfigFile("/.env")
		if err := viper.ReadInConfig(); err != nil {
			// Don't exit on error, just log it. The token might be in the environment.
			fmt.Fprintf(os.Stderr, "Warning: could not read .env file: %v\n", err)
		}
	}

	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		token = viper.GetString("GITHUB_PERSONAL_ACCESS_TOKEN")
	}
	if token != "" {
		os.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", token)
		viper.Set("GITHUB_PERSONAL_ACCESS_TOKEN", token)
	}
}

func sendErrorAndExit(message string, err error) {
	errorResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    -32000,
			"message": message,
			"data": map[string]interface{}{
				"success": false,
				"error":   err,
			},
		},
		"id": nil,
	}

	if err != nil {
		errorResponse["error"].(map[string]interface{})["data"].(map[string]interface{})["error"] = err.Error()
	}

	jsonBytes, _ := json.Marshal(errorResponse)
	fmt.Fprintf(os.Stderr, "%s\n", string(jsonBytes))
	os.Exit(1)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		sendErrorAndExit("Command execution failed", err)
	}
}
