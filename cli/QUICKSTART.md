# Mogoly CLI - Quick Start Guide

Congratulations! You now have a fully functional Mogoly CLI with daemon socket support.

## Installation

```bash
cd /home/doni/Documents/Projects/Perso/Mogoly/cli
go build -o mogoly
sudo mv mogoly /usr/local/bin/
```

## Quick Start

### 1. Start the Daemon

```bash
# Start in foreground
mogoly daemon start

# Or start in background
mogoly daemon start --detach
```

### 2. Check Daemon Status

```bash
mogoly daemon status
```

### 3. Create a Database Service

```bash
# Create PostgreSQL database
mogoly cloud create mydb --type postgres --version 14

# Create MySQL database
mogoly cloud create mysqldb --type mysql --username root --password secret

# Create MongoDB
mogoly cloud create mongodb --type mongodb
```

### 4. List Services

```bash
mogoly cloud list
```

### 5. Manage Services

```bash
# View logs
mogoly cloud logs mydb --tail 100

# Inspect details
mogoly cloud inspect mydb

# Stop service
mogoly cloud stop mydb

# Start service
mogoly cloud start mydb

# Restart service
mogoly cloud restart mydb

# Delete service
mogoly cloud delete mydb
```

### 6. View Daemon Logs

```bash
# Show last 100 lines
mogoly daemon logs --tail 100

# Follow logs in real-time
mogoly daemon logs --follow
```

### 7. Stop the Daemon

```bash
mogoly daemon stop
```

## Command Reference

### Cloud Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `cloud create` | `c create` | Create database service |
| `cloud list` | `c ls` | List all services |
| `cloud start` | `c up` | Start stopped service |
| `cloud stop` | `c down` | Stop running service |
| `cloud restart` | `c reload` | Restart service |
| `cloud delete` | `c rm` | Delete service |
| `cloud logs` | `c log` | View service logs |
| `cloud inspect` | `c info` | Inspect service details |

### Daemon Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `daemon start` | `d start` | Start daemon |
| `daemon stop` | `d stop` | Stop daemon |
| `daemon restart` | `d reload` | Restart daemon |
| `daemon status` | `d ps` | Check status |
| `daemon logs` | `d log` | View logs |

### Global Flags

- `--config, -c`: Configuration file path
- `--socket, -s`: Custom socket path
- `--no-color`: Disable colored output
- `--debug`: Enable debug logging
- `--version`: Show version

## Examples

### Create PostgreSQL with Custom Settings

```bash
mogoly cloud create proddb \
  --type postgres \
  --version 15 \
  --username admin \
  --password secure123 \
  --database production
```

### Using Aliases

```bash
# Short form
mogoly c create db -t postgres -v 14
mogoly c ls
mogoly c log db -t 50
mogoly c rm db
```

### Daemon Management

```bash
# Start daemon
mogoly d start -d

# Check status
mogoly d ps

# View logs
mogoly d log -t 200 -f
```

## Configuration File

Create `~/.mogoly/config.yaml`:

```yaml
daemon:
  socket_path: ~/.mogoly/mogoly.sock
  log_level: info

cloud:
  docker_host: unix:///var/run/docker.sock
  network_name: mogoly-network

defaults:
  postgres_version: "15"
  mysql_version: "8.0"
```

## Troubleshooting

### Daemon Won't Start

```bash
# Check if daemon is already running
mogoly daemon status

# View logs for errors
mogoly daemon logs --tail 50

# Try starting in foreground to see errors
mogoly daemon start
```

### Cannot Connect to Daemon

```bash
# Check socket path
ls -la ~/.mogoly/mogoly.sock

# Try custom socket
mogoly --socket /tmp/mogoly.sock daemon status
```

### Service Creation Fails

```bash
# Check Docker is running
docker ps

# View detailed logs
mogoly daemon logs --follow
```

## Next Steps

- Set up shell completion: `mogoly completion bash > /etc/bash_completion.d/mogoly`
- Create database backups
- Configure load balancers (coming soon)
- Integrate with CI/CD

## Architecture

```
~/.mogoly/
├── config.yaml      # Configuration file
├── daemon.json      # Daemon state
├── mogoly.pid       # Process ID
├── mogoly.sock      # Unix socket
└── mogoly.log       # Daemon logs
```

Happy managing! 🚀
