package statesummary

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/backend/execution/statemanager/mock"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
)

func TestStateSummaryInternal(t *testing.T) {

	var (
		addresses = []types.EthAddress{
			types.DummyAddress(32),
			types.DummyAddress(64),
			types.DummyAddress(54),
			types.DummyAddress(23),
		}

		storageKeys = []types.FullBytes32{
			types.DummyFullByte(102),
			types.DummyFullByte(1002),
			types.DummyFullByte(1012),
			types.DummyFullByte(1023),
		}

		storageValue = []types.FullBytes32{
			types.DummyFullByte(202),
			types.DummyFullByte(2002),
			types.DummyFullByte(2012),
			types.DummyFullByte(2023),
		}
	)

	var ss StateSummary

	getInitialState := func() mock.State {
		state := mock.State{}
		state.InsertEOA(addresses[0], 0, big.NewInt(400))
		state.InsertEOA(addresses[1], 0, big.NewInt(200))
		state.InsertContract(addresses[2], types.DummyBytes32(67), types.DummyFullByte(56), 100)
		state.InsertContract(addresses[3], types.DummyBytes32(76), types.DummyFullByte(57), 102)
		state.SetStorage(addresses[2], storageKeys[0], storageValue[0])
		state.SetStorage(addresses[2], storageKeys[1], storageValue[1])
		state.SetStorage(addresses[3], storageKeys[2], storageValue[2])
		state.SetStorage(addresses[3], storageKeys[3], storageValue[3])
		return state
	}

	testCases := []struct {
		Explainer     string
		StateLogsGens func(initState mock.State) [][]mock.StateAccessLog
	}{
		{
			Explainer: "Reverted transaction",
			StateLogsGens: func(initState mock.State) [][]mock.StateAccessLog {
				return mock.NewStateLogBuilder(100, initState).
					WithAddress(addresses[0]).
					IncNonce().
					WriteBalance(big.NewInt(100)).
					Done()
			},
		},
		{
			Explainer: "Update contract and its storage",
			StateLogsGens: func(initState mock.State) [][]mock.StateAccessLog {
				return mock.NewStateLogBuilder(100, initState).
					WithAddress(addresses[2]).
					ReadStorage(storageKeys[0]).
					WriteStorage(storageKeys[1], storageValue[0]).
					Done()
			},
		},
		{
			Explainer: "Reading two contracts",
			StateLogsGens: func(initState mock.State) [][]mock.StateAccessLog {
				return mock.NewStateLogBuilder(100, initState).
					WithAddress(addresses[2]).
					ReadStorage(storageKeys[0]).
					ReadStorage(storageKeys[1]).
					ReadStorage(storageKeys[2]).
					ReadStorage(storageKeys[3]).
					WithAddress(addresses[3]).
					ReadStorage(storageKeys[0]).
					ReadStorage(storageKeys[1]).
					ReadStorage(storageKeys[2]).
					ReadStorage(storageKeys[3]).
					Done()
			},
		},
		{
			Explainer: "Contract deletion",
			StateLogsGens: func(initState mock.State) [][]mock.StateAccessLog {
				return mock.NewStateLogBuilder(100, initState).
					WithAddress(addresses[2]).
					WriteStorage(storageKeys[1], storageValue[3]).
					WriteStorage(storageKeys[2], storageValue[3]).
					WriteStorage(storageKeys[0], types.FullBytes32{}).
					ReadStorage(storageKeys[3]).
					EraseAccount().
					Done()

			},
		},
		{
			Explainer: "one contract, two blocks",
			StateLogsGens: func(initState mock.State) [][]mock.StateAccessLog {
				return mock.NewStateLogBuilder(100, initState).
					WithAddress(addresses[2]).
					ReadStorage(storageKeys[0]).
					WriteStorage(storageKeys[1], storageValue[0]).
					GoNextBlock().
					WithAddress(addresses[2]).
					ReadStorage(storageKeys[0]).
					WriteStorage(storageKeys[1], storageValue[3]).
					Done()
			},
		},
		{
			Explainer: "Redeployed contract",
			StateLogsGens: func(initState mock.State) [][]mock.StateAccessLog {
				return mock.NewStateLogBuilder(100, initState).
					WithAddress(addresses[2]).
					WriteStorage(storageKeys[1], storageValue[3]).
					WriteStorage(storageKeys[2], storageValue[3]).
					WriteStorage(storageKeys[0], types.FullBytes32{}).
					ReadStorage(storageKeys[3]).
					EraseAccount().
					InitContract(45, types.DummyFullByte(5679), types.DummyBytes32(346)).
					WriteStorage(storageKeys[2], storageValue[1]).
					ReadStorage(storageKeys[0]).
					Done()
			},
		},
		{
			Explainer: "Contract created then deleted in another account",
			StateLogsGens: func(initState mock.State) [][]mock.StateAccessLog {
				return mock.NewStateLogBuilder(100, initState).
					WithAddress(addresses[2]).
					WriteStorage(storageKeys[1], storageValue[3]).
					WriteStorage(storageKeys[2], storageValue[3]).
					WriteStorage(storageKeys[0], types.FullBytes32{}).
					ReadStorage(storageKeys[3]).
					EraseAccount().
					GoNextBlock().
					WithAddress(addresses[2]).
					InitContract(45, types.DummyFullByte(5679), types.DummyBytes32(346)).
					WriteStorage(storageKeys[2], storageValue[1]).
					ReadStorage(storageKeys[0]).
					Done()
			},
		},
		{
			Explainer: "Reading a non-existing contract",
			StateLogsGens: func(initState mock.State) [][]mock.StateAccessLog {
				return mock.NewStateLogBuilder(100, initState).
					WithAddress(types.DummyAddress(0)).
					ReadBalance().Done()
			},
		},
	}

	for i, tCase := range testCases {

		t.Run(fmt.Sprintf("test-case-%v", i), func(t *testing.T) {

			t.Logf("Test case explainer: %v", tCase.Explainer)

			define := func(b *wizard.Builder) {
				ss = DefineStateSummaryModule(b.CompiledIOP, 1<<5)
			}

			prove := func(run *wizard.ProverRuntime) {

				var (
					initState    = getInitialState()
					shomeiState  = mock.InitShomeiState(initState)
					stateLogs    = tCase.StateLogsGens(initState)
					shomeiTraces = mock.StateLogsToShomeiTraces(shomeiState, stateLogs)
				)

				ss.Assign(run, shomeiTraces)
			}

			comp := wizard.Compile(define, dummy.Compile)
			proof := wizard.Prove(comp, prove)
			err := wizard.Verify(comp, proof)

			if err != nil {
				t.Fatalf("verification failed: %v", err)
			}
		})

	}

}
