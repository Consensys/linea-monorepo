/*
Copyright © 2023 Consensys
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"

	"github.com/consensys/accelerated-crypto-monorepo/backend/config"
	"github.com/consensys/accelerated-crypto-monorepo/backend/files"
	"github.com/consensys/accelerated-crypto-monorepo/backend/prover"
	"github.com/consensys/accelerated-crypto-monorepo/backend/prover/dummycircuit"
	"github.com/consensys/accelerated-crypto-monorepo/backend/prover/outercircuit"
	"github.com/consensys/accelerated-crypto-monorepo/backend/prover/plonkutil"
	"github.com/consensys/accelerated-crypto-monorepo/glue"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg"
	"github.com/consensys/gnark/backend/plonk"
	plonkBn254 "github.com/consensys/gnark/backend/plonk/bn254"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fFullCircuit  bool // fullCircuit is a flag to generate assets for full circuit
	fLightCircuit bool // lightCircuit is a flag to generate assets for light circuit
	fTrustedSRS   bool // trustedSRS is a flag to indicate whether the command fetches the KZG SRS or generates a (unsafe) test one.
)

const (
	circuitFile  = "circuit.bin"
	pkFile       = "proving_key.bin"
	vkFile       = "verifying_key.bin"
	solidityFile = "Verifier.sol"
	manifestFile = "manifest.json"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates assets needed for proof verification - solidity smart contract, circuit, proving and verifying keys",
	Run:   generateAssets,
}

func generateAssets(cmd *cobra.Command, args []string) {
	if (fFullCircuit && fLightCircuit) || (!fFullCircuit && !fLightCircuit) {
		checkError(fmt.Errorf("please specify either --full or --light flag"))
	}

	var (
		circuitName string
	)

	// generates the light circuit, no need to use the kzg setup
	if fLightCircuit {
		logrus.Infof("generating the light circuit")
		dummycircuit.GenSetupLight(fPath)
		return
	}

	// the following is for the full circuit only
	logrus.Infof("generating the full circuit")

	circuitName = "full"
	if config.IS_LARGE {
		circuitName = "full-large"
	}
	var setup plonkutil.Setup

	conf := config.MustGetProver()
	proverOptions := prover.ProverOptions{}
	if conf.WithStateManager {
		proverOptions.WithStateManager([][]any{})
	}

	if conf.WithKeccak {
		proverOptions.WithKeccak()
	}

	if conf.WithEcdsa {
		proverOptions.WithEcdsa(glue.NoTxSignatures)
	}

	iop := prover.GetFullIOP(&proverOptions)

	// Setup for the main circuit
	if fTrustedSRS {
		logrus.Infof("using the provided SRS")
		// fetch the SRS from cache or S3 bucket
		// TODO @gbotrel since we call prover.NewSetup, we don't know the size of the SRS.
		// replace with needed size for large (when fFullCircuit is set)
		srs, err := fetchSRS()
		checkError(err)
		setup = outercircuit.GenSetupFromSRS(iop, srs)
	} else {
		logrus.Infof("using the unsafe SRS")
		setup = outercircuit.GenSetupUnsafe(iop)
	}

	// Create manifest
	manifest := NewManifest(setup.SCS.GetNbConstraints(), circuitName)
	if err := manifest.Write(filepath.Join(fPath, manifestFile)); err != nil {
		checkError(err)
	}

	// Write assets to disk
	if err := writeAsset(setup.SCS, circuitFile); err != nil {
		checkError(err)
	}

	// TODO @gbotrel maybe, do not compress the proving key?
	if err := writeAssetRaw(setup.PK, pkFile); err != nil {
		checkError(err)
	}

	if err := writeAsset(setup.VK, vkFile); err != nil {
		checkError(err)
	}

	{
		destination := filepath.Join(fPath, solidityFile)
		fmt.Println("writing", destination, " ...")
		f := files.MustOverwrite(destination)
		defer f.Close()

		if err := setup.VK.ExportSolidity(f); err != nil {
			checkError(err)
		}
	}

	color.Green("done ✅")
}

func writeAsset(asset io.WriterTo, name string) error {
	destination := filepath.Join(fPath, name)
	fmt.Println("writing", destination, " ...")
	f := files.MustOverwrite(destination)
	_, err := asset.WriteTo(f)
	defer f.Close()
	return err
}

func writeAssetRaw(asset plonk.ProvingKey, name string) error {
	destination := filepath.Join(fPath, name)
	fmt.Println("writing", destination, " ...")
	f := files.MustOverwrite(destination)
	buf := bufio.NewWriterSize(f, 50_000_000_000)
	defer f.Close()
	_, err := asset.(*plonkBn254.ProvingKey).WriteRawTo(buf)
	buf.Flush()
	return err
}

func fetchSRS() (*kzg.SRS, error) {
	// TODO @gbotrel add a cache in the home directory to avoid re-downloading the SRS

	const srsS3Bucket string = "zk-uat-prover"
	const srsS3Key string = "ignition-srs/kzg_srs_100800000_bn254_mainignition"

	// download the SRS from S3
	body, err := downloadFromS3(srsS3Bucket, "ignition-srs/kzg_srs_100800000_bn254_mainignition")
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var kzgSRS kzg.SRS
	_, err = kzgSRS.ReadFrom(body)
	if err != nil {
		return nil, err
	}

	return &kzgSRS, nil
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().BoolVar(&fFullCircuit, "full", false, "Generate assets for full circuit")
	generateCmd.Flags().BoolVar(&fLightCircuit, "light", false, "Generate assets for light circuit")
	generateCmd.Flags().BoolVar(&fTrustedSRS, "trusted-srs", false, "Fetch the kzg SRS from cache or S3 bucket. If false, generates a test SRS.")

	generateCmd.MarkFlagsMutuallyExclusive("full", "light")
}
