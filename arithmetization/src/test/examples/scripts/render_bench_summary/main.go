// Parses the per-iteration logs produced by the
// arithmetization-benchmark-zkc-interpreter workflow and renders a
// Markdown summary to stdout (suitable for `>> $GITHUB_STEP_SUMMARY`).
//
// Log file naming (produced by the workflow's timing loop):
//
//	keccak_{opt|base}_<i>.log          - 1 file per (variant, iter)
//	blake_{opt|base}_<i>_<vec>.log     - M files per (variant, iter) where M = -blake-n
//
// Each log is `/usr/bin/time -v zkc execute -v <json> <main.zkc>` output. The
// fields we extract are:
//
//   - "Constraint execution took Xs"            -> constraint_s        (float)
//   - "Constraint execution (N steps) ..."      -> steps               (int, invariant)
//   - "Elapsed (wall clock) time ... : M:SS.ss" -> wall_s              (float)
//   - "Maximum resident set size (kbytes): N"   -> rss_kb              (int)
//   - "TOTAL clock cycles: N"                   -> cycles              (int, invariant)
//
// Aggregation per (variant, iter):
//   - Keccak: identity (1 log per (variant, iter)).
//   - Blake : sum constraint_s/wall_s/cycles/steps across the M vector logs,
//     max RSS, assert all M parsed.
//
// Output:
//   - Per-workload invariant PASS/FAIL lines for cycle and step count.
//   - Per-iter Markdown table.
//   - Aggregate Markdown table with mean / stdev / [min, max] / paired Δ / paired t-stat.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type metrics struct {
	constraintS float64
	steps       uint64
	wallS       float64
	rssKB       uint64
	cycles      uint64
}

type variantMetrics struct {
	byIter map[int]*metrics // iter -> aggregated metrics
}

type workload struct {
	name string
	opt  variantMetrics
	base variantMetrics
}

var (
	reConstraintS   = regexp.MustCompile(`Constraint execution took ([\d.]+)s`)
	reSteps         = regexp.MustCompile(`Constraint execution \((\d+) steps\)`)
	reWall          = regexp.MustCompile(`Elapsed \(wall clock\) time \(h:mm:ss or m:ss\): (\S+)`)
	reRSS           = regexp.MustCompile(`Maximum resident set size \(kbytes\): (\d+)`)
	reCycles        = regexp.MustCompile(`TOTAL clock cycles: (\d+)`)
	reKeccakLogName = regexp.MustCompile(`^keccak_(opt|base)_(\d+)\.log$`)
	reBlakeLogName  = regexp.MustCompile(`^blake_(opt|base)_(\d+)_(\d+)\.log$`)
)

func parseWall(s string) (float64, error) {
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 2: // M:SS.ss
		m, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		sec, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
		return m*60 + sec, nil
	case 3: // H:MM:SS.ss
		h, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		m, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
		sec, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return 0, err
		}
		return h*3600 + m*60 + sec, nil
	}
	return 0, fmt.Errorf("unrecognised elapsed-time format: %q", s)
}

func parseLog(path string) (metrics, error) {
	var m metrics
	data, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	body := string(data)

	if mm := reConstraintS.FindStringSubmatch(body); mm != nil {
		if v, err := strconv.ParseFloat(mm[1], 64); err == nil {
			m.constraintS = v
		} else {
			return m, fmt.Errorf("%s: parse constraint_s: %w", path, err)
		}
	} else {
		return m, fmt.Errorf("%s: constraint_s line not found", path)
	}
	if mm := reSteps.FindStringSubmatch(body); mm != nil {
		if v, err := strconv.ParseUint(mm[1], 10, 64); err == nil {
			m.steps = v
		} else {
			return m, fmt.Errorf("%s: parse steps: %w", path, err)
		}
	} else {
		return m, fmt.Errorf("%s: steps line not found", path)
	}
	if mm := reWall.FindStringSubmatch(body); mm != nil {
		v, err := parseWall(mm[1])
		if err != nil {
			return m, fmt.Errorf("%s: parse wall_s: %w", path, err)
		}
		m.wallS = v
	} else {
		return m, fmt.Errorf("%s: wall_s line not found", path)
	}
	if mm := reRSS.FindStringSubmatch(body); mm != nil {
		v, err := strconv.ParseUint(mm[1], 10, 64)
		if err != nil {
			return m, fmt.Errorf("%s: parse rss: %w", path, err)
		}
		m.rssKB = v
	} else {
		return m, fmt.Errorf("%s: rss line not found", path)
	}
	if mm := reCycles.FindStringSubmatch(body); mm != nil {
		v, err := strconv.ParseUint(mm[1], 10, 64)
		if err != nil {
			return m, fmt.Errorf("%s: parse cycles: %w", path, err)
		}
		m.cycles = v
	} else {
		return m, fmt.Errorf("%s: cycles line not found", path)
	}
	return m, nil
}

