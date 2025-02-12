package controller

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/stretchr/testify/assert"
)

func TestExecBootstrapInFileRegexp(t *testing.T) {
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
		respGLM       = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json"
		respGLL       = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json"
		respMetadataM = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"
		respMetadataL = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"

		respGLWithFailM         = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json"
		respGLWithFailL         = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json"
		respGLWith2FailsM       = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json"
		respGLWith2FailsL       = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json"
		respMetadataWithFailM   = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"
		respMetadataWithFailL   = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"
		respMetadataWith2FailsM = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"
		respMetadataWith2FailsL = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"

		respGLWoEtv       = "responses/102-103-etv-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json"
		respGLWoStv       = "responses/102-103-etv0.2.3-stv-getZkProof_Bootstrap_GLSubmodule.json"
		respMetadataWoEtv = "responses/102-103-etv-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"
		respMetadataWoStv = "responses/102-103-etv0.2.3-stv-getZkProof_Bootstrap_DistMetadata.json"
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
			Fnames:         [][]string{{correctM}, {correctWithFailM}, {correctWith2FailsM}, {missingEtv}, {missingStv}},
			Explainer:      "happy path, case M",
			ExpectedOutput: [][]string{{respGLM, respMetadataM}, {respGLWithFailM, respMetadataWithFailM}, {respGLWith2FailsM, respMetadataWith2FailsM}, {respGLWoEtv, respMetadataWoEtv}, {respGLWoStv, respMetadataWoStv}},
			ExpToLarge:     [][]string{{toLargeM, toLargeM}, {toLargeWithFailM, toLargeWithFailM}, {toLargeWith2FailsM, toLargeWith2FailsM}, {toLargeWoEtv, toLargeWoEtv}, {toLargeWoStv, toLargeWoStv}},
			ExpSuccess:     [][]string{{successM, successM}, {successWithFailM, successWithFailM}, {successWith2FailsM, successWith2FailsM}, {successtWoEtv, successtWoEtv}, {successMWoStv, successMWoStv}},
			ExpFailW2:      [][]string{{failM, failM}, {failWithFailM, failWithFailM}, {failWith2FailsM, failWith2FailsM}, {failtWoEtv, failtWoEtv}, {failMWoStv, failMWoStv}},
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: true,
			Fnames:         [][]string{{correctL}, {correctWithFailL}, {correctWith2FailsL}},
			Explainer:      "happy path, case L",
			ExpectedOutput: [][]string{{respGLL, respMetadataL}, {respGLWithFailL, respMetadataWithFailL}, {respGLWith2FailsL, respMetadataWith2FailsL}},
			ExpSuccess:     [][]string{{successL, successL}, {successWithFailL, successWithFailL}, {successWith2FailsL, successWith2FailsL}},
			ExpFailW2:      [][]string{{failL, failL}, {failWithFailL, failWithFailL}, {failWith2FailsL, failWith2FailsL}},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    [][]string{{correctL}, {correctWithFailL}, {correctWith2FailsL}},
			Explainer: "M does not pick the files reserved for L",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames:    [][]string{{correctM}, {correctWithFailM}, {correctWith2FailsM}},
			Explainer: "L does not pick the files reserved for M",
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    [][]string{{notAPoint}, {badName}},
			Explainer: "M does not pick obviously invalid files",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames:    [][]string{{missingEtv}, {missingStv}, {notAPoint}, {badName}},
			Explainer: "L does not pick obviously invalid files",
		},
	}

	for _, c := range testcase {
		conf := config.Config{}
		conf.Version = "0.1.2"
		conf.ExecBootstrap.CanRunFullLarge = c.Ext == "large"
		conf.ExecBootstrap.RequestsRootDir = []string{""}
		conf.ExecBootstrap.ResponsesRootDir = []string{"", ""}

		def, err := ExecBootstrapDefinition(&conf)
		assert.NoError(t, err)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, def, c)
		})
	}
}

