// Remember for the limitless prover jobs, there will be no large deferrals here
package controller

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
)

// This test ensures that the naming convention is respected by the file-watcher
// for bootstrap jobs: only valid inputs are recognized, and the corresponding
// output files are also generated.
func TestBootstrapInFileRegexp(t *testing.T) {

	conf := config.Config{}
	conf.Version = "0.1.2"

	var (
		correct           = "102-103-etv0.2.3-stv1.2.3-getZkProof.json"
		correctWithFail   = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_77"
		correctWith2Fails = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_77.failure.code_77"
		missingEtv        = "102-103-stv1.2.3-getZkProof.json"
		missingStv        = "102-103-etv0.2.3-getZkProof.json"
		notAPoint         = "102-103-etv0.2.3-getZkProofAjson"
		badName           = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	// The requests in case of success
	var (
		resp         = "requests/102-103-metadata-getZkProof.json"
		respWithFail = "requests/102-103-metadata-getZkProof.json"
		// #nosec G101 -- Not a credential
		respWith2Fails = "requests/102-103-metadata-getZkProof.json"
	)

	// The rename in case it is a success
	var (
		success           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
		successWoStv      = "requests-done/102-103-etv0.2.3-getZkProof.json.success"
		successWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof.json.success"
		successWithFail   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
		successWith2Fails = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
	)

	// The rename in case it is a panic (code = 2)
	var (
		fail           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
		failWoStv      = "requests-done/102-103-etv0.2.3-getZkProof.json.failure.code_2"
		failWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof.json.failure.code_2"
		failWithFail   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
		failWith2Fails = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correct, correctWithFail, correctWith2Fails, missingEtv, missingStv},
			Explainer:      "happy path, Bootstrapper picks valid files",
			ExpectedOutput: []string{resp, respWithFail, respWith2Fails, resp, resp},
			ExpSuccess:     []string{success, successWithFail, successWith2Fails, successWoEtv, successWoStv},
			ExpFailW2:      []string{fail, failWithFail, failWith2Fails, failWoEtv, failWoStv},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    []string{notAPoint, badName},
			Explainer: "Bootstrapper rejects invalid files",
		},
	}

	for _, c := range testcase {
		def := BootstrapDefinition(&conf)
		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, &def, c)
		})
	}
}

// This test ensures that the naming convention is respected by the file-watcher
// for conglomeration jobs: only valid inputs are recognized, and the corresponding
// output files are also generated.
func TestConglomerationInFileRegexp(t *testing.T) {

	conf := config.Config{}
	conf.Version = "0.1.2"

	var (
		correct           = "200-201-metadata-getZkProof.json"
		correctWithFail   = "200-201-metadata-getZkProof.json.failure.code_77"
		correctWith2Fails = "200-201-metadata-getZkProof.json.failure.code_77.failure.code_77"
		notAPoint         = "200-201-metadata-getZkProofAjson"
		badName           = "200-201-getAggregatedProof.json"
	)

	// The requests in case of success
	var (
		resp           = "responses/200-201-getZkProof.json"
		respWithFail   = "responses/200-201-getZkProof.json"
		respWith2Fails = "responses/200-201-getZkProof.json"
	)

	// The rename in case it is a success
	var (
		success           = "requests-done/200-201-metadata-getZkProof.json.success"
		successWithFail   = "requests-done/200-201-metadata-getZkProof.json.success"
		successWith2Fails = "requests-done/200-201-metadata-getZkProof.json.success"
	)

	// The rename in case it is a panic (code = 2)
	var (
		fail           = "requests-done/200-201-metadata-getZkProof.json.failure.code_2"
		failWithFail   = "requests-done/200-201-metadata-getZkProof.json.failure.code_2"
		failWith2Fails = "requests-done/200-201-metadata-getZkProof.json.failure.code_2"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correct, correctWithFail, correctWith2Fails},
			Explainer:      "happy path, Conglomerator picks valid files",
			ExpectedOutput: []string{resp, respWithFail, respWith2Fails},
			ExpSuccess:     []string{success, successWithFail, successWith2Fails},
			ExpFailW2:      []string{fail, failWithFail, failWith2Fails},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    []string{notAPoint, badName},
			Explainer: "Conglomerator rejects invalid files",
		},
	}

	for _, c := range testcase {
		def := ConglomerationDefinition(&conf)
		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, &def, c)
		})
	}
}

