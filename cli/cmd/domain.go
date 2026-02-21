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

	"github.com/DoniLite/Mogoly/cli/actions"
	"github.com/DoniLite/Mogoly/cli/daemon"
	"github.com/spf13/cobra"
)

var (
	domainIsLocal bool
	domainLBName  string
	domainAutoSSL bool
)

// domainCmd represents the domain command
var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Manage custom domains with HTTPS",
	Long: `Manage custom domains for your load balancers with automatic HTTPS support.

Local domains (.local, .test) use self-signed certificates.
Production domains use Let's Encrypt for automatic SSL.

Examples:
  mogoly domain add api.local --local
  mogoly domain add api.myapp.com --lb mygateway --auto-ssl
  mogoly domain list
  mogoly domain remove api.local`,
}

// domainAddCmd adds a custom domain
var domainAddCmd = &cobra.Command{
	Use:   "add [domain]",
	Short: "Add a custom domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		payload := map[string]interface{}{
			"domain":   domain,
			"is_local": domainIsLocal,
			"lb_name":  domainLBName,
			"auto_ssl": domainAutoSSL,
		}

		resp, err := client.SendAction(ctx, actions.ActionDomainAdd, payload)
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to add domain: %s", resp.Error)
		}

		if domainIsLocal {
			fmt.Printf("✓ Local domain '%s' added successfully\n", domain)
			fmt.Printf("  Certificate generated at ~/.mogoly/certs/%s.crt\n", domain)
			fmt.Printf("  Added to /etc/hosts: 127.0.0.1 %s\n", domain)
			fmt.Printf("\n  Test with: curl https://%s\n", domain)
		} else {
			fmt.Printf("✓ Domain '%s' added successfully\n", domain)
			if domainAutoSSL {
				fmt.Printf("  Let's Encrypt certificate will be obtained automatically\n")
			}
		}

		return nil
	},
}

// domainListCmd lists all domains
var domainListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all custom domains",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, actions.ActionDomainList, nil)
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to list domains: %s", resp.Error)
		}

		var domains []map[string]interface{}
		if err := resp.DecodePayload(&domains); err != nil {
			return err
		}

		if len(domains) == 0 {
			fmt.Println("No custom domains configured")
			return nil
		}

		// Print table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "DOMAIN\tTYPE\tSTATUS\tCERTIFICATE")
		for _, domain := range domains {
			domainType := "Production"
			if isLocal, ok := domain["is_local"].(bool); ok && isLocal {
				domainType = "Local"
			}

			certInfo := "Let's Encrypt"
			if domainType == "Local" {
				certInfo = "Self-signed"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				domain["domain"],
				domainType,
				"Active",
				certInfo)
		}
		w.Flush()

		return nil
	},
}

// domainRemoveCmd removes a domain
var domainRemoveCmd = &cobra.Command{
	Use:     "remove [domain]",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a custom domain",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]

		client := daemon.NewClient("")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.SendAction(ctx, actions.ActionDomainRemove, map[string]string{"domain": domain})
		if err != nil {
			return fmt.Errorf("failed to communicate with daemon: %v", err)
		}

		if resp.Error != "" {
			return fmt.Errorf("failed to remove domain: %s", resp.Error)
		}

		fmt.Printf("✓ Domain '%s' removed successfully\n", domain)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(domainCmd)

	// Add subcommands
	domainCmd.AddCommand(domainAddCmd)
	domainCmd.AddCommand(domainListCmd)
	domainCmd.AddCommand(domainRemoveCmd)

	// Flags for add command
	domainAddCmd.Flags().BoolVar(&domainIsLocal, "local", false, "Create local development domain (.local, .test)")
	domainAddCmd.Flags().StringVar(&domainLBName, "lb", "", "Load balancer name to attach domain to")
	domainAddCmd.Flags().BoolVar(&domainAutoSSL, "auto-ssl", false, "Enable automatic SSL via Let's Encrypt")
}
