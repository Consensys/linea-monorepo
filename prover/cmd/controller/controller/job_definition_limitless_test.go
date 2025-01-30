package controller

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/stretchr/testify/assert"
)

// This tests ensures that the naming convention is respected by the file-watcher
// i.e., files with the right naming only are recognized. And the corresponding
// output files are also recognized.
func TestBootstrapSubModInFileRegexp(t *testing.T) {

	var (
		correctM           = "102-103-etv0.2.3-stv1.2.3-getZkProof.json"
		correctL           = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.large"
		correctWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_77"
		correctWithFailL   = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_77"
		correctWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_77.failure.code_77"
		correctWith2FailsL = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_77.failure.code_77"
		missingEtv         = "102-103-stv1.2.3-getZkProof.json"
		missingStv         = "102-103-etv0.2.3-getZkProof.json"
		notAPoint          = "102-103-etv0.2.3-getZkProofAjson"
		badName            = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	// The responses in case of success
	var (
		respM = "responses/102-103-getZkProof_Bootstrap_Submodule.json"
		respL = "responses/102-103-getZkProof_Bootstrap_Submodule.json"
		// #nosec G101 -- Not a credential
		respWithFailM = "responses/102-103-getZkProof_Bootstrap_Submodule.json"
		// #nosec G101 -- Not a credential
		respWithFailL = "responses/102-103-getZkProof_Bootstrap_Submodule.json"
		// #nosec G101 -- Not a credential
		respWith2FailsM = "responses/102-103-getZkProof_Bootstrap_Submodule.json"
		// #nosec G101 -- Not a credential
		respWith2FailsL = "responses/102-103-getZkProof_Bootstrap_Submodule.json"
	)

	// The rename in case it is deferred to the large prover
	var (
		toLargeM           = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_137"
		toLargeWithFailM   = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_137"
		toLargeWith2FailsM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_137"
		toLargeWoEtv       = "requests/102-103-stv1.2.3-getZkProof.json.large.failure.code_137"
		toLargeWoStv       = "requests/102-103-etv0.2.3-getZkProof.json.large.failure.code_137"
	)

	// The rename in case it is a success
	var (
		successM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
		successMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof.json.success"
		successtWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof.json.success"
		successL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.success"
		successWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
		successWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.success"
		successWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
		successWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.success"
	)

	// The rename in case it is a panic (code = 2)
	var (
		failM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
		failMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof.json.failure.code_2"
		failtWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof.json.failure.code_2"
		failL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_2"
		failWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
		failWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_2"
		failWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
		failWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_2"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correctM, correctWithFailM, correctWith2FailsM, missingEtv, missingStv},
			Explainer:      "happy path, case M",
			ExpectedOutput: []string{respM, respWithFailM, respWith2FailsM, respM, respM},
			ExpToLarge:     []string{toLargeM, toLargeWithFailM, toLargeWith2FailsM, toLargeWoEtv, toLargeWoStv},
			ExpSuccess:     []string{successM, successWithFailM, successWith2FailsM, successtWoEtv, successMWoStv},
			ExpFailW2:      []string{failM, failWithFailM, failWith2FailsM, failtWoEtv, failMWoStv},
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correctL, correctWithFailL, correctWith2FailsL},
			Explainer:      "happy path, case L",
			ExpectedOutput: []string{respL, respWithFailL, respWith2FailsL},
			ExpSuccess:     []string{successL, successWithFailL, successWith2FailsL},
			ExpFailW2:      []string{failL, failWithFailL, failWith2FailsL},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    []string{correctL, correctWithFailL, correctWith2FailsL},
			Explainer: "M does not pick the files reserved for L",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames:    []string{correctM, correctWithFailM, correctWith2FailsM},
			Explainer: "L does not pick the files reserved for M",
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    []string{notAPoint, badName},
			Explainer: "M does not pick obviously invalid files",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames:    []string{missingEtv, missingStv, notAPoint, badName},
			Explainer: "L does not pick obviously invalid files",
		},
	}

	for _, c := range testcase {

		//log.Printf("Running testcase:%s \n", c.Explainer)

		conf := config.Config{}
		conf.Version = "0.1.2"
		conf.Bootstrap_Submodule.CanRunFullLarge = c.Ext == "large"

		def, err := BootstrapSubModDefinition(&conf)
		assert.NoError(t, err)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, def, c)
		})
	}
}
