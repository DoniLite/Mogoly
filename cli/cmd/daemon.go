/*
Copyright © 2025 Doni Lite hello@donilite.me
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/DoniLite/Mogoly/cli/daemon"
	"github.com/spf13/cobra"
)

var (
	daemonDetach bool
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:     "daemon",
	Aliases: []string{"d"},
	Short:   "Manage the Mogoly daemon",
	Long: `Manage the Mogoly daemon lifecycle.

Examples:
  mogoly daemon start
  mogoly daemon status
  mogoly daemon logs --tail 50`,
}

// daemonStartCmd starts the daemon
var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Mogoly daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if daemon is already running
		if isDaemonRunning() {
			fmt.Println("Daemon is already running")
			return nil
		}

		if daemonDetach {
			// Start daemon in detached mode
			return startDaemonDetached()
		} else {
			// Start daemon in foreground
			return startDaemonForeground()
		}
	},
}

// daemonStopCmd stops the daemon
var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Mogoly daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		pidPath, err := daemon.GetPIDFilePath()
		if err != nil {
			return err
		}

		// Read PID file
		pidBytes, err := os.ReadFile(pidPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Daemon is not running")
				return nil
			}
			return fmt.Errorf("failed to read PID file: %v", err)
		}

		pid, err := strconv.Atoi(string(pidBytes))
		if err != nil {
			return fmt.Errorf("invalid PID in file: %v", err)
		}

		// Find process
		process, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("failed to find process: %v", err)
		}

		// Send SIGTERM
		if err := process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("failed to stop daemon: %v", err)
		}

		fmt.Println("✓ Daemon stopped successfully")

		// Remove PID file
		os.Remove(pidPath)

		return nil
	},
}

// daemonRestartCmd restarts the daemon
var daemonRestartCmd = &cobra.Command{
	Use:     "restart",
	Aliases: []string{"reload"},
	Short:   "Restart the Mogoly daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Stop daemon
		if isDaemonRunning() {
			if err := daemonStopCmd.RunE(cmd, args); err != nil {
				return err
			}

			// Wait for daemon to stop
			time.Sleep(1 * time.Second)
		}

		// Start daemon
		return daemonStartCmd.RunE(cmd, args)
	},
}

// daemonStatusCmd shows daemon status
var daemonStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"ps"},
	Short:   "Show daemon status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isDaemonRunning() {
			fmt.Println("Daemon is not running")
			return nil
		}

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, daemon.ActionDaemonStatus, nil)
		if err != nil {
			fmt.Println("Daemon is not responding")
			return nil
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to get status: %s", resp.Error)
		}

		var status map[string]interface{}
		if err := resp.DecodePayload(&status); err != nil {
			return err
		}

		fmt.Println("✓ Daemon is running")
		fmt.Printf("  PID:        %.0f\n", status["pid"])
		fmt.Printf("  Socket:     %s\n", status["socket"])
		fmt.Printf("  Started At: %s\n", status["started_at"])

		return nil
	},
}

// daemonLogsCmd shows daemon logs
var daemonLogsCmd = &cobra.Command{
	Use:     "logs",
	Aliases: []string{"log"},
	Short:   "View daemon logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		logPath, err := daemon.GetLogFilePath()
		if err != nil {
			return err
		}

		// Check if log file exists
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			fmt.Println("No logs available")
			return nil
		}

		// Use tail command to show logs
		tailCmd := exec.Command("tail", "-n", strconv.Itoa(tailLines), logPath)
		if followLogs {
			tailCmd = exec.Command("tail", "-f", "-n", strconv.Itoa(tailLines), logPath)
		}

		tailCmd.Stdout = os.Stdout
		tailCmd.Stderr = os.Stderr

		return tailCmd.Run()
	},
}

// Helper functions

func isDaemonRunning() bool {
	pidPath, err := daemon.GetPIDFilePath()
	if err != nil {
		return false
	}

	// Check if PID file exists
	pidBytes, err := os.ReadFile(pidPath)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Try to send signal 0 (no-op) to check if process is alive
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func startDaemonForeground() error {
	server, err := daemon.NewServer("")
	if err != nil {
		return fmt.Errorf("failed to create daemon server: %v", err)
	}

	if err := server.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %v", err)
	}

	fmt.Println("✓ Daemon started successfully")
	fmt.Println("Press Ctrl+C to stop")

	// Wait for shutdown
	server.Wait()

	return nil
}

func startDaemonDetached() error {
	// Get the current executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Start daemon as background process
	cmd := exec.Command(executable, "daemon", "start")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %v", err)
	}

	// Detach from parent
	cmd.Process.Release()

	fmt.Println("✓ Daemon started successfully in background")

	return nil
}

func init() {
	rootCmd.AddCommand(daemonCmd)

	// Add subcommands
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonRestartCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonLogsCmd)

	// Flags
	daemonStartCmd.Flags().BoolVarP(&daemonDetach, "detach", "d", false, "Run daemon in background")
	daemonLogsCmd.Flags().IntVarP(&tailLines, "tail", "t", 100, "Number of lines to show")
	daemonLogsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Follow log output")
}
