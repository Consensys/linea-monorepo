package zkcdriver_test

import (
	"testing"

	zkc_util "github.com/consensys/go-corset/pkg/zkc/util"
	"github.com/consensys/linea-monorepo/prover-ray/utils/files"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/consensys/linea-monorepo/prover-ray/zkcdriver"
)

func TestExample(t *testing.T) {
	sys := wiop.NewSystemf("zkc-test")
	// set an example input
	var (
		// define an example input file
		json = "{\"data\": \"0x0041_0042\" }"
		// construct inputs map
		input zkcdriver.PreReadInputs
	)
	// parse example input
	input.Inputs, input.Err = zkc_util.ParseJsonInputFile([]byte(json))
	// Sanity check
	if input.Err != nil {
		t.Errorf("error parsing program inputs (%v)", input.Err)
		t.FailNow()
	}
	// initialise round
	sys.NewRound()
	// construct ZkC driver
	driver := zkcdriver.NewZkCDriver(
		sys,
		zkcdriver.Settings{},
		files.MustRead("./testdata/zkc_example.bin"))
	// FIXME: not sure how best to instantiate runtime here?
	driver.AssignWithPreRead(&wiop.Runtime{}, input)
	// FIXME: run the prover to complete the test.  For now, I just used
	// go-corset's internal check to illustrate how this can work (e.g. it might
	// be useful for debugging). Run go-corset constraint check
	runZkcConstraintCheck(t, driver, input.Inputs)
}

func runZkcConstraintCheck(t *testing.T, driver *zkcdriver.ZkCDriver, input map[string][]byte) {
	// trace program with given input
	tr, errs := driver.BinaryFile.Trace(input, driver.TracingConfig)
	// sanity check
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("%s", err.Error())
		}
		t.FailNow()
	}
	//
	if errs := driver.BinaryFile.Check(tr, driver.TracingConfig); len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("%s", err.Message())
		}
		t.FailNow()
	}
}
