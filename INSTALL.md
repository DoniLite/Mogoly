# Mogoly Installation & Setup Guide

## Installation

Mogoly **does not require root/admin privileges** to install. Only domain operations need elevated permissions.

### Linux/macOS

```bash
# Option 1: Using Go (recommended for development)
go install github.com/DoniLite/Mogoly/cli/mogoly@latest

# Option 2: Using install script
curl -sSL https://raw.githubusercontent.com/DoniLite/Mogoly/main/install.sh | bash

# Option 3: Download binary
curl -L https://github.com/DoniLite/Mogoly/releases/latest/download/mogoly-linux -o mogoly
chmod +x mogoly
sudo mv mogoly /usr/local/bin/
```

### Windows

```powershell
# Using PowerShell
Invoke-WebRequest -Uri "https://github.com/DoniLite/Mogoly/releases/latest/download/mogoly-windows.exe" -OutFile "mogoly.exe"

# Or using Go
go install github.com/DoniLite/Mogoly/cli/mogoly@latest
```

## What Gets Installed

The installation creates:

```
~/.mogoly/              # Config directory (no sudo needed)
├── config.yaml         # Configuration file
├── daemon.json         # Daemon state
├── mogoly.pid          # Process ID file
├── mogoly.log          # Daemon logs
└── certs/              # Certificate storage
    ├── *.crt           # Domain certificates
    └── *.key           # Private keys
```

**No files outside user directory** - safe installation!

## Permission Requirements

### What DOESN'T Need Sudo

✅ Installation
✅ Starting/stopping daemon
✅ Creating load balancers
✅ Adding backends
✅ Listing domains
✅ Viewing logs
✅ Health checks

### What NEEDS Sudo

⚠️ Adding local domains (modifies `/etc/hosts`)
⚠️ Removing local domains (modifies `/etc/hosts`)

## First-Time Setup

```bash
# 1. Install (no sudo)
curl -sSL https://install.mogoly.dev | bash

# 2. Start daemon (no sudo)
mogoly daemon start

# 3. Add your first domain (needs sudo)
sudo mogoly domain add myapp.local --local

# 4. Create load balancer (no sudo)
mogoly lb create --name myapp --port 443

# 5. Add backend (no sudo)
mogoly lb add-backend myapp --url http://localhost:3000
```

## Sudo-Free Alternative

If you cannot use sudo, use the DNS server approach instead:

```bash
# Start DNS server (no sudo needed)
mogoly dns start --port 5353

# Configure system DNS to point to localhost:5353
# Then use domains without /etc/hosts modification
```

## Platform-Specific Notes

### Linux

```bash
# Hosts file location
/etc/hosts

# Domain commands need sudo
sudo mogoly domain add myapp.local --local

# Or run with pkexec (GUI password prompt)
pkexec mogoly domain add myapp.local --local
```

### macOS

```bash
# Hosts file location
/etc/hosts

# Domain commands need sudo
sudo mogoly domain add myapp.local --local

# Flush DNS after adding domain
sudo dscacheutil -flushcache
```

### Windows

```powershell
# Hosts file location
C:\Windows\System32\drivers\etc\hosts

# Domain commands need Administrator
# Right-click CMD/PowerShell → "Run as Administrator"
mogoly domain add myapp.local --local

# Flush DNS after adding domain
ipconfig /flushdns
```

## Uninstallation

```bash
# Stop daemon
mogoly daemon stop

# Remove domains
mogoly domain list
sudo mogoly domain remove myapp.local  # For each domain

# Remove binary
rm $(which mogoly)

# Remove config (optional)
rm -rf ~/.mogoly
```

All hosts entries are automatically removed when you uninstall!

## Troubleshooting

### "Permission denied" when adding domain

**Problem**: `/etc/hosts` modification requires root

**Solution**: Use `sudo` or `pkexec`
```bash
sudo mogoly domain add myapp.local --local
```

### Can't use sudo

**Solution 1**: Ask admin to add domain once
```bash
sudo mogoly domain add myapp.local --local
# Domain persists until removed
```

**Solution 2**: Use DNS server mode (coming soon)

### Daemon won't start on port 443

**Problem**: Ports <1024 require root on Linux

**Solution**: Use unprivileged port
```bash
mogoly lb create --name myapp --port 8443
# Access via https://myapp.local:8443
```

Or allow daemon to bind privileged ports:
```bash
sudo setcap CAP_NET_BIND_SERVICE=+eip $(which mogoly)
mogoly daemon start  # Now can bind to 443 without sudo
```

## Security Considerations

1. **Installation**: Safe, user-space only
2. **Domain operations**: Requires sudo, modifies system file
3. **Certificates**: Stored with 0600 permissions
4. **Daemon**: Runs as regular user
5. **Hosts entries**: Marked with `# Managed by Mogoly` for easy identification

## Best Practices

1. **Install globally**: Everyone can use `mogoly` command
2. **Domains per-user**: Each user manages their own domains
3. **Shared daemon**: One system-wide daemon can serve all users
4. **Cleanup**: Always remove domains before uninstalling

---

## Summary

| Operation | Requires Sudo | Reason |
|-----------|--------------|--------|
| Installation | ❌ No | User-space only |
| Start daemon | ❌ No | Non-privileged ports |
| Add domain | ✅ Yes | Modifies /etc/hosts |
| Remove domain | ✅ Yes | Modifies /etc/hosts |
| Create LB | ❌ No | Configuration only |
| Add backend | ❌ No | Configuration only |

**Bottom line**: Normal installation, sudo only for domain management! 🎉
