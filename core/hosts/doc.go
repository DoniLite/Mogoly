/*
Package hosts provides cross-platform hosts file management.

# Overview

The hosts package allows adding, removing, and listing domain-to-IP mappings
in the system hosts file. It supports:
  - Linux/macOS: /etc/hosts
  - Windows: C:\Windows\System32\drivers\etc\hosts

All entries created by this package are marked with "# Managed by Mogoly"
for easy identification and cleanup.

# Quick Start

	import "github.com/DoniLite/Mogoly/core/hosts"

	func main() {
		// Create hosts manager
		manager, err := hosts.NewManager()
		if err != nil {
			log.Fatal(err)
		}

		// Add a domain entry (requires sudo/admin)
		err = manager.Add("myapp.local", "127.0.0.1")
		if err != nil {
			log.Fatal(err)
		}

		// Check if domain exists
		if manager.Has("myapp.local") {
			ip, _ := manager.Get("myapp.local")
			fmt.Printf("myapp.local -> %s\n", ip)
		}

		// List all Mogoly-managed entries
		entries := manager.List()
		for domain, ip := range entries {
			fmt.Printf("%s -> %s\n", domain, ip)
		}

		// Remove domain
		manager.Remove("myapp.local")
	}

# Hosts File Format

Entries are added in this format:

	127.0.0.1 myapp.local # Managed by Mogoly

The marker "# Managed by Mogoly" identifies entries that can be safely
managed (updated/removed) by this package. Entries without this marker
are left untouched.

# Permission Requirements

Modifying the hosts file requires administrator privileges:

  - Linux/macOS: Use sudo
  - Windows: Run as Administrator

If permissions are insufficient, a helpful error message is returned:

	permission denied: hosts file modification requires administrator privileges

	Please run with sudo:
	  sudo mogoly domain add myapp.local --local

# Thread Safety

The Manager is safe for concurrent use. Multiple goroutines can
safely add, remove, and list entries simultaneously.

# See Also

  - domain.Manager for high-level domain management with certificates
*/
package hosts
