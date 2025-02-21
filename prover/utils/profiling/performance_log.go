package profiling

import (
	"math"
	"os"
	"runtime"
	"runtime/metrics"
	"runtime/pprof"
	"time"

	"go.uber.org/goleak"
)

// PerformanceLog captures performance metrics
type PerformanceLog struct {
	StartTime time.Time
	StopTime  time.Time

	CpuUsageStats       [3]float64 // [min, avg, max] in percent
	CpuUsageEachSeconds []float64  // CPU usage per second

	// Memory usage per second in GiB
	MemoryInUsePerSecondGiB            []float64
	MemoryAllocatedPerSecondGiB        []float64
	MemoryGCNotDeallocatedPerSecondGiB []float64

	// [min, avg, max] in GiB
	MemoryInUseStatsGiB            [3]float64
	MemoryAllocatedStatsGiB        [3]float64
	MemoryGCNotDeallocatedStatsGiB [3]float64

	cpuProfile *os.File
	memProfile *os.File
}

// performanceMonitor manages the collection of performance metrics
type performanceMonitor struct {
	log       *PerformanceLog
	stopChan  chan struct{}
	startTime time.Time
}

// StartPerformanceMonitor initializes and starts performance monitoring
func StartPerformanceMonitor() *performanceMonitor {
	m := &performanceMonitor{
		log: &PerformanceLog{
			StartTime: time.Now(),
		},
		stopChan:  make(chan struct{}),
		startTime: time.Now(),
	}

	// Start CPU and memory profiling
	if f, err := os.Create("cpu-profile.pb.gz"); err == nil {
		pprof.StartCPUProfile(f)
		m.log.cpuProfile = f
	}
	if f, err := os.Create("mem-profile.pb.gz"); err == nil {
		m.log.memProfile = f
	}

	// Start goroutine leak detection
	goleak.VerifyNone(nil)

	// Start sampling every second
	go m.sample()
	return m
}

// Stop ends performance monitoring and returns the collected metrics
func (m *performanceMonitor) Stop() *PerformanceLog {
	close(m.stopChan)
	m.log.StopTime = time.Now()

	// Stop CPU profiling
	if m.log.cpuProfile != nil {
		pprof.StopCPUProfile()
		m.log.cpuProfile.Close()
	}

	// Write memory profile
	if m.log.memProfile != nil {
		pprof.WriteHeapProfile(m.log.memProfile)
		m.log.memProfile.Close()
	}

	m.calculateStats()
	return m.log
}

// sample collects performance metrics at regular intervals
func (m *performanceMonitor) sample() {
	ticker := time.NewTicker(time.Second)
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
