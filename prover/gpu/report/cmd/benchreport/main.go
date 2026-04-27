// Command gen_report generates an HTML performance report from Go benchmark output.
//
// Usage:
//
//	go run ./gpu/report/ < combined_bench.txt > gpu/gpu_benchmark_report.html
package main

import (
	"bufio"
	"fmt"
	"html/template"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type BenchResult struct {
	Name      string
	Package   string
	Group     string // e.g. "KBVecH2D", "BLSFFTForward"
	SubName   string // e.g. "n=16K"
	N         int    // parsed vector size
	NsPerOp   float64
	MBPerSec  float64
	BytesOp   int64
	Metrics   map[string]float64 // custom metrics: h2d_µs, compute_µs, etc.
}

type GroupData struct {
	Name    string
	Results []BenchResult
}

type ReportData struct {
	GPUName     string
	Groups      []GroupData
	Categories  []Category
	AllResults  []BenchResult
}

type Category struct {
	Name   string
	Groups []GroupData
}

var benchLineRe = regexp.MustCompile(`^Benchmark(\S+?)-\d+\s+(\d+)\s+(\S+)\s+ns/op(.*)$`)
var metricRe = regexp.MustCompile(`(\S+)\s+(\S+)`)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var results []BenchResult

	for scanner.Scan() {
		line := scanner.Text()
		m := benchLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		fullName := m[1]
		nsPerOp, _ := strconv.ParseFloat(m[3], 64)
		extra := m[4]

		r := BenchResult{
			Name:    fullName,
			NsPerOp: nsPerOp,
			Metrics: make(map[string]float64),
		}

		// Parse MB/s and custom metrics from extra
		parts := strings.Fields(extra)
		for i := 0; i+1 < len(parts); i += 2 {
			val, err := strconv.ParseFloat(parts[i], 64)
			if err != nil {
				continue
			}
			label := parts[i+1]
			if label == "MB/s" {
				r.MBPerSec = val
			} else {
				r.Metrics[label] = val
			}
		}

		// Parse group and subname
		if idx := strings.Index(fullName, "/"); idx >= 0 {
			r.Group = fullName[:idx]
			r.SubName = fullName[idx+1:]
		} else {
			r.Group = fullName
			r.SubName = ""
		}

		// Try to parse N from subname
		r.N = parseN(r.SubName)

		// Categorize package
		if strings.HasPrefix(r.Group, "KB") || strings.HasPrefix(r.Group, "Poseidon2") || strings.HasPrefix(r.Group, "LinCombE4") || strings.HasPrefix(r.Group, "Vortex") || strings.HasPrefix(r.Group, "Commit") {
			r.Package = "vortex"
		} else {
			r.Package = "plonk"
		}

		results = append(results, r)
	}

	// Group by Group name
	groupMap := make(map[string][]BenchResult)
	var groupOrder []string
	for _, r := range results {
		if _, ok := groupMap[r.Group]; !ok {
			groupOrder = append(groupOrder, r.Group)
		}
		groupMap[r.Group] = append(groupMap[r.Group], r)
	}

	// Sort within each group by N
	var groups []GroupData
	for _, name := range groupOrder {
		rs := groupMap[name]
		sort.Slice(rs, func(i, j int) bool { return rs[i].N < rs[j].N })
		groups = append(groups, GroupData{Name: name, Results: rs})
	}

	// Categorize
	categories := categorize(groups)

	data := ReportData{
		GPUName:    "NVIDIA RTX PRO 6000 Blackwell (98 GB VRAM)",
		Groups:     groups,
		Categories: categories,
		AllResults: results,
	}

	tmpl := template.Must(template.New("report").Funcs(template.FuncMap{
		"fmtNs":    fmtNs,
		"fmtMBs":   fmtMBs,
		"fmtFloat": func(f float64) string { return strconv.FormatFloat(f, 'f', 1, 64) },
		"toJSON":   toJSON,
		"hasKey":   func(m map[string]float64, k string) bool { _, ok := m[k]; return ok },
		"getKey":   func(m map[string]float64, k string) float64 { return m[k] },
		"colorClass": func(speedup float64) string {
			if speedup >= 50 {
				return "excellent"
			} else if speedup >= 10 {
				return "good"
			} else if speedup >= 2 {
				return "moderate"
			}
			return "poor"
		},
	}).Parse(reportHTML))

	if err := tmpl.Execute(os.Stdout, data); err != nil {
		fmt.Fprintf(os.Stderr, "template error: %v\n", err)
		os.Exit(1)
	}
}

