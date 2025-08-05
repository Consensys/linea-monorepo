package main

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"iter"
	"math/big"
	"os"
	"slices"
	"strconv"

	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	fr_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	kzg_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/kzg"
)

const (
	// blobCommitmentVersionKZG is the version byte for the KZG point evaluation precompile.
	blobCommitmentVersionKZG uint8 = 0x01
	// evmBlockSize is the size of the SRS used in the KZG precompile. This
	// defines the polynomial degree and therefore the size of the blob. It is also the expected
	// return value of the POINTEVAL precompile.
	evmBlockSize = 4096
)

var (
	evmBlsModulus = fr_bls12381.Modulus()
)

type pointEvalInput struct {
	versionedHashCase      pointEvalInputHashType
	evaluationPointCase    pointEvalInputScalarType
	claimedValueCase       pointEvalInputScalarType
	commitmentCase         pointEvalInputG1ElementType
	proofCase              pointEvalInputG1ElementType
	expectedBlobSizeCase   pointEvalInputExpectedBlobSizeType
	expectedBlsModulusCase pointEvalInputExpectedModulusType

	inputType pointEvalInputType // it is usually invalid, only in case where previous fields are valid in which case we toggle this value

	VersionedHash      [32]byte
	EvaluationPoint    *big.Int
	ClaimedValue       *big.Int
	Commitment         [bls12381.SizeOfG1AffineCompressed]byte
	Proof              [bls12381.SizeOfG1AffineCompressed]byte
	ExpectedBlobSize   *big.Int
	ExpectedBlsModulus *big.Int

	// arithmetization determines where to send the input to. In case evaluation
	// point or claimed value overflows, or the expected results are incorrect, it is sent to neither.
	ToPointEvalCircuit        bool // if the input is valid, we send it to the point evaluation circuit
	ToPointEvalInvalidCircuit bool // if the input is invalid, we send it to the invalid circuit
}

func (i *pointEvalInput) generate() (bool, error) {
	// we first generate a valid input, then modify the values according to the input type
	randPoly := make(fr_bls12381.Vector, evmBlockSize)
	switch i.inputType {
	case validPointEval:
		randPoly.MustSetRandom()
	case validPointEvalZero:
		for i := range randPoly {
			randPoly[i].SetZero()
		}
	}
	// TODO: we should generate a valid poly such that the claimedValue could overflow
	// TODO: we should generate a valid poly such that the commitment Y is small or large as needed
	kzgCommitment, err := kzg_bls12381.Commit(randPoly, *srsPk)
	if err != nil {
		return false, fmt.Errorf("failed to commit random polynomial: %w", err)
	}
	// TODO: we should mutate bytes such that the encoded commitment is as needed
	i.Commitment = kzgCommitment.Bytes()
	var evaluationPoint fr_bls12381.Element
	switch i.evaluationPointCase {
	case validPointEvalScalar:
		evaluationPoint.MustSetRandom()
		i.EvaluationPoint = evaluationPoint.BigInt(new(big.Int))
	case validPointEvalScalar_Zero:
		evaluationPoint.SetZero()
		i.EvaluationPoint = big.NewInt(0)
	case invalidPointEvalScalar_Overflow:
		i.EvaluationPoint = generateScalar(msmScalarBig)
		evaluationPoint.SetBigInt(i.EvaluationPoint)
	}
	// TODO: we should generate inputs such that the proof Y is small or large as needed
	kzgProof, err := kzg_bls12381.Open(randPoly, evaluationPoint, *srsPk)
	if err != nil {
		return false, fmt.Errorf("failed to open random polynomial: %w", err)
	}
	// TODO: we should mutate bytes such that the encoded proof is as needed
	i.Proof = kzgProof.H.Bytes()
	// TODO: modify according to the test case
	i.ClaimedValue = evaluationPoint.BigInt(new(big.Int))
	// TODO: modify hash according to the test case
	i.VersionedHash = sha256.Sum256(i.Commitment[:])
	i.VersionedHash[0] = blobCommitmentVersionKZG
	// TODO: modify expected blob size according to the test case
	i.ExpectedBlobSize = big.NewInt(evmBlockSize)
	// TODO: modify expected modulus according to the test case
	i.ExpectedBlsModulus = new(big.Int).Set(evmBlsModulus)

	i.ToPointEvalCircuit = true
	i.ToPointEvalInvalidCircuit = false
	return true, nil
}

