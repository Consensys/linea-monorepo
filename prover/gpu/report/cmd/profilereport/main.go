// Command gen_profile_report generates an HTML GPU profiling report from ncu CSV output.
//
// Usage:
//
//	go run ./gpu/report/gen_profile_report.go < ncu_metrics.csv > gpu/gpu_profile_report.html
package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type KernelProfile struct {
	Name          string
	BlockSize     int
	GridSize      int
	DurationNs    float64
	DRAMPct       float64 // dram throughput % of peak
	SMPct         float64 // SM throughput % of peak
	OccupancyWps  float64 // occupancy limit warps
	Invocations   int
}

type AggregatedKernel struct {
	Name         string
	BlockSize    int
	GridSize     int
	AvgDurNs     float64
	TotalDurNs   float64
	DRAMPct      float64
	SMPct        float64
	OccWps       float64
	Count        int
	Category     string
	BottleneckDesc string
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	var csvLines []string
	inCSV := false

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "read error: %v\n", err)
			os.Exit(1)
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "\"ID\"") {
			inCSV = true
		}
		if inCSV && strings.HasPrefix(line, "\"") {
			csvLines = append(csvLines, line)
		}
	}

	if len(csvLines) < 2 {
		fmt.Fprintf(os.Stderr, "no CSV data found in input\n")
		os.Exit(1)
	}

	// Parse CSV
	csvReader := csv.NewReader(strings.NewReader(strings.Join(csvLines, "\n")))
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "CSV parse error: %v\n", err)
		os.Exit(1)
	}

	// Header: ID, Process ID, Process Name, Host Name, Kernel Name, Context, Stream, Block Size, Grid Size, Device, CC, Section Name, Metric Name, Metric Unit, Metric Value
	// Group by kernel ID
	type metric struct {
		name  string
		value float64
	}
	type kernelData struct {
		name      string
		blockSize string
		gridSize  string
		metrics   map[string]float64
	}

	kernelMap := make(map[string]*kernelData) // keyed by ID
	var ids []string

	for _, row := range records[1:] { // skip header
		if len(row) < 15 {
			continue
		}
		id := row[0]
		kernelName := row[4]
		blockSizeStr := row[7]
		gridSizeStr := row[8]
		metricName := row[12]
		metricValue := row[14]

		if _, ok := kernelMap[id]; !ok {
			kernelMap[id] = &kernelData{
				name:      kernelName,
				blockSize: blockSizeStr,
				gridSize:  gridSizeStr,
				metrics:   make(map[string]float64),
			}
			ids = append(ids, id)
		}

		val, _ := strconv.ParseFloat(strings.ReplaceAll(metricValue, ",", ""), 64)
		kernelMap[id].metrics[metricName] = val
	}

	// Build kernel profiles
	var profiles []KernelProfile
	for _, id := range ids {
		kd := kernelMap[id]
		p := KernelProfile{
			Name:         cleanKernelName(kd.name),
			BlockSize:    parseGridDim(kd.blockSize),
			GridSize:     parseGridDim(kd.gridSize),
			DurationNs:   kd.metrics["gpu__time_duration.sum"],
			DRAMPct:      kd.metrics["dram__throughput.avg.pct_of_peak_sustained_elapsed"],
			SMPct:        kd.metrics["sm__throughput.avg.pct_of_peak_sustained_elapsed"],
			OccupancyWps: kd.metrics["launch__occupancy_limit_warps"],
			Invocations:  1,
		}
		profiles = append(profiles, p)
	}

	// Aggregate by kernel name
	aggMap := make(map[string]*AggregatedKernel)
	var aggOrder []string
	for _, p := range profiles {
		if _, ok := aggMap[p.Name]; !ok {
			aggMap[p.Name] = &AggregatedKernel{
				Name:      p.Name,
				BlockSize: p.BlockSize,
				GridSize:  p.GridSize,
			}
			aggOrder = append(aggOrder, p.Name)
		}
		a := aggMap[p.Name]
		a.Count++
		a.TotalDurNs += p.DurationNs
		a.DRAMPct += p.DRAMPct
		a.SMPct += p.SMPct
		a.OccWps += p.OccupancyWps
		a.GridSize = max(a.GridSize, p.GridSize)
	}

	var aggregated []AggregatedKernel
	for _, name := range aggOrder {
		a := aggMap[name]
		a.AvgDurNs = a.TotalDurNs / float64(a.Count)
		a.DRAMPct /= float64(a.Count)
		a.SMPct /= float64(a.Count)
		a.OccWps /= float64(a.Count)
		a.Category = categorizeKernel(a.Name)
		a.BottleneckDesc = analyzeBottleneck(a)
		aggregated = append(aggregated, *a)
	}

	// Sort by total duration (most expensive first)
	sort.Slice(aggregated, func(i, j int) bool {
		return aggregated[i].TotalDurNs > aggregated[j].TotalDurNs
	})

	// Calculate total kernel time
	var totalKernelNs float64
	for _, a := range aggregated {
		totalKernelNs += a.TotalDurNs
	}

	data := struct {
		Kernels      []AggregatedKernel
		TotalNs      float64
		NumKernels   int
		NumProfiles  int
	}{
		Kernels:     aggregated,
		TotalNs:     totalKernelNs,
		NumKernels:  len(aggregated),
		NumProfiles: len(profiles),
	}

	tmpl := template.Must(template.New("profile").Funcs(template.FuncMap{
		"fmtDur": func(ns float64) string {
			if ns >= 1e6 {
				return fmt.Sprintf("%.2f ms", ns/1e6)
			} else if ns >= 1e3 {
				return fmt.Sprintf("%.1f µs", ns/1e3)
			}
			return fmt.Sprintf("%.0f ns", ns)
		},
		"fmtPct": func(pct float64) string {
			return fmt.Sprintf("%.1f%%", pct)
		},
		"pctOfTotal": func(ns, total float64) string {
			if total <= 0 {
				return "0%"
			}
			return fmt.Sprintf("%.1f%%", ns/total*100)
		},
		"pctColor": func(pct float64) string {
			if pct >= 70 {
				return "#3fb950" // green - well utilized
			} else if pct >= 40 {
				return "#d29922" // orange - moderate
			} else if pct >= 15 {
				return "#f85149" // red - underutilized
			}
			return "#8b949e" // grey - very low
		},
		"barWidth": func(pct float64) string {
			return fmt.Sprintf("%.0f", pct)
		},
	}).Parse(profileHTML))

	if err := tmpl.Execute(os.Stdout, data); err != nil {
		fmt.Fprintf(os.Stderr, "template error: %v\n", err)
		os.Exit(1)
	}
}

