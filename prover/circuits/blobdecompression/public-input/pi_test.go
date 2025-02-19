//go:build !fuzzlight

package public_input

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/utils"

	"github.com/consensys/linea-monorepo/prover/maths/field"

	"github.com/consensys/gnark-crypto/ecc"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr/iop"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/test"
	"github.com/stretchr/testify/assert"
)

const (
	msbNbBits            = (fr381.Bits - 1) % 8
	msbMask              = 1<<msbNbBits - 1
	testDataDir          = "../../../../contracts/test/testData/compressedDataEip4844"
	testDataNoEip4844Dir = "../../../../contracts/test/testData/compressedData"
)

func createRandomBlobElems(n int) [][]byte {

	blob := make([][]byte, n)
	for i := range blob {
		blob[i] = make([]byte, 32)
		if _, err := rand.Read(blob[i]); err != nil {
			panic(err)
		}
		blob[i][0] &= msbMask
	}
	return blob
}

// TestInterpolateLagrange tests the EIP4844 consistency check against the gnark iop package
func TestInterpolateLagrange(t *testing.T) {

	assignment := func(unitCircleEvaluations []interface{}, evaluationPoint interface{}) *testInterpolateLagrangeCircuit {

		unitCircleEvaluationsFr := mapSlice(unitCircleEvaluations[:], func(i interface{}) fr381.Element {
			var res fr381.Element
			_, err := res.SetInterface(i)
			assert.NoError(t, err)
			return res
		})

		// randomize an evaluation point that fits in a bls12-377 fr element
		var (
			evaluationPointFr fr381.Element
		)
		_, err := evaluationPointFr.SetInterface(evaluationPoint)
		assert.NoError(t, err)

		assignment := testInterpolateLagrangeCircuit{
			UnitCircleEvaluationsBytes: make([][fr381.Bytes]frontend.Variable, len(unitCircleEvaluations)),
		}
		for i := range unitCircleEvaluations {
			bytes := unitCircleEvaluationsFr[i].Bytes()
			copy(assignment.UnitCircleEvaluationsBytes[i][:], utils.ToVariableSlice(bytes[:]))
		}

		// compute the evaluation using the iop package
		domain := fft.NewDomain(uint64(len(unitCircleEvaluations)))
		poly := iop.NewPolynomial(&unitCircleEvaluationsFr, iop.Form{Basis: iop.Lagrange, Layout: iop.Regular})
		poly.ToCanonical(domain)
		evaluation := poly.Evaluate(evaluationPointFr)
		evaluationBytes := evaluation.Bytes()
		var evaluationInt big.Int
		evaluationInt.SetBytes(evaluationBytes[:])

		scalars, err := internal.Bls12381ScalarToBls12377Scalars(evaluationPointFr)
		assert.NoError(t, err)
		utils.Copy(assignment.EvaluationPoint[:], scalars[:])

		scalars, err = internal.Bls12381ScalarToBls12377Scalars(evaluation)
		assert.NoError(t, err)
		utils.Copy(assignment.Evaluation[:], scalars[:])

		return &assignment
	}

	randomAssignment := func(n int) *testInterpolateLagrangeCircuit {
		unitCircleEvaluations := createRandomBlobElems(n)
		var evaluationPoint fr381.Element
		_, err := evaluationPoint.SetRandom()
		assert.NoError(t, err)
		return assignment(mapSlice(unitCircleEvaluations, func(b []byte) interface{} { return b }), evaluationPoint)
	}

	assignments := []*testInterpolateLagrangeCircuit{
		assignment([]interface{}{0, 4, 0, 0}, 0),
		assignment([]interface{}{"221797350491448557374835382936094284962079105806616932502871687042746686348", "5449307655738973627560541249222884206820629036484003960051650107064657316177"}, "6506134398774570609831295452620385261047212455886876242937577495553156355635"),
		assignment([]interface{}{1, 0}, 3),
	}

	for n := 4096; n > 2; n /= 32 {
		assignments = append(assignments, randomAssignment(n))
	}

	slices.SortFunc(assignments, func(a, b *testInterpolateLagrangeCircuit) int {
		return len(a.UnitCircleEvaluationsBytes) - len(b.UnitCircleEvaluationsBytes)
	})

	for i := range assignments {
		assert.NoError(t, test.IsSolved(
			&testInterpolateLagrangeCircuit{UnitCircleEvaluationsBytes: make([][fr381.Bytes]frontend.Variable, len(assignments[i].UnitCircleEvaluationsBytes))},
			assignments[i], ecc.BLS12_377.ScalarField(),
		))
	}
}

type testInterpolateLagrangeCircuit struct {
	EvaluationPoint            [2]frontend.Variable
	Evaluation                 [2]frontend.Variable
	UnitCircleEvaluationsBytes [][fr381.Bytes]frontend.Variable
}

