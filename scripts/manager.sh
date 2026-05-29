#!/bin/bash

# Manager script to start/stop the discover, collector and metrics processes
# Each process will run in background and log to a file in the logs/ directory
# The PID of each process will be stored in the pids/ directory
# Usage:
#   ./manager.sh start all
#   ./manager.sh start discover
#   ./manager.sh start collector
#   ./manager.sh start metrics
#   ./manager.sh stop all
#   ./manager.sh stop discover
#   ./manager.sh stop collector
#   ./manager.sh stop metrics
#   ./manager.sh status

set -e

APP="./youtube-tracker"

LOG_DIR="logs"
PID_DIR="pids"

mkdir -p "$LOG_DIR"
mkdir -p "$PID_DIR"

start_process() {
    local name=$1
    shift

    local cmd="$APP $*"

    # Verificar si ya existe un proceso ejecutándose
    if pgrep -f "$cmd" > /dev/null; then
        echo "[WARN] $name It's already running"
        return
    fi

    echo "[INFO] Starting $name..."

    nohup $cmd \
        > "$LOG_DIR/$name.log" 2>&1 &

    local pid=$!

    echo $pid > "$PID_DIR/$name.pid"

    echo "[OK] $name started with PID $pid"
}

stop_process() {
    local name=$1

    local pidfile="$PID_DIR/$name.pid"

    if [ ! -f "$pidfile" ]; then
        echo "[WARN] There is no PID for $name"
        return
    fi

    local pid
    pid=$(cat "$pidfile")

    if ps -p "$pid" > /dev/null 2>&1; then
        echo "[INFO] Stopping $name PID $pid"
        kill "$pid"
    else
        echo "[WARN] PID $pid does not exist"
    fi

    rm -f "$pidfile"
}

status_process() {
    local name=$1

    local pidfile="$PID_DIR/$name.pid"

    if [ ! -f "$pidfile" ]; then
        echo "[STOPPED] $name"
        return
    fi

    local pid
    pid=$(cat "$pidfile")

    if ps -p "$pid" > /dev/null 2>&1; then
        echo "[RUNNING] $name PID $pid"
    else
        echo "[STOPPED] $name (pid invalid)"
    fi
}

start_discover() {
    start_process discover \
        discover --interval 30 --extractor --log-level=debug
}

start_collector() {
    start_process collector \
        daily --interval 3 --log-level=debug
}

start_metrics() {
    start_process metrics \
        metrics --interval 30 --log-level=debug
}

usage() {
    echo ""
    echo "Usage:"
    echo "  ./manager.sh start all"
    echo "  ./manager.sh start discover"
    echo "  ./manager.sh start collector"
    echo "  ./manager.sh start metrics"
    echo ""
    echo "  ./manager.sh stop all"
    echo "  ./manager.sh stop discover"
    echo "  ./manager.sh stop collector"
    echo "  ./manager.sh stop metrics"
    echo ""
    echo "  ./manager.sh status"
    echo ""
}

ACTION=$1
TARGET=$2

case "$ACTION" in
    start)
        case "$TARGET" in
            all)
                start_discover
                start_collector
                start_metrics
                ;;
            discover)
                start_discover
                ;;
            collector)
                start_collector
                ;;
            metrics)
                start_metrics
                ;;
            *)
                usage
                ;;
        esac
        ;;

    stop)
        case "$TARGET" in
            all)
                stop_process discover
                stop_process collector
                stop_process metrics
                ;;
            discover)
                stop_process discover
                ;;
            collector)
                stop_process collector
                ;;
            metrics)
                stop_process metrics
                ;;
            *)
                usage
                ;;
        esac
        ;;

    status)
        status_process discover
        status_process collector
        status_process metrics
        ;;

    *)
        usage
        ;;
esac