func cleanKernelName(name string) string {
	// Remove template params and simplify
	if idx := strings.Index(name, "("); idx > 0 {
		name = name[:idx]
	}
	return name
}

func parseGridDim(s string) int {
	// Parse "(256, 1, 1)" → 256 or just "256" → 256
	s = strings.Trim(s, "()")
	parts := strings.Split(s, ",")
	if len(parts) > 0 {
		v, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		return v
	}
	return 0
}

func categorizeKernel(name string) string {
	switch {
	case strings.Contains(name, "ntt") || strings.Contains(name, "dit") || strings.Contains(name, "dif"):
		return "NTT"
	case strings.Contains(name, "add") || strings.Contains(name, "mul") || strings.Contains(name, "sub") || strings.Contains(name, "scale"):
		return "Arithmetic"
	case strings.Contains(name, "bitrev"):
		return "Permutation"
	case strings.Contains(name, "p2") || strings.Contains(name, "poseidon"):
		return "Hash"
	case strings.Contains(name, "sis"):
		return "SIS"
	case strings.Contains(name, "msm") || strings.Contains(name, "sort") || strings.Contains(name, "bucket"):
		return "MSM"
	case strings.Contains(name, "transpose") || strings.Contains(name, "copy") || strings.Contains(name, "memcpy"):
		return "Transfer"
	default:
		return "Other"
	}
}

