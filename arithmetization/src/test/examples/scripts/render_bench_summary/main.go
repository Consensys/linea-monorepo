// Parses the per-iteration logs produced by the
// arithmetization benchmark workflows and renders a
// Markdown summary to stdout (suitable for `>> $GITHUB_STEP_SUMMARY`).
//
// Log file naming (produced by the workflow's timing loop):
//
//	keccak_{opt|base}_<i>.log          - 1 file per (variant, iter)
//	blake_{opt|base}_<i>_<vec>.log     - M files per (variant, iter) where M = -blake-n
//
// Each log is `/usr/bin/time -v <zkc> ... <prog>` output, where `<zkc>` is
// either the RISC-V interpreter (`zkc execute`) or the native compiler/VM
// (`zkc exec`). The fields we extract are:
//
//   - `(Machine|Constraint) execution (N steps) took Xs`
//     -> steps + execution_s (uint + float; steps is invariant)
//     The interpreter emits "Machine execution"; native zkc emits
//     "Constraint execution". We match either.
//   - "Elapsed (wall clock) time ... : M:SS.ss" -> wall_s (float) => unused
//   - "Maximum resident set size (kbytes): N"   -> rss_kb (int)
//
// Aggregation per (variant, iter):
//   - Keccak: identity (1 log per (variant, iter)).
//   - Blake : sum constraint_s/wall_s/steps across the M vector logs,
//     max RSS, assert all M parsed.
//
// Output:
//   - Per-workload invariant PASS/FAIL line for step count.
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
	// Matches both `zkc execute` (RISC-V interpreter, "Machine execution")
	// and `zkc exec`    (native zkc,           "Constraint execution").
	reExecution     = regexp.MustCompile(`(?:Machine|Constraint) execution \((\d+) steps\) took ([\d.]+)s`)
	reWall          = regexp.MustCompile(`Elapsed \(wall clock\) time \(h:mm:ss or m:ss\): (\S+)`)
	reRSS           = regexp.MustCompile(`Maximum resident set size \(kbytes\): (\d+)`)
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

	if mm := reExecution.FindStringSubmatch(body); mm != nil {
		steps, err := strconv.ParseUint(mm[1], 10, 64)
		if err != nil {
			return m, fmt.Errorf("%s: parse steps: %w", path, err)
		}
		secs, err := strconv.ParseFloat(mm[2], 64)
		if err != nil {
			return m, fmt.Errorf("%s: parse execution_s: %w", path, err)
		}
		m.steps = steps
		m.constraintS = secs
	} else {
		return m, fmt.Errorf("%s: execution line not found (neither Machine nor Constraint)", path)
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
		count              int
		constraintS, wallS float64
		steps, rssK        uint64
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

// formatThousands renders an unsigned integer with comma thousands separators
// (e.g. 173005454 -> "173,005,454"). Avoids pulling in golang.org/x/text just
// for this; the helper scripts deliberately have zero external deps.
func formatThousands(n uint64) string {
	s := strconv.FormatUint(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	first := len(s) % 3
	if first == 0 {
		first = 3
	}
	b.WriteString(s[:first])
	for i := first; i < len(s); i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

func renderInvariants(out *strings.Builder, w workload) bool {
	var (
		stepsSeen     uint64
		stepsConflict bool
		first         = true
	)
	for _, v := range []*variantMetrics{&w.opt, &w.base} {
		for _, iter := range sortedIters(v.byIter) {
			m := v.byIter[iter]
			if first {
				stepsSeen = m.steps
				first = false
				continue
			}
			if m.steps != stepsSeen {
				stepsConflict = true
			}
		}
	}
	stepsOK := !stepsConflict
	if stepsOK {
		fmt.Fprintf(out, "- **%s** steps: PASS with machine exec steps = %s\n", w.name, formatThousands(stepsSeen))
	} else {
		fmt.Fprintf(out, "- **%s** steps: FAIL — machine exec steps diverge across variants/iters\n", w.name)
	}
	return stepsOK
}

func renderPerIter(out *strings.Builder, w workload, iters int) {
	fmt.Fprintf(out, "\n#### %s — per iteration\n\n", w.name)
	fmt.Fprintln(out, "| iter | base_wall (s) | opt_wall (s) | Δ (o - b) | % |")
	fmt.Fprintln(out, "| ---: | ---: | ---: | ---: | ---: |")
	for i := 1; i <= iters; i++ {
		bm, bOK := w.base.byIter[i]
		om, oOK := w.opt.byIter[i]
		if !bOK || !oOK {
			fmt.Fprintf(out, "| %d | (missing) | (missing) | – | – |\n", i)
			continue
		}
		// Δ = opt - base, so a NEGATIVE Δ means optim is faster than base
		// (lower wall clock = win). Percentage is still normalised to base.
		dW := om.wallS - bm.wallS
		fmt.Fprintf(out, "| %d | %.2f | %.2f | %+.2f | %+.2f%% |\n",
			i, bm.wallS, om.wallS, dW, pct(dW, bm.wallS))
	}
}

func renderAggregate(out *strings.Builder, w workload) {
	type col struct {
		base, opt []float64
	}
	wall := col{}
	rss := col{}
	for _, i := range sortedIters(w.opt.byIter) {
		om := w.opt.byIter[i]
		bm, ok := w.base.byIter[i]
		if !ok {
			continue
		}
		wall.base = append(wall.base, bm.wallS)
		wall.opt = append(wall.opt, om.wallS)
		rss.base = append(rss.base, float64(bm.rssKB))
		rss.opt = append(rss.opt, float64(om.rssKB))
	}

	row := func(label string, c col, fmtStr string) string {
		bMu, bSd := mean(c.base), stdev(c.base)
		bLo, bHi := minMax(c.base)
		oMu, oSd := mean(c.opt), stdev(c.opt)
		oLo, oHi := minMax(c.opt)
		// Δ mean = opt - base; negative = optim wins on average.
		dMu := oMu - bMu
		return fmt.Sprintf(
			"| %s | "+fmtStr+" ± "+fmtStr+" ["+fmtStr+", "+fmtStr+"] | "+fmtStr+" ± "+fmtStr+" ["+fmtStr+", "+fmtStr+"] | %+.2f | %+.2f%% |\n",
			label, bMu, bSd, bLo, bHi, oMu, oSd, oLo, oHi, dMu, pct(dMu, bMu))
	}

	fmt.Fprintf(out, "\n#### %s — aggregate\n\n", w.name)
	fmt.Fprintln(out, "| metric | baseline (mean ± stdev [min, max]) | optimised (mean ± stdev [min, max]) | Δ mean (o - b) | % |")
	fmt.Fprintln(out, "| --- | --- | --- | ---: | ---: |")
	fmt.Fprint(out, row("wall (s)", wall, "%.2f"))
	fmt.Fprint(out, row("RSS (KB)", rss, "%.0f"))

	// Pass (opt, base) so the t-stat sign matches the Δ sign convention:
	// negative t-stat <=> optim faster on average.
	tW := pairedT(wall.opt, wall.base)
	fmt.Fprintf(out, "\nPaired t-stat (o - b): wall = %s (n=%d)\n",
		formatFloat(tW), len(wall.base))
}

func formatFloat(v float64) string {
	if math.IsNaN(v) {
		return "n/a"
	}
	return fmt.Sprintf("%+.2f", v)
}

// parseWorkloads splits a comma-separated `-workloads` value into the set of
// workloads we will render. Unknown entries are rejected.
func parseWorkloads(s string) (wantKeccak, wantBlake bool, err error) {
	for _, part := range strings.Split(s, ",") {
		switch strings.TrimSpace(strings.ToLower(part)) {
		case "":
			continue
		case "keccak":
			wantKeccak = true
		case "blake":
			wantBlake = true
		default:
			return false, false, fmt.Errorf("unknown workload %q (allowed: keccak, blake)", part)
		}
	}
	if !wantKeccak && !wantBlake {
		return false, false, fmt.Errorf("no workloads requested (use -workloads keccak,blake)")
	}
	return wantKeccak, wantBlake, nil
}

func main() {
	logsDir := flag.String("logs", "", "directory containing the per-iter benchmark logs")
	iters := flag.Int("iters", 5, "number of timed iterations per (workload, variant)")
	blakeN := flag.Int("blake-n", 3, "number of Blake vectors aggregated per (variant, iter)")
	workloads := flag.String("workloads", "keccak,blake", "comma-separated list of workloads to render (keccak,blake)")
	baseRef := flag.String("base-ref", "", "baseline branch/commit ref (informational)")
	optimRef := flag.String("optim-ref", "", "optim-test branch/commit ref (informational)")
	zkcVersion := flag.String("zkc-version", "", "go-corset ref used to build the zkc binary (informational)")
	keccakNVectors := flag.Int("keccak-n-vectors", 0, "number of Keccak vectors batched into one zkc exec (informational, 0 = omit)")
	blakeRounds := flag.Int("blake-rounds", 0, "number of Blake2b compression rounds (informational, 0 = omit)")
	flag.Parse()
	if *logsDir == "" {
		fmt.Fprintln(os.Stderr, "error: -logs is required")
		os.Exit(1)
	}
	wantKeccak, wantBlake, err := parseWorkloads(*workloads)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	kc, bl, err := discover(*logsDir, *iters, *blakeN)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	var out strings.Builder
	out.WriteString("## ZkC benchmark — base vs optim\n\n")

	out.WriteString("### Workflow inputs\n\n")
	if *baseRef != "" {
		fmt.Fprintf(&out, "- base branch ref: `%s`\n", *baseRef)
	}
	if *optimRef != "" {
		fmt.Fprintf(&out, "- optim branch ref: `%s`\n", *optimRef)
	}
	if *zkcVersion != "" {
		fmt.Fprintf(&out, "- zkc version (go-corset ref): `%s`\n", *zkcVersion)
	}
	if wantKeccak && *keccakNVectors > 0 {
		fmt.Fprintf(&out, "- number of Keccak vectors: %d\n", *keccakNVectors)
	}
	if wantBlake && *blakeRounds > 0 {
		fmt.Fprintf(&out, "- number of Blake compression rounds: %d\n", *blakeRounds)
	}
	out.WriteString("\n")

	out.WriteString("### First warm up run\n\n")
	allOK := true
	if wantKeccak && !renderInvariants(&out, kc) {
		allOK = false
	}
	if wantBlake && !renderInvariants(&out, bl) {
		allOK = false
	}
	if !allOK {
		out.WriteString("\n> NOTE: machine exec steps differ across variants/iters — the optim branch may have changed semantics, not just performance.\n")
	}

	out.WriteString("\n### Per-iteration timings\n")
	if wantKeccak {
		renderPerIter(&out, kc, *iters)
	}
	if wantBlake {
		renderPerIter(&out, bl, *iters)
	}

	out.WriteString("\n### Aggregates")
	if wantKeccak {
		renderAggregate(&out, kc)
	}
	if wantBlake {
		renderAggregate(&out, bl)
	}

	fmt.Print(out.String())
}
