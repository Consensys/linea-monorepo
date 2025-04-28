package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

// Data represents a row in the CSV file

type DataWrite struct {
	Limbs            []string
	IsModexpBase     string
	IsModexpExponent string
	IsModexpModulus  string
	IsModexpResult   string
}

func splitIntoPairs(input string) []string {
	if len(input) == 0 {
		return []string{}
	}

	var result []string
	for i := 0; i < len(input); i += 4 {
		var inp string

		if i+4 > len(input) {
			inp = input[i:]
		} else {
			inp = input[i : i+4]
		}

		if inp == "0000" {
			result = append(result, "0x0")
		} else {
			result = append(result, "0x"+inp)
		}
	}

	return result
}

func writeCsv(name string, dataEntries []DataWrite) {
	// Open a new CSV file for writing
	file, err := os.Create(name)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	var header []string
	for i := 0; i < 8; i++ {
		//header = append(header, fmt.Sprintf("MODEXP_LIMBS_%d", i))
		header = append(header, fmt.Sprintf("LIMBS_%d", i))
	}

	//header = append(header, []string{"MODEXP_IS_ACTIVE", "MODEXP_IS_SMALL", "MODEXP_IS_LARGE", "MODEXP_TO_SMALL"}...)
	header = append(header, []string{"IS_MODEXP_BASE", "IS_MODEXP_EXPONENT", "IS_MODEXP_MODULUS", "IS_MODEXP_RESULT"}...)
	// Write the header row

	if err := writer.Write(header); err != nil {
		fmt.Println("Error writing header:", err)
		return
	}

	// Write each data entry to the CSV file
	for _, entry := range dataEntries {
		var row []string
		row = append(row, entry.Limbs...)
		row = append(row, []string{entry.IsModexpBase, entry.IsModexpExponent, entry.IsModexpModulus, entry.IsModexpResult}...)

		if err := writer.Write(row); err != nil {
			fmt.Println("Error writing row:", err)
			return
		}
	}
}

func main() {
	// Open the CSV file
	file, err := os.Open("old/single_4096_bits_input.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all rows from the CSV file
	rows, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV:", err)
		return
	}

	// Ensure the file has a header row
	if len(rows) < 2 {
		fmt.Println("CSV file is empty or missing header row.")
		return
	}

	// Parse the rows into the Data structure
	var mergedEntries []DataWrite

	for i, row := range rows[1:] {
		if len(row) != 5 {
			fmt.Printf("Skipping row %d due to incorrect number of columns\n", i+1)
			continue
		}

		mergedEntries = append(mergedEntries, DataWrite{
			Limbs:            splitIntoPairs(fmt.Sprintf("%032s", row[0][2:])),
			IsModexpBase:     row[1],
			IsModexpExponent: row[2],
			IsModexpModulus:  row[3],
			IsModexpResult:   row[4],
		})

	}

	writeCsv("single_4096_bits_input.csv", mergedEntries)
}