func discover(logsDir string, iters, blakeN int) (kc, bl workload, err error) {
	kc.name = "keccak"
	bl.name = "blake"
	kc.opt.byIter = make(map[int]*metrics)
	kc.base.byIter = make(map[int]*metrics)
	bl.opt.byIter = make(map[int]*metrics)
	bl.base.byIter = make(map[int]*metrics)

	type blakeAgg struct {
		count               int
		constraintS, wallS  float64
		steps, cycles, rssK uint64
	}
	blakeAggs := map[[2]int]*blakeAgg{} // key: [iter, variant=0 opt / 1 base]

	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return kc, bl, fmt.Errorf("read logs dir %s: %w", logsDir, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		full := filepath.Join(logsDir, name)

		if mm := reKeccakLogName.FindStringSubmatch(name); mm != nil {
			variant := mm[1]
			iter, _ := strconv.Atoi(mm[2])
			if iter < 1 || iter > iters {
				continue
			}
			m, perr := parseLog(full)
			if perr != nil {
				return kc, bl, perr
			}
			target := &kc.opt
			if variant == "base" {
				target = &kc.base
			}
			target.byIter[iter] = &m
			continue
		}

		if mm := reBlakeLogName.FindStringSubmatch(name); mm != nil {
			variant := mm[1]
			iter, _ := strconv.Atoi(mm[2])
			if iter < 1 || iter > iters {
				continue
			}
			m, perr := parseLog(full)
			if perr != nil {
				return kc, bl, perr
			}
			vi := 0
			if variant == "base" {
				vi = 1
			}
			key := [2]int{iter, vi}
			agg, ok := blakeAggs[key]
			if !ok {
				agg = &blakeAgg{}
				blakeAggs[key] = agg
			}
			agg.count++
			agg.constraintS += m.constraintS
			agg.wallS += m.wallS
			agg.steps += m.steps
			agg.cycles += m.cycles
			if m.rssKB > agg.rssK {
				agg.rssK = m.rssKB
			}
		}
	}

	for key, a := range blakeAggs {
		if a.count != blakeN {
			return kc, bl, fmt.Errorf("blake iter=%d variant=%d: got %d log(s), expected %d", key[0], key[1], a.count, blakeN)
		}
		m := &metrics{
			constraintS: a.constraintS,
			steps:       a.steps,
			wallS:       a.wallS,
			rssKB:       a.rssK,
			cycles:      a.cycles,
		}
		target := &bl.opt
		if key[1] == 1 {
			target = &bl.base
		}
		target.byIter[key[0]] = m
	}

	return kc, bl, nil
}

func sortedIters(m map[int]*metrics) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var s float64
	for _, v := range xs {
		s += v
	}
	return s / float64(len(xs))
}

func stdev(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	mu := mean(xs)
	var s float64
	for _, v := range xs {
		d := v - mu
		s += d * d
	}
	return math.Sqrt(s / float64(len(xs)-1))
}

