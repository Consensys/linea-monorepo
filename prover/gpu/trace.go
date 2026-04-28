package gpu

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// Trace emits per-phase GPU timing events when LIMITLESS_GPU_PROFILE=1.
// Cheap to call when disabled: a single atomic load and return. The events
// are written as JSONL to $LIMITLESS_GPU_PROFILE_PATH, or to a timestamped
// file under $LIMITLESS_GPU_PROFILE_DIR (default /scratch/runs).
//
// Usage:
//
//	defer gpu.TraceTime("vortex_commit", deviceID, time.Now())
//
// or:
//
//	gpu.TraceEvent("quotient", deviceID, dur, map[string]any{"domain": n})

var (
	traceEnabled atomic.Bool
	traceOnce    sync.Once
	traceMu      sync.Mutex
	traceEnc     *json.Encoder
	traceFile    *os.File
)

func initTrace() {
	traceOnce.Do(func() {
		if os.Getenv("LIMITLESS_GPU_PROFILE") != "1" {
			return
		}
		path := os.Getenv("LIMITLESS_GPU_PROFILE_PATH")
		if path == "" {
			dir := os.Getenv("LIMITLESS_GPU_PROFILE_DIR")
			if dir == "" {
				dir = "/scratch/runs"
			}
			if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec // operator-supplied profiling path
				logrus.Warnf("gpu/trace: mkdir %s: %v (tracing disabled)", dir, err)
				return
			}
			path = fmt.Sprintf("%s/gpu_profile_%s.jsonl", dir, time.Now().UTC().Format("20060102_150405"))
		}
		f, err := os.Create(path) //nolint:gosec // operator-supplied profiling path
		if err != nil {
			logrus.Warnf("gpu/trace: create %s: %v (tracing disabled)", path, err)
			return
		}
		traceFile = f
		traceEnc = json.NewEncoder(f)
		traceEnabled.Store(true)
		logrus.Infof("gpu/trace: writing events to %s", path)
	})
}

// TraceEnabled reports whether GPU phase tracing is on.
func TraceEnabled() bool {
	initTrace()
	return traceEnabled.Load()
}

// TraceTime records a single phase event whose duration is time.Since(start).
// Intended use: `defer gpu.TraceTime("phase", id, time.Now())`.
func TraceTime(phase string, deviceID int, start time.Time) {
	if !TraceEnabled() {
		return
	}
	TraceEvent(phase, deviceID, time.Since(start), nil)
}

// TraceEvent records a phase event with an explicit duration and optional
// extra fields. Extra is merged into the JSONL record at top level.
func TraceEvent(phase string, deviceID int, dur time.Duration, extra map[string]any) {
	if !TraceEnabled() {
		return
	}
	rec := map[string]any{
		"ts":     time.Now().UTC().Format(time.RFC3339Nano),
		"event":  "gpu_phase",
		"phase":  phase,
		"device": deviceID,
		"ms":     float64(dur.Microseconds()) / 1000.0,
	}
	for k, v := range extra {
		if _, exists := rec[k]; !exists {
			rec[k] = v
		}
	}
	traceMu.Lock()
	defer traceMu.Unlock()
	if traceEnc == nil {
		return
	}
	if err := traceEnc.Encode(rec); err != nil {
		logrus.Warnf("gpu/trace: write: %v", err)
	}
}

// TraceClose flushes and closes the trace file. Safe to call multiple times.
func TraceClose() {
	traceMu.Lock()
	defer traceMu.Unlock()
	if traceFile != nil {
		_ = traceFile.Sync()
		_ = traceFile.Close()
		traceFile = nil
		traceEnc = nil
		traceEnabled.Store(false)
	}
}
