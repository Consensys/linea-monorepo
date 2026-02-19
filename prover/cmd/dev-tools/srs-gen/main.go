package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	kzg377 "github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: srs-gen <size> <output-dir>\n")
		fmt.Fprintf(os.Stderr, "  Generates a BLS12-377 canonical SRS of the given size.\n")
		fmt.Fprintf(os.Stderr, "  Example: srs-gen 134217731 ./prover-assets/kzgsrs/\n")
		os.Exit(1)
	}

	size := new(big.Int)
	if _, ok := size.SetString(os.Args[1], 10); !ok {
		fmt.Fprintf(os.Stderr, "invalid size: %s\n", os.Args[1])
		os.Exit(1)
	}

	outDir := os.Args[2]

	tau, err := rand.Int(rand.Reader, ecc.BLS12_377.ScalarField())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate random tau: %v\n", err)
		os.Exit(1)
	}

	sizeU64 := size.Uint64()
	fmt.Printf("Generating BLS12-377 canonical SRS with %d points...\n", sizeU64)
	start := time.Now()

	srs, err := kzg377.NewSRS(sizeU64, tau)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate SRS: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SRS generated in %v\n", time.Since(start))

	filename := fmt.Sprintf("kzg_srs_canonical_%d_bls12377_aleo.memdump", sizeU64)
	outPath := filepath.Join(outDir, filename)

	fmt.Printf("Writing SRS to %s...\n", outPath)
	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file: %v\n", err)
		os.Exit(1)
	}

	err = errors.Join(srs.WriteDump(f), f.Close())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write SRS: %v\n", err)
		os.Exit(1)
	}

	fi, _ := os.Stat(outPath)
	fmt.Printf("Done. File size: %.2f GB\n", float64(fi.Size())/(1<<30))
}