func generatePointEvalInput() iter.Seq2[int, pointEvalInput] {
	return func(yield func(int, pointEvalInput) bool) {
		var id int
		// TODO: currently we generate only valid cases, later all cases
		for _, versionedHashCase := range validPointEvalHashes {
			for _, evaluationPointCase := range validPointEvalScalars {
				for _, claimedValueCase := range validPointEvalScalars {
					for _, commitmentCase := range validPointEvalG1Elements {
						for _, proofCase := range validPointEvalG1Elements {
							for _, expectedBlobSizeCase := range validPointEvalExpectedBlobSizes {
								for _, expectedBlsModulusCase := range validPointEvalExpectedModuli {
									for _, inputType := range validPointEvals {
										input := pointEvalInput{
											versionedHashCase:      versionedHashCase,
											evaluationPointCase:    evaluationPointCase,
											claimedValueCase:       claimedValueCase,
											commitmentCase:         commitmentCase,
											proofCase:              proofCase,
											expectedBlobSizeCase:   expectedBlobSizeCase,
											expectedBlsModulusCase: expectedBlsModulusCase,
											inputType:              inputType,
										}
										ok, err := input.generate()
										if err != nil {
											panic(fmt.Sprintf("failed to generate point evaluation input: %v", err))
										}
										if !ok {
											continue
										}
										if !yield(id, input) {
											return
										}
										id++
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func (i *pointEvalInput) WriteCSV(w *csv.Writer, id int) error {
	dataLimbs := slices.Concat(
		splitVersionedHashToLimbs(i.VersionedHash),
		splitScalarToLimbs(i.EvaluationPoint),
		splitScalarToLimbs(i.ClaimedValue),
		splitG1CompressedLimbs(i.Commitment),
		splitG1CompressedLimbs(i.Proof),
	)
	resultLimbs := slices.Concat(
		splitScalarToLimbs(i.ExpectedBlobSize),
		splitScalarToLimbs(i.ExpectedBlsModulus),
	)

	for index, v := range dataLimbs {
		if err := w.Write([]string{
			strconv.Itoa(id),
			"1",                                   // DATA
			"0",                                   // RSTL
			strconv.Itoa(index),                   // INDEX
			strconv.Itoa(index),                   // CT
			"1",                                   // IS_FIRST_INPUT
			"0",                                   // IS_SECOND_INPUT
			formatBoolAsInt(i.ToPointEvalCircuit), // CIRCUIT_SELECTOR_POINTEVAL
			formatBoolAsInt(i.ToPointEvalInvalidCircuit), // CIRCUIT_SELECTOR_POINTEVAL_FAILURE
			v, // LIMB
		}); err != nil {
			return fmt.Errorf("write versioned hash limb: %w", err)
		}
		index++
	}
	for index, v := range resultLimbs {
		if err := w.Write([]string{
			strconv.Itoa(id),
			"0",                                   // DATA
			"1",                                   // RSTL
			strconv.Itoa(index),                   // INDEX
			strconv.Itoa(index),                   // CT
			"0",                                   // IS_FIRST_INPUT
			"0",                                   // IS_SECOND_INPUT
			formatBoolAsInt(i.ToPointEvalCircuit), // CIRCUIT_SELECTOR_POINTEVAL
			formatBoolAsInt(i.ToPointEvalInvalidCircuit), // CIRCUIT_SELECTOR_POINTEVAL_FAILURE
			v, // LIMB
		}); err != nil {
			return fmt.Errorf("write expected blob size limb: %w", err)
		}
		index++
	}
	return nil
}

func headerPointEval() []string {
	return []string{
		"ID",
		"DATA_POINT_EVALUATION",
		"RSLT_POINT_EVALUATION",
		"INDEX",
		"CT",
		"IS_FIRST_INPUT",
		"IS_SECOND_INPUT",
		"CIRCUIT_SELECTOR_POINT_EVALUATION",
		"CIRCUIT_SELECTOR_POINT_EVALUATION_FAILURE",
		"LIMB",
	}
}

func mainPointEval() error {
	f, err := os.Create(path_pointeval)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(headerPointEval()); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	id := 0
	for _, input := range generatePointEvalInput() {
		if err := input.WriteCSV(w, id); err != nil {
			return fmt.Errorf("write point evaluation input: %w", err)
		}
		id++
	}
	return nil
}