func TestExecGLInFileRegexp(t *testing.T) {
	var (
		correctM           = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json"
		correctL           = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large"
		correctWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.failure.code_77"
		correctWithFailL   = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_77"
		correctWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.failure.code_77.failure.code_77"
		correctWith2FailsL = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_77.failure.code_77"
		missingEtv         = "102-103-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json"
		missingStv         = "102-103-etv0.2.3-getZkProof_Bootstrap_GLSubmodule.json"
		notAPoint          = "102-103-etv0.2.3-getZkProofAjson"
		badName            = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	// The responses in case of success
	var (
		respRndBeaconM = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json"
		respRndBeaconL = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json"
		respGLM        = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json"
		respGLL        = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json"

		respRndBeaconWithFailM   = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json"
		respRndBeaconWithFailL   = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json"
		respRndBeaconWith2FailsM = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json"
		respRndBeaconWith2FailsL = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json"
		respGLWithFailM          = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json"
		respGLWithFailL          = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json"
		respGLWith2FailsM        = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json"
		respGLWith2FailsL        = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json"

		respRndBeaconWoEtv = "responses/102-103-etv-stv1.2.3-getZkProof_GL_RndBeacon.json"
		respRndBeaconWoStv = "responses/102-103-etv0.2.3-stv-getZkProof_GL_RndBeacon.json"
		respGLWoEtv        = "responses/102-103-etv-stv1.2.3-getZkProof_GL.json"
		respGLWoStv        = "responses/102-103-etv0.2.3-stv-getZkProof_GL.json"
	)

	// The rename in case it is deferred to the large prover
	var (
		toLargeM           = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_137"
		toLargeWithFailM   = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_137"
		toLargeWith2FailsM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_137"
		toLargeWoEtv       = "requests/102-103-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_137"
		toLargeWoStv       = "requests/102-103-etv0.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_137"
	)

	// The rename in case it is a success
	var (
		successM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.success"
		successMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_Bootstrap_GLSubmodule.json.success"
		successtWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.success"
		successL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.success"
		successWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.success"
		successWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.success"
		successWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.success"
		successWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.success"
	)

	// The rename in case it is a panic (code = 2)
	var (
		failM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.failure.code_2"
		failMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_Bootstrap_GLSubmodule.json.failure.code_2"
		failtWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.failure.code_2"
		failL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_2"
		failWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.failure.code_2"
		failWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_2"
		failWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.failure.code_2"
		failWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_2"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         [][]string{{correctM}, {correctWithFailM}, {correctWith2FailsM}, {missingEtv}, {missingStv}},
			Explainer:      "happy path, case M",
			ExpectedOutput: [][]string{{respRndBeaconM, respGLM}, {respRndBeaconWithFailM, respGLWithFailM}, {respRndBeaconWith2FailsM, respGLWith2FailsM}, {respRndBeaconWoEtv, respGLWoEtv}, {respRndBeaconWoStv, respGLWoStv}},
			ExpToLarge:     [][]string{{toLargeM, toLargeM}, {toLargeWithFailM, toLargeWithFailM}, {toLargeWith2FailsM, toLargeWith2FailsM}, {toLargeWoEtv, toLargeWoEtv}, {toLargeWoStv, toLargeWoStv}},
			ExpSuccess:     [][]string{{successM, successM}, {successWithFailM, successWithFailM}, {successWith2FailsM, successWith2FailsM}, {successtWoEtv, successtWoEtv}, {successMWoStv, successMWoStv}},
			ExpFailW2:      [][]string{{failM, failM}, {failWithFailM, failWithFailM}, {failWith2FailsM, failWith2FailsM}, {failtWoEtv, failtWoEtv}, {failMWoStv, failMWoStv}},
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: true,
			Fnames:         [][]string{{correctL}, {correctWithFailL}, {correctWith2FailsL}},
			Explainer:      "happy path, case L",
			ExpectedOutput: [][]string{{respRndBeaconL, respGLL}, {respRndBeaconWithFailL, respGLWithFailL}, {respRndBeaconWith2FailsL, respGLWith2FailsL}},
			ExpSuccess:     [][]string{{successL, successL}, {successWithFailL, successWithFailL}, {successWith2FailsL, successWith2FailsL}},
			ExpFailW2:      [][]string{{failL, failL}, {failWithFailL, failWithFailL}, {failWith2FailsL, failWith2FailsL}},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    [][]string{{correctL}, {correctWithFailL}, {correctWith2FailsL}},
			Explainer: "M does not pick the files reserved for L",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames:    [][]string{{correctM}, {correctWithFailM}, {correctWith2FailsM}},
			Explainer: "L does not pick the files reserved for M",
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    [][]string{{notAPoint}, {badName}},
			Explainer: "M does not pick obviously invalid files",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames:    [][]string{{missingEtv}, {missingStv}, {notAPoint}, {badName}},
			Explainer: "L does not pick obviously invalid files",
		},
	}

	for _, c := range testcase {
		conf := config.Config{}
		conf.Version = "0.1.2"
		conf.ExecGL.CanRunFullLarge = c.Ext == "large"
		conf.ExecGL.RequestsRootDir = []string{""}
		conf.ExecGL.ResponsesRootDir = []string{"", ""}

		def, err := ExecGLDefinition(&conf)
		assert.NoError(t, err)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, def, c)
		})
	}
}

