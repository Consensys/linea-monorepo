package vortex_koalabear

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/require"
)

func TestNoSisTransversalHashMatchesReference(t *testing.T) {
	testCases := []struct {
		name    string
		numRows int
		numCols int
	}{
		{name: "rows_387_cols_1024", numRows: 387, numCols: 1024},
		{name: "rows_160_cols_1024", numRows: 160, numCols: 1024},
		{name: "rows_8_cols_1024", numRows: 8, numCols: 1024},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := makeNoSisRows(tc.numRows, tc.numCols)
			params := NewParams(2, tc.numCols/2, tc.numRows, 6, 16)
			got := params.noSisTransversalHash(rows)
			want := referenceNoSisTransversalHash(rows)
			require.Equal(t, want, got)
		})
	}
}

func BenchmarkNoSisTransversalHash(b *testing.B) {
	testCases := []struct {
		name    string
		numRows int
		numCols int
	}{
		{name: "rows_387_cols_16384", numRows: 387, numCols: 1 << 14},
		{name: "rows_160_cols_16384", numRows: 160, numCols: 1 << 14},
		{name: "rows_8_cols_16384", numRows: 8, numCols: 1 << 14},
	}

	for _, tc := range testCases {
		rows := makeNoSisRows(tc.numRows, tc.numCols)
		params := NewParams(2, tc.numCols/2, tc.numRows, 6, 16)

		b.Run(fmt.Sprintf("optimized_%s", tc.name), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = params.noSisTransversalHash(rows)
			}
		})

		b.Run(fmt.Sprintf("reference_%s", tc.name), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = referenceNoSisTransversalHash(rows)
			}
		})
	}
}

func BenchmarkVortexHashPathsByRows(b *testing.B) {
	const numCols = 1 << 13

	// Covers the small, around-threshold, and large row counts seen in limitless.
	rowCounts := []int{
		4, 8, 12, 28,
		64, 80, 96, 104, 112, 123, 128, 136, 140, 159, 160,
		175, 193, 204, 208, 212, 226, 228, 257, 264, 270, 271,
		304, 312, 320, 352, 384, 387, 388, 396, 404, 410, 411,
		436, 440, 448, 454, 480, 492, 512, 576, 640, 768, 832,
		896, 960, 1024, 1152, 1280, 1504, 2208,
	}

	for _, numRows := range rowCounts {
		rows := makeNoSisRows(numRows, numCols)
		params := NewParams(2, numCols, numRows, 6, 16)
		encodedRows := params.EncodeRows(rows)
		paddedNoSisRows := ((numRows + 7) / 8) * 8
		sisChunks := (numRows + 31) / 32

		b.Run(fmt.Sprintf("rows_%d", numRows), func(b *testing.B) {
			b.Run("no_sis_hash_only", func(b *testing.B) {
				b.ReportAllocs()
				b.ReportMetric(float64(paddedNoSisRows), "padded_rows")
				b.ReportMetric(float64(sisChunks), "sis_chunks")
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_ = params.noSisTransversalHash(encodedRows)
				}
			})

			b.Run("sis_hash_only", func(b *testing.B) {
				b.ReportAllocs()
				b.ReportMetric(float64(paddedNoSisRows), "padded_rows")
				b.ReportMetric(float64(sisChunks), "sis_chunks")
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, _ = params.sisTransversalHash(encodedRows)
				}
			})
		})
	}
}

func TestIncrementalHasherMatchesTransversalHash(t *testing.T) {
	testCases := []struct {
		name      string
		numRows   int
		numCols   int
		batchSize int
	}{
		{"batch_equals_total", 32, 1024, 32},
		{"batch_1", 32, 1024, 1},
		{"batch_7_uneven", 32, 1024, 7},
		{"batch_16", 32, 1024, 16},
		{"large_rows_batch_50", 387, 1024, 50},
		{"large_rows_batch_256", 512, 1024, 256},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := makeNoSisRows(tc.numRows, tc.numCols)
			params := NewParams(2, tc.numCols, tc.numRows, 9, 16)
			encoded := params.EncodeRows(rows)
			nbEncodedCols := params.RsParams.NbEncodedColumns()

			// Reference: TransversalHash
			refHashes := make([]field.Element, nbEncodedCols*params.Key.OutputSize())
			refHashes = params.Key.TransversalHash(encoded, refHashes)

			// Streaming: IncrementalHasher
			hasher := params.Key.NewIncrementalHasher(nbEncodedCols)
			for start := 0; start < len(encoded); start += tc.batchSize {
				end := start + tc.batchSize
				if end > len(encoded) {
					end = len(encoded)
				}
				hasher.AbsorbBatch(encoded[start:end])
			}
			streamHashes := hasher.Finalize()

			require.Equal(t, len(refHashes), len(streamHashes))
			for i := range refHashes {
				require.Equal(t, refHashes[i], streamHashes[i],
					"mismatch at index %d (col %d, offset %d)",
					i, i/params.Key.OutputSize(), i%params.Key.OutputSize())
			}
		})
	}
}

