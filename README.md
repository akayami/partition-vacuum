# Partition Cleaner Daemon

A lightweight Go daemon that monitors disk usage on a specified partition and automatically deletes the oldest files from a target directory when free space falls below a configured threshold.

## Features

- Monitors partition free space percentage.
- Deletes files from a specific directory to reclaim space.
- Deletes oldest files first (based on modification time).
- Configurable check interval and minimum free space percentage.
- Safety checks: only deletes regular files, stops once target space is reclaimed.

## Building

To build the daemon, run:

```bash
go build -o partition-cleaner
```

## Usage

Run the binary with the required flags:

```bash
./partition-cleaner -partition <mount_point> -targetDir <cleanup_directory> [options]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-partition` | **Required**. Path to the partition to monitor (e.g., `/`, `/mnt/data`). | |
| `-targetDir` | **Required**. Path to the directory containing files to clean up. | |
| `-minFreePercent` | Minimum percentage of free space to maintain. | `10.0` |
| `-checkInterval` | How often to check disk usage (e.g., `1m`, `30s`, `1h`). | `1m0s` |
| `-dryRun` | Simulate deletion without actually removing files. | `false` |

## Example

Monitor the root partition `/` and maintain at least **15%** free space by deleting old files from `/var/log/myapp`, checking every **5 minutes**:

```bash
./partition-cleaner \
  -partition / \
  -targetDir /var/log/myapp \
  -minFreePercent 15 \
  -checkInterval 5m
```

## How it works

1. The daemon checks the disk usage of the filesystem containing `-partition`.
2. If the free space percentage is below `-minFreePercent`:
   - It calculates the number of bytes needed to reach the target free percentage.
   - It lists all regular files in `-targetDir`.
   - It sorts them by modification time (oldest first).
   - It deletes files one by one until enough space is reclaimed or no more eligible files remain.
3. It sleeps for `-checkInterval` and repeats.