func (c *testInterpolateLagrangeCircuit) Define(api frontend.API) error {
	field, err := emulated.NewField[emulated.BLS12381Fr](api)
	unitCircleEvaluations := mapSlice(c.UnitCircleEvaluationsBytes, func(v [fr381.Bytes]frontend.Variable) *emulated.Element[emulated.BLS12381Fr] {
		var bits [len(v) * 8]frontend.Variable
		for i := range v {
			copy(bits[i*8:], api.ToBinary(v[len(v)-1-i], 8))
		}
		return field.FromBits(bits[:]...)
	})
	evaluationPoint := newElementFromVars(api, c.EvaluationPoint)
	if err != nil {
		return err
	}

	evaluation, err := interpolateLagrangeBls12381(field, unitCircleEvaluations, evaluationPoint)
	if err != nil {
		return err
	}
	evaluationHL := bls12381ScalarToBls12377Scalars(api, evaluation)
	api.AssertIsEqual(c.Evaluation[0], evaluationHL[0])
	api.AssertIsEqual(c.Evaluation[1], evaluationHL[1])

	return nil
}

type blobConsistencyCheckCircuit struct {
	BlobBytes []frontend.Variable   // "the blob" in EIP-4844 parlance
	X         [32]frontend.Variable `gnark:",public"` // "high" and "low"
	Y         [2]frontend.Variable  `gnark:",public"`
}

func (c *blobConsistencyCheckCircuit) Define(api frontend.API) error {
	blobCrumbs := internal.PackedBytesToCrumbs(api, c.BlobBytes, fr381.Bits-1)
	y, err := VerifyBlobConsistency(api, blobCrumbs, c.X)
	if err != nil {
		return err
	}
	api.AssertIsEqual(y[0], c.Y[0])
	api.AssertIsEqual(y[1], c.Y[1])
	return nil
}

type BlobConsistencyCheckTestCase struct {
	Eip4844Enabled bool   `json:"eip4844Enabled"`
	CompressedData string `json:"compressedData"`
	ExpectedX      string `json:"expectedX"`
	ExpectedY      string `json:"expectedY"`
	SnarkHash      string `json:"snarkHash"`
}

func decodeHex(t *testing.T, s string) []byte {
	assert.Equal(t, s[:2], "0x")
	b, err := hex.DecodeString(s[2:])
	assert.NoError(t, err)
	return b
}

func decodeHexHL(t *testing.T, s string) (r [2]frontend.Variable) {
	b := decodeHex(t, s)
	assert.Equal(t, len(b), 32)

	scalars, err := internal.Bls12381ScalarToBls12377Scalars(b)
	assert.NoError(t, err)
	utils.Copy(r[:], scalars[:])

	return
}

func TestVerifyBlobConsistencyIntegration(t *testing.T) {
	circuit := &blobConsistencyCheckCircuit{
		BlobBytes: make([]frontend.Variable, 4096*32),
	}

	loadTestsInFolder := func(path string) {
		files, err := os.ReadDir(path)
		folderName := filepath.Base(path)
		assert.NoError(t, err)
		for _, file := range files {
			folderAndFile := filepath.Join(folderName, file.Name())
			if !strings.HasPrefix(file.Name(), "blocks-") || !strings.HasSuffix(file.Name(), ".json") {
				t.Logf("skipping \"%s\"", folderAndFile)
				continue
			}
			t.Logf("loading \"%s\"", folderAndFile)
			filePath := filepath.Join(path, file.Name())
			fileRaw, err := os.ReadFile(filePath)
			assert.NoError(t, err)
			var testCase BlobConsistencyCheckTestCase
			assert.NoError(t, json.Unmarshal(fileRaw, &testCase))

			if !testCase.Eip4844Enabled {
				t.Logf("skipping \"%s\" because eip4844 is disabled", file.Name())
				continue
			}

			var assignment blobConsistencyCheckCircuit

			blob, err := base64.StdEncoding.DecodeString(testCase.CompressedData)
			assert.NoError(t, err)
			assert.Zero(t, len(blob)%32, "blob not consisting of 32-byte field elements")
			assert.LessOrEqual(t, len(blob), 4096*32, "blob too large")
			blob = append(blob, make([]byte, 4096*32-len(blob))...) // pad if necessary
			assignment.BlobBytes = utils.ToVariableSlice(blob)

			utils.Copy(assignment.X[:], decodeHex(t, testCase.ExpectedX))
			assignment.Y = decodeHexHL(t, testCase.ExpectedY)

			t.Run(folderAndFile, func(t *testing.T) {
				assert.NoError(t, test.IsSolved(circuit, &assignment, ecc.BLS12_377.ScalarField()))
			})
		}
	}

	loadTestsInFolder(testDataDir)
	loadTestsInFolder(testDataNoEip4844Dir)

}

func TestCompileBlobConsistencyCheck(t *testing.T) {
	circuit := &blobConsistencyCheckCircuit{
		BlobBytes: make([]frontend.Variable, 4096*32),
	}
	cs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit, frontend.WithCapacity(8000000))
	assert.NoError(t, err)
	fmt.Println(cs.GetNbConstraints())
}