func TestExecRndBeaconInFileRegexp(t *testing.T) {
	var (
		correctBootstrapMetadataM           = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"
		correctBootstrapMetadataL           = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large"
		correctBootstrapMetadataWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_77"
		correctBootstrapMetadataWithFailL   = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_77"
		correctBootstrapMetadataWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_77.failure.code_77"
		correctBootstrapMetadataWith2FailsL = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_77.failure.code_77"
		missingBootstrapMetadataEtv         = "102-103-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"
		missingBootstrapMetadataStv         = "102-103-etv0.2.3-getZkProof_Bootstrap_DistMetadata.json"

		correctGLM           = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json"
		correctGLL           = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large"
		correctGLWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_77"
		correctGLWithFailL   = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_77"
		correctGLWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_77.failure.code_77"
		correctGLWith2FailsL = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_77.failure.code_77"
		missingGLEtv         = "102-103-stv1.2.3-getZkProof_GL_RndBeacon.json"
		missingGLStv         = "102-103-etv0.2.3-getZkProof_GL_RndBeacon.json"
		notAPoint            = "102-103-etv0.2.3-getZkProofAjson"
		badName              = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	// The responses in case of success
	var (
		respRndBeaconM = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"
		respRndBeaconL = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"

		respRndBeaconWithFailM   = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"
		respRndBeaconWithFailL   = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"
		respRndBeaconWith2FailsM = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"
		respRndBeaconWith2FailsL = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"

		respRndBeaconWoEtv = "responses/102-103-etv-stv1.2.3-getZkProof_RndBeacon.json"
		respRndBeaconWoStv = "responses/102-103-etv0.2.3-stv-getZkProof_RndBeacon.json"
	)

	// The rename in case it is deferred to the large prover
	var (
		toLargeBootstrapMetadataM           = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"
		toLargeBootstrapMetadataWithFailM   = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"
		toLargeBootstrapMetadataWith2FailsM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"
		toLargeBootstrapMetadataWoEtv       = "requests/102-103-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"
		toLargeBootstrapMetadataWoStv       = "requests/102-103-etv0.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"

		toLargeGLM           = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_137"
		toLargeGLWithFailM   = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_137"
		toLargeGLWith2FailsM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_137"
		toLargeGLWoEtv       = "requests/102-103-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_137"
		toLargeGLWoStv       = "requests/102-103-etv0.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_137"
	)

	// The rename in case it is a success
	var (
		successBootstrapMetadataM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
		successBootstrapMetadataMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
		successBootstrapMetadatastWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
		successBootstrapMetadataL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success"
		successBootstrapMetadataWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
		successBootstrapMetadataWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success"
		successBootstrapMetadataWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
		successBootstrapMetadataWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success"

		successGLM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.success"
		successGLMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_GL_RndBeacon.json.success"
		successGLstWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_GL_RndBeacon.json.success"
		successGLL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.success"
		successGLWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.success"
		successGLWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.success"
		successGLWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.success"
		successGLWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.success"
	)

	// The rename in case it is a panic (code = 2)
	var (
		failBootstrapMetadataM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
		failBootstrapMetadataMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
		failBootstrapMetadatastWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
		failBootstrapMetadataL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_2"
		failBootstrapMetadataWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
		failBootstrapMetadataWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_2"
		failBootstrapMetadataWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
		failBootstrapMetadataWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_2"

		failGLM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_2"
		failGLMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_GL_RndBeacon.json.failure.code_2"
		failGLstWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_2"
		failGLL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_2"
		failGLWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_2"
		failGLWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_2"
		failGLWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_2"
		failGLWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_2"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames: [][]string{
				{correctBootstrapMetadataM, correctGLM},
				{correctBootstrapMetadataWithFailM, correctGLWithFailM},
				{correctBootstrapMetadataWith2FailsM, correctGLWith2FailsM},
				{missingBootstrapMetadataEtv, missingGLEtv},
				{missingBootstrapMetadataStv, missingGLStv},
			},
			Explainer: "happy path, case M",
			ExpectedOutput: [][]string{
				{respRndBeaconM},
				{respRndBeaconWithFailM},
				{respRndBeaconWith2FailsM},
				{respRndBeaconWoEtv},
				{respRndBeaconWoStv},
			},
			ExpToLarge: [][]string{
				{toLargeBootstrapMetadataM, toLargeGLM},
				{toLargeBootstrapMetadataWithFailM, toLargeGLWithFailM},
				{toLargeBootstrapMetadataWith2FailsM, toLargeGLWith2FailsM},
				{toLargeBootstrapMetadataWoEtv, toLargeGLWoEtv},
				{toLargeBootstrapMetadataWoStv, toLargeGLWoStv},
			},
			ExpSuccess: [][]string{
				{successBootstrapMetadataM, successGLM},
				{successBootstrapMetadataWithFailM, successGLWithFailM},
				{successBootstrapMetadataWith2FailsM, successGLWith2FailsM},
				{successBootstrapMetadatastWoEtv, successGLstWoEtv},
				{successBootstrapMetadataMWoStv, successGLMWoStv},
			},
			ExpFailW2: [][]string{
				{failBootstrapMetadataM, failGLM},
				{failBootstrapMetadataWithFailM, failGLWithFailM},
				{failBootstrapMetadataWith2FailsM, failGLWith2FailsM},
				{failBootstrapMetadatastWoEtv, failGLstWoEtv},
				{failBootstrapMetadataMWoStv, failGLMWoStv},
			},
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: true,
			Fnames: [][]string{
				{correctBootstrapMetadataL, correctGLL},
				{correctBootstrapMetadataWithFailL, correctGLWithFailL},
				{correctBootstrapMetadataWith2FailsL, correctGLWith2FailsL},
			},
			Explainer: "happy path, case L",
			ExpectedOutput: [][]string{
				{respRndBeaconL},
				{respRndBeaconWithFailL},
				{respRndBeaconWith2FailsL},
			},
			ExpSuccess: [][]string{
				{successBootstrapMetadataL, successGLL},
				{successBootstrapMetadataWithFailL, successGLWithFailL},
				{successBootstrapMetadataWith2FailsL, successGLWith2FailsL},
			},
			ExpFailW2: [][]string{
				{failBootstrapMetadataL, failGLL},
				{failBootstrapMetadataWithFailL, failGLWithFailL},
				{failBootstrapMetadataWith2FailsL, failGLWith2FailsL},
			},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames: [][]string{
				{correctBootstrapMetadataL, correctGLL},
				{correctBootstrapMetadataWithFailL, correctGLWithFailL},
				{correctBootstrapMetadataWith2FailsL, correctGLWith2FailsL},
			},
			Explainer: "M does not pick the files reserved for L",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames: [][]string{
				{correctBootstrapMetadataM, correctGLM},
				{correctBootstrapMetadataWithFailM, correctGLWithFailM},
				{correctBootstrapMetadataWith2FailsM, correctGLWith2FailsM},
			},
			Explainer: "L does not pick the files reserved for M",
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames: [][]string{
				{notAPoint},
				{badName},
			},
			Explainer: "M does not pick obviously invalid files",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames: [][]string{
				{missingBootstrapMetadataEtv, missingGLEtv},
				{missingBootstrapMetadataStv, missingGLStv},
				{notAPoint},
				{badName},
			},
			Explainer: "L does not pick obviously invalid files",
		},
	}

	for _, c := range testcase {
		conf := config.Config{}
		conf.Version = "0.1.2"
		conf.ExecRndBeacon.CanRunFullLarge = c.Ext == "large"
		// conf.ExecRndBeacon.GL.RequestsRootDir = []string{"", ""}

		conf.ExecRndBeacon.RequestsRootDir = []string{"", ""}
		conf.ExecRndBeacon.ResponsesRootDir = []string{""}

		def, err := ExecRndBeaconDefinition(&conf)
		assert.NoError(t, err)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, def, c)
		})
	}
}

