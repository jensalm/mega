package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/spf13/cobra"
)

const trimPrefix = "github.com/michaelquigley/"

// tagline is mega's one line, provisional: the root command's help and the
// launcher's comment. the UI keeps the same line in ui/src/brand.ts.
const tagline = "build your constellation"

var (
	verbose    bool
	configPath string
	dataPath   string
	listenAddr string
)

var rootCmd = &cobra.Command{
	Use:           "mega",
	Short:         tagline,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runServe,
}

func init() {
	dl.Init(dl.DefaultOptions().SetTrimPrefix(trimPrefix))
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Config file path (default $XDG_CONFIG_HOME/mega/mega.yaml)")
	rootCmd.PersistentFlags().StringVar(&dataPath, "data", "", "Data directory (default $XDG_DATA_HOME/mega)")
	rootCmd.PersistentFlags().StringVar(&listenAddr, "listen", "", "Address to serve on, loopback only (default 127.0.0.1:6342)")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		// the commands that serve log to a file when no terminal sees them.
		switch cmd {
		case rootCmd, serveCmd, daemonCmd:
			initLogging(true)
		default:
			initLogging(false)
		}
	}
}

func init() {
	rootCmd.AddCommand(newDesktopCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		dl.Fatalf("%v", err)
	}
}
