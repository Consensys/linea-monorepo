package main

import "log/slog"

const (
	path_add_g1  = "bls_g1_add_input.csv"
	path_add_g2  = "bls_g2_add_input.csv"
	path_msm_g1  = "bls_g1_msm_inputs.csv"
	path_msm_g2  = "bls_g2_msm_inputs.csv"
	path_pairing = "bls_pairing_inputs.csv"
	path_map_g1  = "bls_g1_map_inputs.csv"
	path_map_g2  = "bls_g2_map_inputs.csv"
)

const (
	maxNbPairingInputs = 4
	maxNbMsmInputs     = 4
	nbRepetitionsMap   = 3
)

//go:generate go run .
func main() {
	logger := slog.Default()

	logger.Info("Generating test data for BLS operations")
	logger.Info("Generating ADD test data", "path_g1", path_add_g1, "path_g2", path_add_g2)
	if err := mainAdd(); err != nil {
		logger.Error("Failed to generate ADD test data", "error", err)
		return
	}
	logger.Info("Generating MAP test data", "path_g1", path_msm_g1, "path_g2", path_msm_g2)
	if err := mainMsm(); err != nil {
		logger.Error("Failed to generate MSM test data", "error", err)
		return
	}
	logger.Info("Generating PAIRING test data", "path_pairing", path_pairing)
	if err := mainPairing(); err != nil {
		logger.Error("Failed to generate PAIRING test data", "error", err)
		return
	}
	logger.Info("Generating MAP test data", "path_g1", path_map_g1, "path_g2", path_map_g2)
	if err := mainMap(); err != nil {
		logger.Error("Failed to generate MAP test data", "error", err)
		return
	}
}
