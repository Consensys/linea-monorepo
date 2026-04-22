package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/consensys/linea-monorepo/verifier-risc5/internal/guestabi"
	"github.com/consensys/linea-monorepo/verifier-risc5/internal/workload"
)

type inputSpec struct {
	Words    []string `json:"words"`
	Expected string   `json:"expected"`
}

func parseWord(value string) (uint64, error) {
	parsed, err := strconv.ParseUint(value, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %q: %w", value, err)
	}
	return parsed, nil
}

func defaultWords() []uint64 {
	words := make([]uint64, len(workload.DefaultWords))
	copy(words, workload.DefaultWords[:])
	return words
}

func loadSpec(path string) ([]uint64, uint64, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, fmt.Errorf("read %s: %w", path, err)
	}

	var spec inputSpec
	if err := json.Unmarshal(payload, &spec); err != nil {
		return nil, 0, fmt.Errorf("decode %s: %w", path, err)
	}

	words := defaultWords()
	if len(spec.Words) != 0 {
		words = make([]uint64, len(spec.Words))
		for i, value := range spec.Words {
			word, err := parseWord(value)
			if err != nil {
				return nil, 0, fmt.Errorf("words[%d]: %w", i, err)
			}
			words[i] = word
		}
	}

	if len(words) > guestabi.MaxWords {
		return nil, 0, fmt.Errorf("too many words: got %d, max %d", len(words), guestabi.MaxWords)
	}

	expected := workload.Compute(words)
	if spec.Expected != "" {
		expected, err = parseWord(spec.Expected)
		if err != nil {
			return nil, 0, fmt.Errorf("expected: %w", err)
		}
	}

	return words, expected, nil
}

func encode(words []uint64, expected uint64) []byte {
	buf := make([]byte, int(guestabi.HeaderSize)+len(words)*8)
	binary.LittleEndian.PutUint32(buf[0:], guestabi.Magic)
	binary.LittleEndian.PutUint32(buf[4:], guestabi.Version)
	binary.LittleEndian.PutUint32(buf[8:], uint32(len(words)))
	binary.LittleEndian.PutUint32(buf[12:], 0)
	binary.LittleEndian.PutUint64(buf[16:], expected)

	offset := int(guestabi.HeaderSize)
	for _, word := range words {
		binary.LittleEndian.PutUint64(buf[offset:], word)
		offset += 8
	}

	return buf
}

func main() {
	var inputPath string
	var outputPath string

	flag.StringVar(&inputPath, "in", "inputs/default.json", "input JSON specification")
	flag.StringVar(&outputPath, "out", "build/verifier-input.bin", "output binary blob")
	flag.Parse()

	words, expected, err := loadSpec(inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, encode(words, expected), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", outputPath, err)
		os.Exit(1)
	}

	fmt.Printf("wrote %d words with expected 0x%016x to %s\n", len(words), expected, outputPath)
}
