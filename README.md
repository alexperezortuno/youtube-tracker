# YouTube Tracker

YouTube Tracker is a Go application designed to **detect live YouTube channels**, **collect real-time metrics**, and **persist data in PostgreSQL/TimescaleDB**, using **Redis** as a cache/temporary state store.

It is built to run continuous discovery and metric collection cycles, and it also includes SQL scripts for data analysis and visualization.

---

## Features

- Automatic detection of active YouTube livestreams
- Periodic collection of stream metrics
- Persistence of streams and metrics in a database
- Redis-based storage for active streams
- PostgreSQL + TimescaleDB for time-series data
- Support for multiple channels
- SQL scripts for audience, engagement, and trend analysis
- Local infrastructure with Docker Compose
- Environment-based configuration

---

## Technologies

- **Go 1.26**
- **Redis**
- **PostgreSQL / TimescaleDB**
- **Docker / Docker Compose**
- **YouTube Data API v3**
- **Grafana** for optional visualization

---

## Project Structure

```text
youtube-tracker/
    ├── bin/                    # Compiled binaries
    ├── cmd/                    # Command line applications
    │   ├── daily.go            # Daily metrics collector
    │   ├── discover.go         # Livestream discovery service
    │   ├── metrics.go          # Real-time metrics collector
    │   └── root.go             # Root command and CLI setup
    ├── internal/               # Internal packages
    │   ├── cache/              # Redis cache implementation
    │   ├── collector/          # Metrics collection logic
    │   ├── config/             # Configuration management
    │   ├── discovery/          # Livestream discovery logic
    │   ├── lifecycle/          # Application lifecycle management
    │   ├── models/             # Data models and structures
    │   ├── source/             # YouTube API data sources
    │   └── storage/            # Database storage implementation
    ├── scripts/                # Utility scripts
    │   ├── backup-logs.sh      # Log backup script
    │   ├── database/           # Database migration scripts
    │   ├── install.sh          # Installation script
    │   ├── manager.sh          # Process management script
    │   └── viewers_x_minute.sql # SQL analysis script
    ├── .env                    # Environment configuration
    ├── .gitignore              # Git ignore rules
    ├── channel_names.txt       # List of channel names to track
    ├── channels.txt            # List of channel IDs to track
    ├── docker-compose.yml      # Docker Compose configuration
    ├── Dockerfile              # Docker configuration
    ├── go.mod                  # Go module definition
    ├── go.sum                  # Go module checksums
    ├── main.go                 # Application entry point
    ├── Makefile                # Build automation
    └── README.md               # Project documentation
```

---

## Prerequisites

- Go 1.26 or higher
- Docker and Docker Compose
- A YouTube Data API v3 key
- Access to PostgreSQL/TimescaleDB and Redis, or the ability to run them with Docker

---

## Configuration

The application uses a `.env` file for configuration.

### Environment Variables

```bash
YOUTUBE_API_KEY=your_youtube_api_key 
POSTGRES_URL=postgres://user:pass@localhost:5432/metrics?sslmode=disable 
REDIS_ADDR=localhost:6379 
CHANNEL_IDS=channel_id_1,channel_id_2,channel_id_3
```

> Note: the exact variable names may depend on your current `config.Load()` implementation.

```bash
git clone https://github.com/alexperezortuno/youtube-tracker.git cd youtube-tracker
```

### 2. Configure environment variables

Create a `.env` file in the project root:

```bash
YOUTUBE_API_KEY=your_api_key 
POSTGRES_URL=postgres://user:pass@localhost:5432/metrics?sslmode=disable 
REDIS_ADDR=localhost:6379 
CHANNEL_IDS=channel1,channel2
```

### 3. Start the infrastructure with Docker

```bash
docker-compose up -d
```

This starts:

- Redis on `localhost:6379`
- PostgreSQL/TimescaleDB on `localhost:5432`
- Grafana on `localhost:3000`

### 4. Initialize the database

