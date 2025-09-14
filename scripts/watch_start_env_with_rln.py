#!/usr/bin/env python3
import subprocess
import sys
import time
import threading
from datetime import datetime

STALL_PATTERN = "[dotenv@"
STALL_WINDOW_SEC = 120


def now():
    return datetime.now().strftime("%H:%M:%S")


def run_with_watch():
    proc = subprocess.Popen(
        ["make", "start-env-with-rln"],
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        bufsize=1,
        universal_newlines=True,
    )

    last_line_time = time.time()
    saw_stall_marker_time = None

    def reader():
        nonlocal last_line_time, saw_stall_marker_time
        for line in proc.stdout:
            sys.stdout.write(line)
            sys.stdout.flush()
            last_line_time = time.time()
            if STALL_PATTERN in line:
                # Start stall timer from first observation
                if saw_stall_marker_time is None:
                    saw_stall_marker_time = time.time()

    t = threading.Thread(target=reader, daemon=True)
    t.start()

    try:
        while proc.poll() is None:
            time.sleep(2)
            if saw_stall_marker_time is not None:
                if time.time() - saw_stall_marker_time > STALL_WINDOW_SEC:
                    print(f"[{now()}] Detected potential stall after dotenv injection. Collecting diagnostics...", flush=True)
                    diagnostics()
                    # Reset timer to avoid spamming
                    saw_stall_marker_time = time.time()
        # Ensure reader finishes
        t.join(timeout=2)
        return proc.returncode
    except KeyboardInterrupt:
        proc.terminate()
        try:
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            proc.kill()
        return 130


def sh(cmd):
    try:
        out = subprocess.check_output(cmd, shell=True, text=True, stderr=subprocess.STDOUT)
        return out.strip()
    except subprocess.CalledProcessError as e:
        return e.output.strip()


def diagnostics():
    print("\n===== DIAGNOSTICS BEGIN =====", flush=True)
    print("-- docker compose ps (rln) --", flush=True)
    print(sh("docker compose -f docker/compose-tracing-v2-rln.yml ps"))
    print("\n-- docker ps --", flush=True)
    print(sh("docker ps"))
    print("\n-- sequencer logs (last 2m) --", flush=True)
    print(sh("docker logs sequencer --since=2m | tail -n 200"))
    print("\n-- l2-node-besu logs (last 2m) --", flush=True)
    print(sh("docker logs l2-node-besu --since=2m | tail -n 200"))
    print("===== DIAGNOSTICS END =====\n", flush=True)


if __name__ == "__main__":
    rc = run_with_watch()
    sys.exit(rc)
