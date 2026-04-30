package main

import (
	"debug/elf"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	RISCV_PROGRAM        = "riscV_program"
	RISCV_PROGRAM_LEN    = "riscV_program_length"
	RISCV_INBYTES        = "in_bytes"
	RISCV_INBYTES_LEN    = "in_bytes_length"
	RISCV_PROGRAM_OFFSET = "riscV_program_offset"
	RISCV_INBYTES_OFFSET = "in_bytes_offset"
	RISCV_ENTRY_POINT    = "entry_point"
)

// The purpose of this program is simply to generate a suitable ZkC json input
// file for a given RISC-V binary program.
func main() {
	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, "usage: go run main.go <elfFile> <inputs> <programOffset> <inputsOffset> <entryPoint>")
		os.Exit(1)
	}

	f, err := elf.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	// extract text section
	var text = extractTextBytes(f.Sections)
	// input
	var input []byte
	inputStr := os.Args[2]
	if strings.HasPrefix(inputStr, "0x") || strings.HasPrefix(inputStr, "0X") {
		input, err = hex.DecodeString(inputStr[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error decoding hex input: %v\n", err)
			os.Exit(1)
		}
	} else {
		input = []byte(inputStr)
	}
	// offsets
	var programOffset uint64
	var inputsOffset uint64
	var entryPoint uint64
	programOffset, err = strconv.ParseUint(os.Args[3], 0, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading programOffset: %v\n", err)
		os.Exit(1)
	}
	inputsOffset, err = strconv.ParseUint(os.Args[4], 0, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading inputsOffset: %v\n", err)
		os.Exit(1)
	}
	entryPoint, err = strconv.ParseUint(os.Args[5], 0, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading entryPoint: %v\n", err)
		os.Exit(1)
	}
	printJson(text, input, programOffset, inputsOffset, entryPoint)
}

// Extract the bytes of the text section.
func extractTextBytes(sections []*elf.Section) []byte {
	for _, s := range sections {
		if s.Name == ".text" {
			data, err := s.Data()
			// Sanity check
			if err != nil {
				panic(err.Error())
			}
			//
			return data
		}
	}
	//
	panic("no text section found!")
}

func printJson(text, input []byte, programOffset, inputsOffset, entryPoint uint64) {
	// Convert text bytes into a hex string
	var (
		textString          = hex.EncodeToString(text)
		textLenString       = fmt.Sprintf("%016x", len(text))
		inputsString        = hex.EncodeToString(input)
		inputLenString      = fmt.Sprintf("%016x", len(input))
		programOffsetString = fmt.Sprintf("%016x", programOffset)
		inputsOffsetString  = fmt.Sprintf("%016x", inputsOffset)
		entryPointString    = fmt.Sprintf("%016x", entryPoint)
	)
	//
	if len(text)%4 != 0 {
		panic("text section length not multiple of 4")
	}
	//
	fmt.Println("{")
	fmt.Printf("\t\"%s\": \"0x%s\",\n", RISCV_PROGRAM, textString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", RISCV_PROGRAM_LEN, textLenString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", RISCV_INBYTES, inputsString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", RISCV_INBYTES_LEN, inputLenString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", RISCV_PROGRAM_OFFSET, programOffsetString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", RISCV_INBYTES_OFFSET, inputsOffsetString)
	fmt.Printf("\t\"%s\": \"0x%s\"\n", RISCV_ENTRY_POINT, entryPointString)
	fmt.Println("}")
}
