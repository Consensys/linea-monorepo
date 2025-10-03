package main

import (
	"encoding/csv"
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
)

func main() {
	f, err := os.Open("vortex_benchmarks.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(records[0]) // CSV header

	for _, row := range records[1:] {
		table.Append(row)
	}

	table.Render()
}