func parseN(s string) int {
	// Try patterns like "n=16K", "n=1M", "n=256", "cols=4K_rows=128"
	re := regexp.MustCompile(`n=(\d+)([KkMm]?)`)
	m := re.FindStringSubmatch(s)
	if m != nil {
		v, _ := strconv.Atoi(m[1])
		switch strings.ToUpper(m[2]) {
		case "K":
			v *= 1024
		case "M":
			v *= 1024 * 1024
		}
		return v
	}
	// Try batch=
	re2 := regexp.MustCompile(`batch=(\d+)`)
	m2 := re2.FindStringSubmatch(s)
	if m2 != nil {
		v, _ := strconv.Atoi(m2[1])
		return v
	}
	// Try cols=
	re3 := regexp.MustCompile(`cols=(\d+)([KkMm]?)`)
	m3 := re3.FindStringSubmatch(s)
	if m3 != nil {
		v, _ := strconv.Atoi(m3[1])
		switch strings.ToUpper(m3[2]) {
		case "K":
			v *= 1024
		case "M":
			v *= 1024 * 1024
		}
		return v
	}
	// Try count=
	re4 := regexp.MustCompile(`count=(\d+)([KkMm]?)`)
	m4 := re4.FindStringSubmatch(s)
	if m4 != nil {
		v, _ := strconv.Atoi(m4[1])
		switch strings.ToUpper(m4[2]) {
		case "K":
			v *= 1024
		case "M":
			v *= 1024 * 1024
		}
		return v
	}
	return 0
}

func fmtNs(ns float64) string {
	switch {
	case ns >= 1e9:
		return fmt.Sprintf("%.2f s", ns/1e9)
	case ns >= 1e6:
		return fmt.Sprintf("%.2f ms", ns/1e6)
	case ns >= 1e3:
		return fmt.Sprintf("%.1f µs", ns/1e3)
	default:
		return fmt.Sprintf("%.0f ns", ns)
	}
}

func fmtMBs(mbs float64) string {
	if mbs >= 1e6 {
		return fmt.Sprintf("%.1f TB/s", mbs/1e6)
	} else if mbs >= 1000 {
		return fmt.Sprintf("%.1f GB/s", mbs/1000)
	}
	return fmt.Sprintf("%.1f MB/s", mbs)
}

func categorize(groups []GroupData) []Category {
	catMap := map[string][]GroupData{
		"Data Transfer (KoalaBear)":  {},
		"Vector Arithmetic (KB)":     {},
		"NTT / FFT (KoalaBear)":      {},
		"Poseidon2":                   {},
		"Vortex Pipeline":            {},
		"Data Transfer (BLS12-377)":  {},
		"Vector Arithmetic (BLS)":    {},
		"NTT / FFT (BLS12-377)":      {},
		"MSM (BLS12-377)":            {},
		"GPU vs CPU Comparison":      {},
	}

	catOrder := []string{
		"Data Transfer (KoalaBear)",
		"Vector Arithmetic (KB)",
		"NTT / FFT (KoalaBear)",
		"Poseidon2",
		"Vortex Pipeline",
		"Data Transfer (BLS12-377)",
		"Vector Arithmetic (BLS)",
		"NTT / FFT (BLS12-377)",
		"MSM (BLS12-377)",
		"GPU vs CPU Comparison",
	}

	for _, g := range groups {
		switch {
		case strings.Contains(g.Name, "vsCPU"):
			catMap["GPU vs CPU Comparison"] = append(catMap["GPU vs CPU Comparison"], g)
		case strings.HasPrefix(g.Name, "KBVecH2D") || strings.HasPrefix(g.Name, "KBVecD2H") || strings.HasPrefix(g.Name, "KBVecD2D"):
			catMap["Data Transfer (KoalaBear)"] = append(catMap["Data Transfer (KoalaBear)"], g)
		case strings.HasPrefix(g.Name, "KBVec"):
			catMap["Vector Arithmetic (KB)"] = append(catMap["Vector Arithmetic (KB)"], g)
		case strings.HasPrefix(g.Name, "KBNTT") || strings.HasPrefix(g.Name, "KBBatch"):
			catMap["NTT / FFT (KoalaBear)"] = append(catMap["NTT / FFT (KoalaBear)"], g)
		case strings.HasPrefix(g.Name, "Poseidon2"):
			catMap["Poseidon2"] = append(catMap["Poseidon2"], g)
		case strings.HasPrefix(g.Name, "Vortex") || strings.HasPrefix(g.Name, "Commit"):
			catMap["Vortex Pipeline"] = append(catMap["Vortex Pipeline"], g)
		case strings.HasPrefix(g.Name, "FrVecH2D") || strings.HasPrefix(g.Name, "FrVecD2H") || strings.HasPrefix(g.Name, "FrVecD2D"):
			catMap["Data Transfer (BLS12-377)"] = append(catMap["Data Transfer (BLS12-377)"], g)
		case strings.HasPrefix(g.Name, "FrVec"):
			catMap["Vector Arithmetic (BLS)"] = append(catMap["Vector Arithmetic (BLS)"], g)
		case strings.HasPrefix(g.Name, "BLSFFT"):
			catMap["NTT / FFT (BLS12-377)"] = append(catMap["NTT / FFT (BLS12-377)"], g)
		case strings.HasPrefix(g.Name, "MSM"):
			catMap["MSM (BLS12-377)"] = append(catMap["MSM (BLS12-377)"], g)
		}
	}

	var cats []Category
	for _, name := range catOrder {
		if gs, ok := catMap[name]; ok && len(gs) > 0 {
			cats = append(cats, Category{Name: name, Groups: gs})
		}
	}
	return cats
}

