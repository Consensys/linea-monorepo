package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"regexp"
)

var reDur = `([0-9]+(?:\.[0-9]+)?(?:ns|us|µs|μs|ms|s))`

var (
	paramRe  = regexp.MustCompile(`numRounds=(\d+),\s*nPoly=(\d+),\s*PolySize=(\d+)`)
	encodeRe = regexp.MustCompile(`timeEncoding=` + reDur)
	sisRe    = regexp.MustCompile(`timeSisHashing=` + reDur)
	merkRe   = regexp.MustCompile(`timeMerkleizing=` + reDur)
	updateRe = regexp.MustCompile(`time to update prover state.*:\s*` + reDur)
	coinRe   = regexp.MustCompile(`LC random coin generation time:\s*` + reDur)
	lcRe     = regexp.MustCompile(`time to compute lC-step\s*` + reDur)
	rsRe     = regexp.MustCompile(`time to compute RS encoding of LC\s*` + reDur)
	coin2Re  = regexp.MustCompile(`time to generate coin for column-opening:\s*` + reDur)
	retRe    = regexp.MustCompile(`time to retrieve opened columns and merkle proofs:\s*` + reDur)
)

var baseCols = []string{"numRounds", "nPoly", "PolySize"}
var timings = []string{
	"timeEncoding",
	"timeSisHashing",
	"timeMerkleizing",
	"updateProverState",
	"lcRandomCoin",
	"lcStep",
	"rsEncodingLC",
	"coinForColumnOpening",
	"retrieveOpenedCols",
}

func main() {
	inPath := flag.String("in", "benchmark.log", "input log file")
	outPath := flag.String("out", "vortex_benchmarks.csv", "output CSV file")
	appendFlag := flag.Bool("append", false, "append to output CSV instead of overwrite")
	flag.Parse()

	inFile, err := os.Open(*inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open input: %v\n", err)
		os.Exit(1)
	}
	defer inFile.Close()

	var outFile *os.File
	if *appendFlag {
		outFile, err = os.OpenFile(*outPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		outFile, err = os.Create(*outPath)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "open output: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// write header only when not appending
	if !*appendFlag {
		header := make([]string, 0, len(baseCols)+len(timings))
		header = append(header, baseCols...)
		header = append(header, timings...)
		writer.Write(header)
	}

	scanner := bufio.NewScanner(inFile)
	runParams := map[string]string{}
	roundTimings := map[string]string{}

	writeRound := func() {
		if len(roundTimings) == 0 {
			return
		}
		row := []string{
			runParams["numRounds"], runParams["nPoly"], runParams["PolySize"],
			roundTimings["timeEncoding"],
			roundTimings["timeSisHashing"],
			roundTimings["timeMerkleizing"],
			roundTimings["updateProverState"],
			roundTimings["lcRandomCoin"],
			roundTimings["lcStep"],
			roundTimings["rsEncodingLC"],
			roundTimings["coinForColumnOpening"],
			roundTimings["retrieveOpenedCols"],
		}
		writer.Write(row)
		roundTimings = map[string]string{}
	}

	for scanner.Scan() {
		line := scanner.Text()

		// New run marker
		if m := paramRe.FindStringSubmatch(line); m != nil {
			// write any last round
			writeRound()
			runParams["numRounds"] = m[1]
			runParams["nPoly"] = m[2]
			runParams["PolySize"] = m[3]
			continue
		}

		if m := encodeRe.FindStringSubmatch(line); m != nil {
			// new round starts, write previous round
			writeRound()
			roundTimings["timeEncoding"] = m[1]
		}
		if m := sisRe.FindStringSubmatch(line); m != nil {
			roundTimings["timeSisHashing"] = m[1]
		}
		if m := merkRe.FindStringSubmatch(line); m != nil {
			roundTimings["timeMerkleizing"] = m[1]
		}
		if m := updateRe.FindStringSubmatch(line); m != nil {
			roundTimings["updateProverState"] = m[1]
		}
		if m := coinRe.FindStringSubmatch(line); m != nil {
			roundTimings["lcRandomCoin"] = m[1]
		}
		if m := lcRe.FindStringSubmatch(line); m != nil {
			roundTimings["lcStep"] = m[1]
		}
		if m := rsRe.FindStringSubmatch(line); m != nil {
			roundTimings["rsEncodingLC"] = m[1]
		}
		if m := coin2Re.FindStringSubmatch(line); m != nil {
			roundTimings["coinForColumnOpening"] = m[1]
		}
		if m := retRe.FindStringSubmatch(line); m != nil {
			roundTimings["retrieveOpenedCols"] = m[1]
		}
	}

	// write last round if any
	writeRound()
	fmt.Println("written CSV to", *outPath)
}