Run the required migration or initialization scripts.

If you use the `Makefile` target:

```bash
make db-init
```

---

## Usage

### Run locally

#### Using the Manager Script

The `./scripts/manager.sh` script provides a convenient way to manage the discover, collector, and metrics processes. Each process runs in the background and logs to the `logs/` directory with their PID stored in the `pids/` directory.

##### Available Commands

```bash
# Start all processes
./scripts/manager.sh start all

# Start individual processes
./scripts/manager.sh start discover
./scripts/manager.sh start collector
./scripts/manager.sh start metrics

# Stop all processes
./scripts/manager.sh stop all

# Stop individual processes
./scripts/manager.sh stop discover
./scripts/manager.sh stop collector
./scripts/manager.sh stop metrics

# Check status of all processes
./scripts/manager.sh status
```

##### Process Details

- **discover**: Detects new livestreams with 30-second intervals
  - Command: `./youtube-tracker discover --interval 30 --extractor --log-level=debug`
- **collector**: Collects daily metrics with 3-minute intervals
  - Command: `./youtube-tracker daily --interval 3 --log-level=debug`
- **metrics**: Collects stream metrics with 30-second intervals
  - Command: `./youtube-tracker metrics --interval 30 --log-level=debug`

##### Log Files

Each process logs to:
- `logs/discover.log`
- `logs/collector.log`
- `logs/metrics.log`

##### PID Files

Each process stores its PID in:
- `pids/discover.pid`
- `pids/collector.pid`
- `pids/metrics.pid`

#### Backup Logs Script

The `./scripts/backup-logs.sh` script provides functionality to archive the logs directory and optionally send it to a remote server using rsync.

##### Available Commands

```bash
# Create a backup of logs directory
./scripts/backup-logs.sh

# Create a backup and send it to a remote server using rsync
./scripts/backup-logs.sh --rsync

# Create a backup and clean the logs directory after creating the backup
./scripts/backup-logs.sh --clean

# Create a backup, send it to a remote server, and clean the logs directory
./scripts/backup-logs.sh --rsync --clean
```

##### Script Details

- Archives the `logs/` directory into a timestamped tar.gz file in the `backups/` directory
- Format: `backups/logs_{hostname}_{timestamp}.tar.gz`
- Optional rsync functionality to send backups to a remote server
- Optional log cleaning functionality to truncate log files after backup

##### Configuration

The script uses the following default settings for rsync:
- Remote user: `user`
- Remote host: `10.0.0.10`
- Remote path: `/data/youtube-tracker/logs`

These settings would need to be modified in the script for your specific environment.

---

## Installation

### 1. Clone the repository








---

## Installation

### 1. Clone the repository


## Get Channel ID

```bash
source .env && curl \
  "https://youtube.googleapis.com/youtube/v3/channels?part=id&forUsername=${CHANNEL_NAME}&key=${YOUTUBE_API_KEY}" \
  --header "Accept: application/json" \
  --compressed
```

---

## Get Live Streams

```bash
(source .env && curl \
  "https://www.googleapis.com/youtube/v3/search?part=snippet&q=${VIDEO_ID}&type=channel&key=${YOUTUBE_API_KEY}" \
  --header "Accept: application/json" \
  --compressed)
```

---

## Get Live Streams

```bash
(source .env && curl \
  "https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=${CHANNEL_ID}&eventType=live&type=video&key=${YOUTUBE_API_KEY}" \
  --header "Accept: application/json" \
  --compressed)
```

---

## Discover Streams in channels

```bash
./bin/app discover --interval 30 --log-level=debug
```

## Get stream metrics

```bash
./bin/app metrics --interval 60 --log-level=debug
```

## Get daily metrics

```bash
./bin/app daily --interval 12 --log-level=debug
```

### Install via Curl

On linux:

```bash
curl -sSL https://raw.githubusercontent.com/alexperezortuno/youtube-tracker/master/scripts/install.sh | bash
```

