#!/usr/bin/env bash
set -euo pipefail

# Usage: GL_WORKERS=4 LPP_WORKERS=3  ./cmd/controller/controller/sh/run_limitless.sh > ./cmd/controller/controller/logs/meta.log 2 >&1
# Kill:

# kill -SIGUSR1 $(pgrep -f controller-bootstrapper)
# kill -SIGTERM $(pgrep -f controller-bootstrapper)

# kill -SIGUSR1 $(pgrep -f controller-worker-gl-1)
# kill -SIGTERM $(pgrep -f controller-worker-gl-1)

# kill -SIGUSR1 $(pgrep -f controller-worker-lpp-1)
# kill -SIGTERM $(pgrep -f controller-worker-lpp-1)

# Debug/Diagnostics
# ps aux | grep controller | grep -v grep

# ====== ABSOLUTE PATHS ======
BASE_DIR="/home/ubuntu/linea-monorepo/prover"
BIN="$BASE_DIR/bin/controller"
CONFIG_DIR="$BASE_DIR/config/test_files/test_config"
LOG_DIR="$BASE_DIR/cmd/controller/controller/logs/limitless"

# ====== CONFIG FILES ======
BOOTSTRAP_CFG="$CONFIG_DIR/config-mainnet-bootstrapper.toml"
CONGLO_CFG="$CONFIG_DIR/config-mainnet-conglomerator.toml"
GL_CFG="$CONFIG_DIR/config-mainnet-gl.toml"
LPP_CFG="$CONFIG_DIR/config-mainnet-lpp.toml"

# ====== TUNABLE WORKER COUNTS ======
GL_WORKERS="${GL_WORKERS:-4}"
LPP_WORKERS="${LPP_WORKERS:-3}"

# ====== GLOBAL STATE ======
PIDS=()

# ====== HELPERS ======
start_instance() {
  local cfg="$1"
  local local_id="$2"
  local logf="$LOG_DIR/controller-${local_id}.log"

  mkdir -p "$LOG_DIR"
  echo "→ Starting controller (local-id=${local_id}) with config ${cfg}"
  echo "  Log file: ${logf}"

  # Change process name with exec -a for easier identifiability
  bash -c "exec -a controller-${local_id} \"$BIN\" --config \"$cfg\" --local-id \"$local_id\"" >"$logf" 2>&1 &
  local pid=$!
  PIDS+=("$pid")
  echo "  PID: $pid"
}


shutdown_all() {
  echo ""
  echo "⚙️  Shutting down controllers..."

  # Kill tracked controller PIDs
  for pid in "${PIDS[@]:-}"; do
    if kill -0 "$pid" 2>/dev/null; then
      echo "  Sending SIGTERM to controller pid $pid"
      kill -TERM "$pid" 2>/dev/null || true
    fi
  done

  sleep 2
  for pid in "${PIDS[@]:-}"; do
    if kill -0 "$pid" 2>/dev/null; then
      echo "  pid $pid still alive, sending SIGKILL"
      kill -KILL "$pid" 2>/dev/null || true
    fi
  done

  # Kill any remaining 'prover prove' processes
  echo "⚙️  Killing any remaining 'prover prove' processes..."
  pkill -f "prover prove" 2>/dev/null || true

  # Optional: clean up /tmp/exec-limitless
  # if [ -d "/tmp/exec-limitless" ]; then
  #   echo "⚙️  Cleaning /tmp/exec-limitless..."
  #   rm -rf /tmp/exec-limitless/* || true
  # fi

  echo "✅ Shutdown complete."
}

trap 'shutdown_all; exit 0' INT TERM

# ====== CHECKS ======
if [ ! -x "$BIN" ]; then
  echo "ERROR: controller binary not executable: $BIN" >&2
  exit 1
fi
for cfg in "$BOOTSTRAP_CFG" "$CONGLO_CFG" "$GL_CFG" "$LPP_CFG"; do
  if [ ! -f "$cfg" ]; then
    echo "ERROR: missing config file: $cfg" >&2
    exit 1
  fi
done

mkdir -p "$LOG_DIR"

# ====== START CONTROLLERS ======
echo "🚀 Launching controllers from $BASE_DIR"
echo "  GL_WORKERS=$GL_WORKERS"
echo "  LPP_WORKERS=$LPP_WORKERS"
echo

# Bootstrapper
start_instance "$BOOTSTRAP_CFG" "bootstrapper"

# Conglomerator
start_instance "$CONGLO_CFG" "conglomerator"

# GL workers
for ((i=1; i<=GL_WORKERS; i++)); do
  start_instance "$GL_CFG" "worker-gl-$i"
done

# LPP workers
for ((i=1; i<=LPP_WORKERS; i++)); do
  start_instance "$LPP_CFG" "worker-lpp-$i"
done

echo
echo "✅ All controllers started successfully!"
echo "  PIDs: ${PIDS[*]}"
echo "  Logs are in: $LOG_DIR"
echo
echo "Press CTRL+C to stop all controllers."
echo

# ====== WAIT UNTIL ALL FINISH ======
for pid in "${PIDS[@]}"; do
  wait "$pid" || true
done

echo "🛑 All child controllers exited."
