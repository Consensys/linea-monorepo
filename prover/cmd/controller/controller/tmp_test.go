package controller

// func TestExecRndBeaconInFileRegexp(t *testing.T) {
// 	var (
// 		correctBootstrapMetadataM = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"
// 		// correctBootstrapMetadataL           = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large"
// 		// correctBootstrapMetadataWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_77"
// 		// correctBootstrapMetadataWithFailL   = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_77"
// 		// correctBootstrapMetadataWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_77.failure.code_77"
// 		// correctBootstrapMetadataWith2FailsL = "102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_77.failure.code_77"
// 		// missingBootstrapMetadataEtv         = "102-103-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json"
// 		// missingBootstrapMetadataStv         = "102-103-etv0.2.3-getZkProof_Bootstrap_DistMetadata.json"

// 		correctGLM = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json"
// 		// correctGLL           = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large"
// 		// correctGLWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_77"
// 		// correctGLWithFailL   = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_77"
// 		// correctGLWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_77.failure.code_77"
// 		// correctGLWith2FailsL = "102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_77.failure.code_77"
// 		// missingGLEtv         = "102-103-stv1.2.3-getZkProof_GL_RndBeacon.json"
// 		// missingGLStv         = "102-103-etv0.2.3-getZkProof_GL_RndBeacon.json"
// 		// notAPoint            = "102-103-etv0.2.3-getZkProofAjson"
// 		// badName              = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
// 	)

// 	// The responses in case of success
// 	var (
// 		respRndBeaconM = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"
// 		// respRndBeaconL = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"

// 		// respRndBeaconWithFailM   = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"
// 		// respRndBeaconWithFailL   = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"
// 		// respRndBeaconWith2FailsM = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"
// 		// respRndBeaconWith2FailsL = "responses/102-103-etv0.2.3-stv1.2.3-getZkProof_RndBeacon.json"

// 		// respRndBeaconWoEtv = "responses/102-103-etv-stv1.2.3-getZkProof_RndBeacon.json"
// 		// respRndBeaconWoStv = "responses/102-103-etv0.2.3-stv-getZkProof_RndBeacon.json"
// 	)

// 	// The rename in case it is deferred to the large prover
// 	var (
// 		toLargeBootstrapMetadataM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"
// 		// toLargeBootstrapMetadataWithFailM   = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"
// 		// toLargeBootstrapMetadataWith2FailsM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"
// 		// toLargeBootstrapMetadataWoEtv       = "requests/102-103-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"
// 		// toLargeBootstrapMetadataWoStv       = "requests/102-103-etv0.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_137"

// 		toLargeGLM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_137"
// 		// toLargeGLWithFailM   = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_137"
// 		// toLargeGLWith2FailsM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_137"
// 		// toLargeGLWoEtv       = "requests/102-103-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_137"
// 		// toLargeGLWoStv       = "requests/102-103-etv0.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_137"
// 	)

// 	// The rename in case it is a success
// 	var (
// 		successBootstrapMetadataM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
// 		// successBootstrapMetadataMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
// 		// successBootstrapMetadatastWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
// 		// successBootstrapMetadataL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success"
// 		// successBootstrapMetadataWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
// 		// successBootstrapMetadataWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success"
// 		// successBootstrapMetadataWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success"
// 		// successBootstrapMetadataWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success"

// 		successGLM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.success"
// 		// successGLMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_GL_RndBeacon.json.success"
// 		// successGLstWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_GL_RndBeacon.json.success"
// 		// successGLL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.success"
// 		// successGLWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.success"
// 		// successGLWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.success"
// 		// successGLWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.success"
// 		// successGLWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.success"
// 	)

// 	// The rename in case it is a panic (code = 2)
// 	var (
// 		failBootstrapMetadataM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
// 		// failBootstrapMetadataMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
// 		// failBootstrapMetadatastWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
// 		// failBootstrapMetadataL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_2"
// 		// failBootstrapMetadataWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
// 		// failBootstrapMetadataWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_2"
// 		// failBootstrapMetadataWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2"
// 		// failBootstrapMetadataWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_2"

// 		failGLM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_2"
// 		// failGLMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof_GL_RndBeacon.json.failure.code_2"
// 		// failGLstWoEtv     = "requests-done/102-103-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_2"
// 		// failGLL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_2"
// 		// failGLWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_2"
// 		// failGLWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_2"
// 		// failGLWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_2"
// 		// failGLWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_2"
// 	)