func TestStreamingCommitMatchesNonStreaming(t *testing.T) {
	testCases := []struct {
		name      string
		numRows   int
		numCols   int
		batchSize int
	}{
		{"rows_32_batch_8", 32, 1 << 10, 8},
		{"rows_128_batch_64", 128, 1 << 10, 64},
		{"rows_387_batch_50", 387, 1 << 10, 50},
		{"rows_512_batch_256", 512, 1 << 10, 256},
		{"rows_128_batch_1", 128, 1 << 10, 1},
		{"rows_128_batch_larger_than_total", 128, 1 << 10, 999},
		{"rows_160_default_batch", 160, 1 << 10, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := makeNoSisRows(tc.numRows, tc.numCols)
			params := NewParams(2, tc.numCols, tc.numRows, 9, 16)

			// Reference: non-streaming
			encRef, commitRef, _, hashesRef := params.CommitMerkleWithSIS(rows)

			// Streaming
			encStr, commitStr, _, hashesStr := params.CommitMerkleWithSISStreaming(rows, tc.batchSize)

			// Check identical commitment (Merkle root)
			require.Equal(t, commitRef, commitStr, "Merkle root mismatch")

			// Check identical SIS hashes
			require.Equal(t, hashesRef, hashesStr, "SIS hashes mismatch")

			// Check identical encoded matrix
			require.Equal(t, len(encRef), len(encStr), "encoded matrix row count mismatch")
			for i := range encRef {
				require.Equal(t, encRef[i].Len(), encStr[i].Len(),
					"encoded row %d length mismatch", i)
				for j := 0; j < encRef[i].Len(); j++ {
					require.Equal(t, encRef[i].Get(j), encStr[i].Get(j),
						"encoded matrix mismatch at row %d, col %d", i, j)
				}
			}
		})
	}
}

func TestStreamingL2CommitMatchesNonStreaming(t *testing.T) {
	testCases := []struct {
		name      string
		numRows   int
		numCols   int
		batchSize int
	}{
		{"rows_32_batch_8", 32, 1 << 10, 8},
		{"rows_128_batch_64", 128, 1 << 10, 64},
		{"rows_387_batch_50", 387, 1 << 10, 50},
		{"rows_512_batch_256", 512, 1 << 10, 256},
		{"rows_128_batch_1", 128, 1 << 10, 1},
		{"rows_128_batch_larger_than_total", 128, 1 << 10, 999},
		{"rows_160_default_batch", 160, 1 << 10, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := makeNoSisRows(tc.numRows, tc.numCols)
			params := NewParams(2, tc.numCols, tc.numRows, 9, 16)

			// Reference: non-streaming
			_, commitRef, _, hashesRef := params.CommitMerkleWithSIS(rows)

			// Level 2 streaming (no W' materialization)
			origMatrix, commitL2, _, hashesL2 := params.CommitMerkleWithSISStreamingL2(rows, tc.batchSize)

			// Check identical commitment (Merkle root)
			require.Equal(t, commitRef, commitL2, "Merkle root mismatch")

			// Check identical SIS hashes
			require.Equal(t, hashesRef, hashesL2, "SIS hashes mismatch")

			// Check OriginalMatrix wraps the correct data
			require.Equal(t, len(rows), len(origMatrix.Rows), "original matrix row count mismatch")
			require.NotNil(t, origMatrix.Params, "params should not be nil")
		})
	}
}

func TestOriginalMatrixExtractColumns(t *testing.T) {
	testCases := []struct {
		name    string
		numRows int
		numCols int
	}{
		{"rows_32", 32, 1 << 10},
		{"rows_128", 128, 1 << 10},
		{"rows_387", 387, 1 << 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := makeNoSisRows(tc.numRows, tc.numCols)
			params := NewParams(2, tc.numCols, tc.numRows, 9, 16)

			// Get reference encoded matrix
			encoded := params.EncodeRows(rows)

			// Create OriginalMatrix
			origMatrix := &OriginalMatrix{Rows: rows, Params: &params}

			// Select some column positions
			entryList := []int{0, 5, 42, 100, params.RsParams.NbEncodedColumns() - 1}

			// Extract columns via OriginalMatrix
			gotCols := origMatrix.ExtractColumns(entryList)

			// Reference: extract from encoded matrix
			for j, entry := range entryList {
				for k := 0; k < tc.numRows; k++ {
					require.Equal(t, encoded[k].Get(entry), gotCols[j][k],
						"mismatch at entry %d, row %d", entry, k)
				}
			}

			// Also test ToEncodedMatrix roundtrip
			reEncoded := origMatrix.ToEncodedMatrix()
			require.Equal(t, len(encoded), len(reEncoded))
			for i := range encoded {
				for j := 0; j < encoded[i].Len(); j++ {
					require.Equal(t, encoded[i].Get(j), reEncoded[i].Get(j),
						"ToEncodedMatrix mismatch at row %d, col %d", i, j)
				}
			}
		})
	}
}

