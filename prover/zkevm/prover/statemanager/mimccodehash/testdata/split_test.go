package testdata

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"io"
	"os"
	"testing"
)

type Record struct {
	CFI_ROMLEX  string
	CODEHASH_HI string
	CODEHASH_LO string
}

type RecordWrite struct {
	CFI_ROMLEX string
	CODEHASH   []string
}

func readCSVFile(filePath string) ([]Record, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	rawRecords, err := reader.ReadAll()
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read CSV records from %s: %w", filePath, err)
	}

	if len(rawRecords) == 0 {
		return nil, nil
	}

	dataRows := rawRecords[1:]

	var records []Record

	for _, row := range dataRows {
		record := Record{
			CFI_ROMLEX:  row[0],
			CODEHASH_HI: row[1],
			CODEHASH_LO: row[2],
		}
		records = append(records, record)
	}

	return records, nil
}

func writeCSVFile(filePath string, records []RecordWrite) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"CFI_ROMLEX"}
	for i := range common.NbLimbU256 {
		header = append(header, fmt.Sprintf("CODEHASH_%d", i))
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header to %s: %w", filePath, err)
	}

	for _, record := range records {
		row := append([]string{record.CFI_ROMLEX}, record.CODEHASH[:]...)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write record %v to %s: %w", record, filePath, err)
		}
	}

	return nil
}

func split(romlexRecord []Record) (res []RecordWrite) {
	for _, record := range romlexRecord {
		codehash := record.CODEHASH_HI[2:] + record.CODEHASH_LO[2:]
		codehashBytes, _ := hex.DecodeString(codehash)

		codehashes := make([]string, 16)
		for i, limbBytes := range common.SplitBytes(codehashBytes) {
			codehashes[i] = fmt.Sprintf("0x%s", hex.EncodeToString(limbBytes))
		}

		res = append(res, RecordWrite{
			CFI_ROMLEX: record.CFI_ROMLEX,
			CODEHASH:   codehashes,
		})
	}

	return res
}

func TestSplit(t *testing.T) {
	records, _ := readCSVFile("romlex_input.csv")
	recordsToWrite := split(records)
	writeCSVFile("romlex_input_new.csv", recordsToWrite)
	fmt.Println(recordsToWrite)
}
