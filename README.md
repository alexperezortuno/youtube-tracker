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
    ├── cmd/ 
    │ └── main.go 
    ├── internal/ 
    │ ├── cache/ 
    │ ├── collector/ 
    │ ├── config/ 
    │ ├── discovery/ 
    │ ├── lifecycle/ 
    │ ├── models/ 
    │ ├── source/ 
    │ └── storage/ 
    ├── scripts/ 
    │ ├── migrations/ 
    │ └── *.sql 
    ├── docker-compose.yml 
    ├── Makefile 
    ├── go.mod 
    └── README.md
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


---

## Installation

### 1. Clone the repository








---

## Installation

### 1. Clone the repository


## Get Channel ID

```bash
(source .env && curl \
  "https://youtube.googleapis.com/youtube/v3/channels?part=id&forUsername=${CHANNEL_NAME}&key=${YOUTUBE_API_KEY}" \
  --header "Accept: application/json" \
  --compressed)"
```

```bash
(source .env && echo "curl https://www.googleapis.com/youtube/v3/channels?part=id&forUsername=${CHANNEL_NAME}&key=${YOUTUBE_API_KEY} --header 'Accept: application/json' --compressed")
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