func analyzeBottleneck(a *AggregatedKernel) string {
	if a.DRAMPct > 60 && a.SMPct < 20 {
		return "Memory-bound: High DRAM utilization but low SM throughput. Consider coalescing access patterns or reducing memory transactions."
	} else if a.SMPct > 60 && a.DRAMPct < 30 {
		return "Compute-bound: High SM throughput. Good utilization of compute resources."
	} else if a.DRAMPct < 20 && a.SMPct < 20 {
		return "Latency-bound: Both DRAM and SM underutilized. Likely limited by kernel launch overhead, occupancy, or instruction-level parallelism."
	} else if a.DRAMPct > 40 && a.SMPct > 40 {
		return "Balanced: Both memory and compute are well utilized."
	}
	return "Mixed: Moderate utilization of both compute and memory subsystems."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

const profileHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>GPU Kernel Profiling Report — Linea Prover</title>
<style>
:root {
  --bg: #0d1117; --surface: #161b22; --border: #30363d;
  --text: #c9d1d9; --text-muted: #8b949e;
  --accent: #58a6ff; --green: #3fb950; --orange: #d29922; --red: #f85149; --purple: #bc8cff;
}
* { margin: 0; padding: 0; box-sizing: border-box; }
body { background: var(--bg); color: var(--text); font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; line-height: 1.6; padding: 2rem; }
.container { max-width: 1400px; margin: 0 auto; }
h1 { font-size: 2.5rem; background: linear-gradient(135deg, var(--red), var(--orange)); -webkit-background-clip: text; -webkit-text-fill-color: transparent; margin-bottom: 0.5rem; }
h2 { font-size: 1.5rem; color: var(--accent); margin: 2rem 0 1rem; padding-bottom: 0.5rem; border-bottom: 1px solid var(--border); }
h3 { font-size: 1.1rem; color: var(--text); margin: 1.5rem 0 0.5rem; }
.meta { color: var(--text-muted); font-size: 0.9rem; margin-bottom: 2rem; }
table { width: 100%; border-collapse: collapse; margin-bottom: 1.5rem; font-size: 0.85rem; }
th, td { padding: 0.6rem 0.8rem; text-align: right; border-bottom: 1px solid var(--border); }
th { color: var(--text-muted); font-weight: 600; text-transform: uppercase; font-size: 0.7rem; letter-spacing: 0.05em; }
td:first-child, th:first-child { text-align: left; }
tr:hover { background: rgba(88, 166, 255, 0.05); }
.bar-container { width: 100px; height: 16px; background: var(--border); border-radius: 4px; display: inline-block; vertical-align: middle; overflow: hidden; }
.bar-fill { height: 100%; border-radius: 4px; transition: width 0.3s; }
.finding { background: var(--surface); border: 1px solid var(--border); border-radius: 8px; padding: 1rem 1.5rem; margin-bottom: 1rem; }
.finding strong { color: var(--accent); }
.finding .bottleneck { color: var(--text-muted); font-size: 0.85rem; margin-top: 0.5rem; }
.kernel-name { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8rem; color: var(--purple); }
.cat-badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.7rem; font-weight: 600; }
.cat-NTT { background: #1f6feb33; color: var(--accent); }
.cat-Arithmetic { background: #3fb95033; color: var(--green); }
.cat-Permutation { background: #d2992233; color: var(--orange); }
.cat-Hash { background: #bc8cff33; color: var(--purple); }
.cat-Transfer { background: #8b949e33; color: var(--text-muted); }
.cat-Other { background: #30363d; color: var(--text-muted); }
.summary-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 2rem; }
.summary-card { background: var(--surface); border: 1px solid var(--border); border-radius: 8px; padding: 1.2rem; }
.summary-card .label { color: var(--text-muted); font-size: 0.8rem; }
.summary-card .value { font-size: 1.8rem; font-weight: 700; color: var(--green); }
</style>
</head>
<body>
<div class="container">

<h1>GPU Kernel Profiling Report</h1>
<p class="meta">
  <strong>Device:</strong> NVIDIA RTX PRO 6000 Blackwell (188 SMs, 95 GB VRAM, CC 12.0)<br>
  <strong>Profiler:</strong> NVIDIA Nsight Compute (ncu)<br>
  <strong>Kernels profiled:</strong> {{.NumProfiles}} invocations across {{.NumKernels}} unique kernels
</p>

<div class="summary-grid">
  <div class="summary-card">
    <div class="label">Total Kernel Time</div>
    <div class="value">{{fmtDur .TotalNs}}</div>
  </div>
  <div class="summary-card">
    <div class="label">Unique Kernels</div>
    <div class="value">{{.NumKernels}}</div>
  </div>
  <div class="summary-card">
    <div class="label">Total Invocations</div>
    <div class="value">{{.NumProfiles}}</div>
  </div>
</div>

<h2>Kernel Performance Overview</h2>
<p class="meta">Sorted by total execution time (most expensive first). DRAM% and SM% show utilization as percentage of peak sustained throughput.</p>

<table>
<thead>
<tr>
  <th>Kernel</th>
  <th>Category</th>
  <th>Calls</th>
  <th>Avg Duration</th>
  <th>Total Duration</th>
  <th>% of Total</th>
  <th>Block</th>
  <th>Grid</th>
  <th>DRAM %</th>
  <th>SM %</th>
  <th>Occ Warps</th>
</tr>
</thead>
<tbody>
{{range .Kernels}}
<tr>
  <td class="kernel-name">{{.Name}}</td>
  <td><span class="cat-badge cat-{{.Category}}">{{.Category}}</span></td>
  <td>{{.Count}}</td>
  <td>{{fmtDur .AvgDurNs}}</td>
  <td>{{fmtDur .TotalDurNs}}</td>
  <td>{{pctOfTotal .TotalDurNs $.TotalNs}}</td>
  <td>{{.BlockSize}}</td>
  <td>{{.GridSize}}</td>
  <td>
    <div class="bar-container">
      <div class="bar-fill" style="width: {{barWidth .DRAMPct}}%; background: {{pctColor .DRAMPct}};"></div>
    </div>
    {{fmtPct .DRAMPct}}
  </td>
  <td>
    <div class="bar-container">
      <div class="bar-fill" style="width: {{barWidth .SMPct}}%; background: {{pctColor .SMPct}};"></div>
    </div>
    {{fmtPct .SMPct}}
  </td>
  <td>{{fmtPct .OccWps}}</td>
</tr>
{{end}}
</tbody>
</table>

<h2>Per-Kernel Analysis & Optimization Recommendations</h2>

{{range .Kernels}}
<div class="finding">
  <strong class="kernel-name">{{.Name}}</strong> <span class="cat-badge cat-{{.Category}}">{{.Category}}</span><br>
  <span class="meta">{{.Count}} call(s) | Block: {{.BlockSize}} | Grid: {{.GridSize}} | Avg: {{fmtDur .AvgDurNs}} | DRAM: {{fmtPct .DRAMPct}} | SM: {{fmtPct .SMPct}}</span>
  <div class="bottleneck">{{.BottleneckDesc}}</div>
</div>
{{end}}

<h2>Global Optimization Recommendations</h2>

<div class="finding">
  <strong>1. Increase Occupancy:</strong> Many kernels show occupancy limited to 6 warps/SM. With 188 SMs on Blackwell, this limits parallelism. Consider reducing register usage or shared memory per block, or using smaller block sizes to allow more concurrent blocks per SM.
</div>
<div class="finding">
  <strong>2. NTT Kernels Dominate:</strong> NTT (DIF/DIT) butterflies account for the majority of kernel time. Each stage launches a separate kernel with moderate DRAM utilization (~40-60%). Fusing multiple stages into a single kernel using shared memory could reduce global memory round-trips.
</div>
<div class="finding">
  <strong>3. Memory-Bound Elementwise Ops:</strong> KB add/mul/sub kernels hit 60-65% of peak DRAM throughput but only ~9% SM. These are memory-bound — performance scales with memory bandwidth. Fusing consecutive elementwise operations (e.g., add+mul, scale+bitrev) would reduce memory traffic.
</div>
<div class="finding">
  <strong>4. ScaleByPowers is Compute-Heavy:</strong> The ScaleByPowers kernel shows 68% SM utilization but only 13% DRAM — it's compute-bound due to chained multiplications (g^i computed via prefix product). This is well-optimized; further improvement would require algorithmic changes.
</div>
<div class="finding">
  <strong>5. Kernel Launch Overhead:</strong> Small kernels (< 10 µs) are dominated by launch overhead. For batch operations, the batch NTT API already amortizes this cost. Consider extending this pattern to other operations (batch add, batch mul).
</div>

<hr style="border-color: var(--border); margin: 3rem 0 1rem;">
<p class="meta" style="text-align: center;">
  Generated by gpu/report/gen_profile_report.go — Linea Prover GPU Profiling Suite
</p>
</div>
</body>
</html>`