func minMax(xs []float64) (float64, float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	lo, hi := xs[0], xs[0]
	for _, v := range xs[1:] {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	return lo, hi
}

func pairedT(base, opt []float64) float64 {
	if len(base) != len(opt) || len(base) < 2 {
		return math.NaN()
	}
	diffs := make([]float64, len(base))
	for i := range base {
		diffs[i] = base[i] - opt[i]
	}
	mu := mean(diffs)
	sd := stdev(diffs)
	if sd == 0 {
		return math.NaN()
	}
	return mu / (sd / math.Sqrt(float64(len(diffs))))
}

func pct(num, denom float64) float64 {
	if denom == 0 {
		return 0
	}
	return 100 * num / denom
}

func renderInvariants(out *strings.Builder, w workload) bool {
	var (
		stepsSeen     uint64
		cyclesSeen    uint64
		stepsConflict bool
		cyclesMixed   bool
		first         = true
	)
	for _, v := range []*variantMetrics{&w.opt, &w.base} {
		for _, iter := range sortedIters(v.byIter) {
			m := v.byIter[iter]
			if first {
				stepsSeen = m.steps
				cyclesSeen = m.cycles
				first = false
				continue
			}
			if m.steps != stepsSeen {
				stepsConflict = true
			}
			if m.cycles != cyclesSeen {
				cyclesMixed = true
			}
		}
	}
	stepsOK := !stepsConflict
	cyclesOK := !cyclesMixed
	if stepsOK {
		fmt.Fprintf(out, "- **%s** steps:  PASS (all variants/iters = `%d`)\n", w.name, stepsSeen)
	} else {
		fmt.Fprintf(out, "- **%s** steps:  FAIL (divergent across variants/iters)\n", w.name)
	}
	if cyclesOK {
		fmt.Fprintf(out, "- **%s** cycles: PASS (all variants/iters = `%d`)\n", w.name, cyclesSeen)
	} else {
		fmt.Fprintf(out, "- **%s** cycles: FAIL (divergent across variants/iters)\n", w.name)
	}
	return stepsOK && cyclesOK
}

func renderPerIter(out *strings.Builder, w workload, iters int) {
	fmt.Fprintf(out, "\n#### %s — per iteration\n\n", w.name)
	fmt.Fprintln(out, "| iter | base_wall (s) | opt_wall (s) | Δ (b - o) | % | base_cstr (s) | opt_cstr (s) | Δ (b - o) | % |")
	fmt.Fprintln(out, "| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |")
	for i := 1; i <= iters; i++ {
		bm, bOK := w.base.byIter[i]
		om, oOK := w.opt.byIter[i]
		if !bOK || !oOK {
			fmt.Fprintf(out, "| %d | (missing) | (missing) | – | – | (missing) | (missing) | – | – |\n", i)
			continue
		}
		dW := bm.wallS - om.wallS
		dC := bm.constraintS - om.constraintS
		fmt.Fprintf(out, "| %d | %.2f | %.2f | %+.2f | %+.2f%% | %.2f | %.2f | %+.2f | %+.2f%% |\n",
			i, bm.wallS, om.wallS, dW, pct(dW, bm.wallS),
			bm.constraintS, om.constraintS, dC, pct(dC, bm.constraintS))
	}
}

func renderAggregate(out *strings.Builder, w workload) {
	type col struct {
		base, opt []float64
	}
	wall := col{}
	cstr := col{}
	rss := col{}
	for _, i := range sortedIters(w.opt.byIter) {
		om := w.opt.byIter[i]
		bm, ok := w.base.byIter[i]
		if !ok {
			continue
		}
		wall.base = append(wall.base, bm.wallS)
		wall.opt = append(wall.opt, om.wallS)
		cstr.base = append(cstr.base, bm.constraintS)
		cstr.opt = append(cstr.opt, om.constraintS)
		rss.base = append(rss.base, float64(bm.rssKB))
		rss.opt = append(rss.opt, float64(om.rssKB))
	}

	row := func(label string, c col, fmtStr string) string {
		bMu, bSd := mean(c.base), stdev(c.base)
		bLo, bHi := minMax(c.base)
		oMu, oSd := mean(c.opt), stdev(c.opt)
		oLo, oHi := minMax(c.opt)
		dMu := bMu - oMu
		return fmt.Sprintf(
			"| %s | "+fmtStr+" ± "+fmtStr+" ["+fmtStr+", "+fmtStr+"] | "+fmtStr+" ± "+fmtStr+" ["+fmtStr+", "+fmtStr+"] | %+.2f | %+.2f%% |\n",
			label, bMu, bSd, bLo, bHi, oMu, oSd, oLo, oHi, dMu, pct(dMu, bMu))
	}

	fmt.Fprintf(out, "\n#### %s — aggregate\n\n", w.name)
	fmt.Fprintln(out, "| metric | baseline (mean ± stdev [min, max]) | optimised (mean ± stdev [min, max]) | Δ mean | % |")
	fmt.Fprintln(out, "| --- | --- | --- | ---: | ---: |")
	fmt.Fprint(out, row("wall (s)", wall, "%.2f"))
	fmt.Fprint(out, row("constraint (s)", cstr, "%.2f"))
	fmt.Fprint(out, row("RSS (KB)", rss, "%.0f"))

	tW := pairedT(wall.base, wall.opt)
	tC := pairedT(cstr.base, cstr.opt)
	fmt.Fprintf(out, "\nPaired t-stat: wall = %s, constraint = %s (n=%d)\n",
		formatFloat(tW), formatFloat(tC), len(wall.base))
}

func formatFloat(v float64) string {
	if math.IsNaN(v) {
		return "n/a"
	}
	return fmt.Sprintf("%+.2f", v)
}

func main() {
	logsDir := flag.String("logs", "", "directory containing the per-iter benchmark logs")
	iters := flag.Int("iters", 5, "number of timed iterations per (workload, variant)")
	blakeN := flag.Int("blake-n", 3, "number of Blake vectors aggregated per (variant, iter)")
	flag.Parse()
	if *logsDir == "" {
		fmt.Fprintln(os.Stderr, "error: -logs is required")
		os.Exit(1)
	}

	kc, bl, err := discover(*logsDir, *iters, *blakeN)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	var out strings.Builder
	out.WriteString("## ZkC interpreter benchmark — base vs optim\n\n")
	out.WriteString("### Invariant sanity checks\n\n")
	allOK := true
	if !renderInvariants(&out, kc) {
		allOK = false
	}
	if !renderInvariants(&out, bl) {
		allOK = false
	}
	if !allOK {
		out.WriteString("\n> NOTE: at least one invariant check FAILED — the optim branch may have changed semantics, not just performance.\n")
	}

	out.WriteString("\n### Per-iteration timings\n")
	renderPerIter(&out, kc, *iters)
	renderPerIter(&out, bl, *iters)

	out.WriteString("\n### Aggregates")
	renderAggregate(&out, kc)
	renderAggregate(&out, bl)

	fmt.Print(out.String())
}
