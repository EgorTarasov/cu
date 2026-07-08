package command

import (
	"time"

	"github.com/EgorTarasov/cu/internal/telemetry"
	"github.com/EgorTarasov/cu/internal/update"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Telemetry state for the single command this process runs. Set in
// PersistentPreRun; read from PersistentPostRun and exitErrf (which bypasses
// PostRun via os.Exit).
var (
	cmdStartTime = time.Now()
	cmdPath      = "cu"
	cmdFlags     []string
)

var RootCmd = &cobra.Command{
	Use:   "cu",
	Short: "Central University CLI Tool",
	Long: `CU is a command-line tool for interacting with Central University services.
It provides access to courses, authentication, and data synchronization.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		cmdStartTime = time.Now()
		cmdPath = cmd.CommandPath()
		cmd.Flags().Visit(func(f *pflag.Flag) {
			cmdFlags = append(cmdFlags, f.Name)
		})
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		telemetry.Default().CommandExecuted(telemetry.CommandEvent{
			Command:  cmdPath,
			Flags:    cmdFlags,
			Duration: time.Since(cmdStartTime),
			Success:  true,
		})
		update.CheckForUpdate()
		telemetry.Default().Flush(telemetry.DefaultFlushTimeout)
	},
}

// ReportUsageError records a command that never reached PersistentPreRun —
// unknown command, bad flags. Called from main on RootCmd.Execute error.
func ReportUsageError() {
	reportCommandFailure("usage")
}

// reportCommandFailure sends the failure event and flushes; callers that
// os.Exit afterwards (exitErrf and friends) must call it first, because
// PersistentPostRun never runs for them.
func reportCommandFailure(errorKind string) {
	t := telemetry.Default()
	t.CommandExecuted(telemetry.CommandEvent{
		Command:   cmdPath,
		Flags:     cmdFlags,
		Duration:  time.Since(cmdStartTime),
		Success:   false,
		ErrorKind: errorKind,
	})
	t.Flush(telemetry.DefaultFlushTimeout)
}