// 	testcase := []inpFileNamesCases{
// 		{
// 			Ext: "", Fail: "code", ShouldMatch: true,
// 			Fnames: [][]string{
// 				{correctBootstrapMetadataM, correctGLM},
// 				// {correctBootstrapMetadataWithFailM, correctGLWithFailM},
// 				// {correctBootstrapMetadataWith2FailsM, correctGLWith2FailsM},
// 				// {missingBootstrapMetadataEtv, missingGLEtv},
// 				// {missingBootstrapMetadataStv, missingGLStv},
// 			},
// 			Explainer: "happy path, case M",
// 			ExpectedOutput: [][]string{
// 				{respRndBeaconM},
// 				// {respRndBeaconWithFailM},
// 				// {respRndBeaconWith2FailsM},
// 				// {respRndBeaconWoEtv},
// 				// {respRndBeaconWoStv},
// 			},
// 			ExpToLarge: [][]string{
// 				{toLargeBootstrapMetadataM, toLargeGLM},
// 				// {toLargeBootstrapMetadataWithFailM, toLargeGLWithFailM},
// 				// {toLargeBootstrapMetadataWith2FailsM, toLargeGLWith2FailsM},
// 				// {toLargeBootstrapMetadataWoEtv, toLargeGLWoEtv},
// 				// {toLargeBootstrapMetadataWoStv, toLargeGLWoStv},
// 			},
// 			ExpSuccess: [][]string{
// 				{successBootstrapMetadataM, successGLM},
// 				// {successBootstrapMetadataWithFailM, successGLWithFailM},
// 				// {successBootstrapMetadataWith2FailsM, successGLWith2FailsM},
// 				// {successBootstrapMetadatastWoEtv, successGLstWoEtv},
// 				// {successBootstrapMetadataMWoStv, successGLMWoStv},
// 			},
// 			ExpFailW2: [][]string{
// 				{failBootstrapMetadataM, failGLM},
// 				// {failBootstrapMetadataWithFailM, failGLWithFailM},
// 				// {failBootstrapMetadataWith2FailsM, failGLWith2FailsM},
// 				// {failBootstrapMetadatastWoEtv, failGLstWoEtv},
// 				// {failBootstrapMetadataMWoStv, failGLMWoStv},
// 			},
// 		},
// 		// {
// 		// 	Ext: "large", Fail: "code", ShouldMatch: true,
// 		// 	Fnames: [][]string{
// 		// 		{correctBootstrapMetadataL, correctGLL},
// 		// 		{correctBootstrapMetadataWithFailL, correctGLWithFailL},
// 		// 		{correctBootstrapMetadataWith2FailsL, correctGLWith2FailsL},
// 		// 	},
// 		// 	Explainer: "happy path, case L",
// 		// 	ExpectedOutput: [][]string{
// 		// 		{respRndBeaconL},
// 		// 		{respRndBeaconWithFailL},
// 		// 		{respRndBeaconWith2FailsL},
// 		// 	},
// 		// 	ExpSuccess: [][]string{
// 		// 		{successBootstrapMetadataL, successGLL},
// 		// 		{successBootstrapMetadataWithFailL, successGLWithFailL},
// 		// 		{successBootstrapMetadataWith2FailsL, successGLWith2FailsL},
// 		// 	},
// 		// 	ExpFailW2: [][]string{
// 		// 		{failBootstrapMetadataL, failGLL},
// 		// 		{failBootstrapMetadataWithFailL, failGLWithFailL},
// 		// 		{failBootstrapMetadataWith2FailsL, failGLWith2FailsL},
// 		// 	},
// 		// },
// 		// {
// 		// 	Ext: "", Fail: "code", ShouldMatch: false,
// 		// 	Fnames: [][]string{
// 		// 		{correctBootstrapMetadataL, correctGLL},
// 		// 		{correctBootstrapMetadataWithFailL, correctGLWithFailL},
// 		// 		{correctBootstrapMetadataWith2FailsL, correctGLWith2FailsL},
// 		// 	},
// 		// 	Explainer: "M does not pick the files reserved for L",
// 		// },
// 		// {
// 		// 	Ext: "large", Fail: "code", ShouldMatch: false,
// 		// 	Fnames: [][]string{
// 		// 		{correctBootstrapMetadataM, correctGLM},
// 		// 		{correctBootstrapMetadataWithFailM, correctGLWithFailM},
// 		// 		{correctBootstrapMetadataWith2FailsM, correctGLWith2FailsM},
// 		// 	},
// 		// 	Explainer: "L does not pick the files reserved for M",
// 		// },
// 		// {
// 		// 	Ext: "", Fail: "code", ShouldMatch: false,
// 		// 	Fnames: [][]string{
// 		// 		{notAPoint},
// 		// 		{badName},
// 		// 	},
// 		// 	Explainer: "M does not pick obviously invalid files",
// 		// },
// 		// {
// 		// 	Ext: "large", Fail: "code", ShouldMatch: false,
// 		// 	Fnames: [][]string{
// 		// 		{missingBootstrapMetadataEtv, missingGLEtv},
// 		// 		{missingBootstrapMetadataStv, missingGLStv},
// 		// 		{notAPoint},
// 		// 		{badName},
// 		// 	},
// 		// 	Explainer: "L does not pick obviously invalid files",
// 		// },
// 	}

// 	for _, c := range testcase {
// 		conf := config.Config{}
// 		conf.Version = "0.1.2"
// 		conf.RndBeacon.CanRunFullLarge = c.Ext == "large"

// 		def, err := ExecRndBeaconDefinition(&conf)
// 		assert.NoError(t, err)

// 		t.Run(c.Explainer, func(t *testing.T) {
// 			runInpFileTestCase(t, def, c)
// 		})
// 	}
// }
