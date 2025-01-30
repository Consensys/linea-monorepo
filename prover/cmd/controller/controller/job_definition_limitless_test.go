package controller

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapDefinition(t *testing.T) {
	var (
		correctM           = "102-103-etv0.2.3-stv1.2.3-getZkProof.json"
		correctWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_77"
		correctWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_77.failure.code_77"
		missingEtv         = "102-103-stv1.2.3-getZkProof.json"
		missingStv         = "102-103-etv0.2.3-getZkProof.json"
		notAPoint          = "102-103-etv0.2.3-getZkProofAjson"
		badName            = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correctM, correctWithFailM, correctWith2FailsM, missingEtv, missingStv},
			Explainer:      "happy path, case M",
			ExpectedOutput: []string{bootstrapSubmoduleFile, bootstrapSubmoduleFile, bootstrapSubmoduleFile, bootstrapSubmoduleFile, bootstrapSubmoduleFile},
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
		conf.Bootstrap.CanRunFullLarge = c.Ext == "large"

		def, err := BootstrapDefinition(&conf)
		assert.NoError(t, err)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCaseLimitless(t, def, c)
		})
	}
}

func TestGLExecutionDefinition(t *testing.T) {
	var (
		correctM           = "102-103-etv0.2.3-stv1.2.3-getZKProof_Bootstrap_Submodule.json"
		correctWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZKProof_Bootstrap_Submodule.json.failure.code_77"
		correctWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZKProof_Bootstrap_Submodule.json.failure.code_77.failure.code_77"
		missingEtv         = "102-103-stv1.2.3-getZKProof_Bootstrap_Submodule.json"
		missingStv         = "102-103-etv0.2.3-getZKProof_Bootstrap_Submodule.json"
		notAPoint          = "102-103-etv0.2.3-getZKProof_Bootstrap_SubmoduleAjson"
		badName            = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correctM, correctWithFailM, correctWith2FailsM, missingEtv, missingStv},
			Explainer:      "happy path, case M",
			ExpectedOutput: []string{glBeaconFile, glBeaconFile, glBeaconFile, glBeaconFile, glBeaconFile},
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
		conf.GLExecution.CanRunFullLarge = c.Ext == "large"

		def, err := GLExecutionDefinition(&conf)
		assert.NoError(t, err)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCaseLimitless(t, def, c)
		})
	}
}

func TestRandomBeaconDefinition(t *testing.T) {
	var (
		correctM           = "102-103-etv0.2.3-stv1.2.3-getZKProof_Bootstrap_DistMetadata.json"
		correctWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZKProof_Bootstrap_DistMetadata.json.failure.code_77"
		correctWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZKProof_Bootstrap_DistMetadata.json.failure.code_77.failure.code_77"
		missingEtv         = "102-103-stv1.2.3-getZKProof_Bootstrap_DistMetadata.json"
		missingStv         = "102-103-etv0.2.3-getZKProof_Bootstrap_DistMetadata.json"
		notAPoint          = "102-103-etv0.2.3-getZKProof_Bootstrap_DistMetadataAjson"
		badName            = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correctM, correctWithFailM, correctWith2FailsM, missingEtv, missingStv},
			Explainer:      "happy path, case M",
			ExpectedOutput: []string{randomBeaconOutputFile, randomBeaconOutputFile, randomBeaconOutputFile, randomBeaconOutputFile, randomBeaconOutputFile},
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
		conf.RandomBeacon.Bootstrap.CanRunFullLarge = c.Ext == "large"
		conf.RandomBeacon.GL.CanRunFullLarge = c.Ext == "large"

		def, err := RandomBeaconDefinition(&conf)
		assert.NoError(t, err)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCaseLimitless(t, def, c)
		})
	}
}

func TestLPPExecutionDefinition(t *testing.T) {
	var (
		correctM           = "102-103-etv0.2.3-stv1.2.3-getZKProof_RndBeacon.json"
		correctWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZKProof_RndBeacon.json.failure.code_77"
		correctWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZKProof_RndBeacon.json.failure.code_77.failure.code_77"
		missingEtv         = "102-103-stv1.2.3-getZKProof_RndBeacon.json"
		missingStv         = "102-103-etv0.2.3-getZKProof_RndBeacon.json"
		notAPoint          = "102-103-etv0.2.3-getZKProof_RndBeaconAjson"
		badName            = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correctM, correctWithFailM, correctWith2FailsM, missingEtv, missingStv},
			Explainer:      "happy path, case M",
			ExpectedOutput: []string{lppOutputFile, lppOutputFile, lppOutputFile, lppOutputFile, lppOutputFile},
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
		conf.LPPExecution.CanRunFullLarge = c.Ext == "large"

		def, err := LPPExecutionDefinition(&conf)
		assert.NoError(t, err)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCaseLimitless(t, def, c)
		})
	}
}

func TestConglomerationDefinition(t *testing.T) {
	var (
		correctM           = "102-103-etv0.2.3-stv1.2.3-getZKProof_GL.json"
		correctWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZKProof_GL.json.failure.code_77"
		correctWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZKProof_GL.json.failure.code_77.failure.code_77"
		missingEtv         = "102-103-stv1.2.3-getZKProof_GL.json"
		missingStv         = "102-103-etv0.2.3-getZKProof_GL.json"
		notAPoint          = "102-103-etv0.2.3-getZKProof_GLAjson"
		badName            = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correctM, correctWithFailM, correctWith2FailsM, missingEtv, missingStv},
			Explainer:      "happy path, case M",
			ExpectedOutput: []string{conglomerationOutputFile, conglomerationOutputFile, conglomerationOutputFile, conglomerationOutputFile, conglomerationOutputFile},
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
		conf.Conglomeration.GL.CanRunFullLarge = c.Ext == "large"
		conf.Conglomeration.LPP.CanRunFullLarge = c.Ext == "large"

		def, err := ConglomerationDefinition(&conf)
		assert.NoError(t, err)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCaseLimitless(t, def, c)
		})
	}
}

func runInpFileTestCaseLimitless(t *testing.T, def *JobDefinition, c inpFileNamesCases) {
	for i, fname := range c.Fnames {
		job, err := NewJob(def, fname)

		if c.ShouldMatch {
			if !assert.NoError(t, err, fname) {
				continue
			}

			resp, err := job.ResponseFile()
			if assert.NoErrorf(t, err, "cannot produce a response for job %s", fname) {
				assert.Equal(t, c.ExpectedOutput[i], resp, "wrong output file")
			}

			if len(c.ExpToLarge) > 0 {
				toLarge, err := job.DeferToLargeFile(
					Status{ExitCode: 137},
				)

				if assert.NoError(t, err, "cannot produce name for the too large job") {
					assert.Equal(t, c.ExpToLarge[i], toLarge)
				}
			}

			if len(c.ExpSuccess) > 0 {
				toSuccess := job.DoneFile(Status{ExitCode: 0})
				assert.Equal(t, c.ExpSuccess[i], toSuccess)
			}

			if len(c.ExpFailW2) > 0 {
				toFail2 := job.DoneFile(Status{ExitCode: 2})
				assert.Equal(t, c.ExpFailW2[i], toFail2)
			}

		} else {
			assert.Errorf(
				t, err, fname,
				"%v should not match %s",
				fname, def.InputFileRegexp.String(),
			)
		}
	}
}
