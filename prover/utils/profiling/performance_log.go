package profiling

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"runtime"
	"runtime/metrics"
	"runtime/pprof"
	"time"

	"github.com/sirupsen/logrus"
)

// PerformanceLog captures performance metrics
type PerformanceLog struct {
	ProfilePath    string
	FlameGraphPath string

	StartTime time.Time
	StopTime  time.Time

	// CPU Usage Stats
	CpuUsageEachSeconds []float64  // CPU usage per second
	CpuUsageStats       [3]float64 // [min, avg, max] in percent

	// Memory usage per second in GiB
	MemoryInUsePerSecondGiB            []float64
	MemoryAllocatedPerSecondGiB        []float64
	MemoryGCNotDeallocatedPerSecondGiB []float64

	// Memory usage stats [min, avg, max] in GiB
	MemoryInUseStatsGiB            [3]float64
	MemoryAllocatedStatsGiB        [3]float64
	MemoryGCNotDeallocatedStatsGiB [3]float64

	// Profiling files
	cpuProfile *os.File
	memProfile *os.File
}

// performanceMonitor manages the collection of performance metrics
type performanceMonitor struct {
	log      *PerformanceLog
	stopChan chan struct{}
}

// StartPerformanceMonitor initializes and starts performance monitoring
// and samples at the specified sampleRate. If flame graph is not needed
// then empty string value may be passed
func StartPerformanceMonitor(sampleRate time.Duration, profilePath, flameGraphPath string) (*performanceMonitor, error) {

	if profilePath == "" {
		return nil, errors.New("empty profile path specified")
	}

	m := &performanceMonitor{
		log: &PerformanceLog{
			ProfilePath:    profilePath,
			FlameGraphPath: flameGraphPath,

			StartTime: time.Now(),
		},
		stopChan: make(chan struct{}),
	}

	// Ensure the flame graph path exists
	if err := os.MkdirAll(profilePath, 0755); err != nil {
		return nil, err
	}

	if flameGraphPath != "" {
		if err := os.MkdirAll(flameGraphPath, 0755); err != nil {
			return nil, err
		}
	}

	// Start CPU profiling
	cpuProfilePath := path.Join(m.log.ProfilePath, "cpu-profile.pb.gz")
	if f, err := os.Create(cpuProfilePath); err != nil {
		return nil, err
	} else {
		pprof.StartCPUProfile(f)
		m.log.cpuProfile = f
	}

	// Start memory profiling
	memProfilePath := path.Join(m.log.ProfilePath, "mem-profile.pb.gz")
	if f, err := os.Create(memProfilePath); err != nil {
		return nil, err
	} else {
		m.log.memProfile = f
	}

	// Start sampling every `sampleRate`
	go m.sample(sampleRate)
	return m, nil
}

// Stop ends performance monitoring and returns the collected metrics
func (m *performanceMonitor) Stop() (*PerformanceLog, error) {
	close(m.stopChan)
	m.log.StopTime = time.Now()

	// Stop CPU profiling
	if m.log.cpuProfile != nil {
		pprof.StopCPUProfile()
		m.log.cpuProfile.Close()

		// Generate Flame graph for CPU
		if m.log.FlameGraphPath != "" {
			if err := m.log.generateFlameGraph("cpu-profile.pb.gz", "cpu-flamegraph.svg"); err != nil {
				return nil, fmt.Errorf("failed to generate CPU flame graph: %v", err)
			}
		}
	}

	// Write memory profile
	if m.log.memProfile != nil {
		pprof.WriteHeapProfile(m.log.memProfile)
		m.log.memProfile.Close()

		// Generate Flame graph for Memory
		if m.log.FlameGraphPath != "" {
			if err := m.log.generateFlameGraph("mem-profile.pb.gz", "mem-flamegraph.svg"); err != nil {
				return nil, fmt.Errorf("failed to generate Memory flame graph: %v", err)
			}
		}
	}

	m.calculateStats()
	return m.log, nil
}

