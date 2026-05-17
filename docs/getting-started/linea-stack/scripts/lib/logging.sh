#!/usr/bin/env sh
# Shared terminal output helpers for the Lineth quickstart scripts.
# POSIX sh: this file is sourced by both Alpine and Debian-based containers.

: "${LINETH_LOG_CONTEXT:=lineth}"

if [ -z "${NO_COLOR:-}" ] && { [ -t 1 ] || [ "${LINETH_COLOR:-auto}" = "always" ]; }; then
  LINETH_COLOR_ENABLED="1"
  LINETH_RESET="$(printf '\033[0m')"
  LINETH_BOLD="$(printf '\033[1m')"
  LINETH_DIM="$(printf '\033[2m')"
  LINETH_BLUE="$(printf '\033[38;5;45m')"
  LINETH_SHADOW="$(printf '\033[38;5;24m')"
  LINETH_BADGE="$(printf '\033[38;5;16m\033[48;5;122m')"
  LINETH_GREEN="$(printf '\033[38;5;82m')"
  LINETH_YELLOW="$(printf '\033[38;5;220m')"
  LINETH_RED="$(printf '\033[38;5;203m')"
else
  LINETH_COLOR_ENABLED=""
  LINETH_RESET=""
  LINETH_BOLD=""
  LINETH_DIM=""
  LINETH_BLUE=""
  LINETH_SHADOW=""
  LINETH_BADGE=""
  LINETH_GREEN=""
  LINETH_YELLOW=""
  LINETH_RED=""
fi

LINETH_SECTION_INDEX="${LINETH_SECTION_INDEX:-0}"

lineth_banner() {
  subtitle="${1:-quickstart (Sepolia L1)}"

  printf '\n'
  if [ -n "$LINETH_COLOR_ENABLED" ]; then
    printf '%s' "$LINETH_SHADOW"
    sed 's/^/ /' <<'EOF'
██╗     ██╗███╗   ██╗███████╗████████╗██╗  ██╗
██║     ██║████╗  ██║██╔════╝╚══██╔══╝██║  ██║
██║     ██║██╔██╗ ██║█████╗     ██║   ███████║
██║     ██║██║╚██╗██║██╔══╝     ██║   ██╔══██║
███████╗██║██║ ╚████║███████╗   ██║   ██║  ██║
╚══════╝╚═╝╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝  ╚═╝
EOF
    printf '\033[6A\r%s%s' "$LINETH_BLUE" "$LINETH_BOLD"
  else
    printf '%s' "$LINETH_BLUE"
  fi
  cat <<'EOF'
██╗     ██╗███╗   ██╗███████╗████████╗██╗  ██╗
██║     ██║████╗  ██║██╔════╝╚══██╔══╝██║  ██║
██║     ██║██╔██╗ ██║█████╗     ██║   ███████║
██║     ██║██║╚██╗██║██╔══╝     ██║   ██╔══██║
███████╗██║██║ ╚████║███████╗   ██║   ██║  ██║
╚══════╝╚═╝╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝  ╚═╝
EOF
  printf '%s' "$LINETH_RESET"
  printf '  %s lineth stack %s  %s%s%s\n' "$LINETH_BADGE" "$LINETH_RESET" "$LINETH_DIM" "$subtitle" "$LINETH_RESET"
}

lineth_section() {
  LINETH_SECTION_INDEX=$((LINETH_SECTION_INDEX + 1))
  printf '\n%s[%s] %s%s\n' "$LINETH_BLUE" "$LINETH_SECTION_INDEX" "$*" "$LINETH_RESET"
}

lineth_line() {
  color="$1"
  label="$2"
  shift 2
  printf '  %s%-5s%s %s\n' "$color" "$label" "$LINETH_RESET" "$*"
}

lineth_ok() { lineth_line "$LINETH_GREEN" "ok" "$*"; }
lineth_info() { lineth_line "$LINETH_DIM" "info" "$*"; }
lineth_warn() { lineth_line "$LINETH_YELLOW" "warn" "$*"; }
lineth_error() { lineth_line "$LINETH_RED" "error" "$*"; }

lineth_kv() {
  key="$1"
  shift
  printf '  %-32s %s\n' "$key" "$*"
}

lineth_indent() {
  sed 's/^/  /'
}

lineth_clean_prefixes() {
  sed -E 's/^\[[^]]+\][[:space:]]*(ERROR:[[:space:]]*)?//'
}

lineth_child_output() {
  lineth_clean_prefixes | lineth_indent
}

lineth_run_stream() {
  tmp="$(mktemp)"
  if "$@" > "$tmp" 2>&1; then
    status=0
  else
    status=$?
  fi
  lineth_child_output < "$tmp"
  rm -f "$tmp"
  return "$status"
}

lineth_die() {
  lineth_error "$*"
  exit 1
}
