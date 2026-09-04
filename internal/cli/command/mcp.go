package command

import (
	"fmt"
	"os"

	cuGw "github.com/EgorTarasov/cu/internal/gateway/cu"
	mcpsrv "github.com/EgorTarasov/cu/internal/mcp"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server (stdio transport)",
	Long: `Start a Model Context Protocol server that exposes Central University tools for LLM clients.

The server communicates over stdin/stdout using JSON-RPC.
Configure it in your MCP client (e.g. Claude Code) as:
  {
    "mcpServers": {
      "cuni": { "command": "cuni", "args": ["mcp"] }
    }
  }`,
	Run: func(cmd *cobra.Command, _ []string) {
		client := mustClient()

		var gitlab mcpsrv.GitLabClient
		if gc, err := cuGw.NewGitLabClientFromEnv(); err == nil {
			gitlab = gc
		}

		srv := mcpsrv.NewServer(client, gitlab)
		if err := srv.Run(cmd.Context()); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	},
}