// This test ensures that the naming convention is respected by the file-watcher
// for GL module jobs: only valid inputs are recognized. GL jobs do not produce
// any output artifact (only /dev/null), so ExpectedOutput is empty, but success
// and failure renames still happen.
func TestGLModsInFileRegexp(t *testing.T) {
	conf := config.Config{}
	conf.Version = "0.1.2"
	conf.ExecutionLimitless.WitnessDir = "/tmp/exec-limitless/witness"

	var (
		correct           = "22504197-22504198-seg-0-mod-0-gl-wit.bin"
		correctWithFail   = "22504197-22504198-seg-0-mod-0-gl-wit.bin.failure.code_77"
		correctWith2Fails = "22504197-22504198-seg-0-mod-0-gl-wit.bin.failure.code_77.failure.code_77"
		notAPoint         = "22504197-22504198-seg-0-mod-0-gl-witA.bin"
		badName           = "22504197-22504198-gl-wit.json"
	)

	var (
		resp = "/dev/null"
	)

	for _, module := range config.ALL_MODULES {
		testcase := []inpFileNamesCases{
			{
				Ext: "", Fail: "code", ShouldMatch: true,
				Fnames:    []string{correct, correctWithFail, correctWith2Fails},
				Explainer: fmt.Sprintf("happy path, GL module %s picks valid files", module),
				ExpectedOutput: []string{
					resp, resp, resp, // no real output files, GL jobs write to /dev/null
				},
				ExpSuccess: []string{
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module, "requests-done", correct+".success"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module, "requests-done", correct+".success"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module, "requests-done", correct+".success"),
				},
				ExpFailW2: []string{
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module, "requests-done", correct+".failure.code_2"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module, "requests-done", correct+".failure.code_2"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module, "requests-done", correct+".failure.code_2"),
				},
			},
			{
				Ext: "", Fail: "code", ShouldMatch: false,
				Fnames:    []string{notAPoint, badName},
				Explainer: fmt.Sprintf("GL module %s rejects invalid files", module),
			},
		}

		for _, c := range testcase {
			def := GLDefinitionForModule(&conf, module)
			t.Run(c.Explainer, func(t *testing.T) {
				runInpFileTestCase(t, &def, c)
			})
		}
	}
}

// This test ensures that the naming convention is respected by the file-watcher
// for LPP module jobs: only valid inputs are recognized. LPP jobs do not produce
// any output artifact (only /dev/null), so ExpectedOutput is empty, but success
// and failure renames still happen.
func TestLPPModsInFileRegexp(t *testing.T) {
	conf := config.Config{}
	conf.Version = "0.1.2"
	conf.ExecutionLimitless.WitnessDir = "/tmp/exec-limitless/witness"

	var (
		correct           = "22504197-22504198-seg-0-mod-0-lpp-wit.bin"
		correctWithFail   = "22504197-22504198-seg-0-mod-0-lpp-wit.bin.failure.code_77"
		correctWith2Fails = "22504197-22504198-seg-0-mod-0-lpp-wit.bin.failure.code_77.failure.code_77"
		notAPoint         = "22504197-22504198-seg-0-mod-0-lpp-witA.bin"
		badName           = "22504197-22504198-lpp-wit.json"
	)

	var (
		resp = "/dev/null"
	)

	for _, module := range config.ALL_MODULES {
		testcase := []inpFileNamesCases{
			{
				Ext: "", Fail: "code", ShouldMatch: true,
				Fnames:    []string{correct, correctWithFail, correctWith2Fails},
				Explainer: fmt.Sprintf("happy path, lpp module %s picks valid files", module),
				ExpectedOutput: []string{
					resp, resp, resp, // no real output files, lpp jobs write to /dev/null
				},
				ExpSuccess: []string{
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module, "requests-done", correct+".success"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module, "requests-done", correct+".success"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module, "requests-done", correct+".success"),
				},
				ExpFailW2: []string{
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module, "requests-done", correct+".failure.code_2"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module, "requests-done", correct+".failure.code_2"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module, "requests-done", correct+".failure.code_2"),
				},
			},
			{
				Ext: "", Fail: "code", ShouldMatch: false,
				Fnames:    []string{notAPoint, badName},
				Explainer: fmt.Sprintf("LPP module %s rejects invalid files", module),
			},
		}

		for _, c := range testcase {
			def := LPPDefinitionForModule(&conf, module)
			t.Run(c.Explainer, func(t *testing.T) {
				runInpFileTestCase(t, &def, c)
			})
		}
	}
}