func BenchmarkCommitMerkleWithSIS(b *testing.B) {
	const numCols = 1 << 13

	rowCounts := []int{256, 512, 768, 1024, 1504, 2208}

	for _, numRows := range rowCounts {
		rows := makeNoSisRows(numRows, numCols)
		params := NewParams(2, numCols, numRows, 9, 16)

		b.Run(fmt.Sprintf("full_commit/rows_%d", numRows), func(b *testing.B) {
			b.ReportAllocs()
			b.ReportMetric(float64(numRows), "rows")
			b.ReportMetric(float64(numCols*2), "encoded_cols")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _, _ = params.CommitMerkleWithSIS(rows)
			}
		})

		b.Run(fmt.Sprintf("encode_rows_only/rows_%d", numRows), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = params.EncodeRows(rows)
			}
		})

		b.Run(fmt.Sprintf("sis_hash_only/rows_%d", numRows), func(b *testing.B) {
			encoded := params.EncodeRows(rows)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = params.sisTransversalHash(encoded)
			}
		})
	}
}

func BenchmarkCommitMerkleWithSISStreaming(b *testing.B) {
	const numCols = 1 << 13

	rowCounts := []int{256, 512, 768, 1024, 1504, 2208}

	for _, numRows := range rowCounts {
		rows := makeNoSisRows(numRows, numCols)
		params := NewParams(2, numCols, numRows, 9, 16)

		batchDivisors := []int{1, 2, 4, 8, 16}
		for _, div := range batchDivisors {
			batchSize := numRows / div
			if batchSize < 1 {
				batchSize = 1
			}
			b.Run(fmt.Sprintf("rows_%d/batch_%d", numRows, batchSize), func(b *testing.B) {
				b.ReportAllocs()
				b.ReportMetric(float64(numRows), "rows")
				b.ReportMetric(float64(batchSize), "batch_size")
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, _, _, _ = params.CommitMerkleWithSISStreaming(rows, batchSize)
				}
			})
		}
	}
}

func BenchmarkCommitMerkleWithSISStreamingL2(b *testing.B) {
	const numCols = 1 << 13

	rowCounts := []int{256, 512, 768, 1024, 1504, 2208}

	for _, numRows := range rowCounts {
		rows := makeNoSisRows(numRows, numCols)
		params := NewParams(2, numCols, numRows, 9, 16)

		batchDivisors := []int{1, 2, 4, 8}
		for _, div := range batchDivisors {
			batchSize := numRows / div
			if batchSize < 1 {
				batchSize = 1
			}
			b.Run(fmt.Sprintf("rows_%d/batch_%d", numRows, batchSize), func(b *testing.B) {
				b.ReportAllocs()
				b.ReportMetric(float64(numRows), "rows")
				b.ReportMetric(float64(batchSize), "batch_size")
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, _, _, _ = params.CommitMerkleWithSISStreamingL2(rows, batchSize)
				}
			})
		}
	}
}

func referenceNoSisTransversalHash(v []smartvectors.SmartVector) []field.Octuplet {
	nbRows := len(v)
	nbCols := v[0].Len()
	res := make([]field.Octuplet, nbCols)
	curCol := make([]field.Element, nbRows)
	h := poseidon2_koalabear.NewMDHasher()
	for col := 0; col < nbCols; col++ {
		for row := 0; row < nbRows; row++ {
			curCol[row] = v[row].Get(col)
		}
		h.WriteElements(curCol...)
		res[col] = h.SumElement()
		h.Reset()
	}
	return res
}

func makeNoSisRows(numRows, numCols int) []smartvectors.SmartVector {
	rows := make([]smartvectors.SmartVector, numRows)
	for row := 0; row < numRows; row++ {
		switch {
		case row%17 == 0:
			rows[row] = smartvectors.NewConstant(field.NewElement(uint64(row+1)), numCols)
		case row%11 == 0:
			windowLen := numCols / 4
			window := make([]field.Element, windowLen)
			for i := range window {
				window[i] = field.NewElement(uint64((row+1)*(i+3) + 7))
			}
			rows[row] = smartvectors.RightPadded(window, field.NewElement(uint64(row+5)), numCols)
		default:
			vec := make([]field.Element, numCols)
			for i := range vec {
				vec[i] = field.NewElement(uint64((row+1)*(i+1) + 13))
			}
			rows[row] = smartvectors.NewRegular(vec)
		}
	}
	return rows
}
