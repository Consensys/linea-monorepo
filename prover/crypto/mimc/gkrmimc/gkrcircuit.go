package gkrmimc

import (
	"strconv"
	"strings"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	cGkr "github.com/consensys/gnark-crypto/ecc/bn254/fr/gkr"
	"github.com/consensys/gnark/constraint"
	cs "github.com/consensys/gnark/constraint/bn254"
	"github.com/consensys/gnark/frontend"
	gGkr "github.com/consensys/gnark/std/gkr"
	"github.com/consensys/gnark/std/hash"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/multicommit"
	"github.com/sirupsen/logrus"
)

var (
	prefetchSize     = 4 // all layers before the "actual" MiMc round layers
	numGates     int = prefetchSize + len(mimc.Constants)
	gateNames    []string
)

func init() {
	// Registers the names of the GKR gates into the
	// global gkr registry.
	createGateNames()
	registerGates()

	// Registers the mimc hash function in the hash builder
	// registry.
	hash.BuilderRegistry["mimc"] = func(api frontend.API) (hash.FieldHasher, error) {
		h, err := gmimc.NewMiMC(api)
		return &h, err
	}

	// Registers the hasher to be used in the GKR prover
	cs.HashBuilderRegistry["mimc"] = mimc.NewMiMC
}

func writePaddedHex(sbb *strings.Builder, n, nbDigits int) {
	hex := strconv.FormatInt(int64(n), 16)
	sbb.WriteString(strings.Repeat("0", nbDigits-len(hex)))
	sbb.WriteString(hex)
}

func createGateNames() {
	nbDigits := 0
	for i := numGates; i > 0; i /= 16 {
		nbDigits++
	}

	gateNames = make([]string, numGates)

	gateNames[0] = "input" // no name necessary. will map to nil. this is just for clarity
	gateNames[1] = "input"
	gateNames[2] = "identity" // actual identity gate
	gateNames[3] = "identity"

	for i := 4; i < numGates; i++ {
		var gateNameBuilder strings.Builder
		gateNameBuilder.WriteString("mimc")
		writePaddedHex(&gateNameBuilder, i, nbDigits)
		gateNames[i] = gateNameBuilder.String()
	}
}

func registerGates() {
	for i := 4; i < numGates-1; i++ {
		name := gateNames[i]
		gGkr.Gates[name] = NewRoundGateGnark(mimc.Constants[i-prefetchSize])
		cGkr.Gates[name] = NewRoundGateCrypto(mimc.Constants[i-prefetchSize])
	}

	name := gateNames[numGates-1]
	gGkr.Gates[name] = NewFinalRoundGateGnark(mimc.Constants[len(mimc.Constants)-1])
	cGkr.Gates[name] = NewFinalRoundGateCrypto(mimc.Constants[len(mimc.Constants)-1])
}

// gkrMiMC a gadget that defines a MiMC
func gkrMiMC(gkr *gGkr.API, initStates, blocks []frontend.Variable) (constraint.GkrVariable, error) {

	var err error
	v := make([]constraint.GkrVariable, numGates-1)
	if v[0], err = gkr.Import(initStates); err != nil {
		panic(err)
	}
	if v[1], err = gkr.Import(blocks); err != nil {
		panic(err)
	}

	v[2] = gkr.NamedGate("identity", v[0])
	v[3] = gkr.NamedGate("identity", v[1])

	for i := 4; i < numGates-1; i++ {
		v[i] = gkr.NamedGate(gateNames[i], v[2], v[i-1])
	}

	res := gkr.NamedGate(gateNames[numGates-1], v[2], v[3], v[numGates-2])

	return res, nil
}

// Toy compression: just summing them up. TODO: Replace with actual compression
// instances are the inner indexes
func checkWithGkr(api frontend.API, initStates, blocks, allegedNewState []frontend.Variable) {

	gkr := gGkr.NewApi()

	D, err := gkrMiMC(gkr, initStates, blocks)
	if err != nil {
		panic(err)
	}

	// This creates a placeholder that will contain the GKR assignment
	var solution gGkr.Solution
	if solution, err = gkr.Solve(api); err != nil {
		panic(err)
	}

	// builds the string of data that we need for the initial randomness. Note
	// that since this strings of data contains the full transcript of the rest
	// of the protocol
	feedToInitialRand := append(initStates, blocks...)
	feedToInitialRand = append(feedToInitialRand, allegedNewState...)

	multicommit.WithCommitment(
		api,
		func(api frontend.API, initialChallenge frontend.Variable) error {
			logrus.Infof("defining the constraints of the GKR verifier")

			// "mimc" means that we are using MiMC hashes to compute the FS challenges
			// this part is responsible for verifying the GKR proof.
			err = solution.Verify("mimc", initialChallenge)
			if err != nil {
				panic(err)
			}

			logrus.Infof("defining the constraints of the GKR verifier : done")

			// Export the last gkr layer as an array of frontend variable
			d := solution.Export(D)
			if len(d) != len(allegedNewState) {
				utils.Panic("length mismatch %v != %v", len(d), len(allegedNewState))
			}

			for i := range d {
				// Ensures GKR and
				api.AssertIsEqual(d[i], allegedNewState[i])
			}

			return nil
		},
		feedToInitialRand...,
	)

}