func TestExecLPPInFileRegexp(t *testing.T) {
	var (
		correctM           = "102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"
		correctL           = "102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.large"
		correctWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.failure.code_77"
		correctWithFailL   = "102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.large.failure.code_77"
		correctWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.failure.code_77.failure.code_77"
		correctWith2FailsL = "102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.large.failure.code_77.failure.code_77"
		missingEtv         = "102-103-stv1.2.3-getZkProof_RndBeacon.json"
		missingStv         = "102-103-etv0.2.3-getZkProof_RndBeacon.json"
		notAPoint          = "102-103-etv0.2.3-getZkProofAjson"
		badName            = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	// The responses in case of success
	var (
		respLPPM = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json"
		respLPPL = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json"

		respLPPWithFailM   = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json"
		respLPPWithFailL   = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json"
		respLPPWith2FailsM = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json"
		respLPPWith2FailsL = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json"

		respLPPWoEtv = "responses/102-103-etv-stv1.2.3-getZkProof_LPP.json"
		respLPPWoStv = "responses/102-103-etv0.2.3-stv-getZkProof_LPP.json"
	)

	// The rename in case it is deferred to the large prover
	var (
		toLargeM           = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.large.failure.code_137"
		toLargeWithFailM   = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.large.failure.code_137"
		toLargeWith2FailsM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.large.failure.code_137"
		toLargeWoEtv       = "requests/102-103-stv1.2.3-getZkProof_RndBeacon.json.large.failure.code_137"
		toLargeWoStv       = "requests/102-103-etv0.2.3-getZkProof_RndBeacon.json.large.failure.code_137"
	)

	// The rename in case it is a success
	var (
		successM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.success"
		successMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_RndBeacon.json.success"
		successtWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof_RndBeacon.json.success"
		successL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.large.success"
		successWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.success"
		successWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.large.success"
		successWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.success"
		successWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.large.success"
	)

	// The rename in case it is a panic (code = 2)
	var (
		failM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.failure.code_2"
		failMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_RndBeacon.json.failure.code_2"
		failtWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof_RndBeacon.json.failure.code_2"
		failL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.large.failure.code_2"
		failWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.failure.code_2"
		failWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.large.failure.code_2"
		failWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.failure.code_2"
		failWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json.large.failure.code_2"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         [][]string{{correctM}, {correctWithFailM}, {correctWith2FailsM}, {missingEtv}, {missingStv}},
			Explainer:      "happy path, case M",
			ExpectedOutput: [][]string{{respLPPM}, {respLPPWithFailM}, {respLPPWith2FailsM}, {respLPPWoEtv}, {respLPPWoStv}},
			ExpToLarge:     [][]string{{toLargeM}, {toLargeWithFailM}, {toLargeWith2FailsM}, {toLargeWoEtv}, {toLargeWoStv}},
			ExpSuccess:     [][]string{{successM}, {successWithFailM}, {successWith2FailsM}, {successtWoEtv}, {successMWoStv}},
			ExpFailW2:      [][]string{{failM}, {failWithFailM}, {failWith2FailsM}, {failtWoEtv}, {failMWoStv}},
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: true,
			Fnames:         [][]string{{correctL}, {correctWithFailL}, {correctWith2FailsL}},
			Explainer:      "happy path, case L",
			ExpectedOutput: [][]string{{respLPPL}, {respLPPWithFailL}, {respLPPWith2FailsL}},
			ExpSuccess:     [][]string{{successL}, {successWithFailL}, {successWith2FailsL}},
			ExpFailW2:      [][]string{{failL}, {failWithFailL}, {failWith2FailsL}},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    [][]string{{correctL}, {correctWithFailL}, {correctWith2FailsL}},
			Explainer: "M does not pick the files reserved for L",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames:    [][]string{{correctM}, {correctWithFailM}, {correctWith2FailsM}},
			Explainer: "L does not pick the files reserved for M",
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    [][]string{{notAPoint}, {badName}},
			Explainer: "M does not pick obviously invalid files",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames:    [][]string{{missingEtv}, {missingStv}, {notAPoint}, {badName}},
			Explainer: "L does not pick obviously invalid files",
		},
	}

	for _, c := range testcase {
		conf := config.Config{}
		conf.Version = "0.1.2"
		conf.ExecLPP.CanRunFullLarge = c.Ext == "large"
		conf.ExecLPP.RequestsRootDir = []string{""}
		conf.ExecLPP.ResponsesRootDir = []string{""}

		def, err := ExecLPPDefinition(&conf)
		assert.NoError(t, err)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, def, c)
		})
	}
}

