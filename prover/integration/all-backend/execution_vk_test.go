package allbackend

import (
	"context"
	"sync"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
	plonkBls12377 "github.com/consensys/gnark/backend/plonk/bls12-377"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/stretchr/testify/require"
)

func TestExecutionVk(t *testing.T) {
	CdProver(t)
	cfg, err := config.NewConfigFromFile("./integration/all-backend/config-integration-light.toml")
	require.NoError(t, err)

	var (
		wg             sync.WaitGroup
		execKey, piKey kzg.VerifyingKey
	)

	wg.Add(2)
	loadKzg := func(res *kzg.VerifyingKey, circId circuits.CircuitID) {
		setup, err := circuits.LoadSetup(cfg, circId)
		require.NoError(t, err)
		*res = setup.VerifyingKey.(*plonkBls12377.VerifyingKey).Kzg
		wg.Done()
	}

	go loadKzg(&execKey, circuits.ExecutionCircuitID)
	go loadKzg(&piKey, circuits.PublicInputInterconnectionCircuitID)

	srsStore, err := circuits.NewSRSStore(cfg.PathForSRS())
	require.NoError(t, err)

	canonical, _, err := srsStore.GetSRS(context.TODO(), &SrsSpec{
		Id:      ecc.BLS12_377,
		MaxSize: 16,
	})
	require.NoError(t, err)
	typedCanonical := canonical.(*kzg.SRS)

	wg.Wait()

	matchWithSrs := func(name string, vk *kzg.VerifyingKey) {
		require.Equal(t, typedCanonical.Vk.G1.X, vk.G1.X, name)
	}

	matchWithSrs("public input", &piKey)
	matchWithSrs("execution", &execKey) // THIS LINE FAILS
}