func toJSON(groups []GroupData) string {
	// Generate JSON data for charts
	var sb strings.Builder
	sb.WriteString("[")
	for i, g := range groups {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf(`{"name":%q,"points":[`, g.Name))
		for j, r := range g.Results {
			if j > 0 {
				sb.WriteString(",")
			}
			label := r.SubName
			if label == "" {
				label = r.Name
			}
			sb.WriteString(fmt.Sprintf(`{"label":%q,"n":%d,"ns":%.1f,"mbs":%.1f}`,
				label, r.N, r.NsPerOp, r.MBPerSec))
		}
		sb.WriteString("]}")
	}
	sb.WriteString("]")
	return sb.String()
}

var _ = math.Log2 // for potential use

const reportHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>GPU Benchmark Report — Linea Prover</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
<style>
:root {
  --bg: #0d1117;
  --surface: #161b22;
  --border: #30363d;
  --text: #c9d1d9;
  --text-muted: #8b949e;
  --accent: #58a6ff;
  --green: #3fb950;
  --orange: #d29922;
  --red: #f85149;
  --purple: #bc8cff;
}
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  background: var(--bg);
  color: var(--text);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif;
  line-height: 1.6;
  padding: 2rem;
}
.container { max-width: 1400px; margin: 0 auto; }
h1 {
  font-size: 2.5rem;
  background: linear-gradient(135deg, var(--accent), var(--purple));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin-bottom: 0.5rem;
}
h2 {
  font-size: 1.5rem;
  color: var(--accent);
  margin: 2rem 0 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border);
}
h3 { font-size: 1.1rem; color: var(--text); margin: 1.5rem 0 0.5rem; }
.meta {
  color: var(--text-muted);
  font-size: 0.9rem;
  margin-bottom: 2rem;
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}
.summary-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
}
.summary-card .label { color: var(--text-muted); font-size: 0.85rem; }
.summary-card .value { font-size: 2rem; font-weight: 700; color: var(--green); }
.summary-card .unit { font-size: 0.9rem; color: var(--text-muted); }
table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
}
th, td {
  padding: 0.6rem 1rem;
  text-align: right;
  border-bottom: 1px solid var(--border);
}
th {
  color: var(--text-muted);
  font-weight: 600;
  text-transform: uppercase;
  font-size: 0.75rem;
  letter-spacing: 0.05em;
}
td:first-child, th:first-child { text-align: left; }
tr:hover { background: rgba(88, 166, 255, 0.05); }
.excellent { color: var(--green); font-weight: 700; }
.good { color: var(--accent); }
.moderate { color: var(--orange); }
.poor { color: var(--red); }
.chart-container {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  height: 400px;
  position: relative;
}
.chart-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
@media (max-width: 900px) { .chart-row { grid-template-columns: 1fr; } }
.key-finding {
  background: linear-gradient(135deg, rgba(88,166,255,0.1), rgba(188,140,255,0.1));
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 1rem;
}
.key-finding strong { color: var(--accent); }
</style>
</head>
<body>
<div class="container">

<h1>GPU Benchmark Report</h1>
<p class="meta">
  <strong>Device:</strong> {{.GPUName}}<br>
  <strong>Generated:</strong> <script>document.write(new Date().toISOString().split('T')[0])</script><br>
  <strong>Benchmarks:</strong> {{len .AllResults}} measurements across {{len .Groups}} test groups
