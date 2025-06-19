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
	CFI      string
	ACC      string
	NBYTES   string
	COUNTER  string
	CODESIZE string
}

type RecordWrite struct {
	CFI      string
	ACC      string
	NBYTES   string
	COUNTER  string
	CODESIZE []string
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
			CFI:      row[0],
			ACC:      row[1],
			NBYTES:   row[2],
			COUNTER:  row[3],
			CODESIZE: row[4],
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

	header := []string{"CFI", "ACC", "NBYTES", "COUNTER"}
	for i := range common.NbLimbU32 {
		header = append(header, fmt.Sprintf("CODESIZE_%d", i))
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header to %s: %w", filePath, err)
	}

	for _, record := range records {
		row := append([]string{record.CFI, record.ACC, record.NBYTES, record.COUNTER}, record.CODESIZE[:]...)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write record %v to %s: %w", record, filePath, err)
		}
	}

	return nil
}

func split(romlexRecord []Record) (res []RecordWrite) {
	for _, record := range romlexRecord {
		codesizeBytes, _ := hex.DecodeString(record.CODESIZE[2:])
		codesizes := make([]string, common.NbLimbU32)
		codesizeLimbs := common.SplitBytes(codesizeBytes)

		padBytes := make([][]byte, common.NbLimbU32-len(codesizeLimbs))
		padBytes = append(padBytes, codesizeLimbs...)

		for i := 0; i < common.NbLimbU32; i++ {
			limbByte := padBytes[i]
			if len(limbByte) == 0 {
				limbByte = []byte{0}
			}

			codesizes[i] = fmt.Sprintf("0x%s", hex.EncodeToString(limbByte))
		}

		res = append(res, RecordWrite{
			CFI:      record.CFI,
			ACC:      record.ACC,
			NBYTES:   record.NBYTES,
			COUNTER:  record.COUNTER,
			CODESIZE: codesizes,
		})
	}

	return res
}

func TestSplit(t *testing.T) {
	records, _ := readCSVFile("rom_input.csv")
	recordsToWrite := split(records)
	writeCSVFile("rom_input_new.csv", recordsToWrite)
	fmt.Println(recordsToWrite)
}