// sample collects performance metrics at regular intervals
func (m *performanceMonitor) sample(duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	var memStats runtime.MemStats
	var lastSampleTime = time.Now()
	var lastCPUTime = getCPUTime()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			newCPUTime := getCPUTime()

			// Calculate CPU usage
			timeDelta := now.Sub(lastSampleTime).Seconds()
			cpuDelta := newCPUTime - lastCPUTime
			cpuUsage := (cpuDelta / timeDelta) * 100

			m.log.CpuUsageEachSeconds = append(m.log.CpuUsageEachSeconds, cpuUsage)

			// Get memory stats
			runtime.ReadMemStats(&memStats)

			memInUseGiB := float64(memStats.HeapInuse) / (1024 * 1024 * 1024)
			memAllocatedGiB := float64(memStats.TotalAlloc) / (1024 * 1024 * 1024)
			memGCNotDeallocatedGiB := float64(memStats.HeapIdle-memStats.HeapReleased) / (1024 * 1024 * 1024)

			m.log.MemoryInUsePerSecondGiB = append(m.log.MemoryInUsePerSecondGiB, memInUseGiB)
			m.log.MemoryAllocatedPerSecondGiB = append(m.log.MemoryAllocatedPerSecondGiB, memAllocatedGiB)
			m.log.MemoryGCNotDeallocatedPerSecondGiB = append(m.log.MemoryGCNotDeallocatedPerSecondGiB, memGCNotDeallocatedGiB)

			lastSampleTime = now
			lastCPUTime = newCPUTime

		case <-m.stopChan:
			return
		}
	}
}

// getCPUTime uses "runtime/metrics" for accurate CPU time measurement
func getCPUTime() float64 {
	const metric = "/cpu/classes/total:cpu-seconds"
	sample := []metrics.Sample{{Name: metric}}
	metrics.Read(sample)

	if sample[0].Value.Kind() == metrics.KindBad {
		return 0
	}
	return sample[0].Value.Float64()
}

// calculateStats computes min, avg, and max for CPU and memory usage
func (m *performanceMonitor) calculateStats() {
	if len(m.log.CpuUsageEachSeconds) > 0 {
		min, avg, max := calculateMinAvgMax(m.log.CpuUsageEachSeconds)
		m.log.CpuUsageStats = [3]float64{min, avg, max}
	}

	if len(m.log.MemoryInUsePerSecondGiB) > 0 {
		min, avg, max := calculateMinAvgMax(m.log.MemoryInUsePerSecondGiB)
		m.log.MemoryInUseStatsGiB = [3]float64{min, avg, max}
	}

	if len(m.log.MemoryAllocatedPerSecondGiB) > 0 {
		min, avg, max := calculateMinAvgMax(m.log.MemoryAllocatedPerSecondGiB)
		m.log.MemoryAllocatedStatsGiB = [3]float64{min, avg, max}
	}

	if len(m.log.MemoryGCNotDeallocatedPerSecondGiB) > 0 {
		min, avg, max := calculateMinAvgMax(m.log.MemoryGCNotDeallocatedPerSecondGiB)
		m.log.MemoryGCNotDeallocatedStatsGiB = [3]float64{min, avg, max}
	}
}

// calculateMinAvgMax computes min, avg, and max for a slice of float64 values
func calculateMinAvgMax(values []float64) (min, avg, max float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}

	min = math.Inf(1)
	max = math.Inf(-1)
	sum := 0.0

	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}

	avg = sum / float64(len(values))
	return min, avg, max
}

// PrintPerformanceLog prints the collected performance metrics
func (pl *PerformanceLog) PrintMetrics() {
	logrus.Printf("Start Time: %v\n", pl.StartTime)
	logrus.Printf("Stop Time: %v\n", pl.StopTime)
	logrus.Printf("CPU Usage Stats: min=%.2f%%, avg=%.2f%%, max=%.2f%%\n",
		pl.CpuUsageStats[0], pl.CpuUsageStats[1], pl.CpuUsageStats[2])
	logrus.Printf("Memory In Use Stats: min=%.2f GiB, avg=%.2f GiB, max=%.2f GiB\n",
		pl.MemoryInUseStatsGiB[0], pl.MemoryInUseStatsGiB[1], pl.MemoryInUseStatsGiB[2])
	logrus.Printf("Memory Allocated Stats: min=%.2f GiB, avg=%.2f GiB, max=%.2f GiB\n",
		pl.MemoryAllocatedStatsGiB[0], pl.MemoryAllocatedStatsGiB[1], pl.MemoryAllocatedStatsGiB[2])
	logrus.Printf("Memory GC Not Deallocated Stats: min=%.2f GiB, avg=%.2f GiB, max=%.2f GiB\n",
		pl.MemoryGCNotDeallocatedStatsGiB[0], pl.MemoryGCNotDeallocatedStatsGiB[1], pl.MemoryGCNotDeallocatedStatsGiB[2])
}
