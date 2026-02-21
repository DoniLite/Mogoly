/*
Copyright © 2025 Doni Lite hello@donilite.me
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	socketPath string
	noColor    bool
	debugMode  bool
	version    = "0.1.0"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mogoly",
	Short: "Mogoly - Load Balancer and Cloud Service Manager",
	Long: `Mogoly is a powerful CLI tool for managing load balancers and cloud database services.

It provides:
  • Load balancing with health checking and automatic failover
  • Docker-based database services (PostgreSQL, MySQL, MongoDB, Redis, MariaDB)
  • WebSocket-based daemon for real-time management
  • Cross-platform support (Linux, macOS, Windows)

Examples:
  mogoly daemon start              Start the daemon
  mogoly cloud create mydb -t postgres
  mogoly cloud list
  mogoly cloud logs mydb --tail 100`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global persistent flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.mogoly/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&socketPath, "socket", "s", "", "daemon socket path")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug logging")

	// Version template
	rootCmd.SetVersionTemplate(`Mogoly version {{.Version}}
`)
}
