package controller

// Remember for the limitless prover jobs, there will be no large deferrals here

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
)

// This test ensures that the naming convention is respected by the file-watcher
// for bootstrap jobs: only valid inputs are recognized, and the corresponding
// output files are also generated.
func TestBootstrapInFileRegexp(t *testing.T) {

	var (
		correctM           = "102-103-etv0.2.3-stv1.2.3-getZkProof.json"
		correctWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_77"
		correctWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_77.failure.code_77"
		missingEtv         = "102-103-stv1.2.3-getZkProof.json"
		missingStv         = "102-103-etv0.2.3-getZkProof.json"
		notAPoint          = "102-103-etv0.2.3-getZkProofAjson"
		badName            = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	// The requests in case of success
	var (
		respM         = "requests/102-103-metadata-getZkProof.json"
		respWithFailM = "requests/102-103-metadata-getZkProof.json"
		// #nosec G101 -- Not a credential
		respWith2FailsM = "requests/102-103-metadata-getZkProof.json"
	)

	// The rename in case it is a success
	var (
		successM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
		successMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof.json.success"
		successtWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof.json.success"
		successWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
		successWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
	)

	// The rename in case it is a panic (code = 2)
	var (
		failM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
		failMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof.json.failure.code_2"
		failtWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof.json.failure.code_2"
		failWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
		failWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correctM, correctWithFailM, correctWith2FailsM, missingEtv, missingStv},
			Explainer:      "happy path, case M",
			ExpectedOutput: []string{respM, respWithFailM, respWith2FailsM, respM, respM},
			ExpSuccess:     []string{successM, successWithFailM, successWith2FailsM, successtWoEtv, successMWoStv},
			ExpFailW2:      []string{failM, failWithFailM, failWith2FailsM, failtWoEtv, failMWoStv},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    []string{notAPoint, badName},
			Explainer: "M does not pick obviously invalid files",
		},
	}

	for _, c := range testcase {

		conf := config.Config{}
		conf.Version = "0.1.2"

		def := BootstrapDefinition(&conf)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, &def, c)
		})
	}
}
