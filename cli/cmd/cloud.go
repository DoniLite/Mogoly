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
	cloudType     string
	cloudVersion  string
	cloudUsername string
	cloudPassword string
	cloudDatabase string
	cloudDomain   string
	tailLines     int
	followLogs    bool
	outputFormat  string
)

// cloudCmd represents the cloud command
var cloudCmd = &cobra.Command{
	Use:     "cloud",
	Aliases: []string{"c"},
	Short:   "Manage cloud database services",
	Long: `Manage cloud database services including PostgreSQL, MySQL, MongoDB, Redis, and MariaDB.

Examples:
  mogoly cloud create mydb --type postgres --version 14
  mogoly cloud list
  mogoly cloud logs mydb --tail 100`,
}

// cloudCreateCmd creates a new database service
var cloudCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new database service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		payload := daemon.CloudCreatePayload{
			Name:         name,
			Type:         cloudType,
			Version:      cloudVersion,
			Username:     cloudUsername,
			Password:     cloudPassword,
			DatabaseName: cloudDatabase,
		}

		resp, err := client.SendAction(ctx, daemon.ActionCloudCreate, payload)
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to create service: %s", resp.Error)
		}

		fmt.Printf("✓ Service '%s' created successfully\n", name)
		return nil
	},
}

// cloudListCmd lists all database services
var cloudListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all database services",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, daemon.ActionCloudList, nil)
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to list services: %s", resp.Error)
		}

		// Parse response
		var instances []map[string]interface{}
		if err := resp.DecodePayload(&instances); err != nil {
			return err
		}

		if len(instances) == 0 {
			fmt.Println("No services found")
			return nil
		}

		// Print table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tSTATUS\tPORT\tCREATED")
		for _, inst := range instances {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.0f\t%s\n",
				inst["id"],
				inst["name"],
				inst["type"],
				inst["status"],
				inst["external_port"],
				inst["created_at"])
		}
		w.Flush()

		return nil
	},
}

// cloudStartCmd starts a stopped service
var cloudStartCmd = &cobra.Command{
	Use:     "start [id]",
	Aliases: []string{"up"},
	Short:   "Start a stopped service",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, daemon.ActionCloudStart, map[string]string{"id": id})
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to start service: %s", resp.Error)
		}

		fmt.Printf("✓ Service '%s' started successfully\n", id)
		return nil
	},
}

// cloudStopCmd stops a running service
var cloudStopCmd = &cobra.Command{
	Use:     "stop [id]",
	Aliases: []string{"down"},
	Short:   "Stop a running service",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, daemon.ActionCloudStop, map[string]string{"id": id})
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to stop service: %s", resp.Error)
		}

		fmt.Printf("✓ Service '%s' stopped successfully\n", id)
		return nil
	},
}

// cloudRestartCmd restarts a service
var cloudRestartCmd = &cobra.Command{
	Use:     "restart [id]",
	Aliases: []string{"reload"},
	Short:   "Restart a service",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, daemon.ActionCloudRestart, map[string]string{"id": id})
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to restart service: %s", resp.Error)
		}

		fmt.Printf("✓ Service '%s' restarted successfully\n", id)
		return nil
	},
}

// cloudDeleteCmd deletes a service
var cloudDeleteCmd = &cobra.Command{
	Use:     "delete [id]",
	Aliases: []string{"rm"},
	Short:   "Delete a service",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, daemon.ActionCloudDelete, map[string]string{"id": id})
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to delete service: %s", resp.Error)
		}

		fmt.Printf("✓ Service '%s' deleted successfully\n", id)
		return nil
	},
}

// cloudLogsCmd shows service logs
var cloudLogsCmd = &cobra.Command{
	Use:     "logs [id]",
	Aliases: []string{"log"},
	Short:   "View service logs",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		payload := daemon.CloudLogsPayload{
			ID:        id,
			TailLines: tailLines,
			Follow:    followLogs,
		}

		// For logs, we might use GetLogs or keep SendAction depending on streaming needs.
		// Previous implementation used SendRequest and DecodeData.
		// Let's use SendAction for consistency as client.GetLogs wasn't fully standardized in previous steps
		// or maybe I should use GetLogs if I added it. I remember adding GetLogs in Client.
		// However, to keep changes minimal and consistent with other commands first, I will use SendAction.
		// Wait, I replaced StreamLogs with GetLogs in client.go. The cli usage here was just a fetch.
		// Let's use SendAction for now as it returns *sync.Message.
		resp, err := client.SendAction(ctx, daemon.ActionCloudLogs, payload)
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to get logs: %s", resp.Error)
		}

		var result map[string]string
		if err := resp.DecodePayload(&result); err != nil {
			return err
		}

		fmt.Println(result["logs"])
		return nil
	},
}

// cloudInspectCmd inspects service details
var cloudInspectCmd = &cobra.Command{
	Use:     "inspect [id]",
	Aliases: []string{"info"},
	Short:   "Inspect service details",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, daemon.ActionCloudInspect, map[string]string{"id": id})
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to inspect service: %s", resp.Error)
		}

		var instance map[string]interface{}
		if err := resp.DecodePayload(&instance); err != nil {
			return err
		}

		// Print details
		fmt.Printf("ID:           %s\n", instance["id"])
		fmt.Printf("Name:         %s\n", instance["name"])
		fmt.Printf("Type:         %s\n", instance["type"])
		fmt.Printf("Status:       %s\n", instance["status"])
		fmt.Printf("Port:         %.0f\n", instance["external_port"])
		fmt.Printf("Container ID: %s\n", instance["container_id"])
		fmt.Printf("Created At:   %s\n", instance["created_at"])

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cloudCmd)

	// Add subcommands
	cloudCmd.AddCommand(cloudCreateCmd)
	cloudCmd.AddCommand(cloudListCmd)
	cloudCmd.AddCommand(cloudStartCmd)
	cloudCmd.AddCommand(cloudStopCmd)
	cloudCmd.AddCommand(cloudRestartCmd)
	cloudCmd.AddCommand(cloudDeleteCmd)
	cloudCmd.AddCommand(cloudLogsCmd)
	cloudCmd.AddCommand(cloudInspectCmd)

	// Flags for create command
	cloudCreateCmd.Flags().StringVarP(&cloudType, "type", "t", "postgres", "Database type (postgres, mysql, mongodb, redis, mariadb)")
	cloudCreateCmd.Flags().StringVarP(&cloudVersion, "version", "v", "", "Database version")
	cloudCreateCmd.Flags().StringVarP(&cloudUsername, "username", "u", "admin", "Database username")
	cloudCreateCmd.Flags().StringVarP(&cloudPassword, "password", "p", "", "Database password")
	cloudCreateCmd.Flags().StringVarP(&cloudDatabase, "database", "d", "", "Database name")
	cloudCreateCmd.Flags().StringVar(&cloudDomain, "domain", "", "Custom domain")

	// Flags for logs command
	cloudLogsCmd.Flags().IntVarP(&tailLines, "tail", "t", 100, "Number of lines to show")
	cloudLogsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Follow log output")

	// Global flags
	cloudCmd.PersistentFlags().StringVarP(&outputFormat, "format", "o", "table", "Output format (table, json, yaml)")
}
