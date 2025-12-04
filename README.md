# Partition Vacuum Daemon

A lightweight Go daemon that monitors disk usage on a specified partition and automatically deletes the oldest files from a target directory when free space falls below a configured threshold.

## Features

- Monitors partition free space percentage.
- Supports minimum free space as a percentage or absolute value (e.g., 10GB).
- Deletes files from a specific directory to reclaim space.
- Deletes oldest files first (based on modification time).
- Configurable check interval and minimum free space.
- Safety checks: only deletes regular files, stops once target space is reclaimed.

## Building

To build the daemon, run:

```bash
go build -o partition-vacuum
```

## Usage

Run the binary with the required flags:

```bash
./partition-vacuum -partition <mount_point> -targetDir <cleanup_directory> [options]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-partition` | **Required**. Path to the partition to monitor (e.g., `/`, `/mnt/data`). | |
| `-targetDir` | **Required**. Path to the directory containing files to clean up. | |
| `-minFreePercent` | Minimum percentage of free space to maintain. | `10.0` |
| `-minFreeBytes` | Minimum absolute free space to maintain (e.g., `10GB`, `500MB`). | |
| `-checkInterval` | How often to check disk usage (e.g., `1m`, `30s`, `1h`). | `1m0s` |
| `-dryRun` | Simulate deletion without actually removing files. | `false` |

> **Note**: When both `-minFreePercent` and `-minFreeBytes` are specified, cleanup triggers if **either** threshold is breached, and the target free space is the **larger** of the two values.

## Example

Monitor the root partition `/` and maintain at least **15%** free space by deleting old files from `/var/log/myapp`, checking every **5 minutes**:

```bash
./partition-vacuum \
  -partition / \
  -targetDir /var/log/myapp \
  -minFreePercent 15 \
  -checkInterval 5m
```

## How it works

1. The daemon checks the disk usage of the filesystem containing `-partition`.
2. If the free space is below the configured threshold (percentage or absolute):
   - It calculates the number of bytes needed to reach the target.
   - It lists all regular files in `-targetDir`.
   - It sorts them by modification time (oldest first).
   - It deletes files one by one until enough space is reclaimed or no more eligible files remain.
3. It sleeps for `-checkInterval` and repeats.

## Configuration File

For more complex setups with multiple directories, use a TOML configuration file:

```toml
[global]
check_interval = "5m"
min_free_percent = 15.0
min_free_bytes = "10GB"   # Absolute minimum
human_readable = true
dry_run = false

[[location]]
target_dirs = ["/var/log/myapp"]
min_free_bytes = "5GB"    # Override for this location
```

Supported byte size formats: `B`, `KB`, `MB`, `GB`, `TB`, `PB` (e.g., `500MB`, `1.5GB`, `10G`).