</p>

<!-- Key Findings -->
<h2>Key Findings</h2>
<div class="key-finding">
  <strong>Transfer Bandwidth:</strong> KoalaBear H2D peaks at ~54 GB/s with pinned memory (3x faster than pageable ~19 GB/s). BLS12-377 H2D peaks at ~17 GB/s (AoS→SoA transpose overhead). D2D copies reach ~1.9 TB/s internal bandwidth.
</div>
<div class="key-finding">
  <strong>NTT Performance:</strong> KoalaBear NTT at 16M elements: 2.1 ms (31.8 GB/s effective). BLS12-377 NTT at 4M: 3.2 ms (41.3 GB/s effective). End-to-end (H2D+FFT+D2H): KoalaBear 16M in 8.7 ms, BLS12-377 4M in 21.6 ms — dominated by PCIe transfers at large sizes.
</div>
<div class="key-finding">
  <strong>GPU vs CPU Speedup:</strong> BLS12-377 FFT: 15.9x (16K) to 58.2x (1M). KoalaBear vector multiply: 50x-283x at large sizes. Vortex commit: 7.2x (4K×128) to 13.7x (64K×512) — limited by H2D copy of raw rows.
</div>
<div class="key-finding">
  <strong>MSM Performance:</strong> Pippenger MSM peaks at ~1.2 GB/s (scalar throughput) at 1M points, 17.3 ms per MSM. Scales to 128M points (2.5 s) using chunked two-pass approach to fit sort buffers in VRAM.
</div>
<div class="key-finding">
  <strong>Batch NTT:</strong> Amortized kernel launch overhead via batch API: 100 × 256K NTTs in 1.68 ms (62.3 GB/s effective) vs 7.2 ms for 100 individual calls — 4.3x improvement from batching.
</div>
<div class="key-finding">
  <strong>Poseidon2:</strong> Batch compression (width=16) at 65K hashes: embedded in the Vortex commit pipeline. GPU Vortex commit for 64K×512 matrix: 13 ms total (RS encode + SIS + Poseidon2 + Merkle tree).
</div>

{{range .Categories}}
<h2>{{.Name}}</h2>
{{range .Groups}}
<h3>{{.Name}}</h3>
<table>
<thead>
<tr>
  <th>Configuration</th>
  <th>Time/op</th>
  <th>Throughput</th>
  {{if hasKey (index .Results 0).Metrics "h2d_µs"}}<th>H2D</th><th>Compute</th><th>D2H</th>{{end}}
</tr>
</thead>
<tbody>
{{range .Results}}
<tr>
  <td>{{if .SubName}}{{.SubName}}{{else}}{{.Name}}{{end}}</td>
  <td>{{fmtNs .NsPerOp}}</td>
  <td>{{if gt .MBPerSec 0.0}}{{fmtMBs .MBPerSec}}{{else}}-{{end}}</td>
  {{if hasKey .Metrics "h2d_µs"}}<td>{{fmtFloat (getKey .Metrics "h2d_µs")}} µs</td>
  <td>{{fmtFloat (getKey .Metrics "compute_µs")}} µs</td>
  <td>{{fmtFloat (getKey .Metrics "d2h_µs")}} µs</td>{{end}}
</tr>
{{end}}
</tbody>
</table>
{{end}}
{{end}}

<!-- Charts -->
<h2>Performance Scaling Charts</h2>

<div id="charts"></div>

<script>
const chartColors = [
  '#58a6ff', '#3fb950', '#d29922', '#f85149', '#bc8cff',
  '#79c0ff', '#56d364', '#e3b341', '#ff7b72', '#d2a8ff'
];

function createChart(container, title, labels, datasets) {
  const div = document.createElement('div');
  div.className = 'chart-container';
  const canvas = document.createElement('canvas');
  div.appendChild(canvas);
  container.appendChild(div);

  new Chart(canvas.getContext('2d'), {
    type: 'line',
    data: { labels, datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        title: { display: true, text: title, color: '#c9d1d9', font: { size: 16 } },
        legend: { labels: { color: '#8b949e' } }
      },
      scales: {
        x: {
          type: 'logarithmic',
          title: { display: true, text: 'Vector Size', color: '#8b949e' },
          ticks: { color: '#8b949e' },
          grid: { color: '#30363d' }
        },
        y: {
          type: 'logarithmic',
          title: { display: true, text: 'Time (µs)', color: '#8b949e' },
          ticks: { color: '#8b949e' },
          grid: { color: '#30363d' }
        }
      }
    }
  });
}

