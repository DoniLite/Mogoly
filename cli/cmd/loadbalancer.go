/*
Copyright © 2025 Doni Lite hello@donilite.me
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/DoniLite/Mogoly/cli/daemon"
	"github.com/spf13/cobra"
)

var (
	lbConfigPath string
	lbName       string
	backendURL   string
	backendName  string
)

// lbCmd represents the load balancer command
var lbCmd = &cobra.Command{
	Use:     "lb",
	Aliases: []string{"loadbalancer"},
	Short:   "Manage load balancers",
	Long: `Manage load balancer instances, backends, and health checks.

Examples:
  mogoly lb create --name api-gateway --config config.yaml
  mogoly lb list
  mogoly lb add-backend api-gateway --url http://localhost:8081
  mogoly lb health api-gateway`,
}

// lbCreateCmd creates a new load balancer
var lbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new load balancer",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		payload := daemon.LBCreatePayload{
			Name:       lbName,
			ConfigPath: lbConfigPath,
		}

		resp, err := client.SendAction(ctx, daemon.ActionLBCreate, payload)
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to create load balancer: %s", resp.Error)
		}

		fmt.Printf("✓ Load balancer '%s' created successfully\n", lbName)
		return nil
	},
}

// lbListCmd lists all load balancers
var lbListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all load balancers",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, daemon.ActionLBList, nil)
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to list load balancers: %s", resp.Error)
		}

		// Parse response
		var lbs []map[string]interface{}
		if err := resp.DecodePayload(&lbs); err != nil {
			return err
		}

		if len(lbs) == 0 {
			fmt.Println("No load balancers found")
			return nil
		}

		// Print table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tHOST\tPORT\tBACKENDS\tSTATUS")
		for _, lb := range lbs {
			fmt.Fprintf(w, "%s\t%s\t%.0f\t%.0f\t%s\n",
				lb["name"],
				lb["host"],
				lb["port"],
				lb["backends_count"],
				lb["status"])
		}
		w.Flush()

		return nil
	},
}

// lbAddBackendCmd adds a backend to a load balancer
var lbAddBackendCmd = &cobra.Command{
	Use:     "add-backend [lb-name]",
	Aliases: []string{"add"},
	Short:   "Add a backend server to a load balancer",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lbName := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		payload := daemon.LBAddBackendPayload{
			LBName:      lbName,
			BackendName: backendName,
			BackendURL:  backendURL,
		}

		resp, err := client.SendAction(ctx, daemon.ActionLBAddBackend, payload)
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to add backend: %s", resp.Error)
		}

		fmt.Printf("✓ Backend '%s' added to load balancer '%s'\n", backendName, lbName)
		return nil
	},
}

// lbRemoveBackendCmd removes a backend from a load balancer
var lbRemoveBackendCmd = &cobra.Command{
	Use:     "remove-backend [lb-name] [backend-name]",
	Aliases: []string{"rm"},
	Short:   "Remove a backend server from a load balancer",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		lbName := args[0]
		backendName := args[1]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		payload := map[string]string{
			"lb_name":      lbName,
			"backend_name": backendName,
		}

		resp, err := client.SendAction(ctx, daemon.ActionLBRemoveBackend, payload)
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to remove backend: %s", resp.Error)
		}

		fmt.Printf("✓ Backend '%s' removed from load balancer '%s'\n", backendName, lbName)
		return nil
	},
}

// lbHealthCmd checks health of a load balancer
var lbHealthCmd = &cobra.Command{
	Use:     "health [lb-name]",
	Aliases: []string{"status"},
	Short:   "Check health status of a load balancer",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lbName := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, daemon.ActionLBHealth, map[string]string{"lb_name": lbName})
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to check health: %s", resp.Error)
		}

		var health map[string]interface{}
		if err := resp.DecodePayload(&health); err != nil {
			return err
		}

		fmt.Printf("Load Balancer: %s\n", lbName)
		fmt.Printf("Status: %s\n", health["status"])

		if pass, ok := health["healthy_backends"].([]interface{}); ok && len(pass) > 0 {
			fmt.Println("\nHealthy Backends:")
			for _, b := range pass {
				if backend, ok := b.(map[string]interface{}); ok {
					fmt.Printf("  ✓ %s (%s)\n", backend["name"], backend["url"])
				}
			}
		}

		if fail, ok := health["unhealthy_backends"].([]interface{}); ok && len(fail) > 0 {
			fmt.Println("\nUnhealthy Backends:")
			for _, b := range fail {
				if backend, ok := b.(map[string]interface{}); ok {
					fmt.Printf("  ✗ %s (%s)\n", backend["name"], backend["url"])
				}
			}
		}

		return nil
	},
}

// lbStartCmd starts a load balancer
var lbStartCmd = &cobra.Command{
	Use:     "start [lb-name]",
	Aliases: []string{"up"},
	Short:   "Start a load balancer",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lbName := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, daemon.ActionLBStart, map[string]string{"lb_name": lbName})
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to start load balancer: %s", resp.Error)
		}

		fmt.Printf("✓ Load balancer '%s' started successfully\n", lbName)
		return nil
	},
}

// lbStopCmd stops a load balancer
var lbStopCmd = &cobra.Command{
	Use:     "stop [lb-name]",
	Aliases: []string{"down"},
	Short:   "Stop a load balancer",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lbName := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, daemon.ActionLBStop, map[string]string{"lb_name": lbName})
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to stop load balancer: %s", resp.Error)
		}

		fmt.Printf("✓ Load balancer '%s' stopped successfully\n", lbName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lbCmd)

	// Add subcommands
	lbCmd.AddCommand(lbCreateCmd)
	lbCmd.AddCommand(lbListCmd)
	lbCmd.AddCommand(lbAddBackendCmd)
	lbCmd.AddCommand(lbRemoveBackendCmd)
	lbCmd.AddCommand(lbHealthCmd)
	lbCmd.AddCommand(lbStartCmd)
	lbCmd.AddCommand(lbStopCmd)

	// Flags for create command
	lbCreateCmd.Flags().StringVarP(&lbName, "name", "n", "", "Load balancer name (required)")
	lbCreateCmd.MarkFlagRequired("name")
	lbCreateCmd.Flags().StringVarP(&lbConfigPath, "config", "c", "", "Configuration file path")

	// Flags for add-backend command
	lbAddBackendCmd.Flags().StringVarP(&backendURL, "url", "u", "", "Backend URL (required)")
	lbAddBackendCmd.MarkFlagRequired("url")
	lbAddBackendCmd.Flags().StringVarP(&backendName, "name", "n", "", "Backend name")

	// Global flags
	lbCmd.PersistentFlags().StringVarP(&outputFormat, "format", "o", "table", "Output format (table, json, yaml)")
}