func TestVectorIopCompatibility(t *testing.T) {
	var testCase BlobConsistencyCheckTestCase
	fileRaw, err := os.ReadFile(filepath.Join(testDataDir, "blocks-1-46.json"))
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(fileRaw, &testCase))

	domain := fft.NewDomain(4096)
	blob, err := base64.StdEncoding.DecodeString(testCase.CompressedData)
	assert.NoError(t, err)
	blob = append(blob, make([]byte, 4096*32-len(blob))...) // pad if necessary
	blobElems := make([]fr381.Element, 4096)
	for i := range blobElems {
		assert.NoError(t, blobElems[bitReverse(i, 12)].SetBytesCanonical(blob[i*32:(i+1)*32]), i)
	}
	poly := iop.NewPolynomial(&blobElems, iop.Form{Basis: iop.Lagrange, Layout: iop.Regular})
	poly.ToCanonical(domain)

	var evaluationPoint fr381.Element
	evaluationPoint.SetBytes(decodeHex(t, testCase.ExpectedX))

	y := poly.Evaluate(evaluationPoint)

	// expectedY is re-encoded into a field so that we can compare y and expectedY
	// in text form without being affected by the left-most zeroes.
	expectedY := field.NewFromString(testCase.ExpectedY)
	assert.Equal(t, expectedY.Text(16), y.Text(16))
}

func TestFrConversions(t *testing.T) {
	testCases := []interface{}{
		"452312848583266388373324160190187140051835877600158453279131187530910662656", // 1 << 248
		"2657916159713343612780529581821191720363617578274368076333190323772338857867",
		append([]byte{64}, make([]byte, 31)...), // 2^254 or 0x40000...
	}

	for i := 0; i < 100; i++ {
		var b [32]byte
		_, err := rand.Read(b[:])
		assert.NoError(t, err)
		testCases = append(testCases, b[:])
	}

	options := []test.TestingOption{
		test.WithCurves(ecc.BLS12_377), test.WithBackends(backend.PLONK),
		test.NoTestEngine(),
	}

	var twoTo252 fr381.Element
	{
		i := big.NewInt(1)
		i.Lsh(i, 252)
		twoTo252.SetBigInt(i)
	}

	for _, testCase := range testCases {
		xPartitioned, err := internal.Bls12381ScalarToBls12377Scalars(testCase)
		assert.NoError(t, err)

		var xBack, tmp fr381.Element
		_, err = xBack.SetInterface(xPartitioned[0])
		assert.NoError(t, err)
		xBack.Mul(&xBack, &twoTo252)
		_, err = tmp.SetInterface(xPartitioned[1])
		assert.NoError(t, err)
		xBack.Add(&xBack, &tmp)

		_, err = tmp.SetInterface(testCase)
		assert.NoError(t, err)

		assert.Equal(t, tmp, xBack, fmt.Sprintf("out-of-snark conversion round-trip failed on %s or 0x%s", tmp.Text(10), tmp.Text(16)))
		var assignment testFrConversionCircuit
		utils.Copy(assignment.X[:], xPartitioned[:])
		options = append(options, test.WithValidAssignment(&assignment))
	}

	test.NewAssert(t).CheckCircuit(&testFrConversionCircuit{}, options...)
}

type testFrConversionCircuit struct {
	X [2]frontend.Variable
}

func (c *testFrConversionCircuit) Define(api frontend.API) error {
	x := newElementFromVars(api, c.X)
	xBack := bls12381ScalarToBls12377Scalars(api, x)
	api.AssertIsEqual(c.X[0], xBack[0])
	api.AssertIsEqual(c.X[1], xBack[1])
	return nil
}

func TestPackCrumbEmulated(t *testing.T) {
	var bytes [fr381.Bytes]byte
	_, err := rand.Read(bytes[:])
	assert.NoError(t, err)
	bytes[0] &= msbMask
	var assignment testPackCrumbEmulatedCircuit
	copy(assignment.Bytes[:], utils.ToVariableSlice(bytes[:]))
	test.NewAssert(t).CheckCircuit(
		&testPackCrumbEmulatedCircuit{}, test.WithValidAssignment(&assignment), test.WithCurves(ecc.BLS12_377), test.WithBackends(backend.PLONK),
		test.NoTestEngine(),
	)
}

type testPackCrumbEmulatedCircuit struct {
	Bytes [fr381.Bytes]frontend.Variable // big endian
}

func (c *testPackCrumbEmulatedCircuit) Define(api frontend.API) error {
	var bits [fr381.Bits - 1]frontend.Variable // big endian
	copy(bits[:], api.ToBinary(c.Bytes[0], msbNbBits))
	for i := 1; i < len(c.Bytes); i++ {
		copy(bits[msbNbBits+8*i-8:], api.ToBinary(c.Bytes[i], 8))
	}

	var crumbs [len(bits) / 2]frontend.Variable
	for i := range crumbs {
		crumbs[i] = api.Add(api.Mul(bits[2*i], 2), bits[2*i+1])
	}

	slices.Reverse(bits[:]) // now little endian

	field, err := emulated.NewField[emulated.BLS12381Fr](api)
	if err != nil {
		return err
	}
	expected := field.FromBits(bits[:]...)

	actual := packCrumbsEmulated(api, crumbs[:])
	if len(actual) != 1 {
		return fmt.Errorf("expected one element, got %d", len(actual))
	}

	field.AssertIsEqual(expected, actual[0])

	return nil
}