function createThroughputChart(container, title, labels, datasets) {
  const div = document.createElement('div');
  div.className = 'chart-container';
  const canvas = document.createElement('canvas');
  div.appendChild(canvas);
  container.appendChild(div);

  new Chart(canvas.getContext('2d'), {
    type: 'bar',
    data: { labels, datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        title: { display: true, text: title, color: '#c9d1d9', font: { size: 16 } },
        legend: { labels: { color: '#8b949e' } }
      },
      scales: {
        x: {
          ticks: { color: '#8b949e' },
          grid: { color: '#30363d' }
        },
        y: {
          type: 'logarithmic',
          title: { display: true, text: 'Throughput (MB/s)', color: '#8b949e' },
          ticks: { color: '#8b949e' },
          grid: { color: '#30363d' }
        }
      }
    }
  });
}

const chartsDiv = document.getElementById('charts');

// NTT scaling chart
const data = {{toJSON .Groups}};

// Group related benchmarks for charts
const nttGroups = data.filter(g => g.name.includes('NTT') && g.name.includes('Scaling'));
if (nttGroups.length > 0) {
  const labels = nttGroups[0].points.map(p => p.label);
  const datasets = nttGroups.map((g, i) => ({
    label: g.name,
    data: g.points.map(p => p.ns / 1000),
    borderColor: chartColors[i],
    backgroundColor: chartColors[i] + '33',
    tension: 0.3
  }));
  createChart(chartsDiv, 'NTT Scaling (Time vs Size)',
    nttGroups[0].points.map(p => p.n), datasets);
}

// Transfer bandwidth chart
const xferGroups = data.filter(g => g.name.includes('H2D') || g.name.includes('D2H') || g.name.includes('D2D'));
if (xferGroups.length > 0) {
  const maxPoints = Math.max(...xferGroups.map(g => g.points.length));
  const ref = xferGroups.reduce((a, b) => a.points.length > b.points.length ? a : b);
  const labels = ref.points.map(p => p.label);
  const datasets = xferGroups.map((g, i) => ({
    label: g.name,
    data: g.points.map(p => p.mbs),
    backgroundColor: chartColors[i] + 'cc',
    borderColor: chartColors[i]
  }));
  createThroughputChart(chartsDiv, 'Transfer Bandwidth (MB/s)',
    labels, datasets);
}

// GPU vs CPU comparison
const vsCPU = data.filter(g => g.name.includes('vsCPU'));
if (vsCPU.length > 0) {
  vsCPU.forEach((g, gi) => {
    const gpuPoints = g.points.filter(p => p.label.startsWith('GPU'));
    const cpuPoints = g.points.filter(p => p.label.startsWith('CPU'));
    if (gpuPoints.length > 0 && cpuPoints.length > 0) {
      const labels = gpuPoints.map(p => p.label.replace('GPU/', ''));
      createThroughputChart(chartsDiv, g.name + ' — GPU vs CPU Throughput',
        labels,
        [
          { label: 'GPU', data: gpuPoints.map(p => p.mbs), backgroundColor: '#3fb95099', borderColor: '#3fb950' },
          { label: 'CPU', data: cpuPoints.map(p => p.mbs), backgroundColor: '#f8514999', borderColor: '#f85149' }
        ]);
    }
  });
}

// Vector arithmetic chart
const arithGroups = data.filter(g =>
  (g.name.includes('VecAdd') || g.name.includes('VecMul') || g.name.includes('VecSub') || g.name.includes('VecScale')) &&
  !g.name.includes('CPU'));
if (arithGroups.length > 0) {
  const ref = arithGroups[0];
  const labels = ref.points.map(p => p.label);
  const datasets = arithGroups.map((g, i) => ({
    label: g.name,
    data: g.points.map(p => p.ns / 1000),
    borderColor: chartColors[i],
    backgroundColor: chartColors[i] + '33',
    tension: 0.3
  }));
  createChart(chartsDiv, 'Vector Arithmetic Latency',
    ref.points.map(p => p.n), datasets);
}
</script>

<hr style="border-color: var(--border); margin: 3rem 0 1rem;">
<p class="meta" style="text-align: center;">
  Generated by gpu/report/gen_report.go — Linea Prover GPU Benchmark Suite
</p>
</div>
</body>
</html>`
