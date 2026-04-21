package command

import (
	"fmt"
	"os"

	cuGw "cu-sync/internal/gateway/cu"
	mcpsrv "cu-sync/internal/mcp"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server (stdio transport)",
	Long: `Start a Model Context Protocol server that exposes CU tools for LLM clients.

The server communicates over stdin/stdout using JSON-RPC.
Configure it in your MCP client (e.g. Claude Code) as:
  {
    "mcpServers": {
      "cu": { "command": "cu", "args": ["mcp"] }
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