func TestExecConglomerationInFileRegexp(t *testing.T) {
	var (
		correctBootstrapMetadataM           = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"
		correctBootstrapMetadataL           = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large"
		correctBootstrapMetadataWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_77"
		correctBootstrapMetadataWithFailL   = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_77"
		correctBootstrapMetadataWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_77.failure.code_77"
		correctBootstrapMetadataWith2FailsL = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_77.failure.code_77"
		missingBootstrapMetadataEtv         = "102-103-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"
		missingBootstrapMetadataStv         = "102-103-etv0.2.3-getZkProof_Bootstrap_DistMetadata.json"

		correctGLM           = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json"
		correctGLL           = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.large"
		correctGLWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.failure.code_77"
		correctGLWithFailL   = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.large.failure.code_77"
		correctGLWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.failure.code_77.failure.code_77"
		correctGLWith2FailsL = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.large.failure.code_77.failure.code_77"
		missingGLEtv         = "102-103-stv1.2.3-getZkProof_GL.json"
		missingGLStv         = "102-103-etv0.2.3-getZkProof_GL.json"

		correctLPPM           = "102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json"
		correctLPPL           = "102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.large"
		correctLPPWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.failure.code_77"
		correctLPPWithFailL   = "102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.large.failure.code_77"
		correctLPPWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.failure.code_77.failure.code_77"
		correctLPPWith2FailsL = "102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.large.failure.code_77.failure.code_77"
		missingLPPEtv         = "102-103-stv1.2.3-getZkProof_LPP.json"
		missingLPPStv         = "102-103-etv0.2.3-getZkProof_LPP.json"
		notAPoint             = "102-103-etv0.2.3-getZkProofAjson"
		badName               = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	// The responses in case of success
	var (
		respConglomerateM = "responses/102-103-getZkProof.json"
		respConglomerateL = "responses/102-103-getZkProof.json"

		respConglomerateWithFailM   = "responses/102-103-getZkProof.json"
		respConglomerateWithFailL   = "responses/102-103-getZkProof.json"
		respConglomerateWith2FailsM = "responses/102-103-getZkProof.json"
		respConglomerateWith2FailsL = "responses/102-103-getZkProof.json"

		respConglomerateWoEtv = "responses/102-103-getZkProof.json"
		respConglomerateWoStv = "responses/102-103-getZkProof.json"
	)

	// The rename in case it is deferred to the large prover
	var (
		toLargeBootstrapMetadataM           = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"
		toLargeBootstrapMetadataWithFailM   = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"
		toLargeBootstrapMetadataWith2FailsM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"
		toLargeBootstrapMetadataWoEtv       = "requests/102-103-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"
		toLargeBootstrapMetadataWoStv       = "requests/102-103-etv0.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"

		toLargeGLM           = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.large.failure.code_137"
		toLargeGLWithFailM   = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.large.failure.code_137"
		toLargeGLWith2FailsM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.large.failure.code_137"
		toLargeGLWoEtv       = "requests/102-103-stv1.2.3-getZkProof_GL.json.large.failure.code_137"
		toLargeGLWoStv       = "requests/102-103-etv0.2.3-getZkProof_GL.json.large.failure.code_137"

		toLargeLPPM           = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.large.failure.code_137"
		toLargeLPPWithFailM   = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.large.failure.code_137"
		toLargeLPPWith2FailsM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.large.failure.code_137"
		toLargeLPPWoEtv       = "requests/102-103-stv1.2.3-getZkProof_LPP.json.large.failure.code_137"
		toLargeLPPWoStv       = "requests/102-103-etv0.2.3-getZkProof_LPP.json.large.failure.code_137"
	)

	// The rename in case it is a success
	var (
		successBootstrapMetadataM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
		successBootstrapMetadataMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
		successBootstrapMetadatastWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
		successBootstrapMetadataL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success"
		successBootstrapMetadataWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
		successBootstrapMetadataWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success"
		successBootstrapMetadataWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
		successBootstrapMetadataWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success"

		successGLM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.success"
		successGLMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_GL.json.success"
		successGLstWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_GL.json.success"
		successGLL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.large.success"
		successGLWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.success"
		successGLWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.large.success"
		successGLWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.success"
		successGLWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.large.success"

		successLPPM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.success"
		successLPPMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_LPP.json.success"
		successLPPstWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_LPP.json.success"
		successLPPL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.large.success"
		successLPPWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.success"
		successLPPWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.large.success"
		successLPPWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.success"
		successLPPWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.large.success"
	)

	// The rename in case it is a panic (code = 2)
	var (
		failBootstrapMetadataM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
		failBootstrapMetadataMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
		failBootstrapMetadatastWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
		failBootstrapMetadataL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_2"
		failBootstrapMetadataWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
		failBootstrapMetadataWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_2"
		failBootstrapMetadataWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
		failBootstrapMetadataWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_2"

		failGLM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.failure.code_2"
		failGLMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_GL.json.failure.code_2"
		failGLstWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_GL.json.failure.code_2"
		failGLL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.large.failure.code_2"
		failGLWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.failure.code_2"
		failGLWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.large.failure.code_2"
		failGLWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.failure.code_2"
		failGLWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL.json.large.failure.code_2"

		failLPPM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.failure.code_2"
		failLPPMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_LPP.json.failure.code_2"
		failLPPstWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_LPP.json.failure.code_2"
		failLPPL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.large.failure.code_2"
		failLPPWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.failure.code_2"
		failLPPWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.large.failure.code_2"
		failLPPWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.failure.code_2"
		failLPPWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_LPP.json.large.failure.code_2"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames: [][]string{
				{correctBootstrapMetadataM, correctGLM, correctLPPM},
				{correctBootstrapMetadataWithFailM, correctGLWithFailM, correctLPPWithFailM},
				{correctBootstrapMetadataWith2FailsM, correctGLWith2FailsM, correctLPPWith2FailsM},
				{missingBootstrapMetadataEtv, missingGLEtv, missingLPPEtv},
				{missingBootstrapMetadataStv, missingGLStv, missingLPPStv},
			},
			Explainer: "happy path, case M",
			ExpectedOutput: [][]string{
				{respConglomerateM},
				{respConglomerateWithFailM},
				{respConglomerateWith2FailsM},
				{respConglomerateWoEtv},
				{respConglomerateWoStv},
			},
			ExpToLarge: [][]string{
				{toLargeBootstrapMetadataM, toLargeGLM, toLargeLPPM},
				{toLargeBootstrapMetadataWithFailM, toLargeGLWithFailM, toLargeLPPWithFailM},
				{toLargeBootstrapMetadataWith2FailsM, toLargeGLWith2FailsM, toLargeLPPWith2FailsM},
				{toLargeBootstrapMetadataWoEtv, toLargeGLWoEtv, toLargeLPPWoEtv},
				{toLargeBootstrapMetadataWoStv, toLargeGLWoStv, toLargeLPPWoStv},
			},
			ExpSuccess: [][]string{
				{successBootstrapMetadataM, successGLM, successLPPM},
				{successBootstrapMetadataWithFailM, successGLWithFailM, successLPPWithFailM},
				{successBootstrapMetadataWith2FailsM, successGLWith2FailsM, successLPPWith2FailsM},
				{successBootstrapMetadatastWoEtv, successGLstWoEtv, successLPPstWoEtv},
				{successBootstrapMetadataMWoStv, successGLMWoStv, successLPPMWoStv},
			},
			ExpFailW2: [][]string{
				{failBootstrapMetadataM, failGLM, failLPPM},
				{failBootstrapMetadataWithFailM, failGLWithFailM, failLPPWithFailM},
				{failBootstrapMetadataWith2FailsM, failGLWith2FailsM, failLPPWith2FailsM},
				{failBootstrapMetadatastWoEtv, failGLstWoEtv, failLPPstWoEtv},
				{failBootstrapMetadataMWoStv, failGLMWoStv, failLPPMWoStv},
			},
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: true,
			Fnames: [][]string{
				{correctBootstrapMetadataL, correctGLL, correctLPPL},
				{correctBootstrapMetadataWithFailL, correctGLWithFailL, correctLPPWithFailL},
				{correctBootstrapMetadataWith2FailsL, correctGLWith2FailsL, correctLPPWith2FailsL},
			},
			Explainer: "happy path, case L",
			ExpectedOutput: [][]string{
				{respConglomerateL},
				{respConglomerateWithFailL},
				{respConglomerateWith2FailsL},
			},
			ExpSuccess: [][]string{
				{successBootstrapMetadataL, successGLL, successLPPL},
				{successBootstrapMetadataWithFailL, successGLWithFailL, successLPPWithFailL},
				{successBootstrapMetadataWith2FailsL, successGLWith2FailsL, successLPPWith2FailsL},
			},
			ExpFailW2: [][]string{
				{failBootstrapMetadataL, failGLL, failLPPL},
				{failBootstrapMetadataWithFailL, failGLWithFailL, failLPPWithFailL},
				{failBootstrapMetadataWith2FailsL, failGLWith2FailsL, failLPPWith2FailsL},
			},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames: [][]string{
				{correctBootstrapMetadataL, correctGLL, correctLPPL},
				{correctBootstrapMetadataWithFailL, correctGLWithFailL, correctLPPWithFailL},
				{correctBootstrapMetadataWith2FailsL, correctGLWith2FailsL, correctLPPWith2FailsL},
			},
			Explainer: "M does not pick the files reserved for L",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames: [][]string{
				{correctBootstrapMetadataM, correctGLM, correctLPPM},
				{correctBootstrapMetadataWithFailM, correctGLWithFailM, correctLPPWithFailM},
				{correctBootstrapMetadataWith2FailsM, correctGLWith2FailsM, correctLPPWith2FailsM},
			},
			Explainer: "L does not pick the files reserved for M",
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames: [][]string{
				{notAPoint},
				{badName},
			},
			Explainer: "M does not pick obviously invalid files",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames: [][]string{
				{missingBootstrapMetadataEtv, missingGLEtv, missingLPPEtv},
				{missingBootstrapMetadataStv, missingGLStv, missingLPPStv},
				{notAPoint},
				{badName},
			},
			Explainer: "L does not pick obviously invalid files",
		},
	}

	for _, c := range testcase {
		conf := config.Config{}
		conf.Version = "0.1.2"
		conf.ExecConglomeration.CanRunFullLarge = c.Ext == "large"
		// conf.ExecConglomeration.GL.RequestsRootDir = []string{""}
		// conf.ExecConglomeration.LPP.RequestsRootDir = []string{""}
		// conf.ExecConglomeration.BootstrapMetadata.RequestsRootDir = []string{""}
		conf.ExecConglomeration.RequestsRootDir = []string{"", "", ""}
		conf.ExecConglomeration.ResponsesRootDir = []string{""}

		def, err := ExecConglomerationDefinition(&conf)
		assert.NoError(t, err)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, def, c)
		})
	}
}
