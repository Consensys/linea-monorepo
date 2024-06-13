package plonk

import (
	"math/big"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	globalCs "github.com/consensys/gnark/constraint"
	cs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	fcs "github.com/consensys/gnark/frontend/cs"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// Struct gathering synchronization variable for the prover
type solverSync struct {
	// The commitment witness
	comChan chan []field.Element
	// The commitment value
	randChan chan field.Element
	// The final solution
	solChan chan *cs.SparseR1CSSolution
}

// This function is responsible for scheduling the assignment of the Wizard
// columns related to the currently compiled Plonk circuit. It is used
// specifically for when we use BBS commitment as part of the circuit.
//
// In essence, the function works by starting the Plonk solver in the
// background. From then on, the function registers two prover functions, one
// for the first round where we assign the BBS commitment polynomial and one
// for the second round where we assign the LRO polynomials. The solver
// goroutine terminates at the end of the second round and is paused between the
// two rounds.
func (ctx *compilationCtx) RegisterCommitProver() {

	commitHintID := solver.GetHintID(fcs.Bsb22CommitmentComputePlaceholder)

	// The first function bootstraps the solver and return once the commitment has been assigned
	commitProver := func(run *wizard.ProverRuntime) {

		for i := 0; i < ctx.nbInstances; i++ {

			// Initialize the channels
			solSync := solverSync{
				comChan:  make(chan []field.Element, 1),
				randChan: make(chan field.Element, 1),
				solChan:  make(chan *cs.SparseR1CSSolution, 1),
			}

			// Store the channels in the runtime so that we can
			// access them in later rounds
			run.State.InsertNew(ctx.Sprintf("SOLSYNC_%v", i), solSync)

			// Let the assigner return an assignment
			assignment := ctx.Plonk.WitnessAssigner[i]()

			// Check that both the assignment and the base
			// circuit have the same type
			if reflect.TypeOf(ctx.Plonk.Circuit) != reflect.TypeOf(assignment) {
				utils.Panic("circuit and assignment do not have the same type")
			}

			// Also assigns the public witness
			publicWitness, err := frontend.NewWitness(assignment, ecc.BLS12_377.ScalarField(), frontend.PublicOnly())
			if err != nil {
				utils.Panic("Could not parse the assignment into a public witness")
			}

			// Converts it as a smart-vector
			pubWitSV := smartvectors.RightZeroPadded(
				[]field.Element(publicWitness.Vector().(fr.Vector)),
				ctx.DomainSize(),
			)

			// Assign the public witness
			run.AssignColumn(ctx.Columns.PI[i].GetColID(), pubWitSV)

			// Parse it as witness
			witness, err := frontend.NewWitness(assignment, ecc.BLS12_377.ScalarField())
			if err != nil {
				utils.Panic("Could not parse the assignment into a witness")
			}

			// Start the solver in the background
			go func() {
				// Solve the circuit
				sol_, err := ctx.Plonk.SPR.Solve(
					witness,
					// Inject our special hint for the commitment. It's goal is to
					// force the solver to pause once the commitment
					solver.OverrideHint(
						commitHintID,
						ctx.solverCommitmentHint(solSync.comChan, solSync.randChan),
					),
				)
				if err != nil {
					utils.Panic("Error in the solver: %v", err)
				}

				// Once the solver has finished, return the solution
				// in the dedicated channel and terminate the solver task
				solSync.solChan <- sol_.(*cs.SparseR1CSSolution)
				close(solSync.solChan)
			}()

			// Get the commitment from the chan once ready
			com := <-solSync.comChan

			// And assign it in the runtime
			run.AssignColumn(ctx.Columns.Cp[i].GetColID(), smartvectors.NewRegular(com))
		}

	}

	// The prover part responsible for assigning the LRO polynomials
	lroProver := func(run *wizard.ProverRuntime) {

		for i := 0; i < ctx.nbInstances; i++ {
			// Retrive the solsync
			solsync := run.State.MustGet(ctx.Sprintf("SOLSYNC_%v", i)).(solverSync)
			run.State.TryDel(ctx.Sprintf("SOLSYNC_%v", i))

			// Inject the coin which will be assigned to the randomness
			solsync.randChan <- run.GetRandomCoinField(ctx.Columns.Hcp.Name)
			close(solsync.randChan)

			// And we block until the solver has completely finished
			// and returns a solution
			solution := <-solsync.solChan

			// And finally, we assign L, R, O from it
			run.AssignColumn(ctx.Columns.L[i].GetColID(), smartvectors.NewRegular(solution.L))
			run.AssignColumn(ctx.Columns.R[i].GetColID(), smartvectors.NewRegular(solution.R))
			run.AssignColumn(ctx.Columns.O[i].GetColID(), smartvectors.NewRegular(solution.O))
		}

	}

	ctx.comp.SubProvers.AppendToInner(ctx.round, commitProver)
	ctx.comp.SubProvers.AppendToInner(ctx.round+1, lroProver)

}

// Computes the replacement hint that we pass to the gnark's solver in place of
// the default BBS22 initial challenge hint. This hint will be passed to the
// gnark Solver. Instead of computing and hashing a group element as in the
// BBS22 paper. We instead use the FS mechanism that is embedded in the wizard.
// As a reminder, the shape of the committed polynomial is as follows: it is all
// zero except in the position containing committed polynomials.
//
// To proceed we need to allocate the column outside of the Solver function; the
// assignment of the column cannot be done at the same time as the rest of the
// Plonk witness. Thus, the function will only extract the corresponding column
// , pass it to a channel and pause. It will resume in a later stage of the
// Wizard proving runtime to complete the solving once the the challenge to
// return is available.
func (ctx *compilationCtx) solverCommitmentHint(
	// Channel through which the committed poly is obtained
	pi2Chan chan []field.Element,
	// Channel through which the randomness is injected back
	randChan chan field.Element,
) func(_ *big.Int, ins, outs []*big.Int) error {

	return func(_ *big.Int, ins, outs []*big.Int) error {
		// pi2 is meant to store a copy of the BBS22 "commitment" which are
		// collecting in the following lines of code. The polynomial is
		// constructed as follows. All "non-committed" wires are zero and
		// the only non-committed values are
		var (
			pi2    = make([]field.Element, ctx.DomainSize())
			spr    = ctx.Plonk.SPR
			offset = spr.GetNbPublicVariables()
			// The first input of the function Hint function does not correspond
			// to a committed wire but to a position to use in the
			// `PlonkCommitments` of the circuit. My guess is that it is used
			// in the multi-round BBS22 case (which the current implementation
			// of Plonk in Wizard does not support). We still reflect that here
			// in case we want to support it in the future.
			comDepth = int(ins[0].Int64())
		)

		// Trims the above-mentionned comdepth
		ins = ins[1:]

		// We use the first commit ID because, there is allegedly only one
		// commitment
		sprCommittedIDs := spr.CommitmentInfo.(globalCs.PlonkCommitments)[comDepth].Committed
		for i := range ins {
			pi2[offset+sprCommittedIDs[i]].SetBigInt(ins[i])
		}

		// Sends the commitment value to the runtime
		pi2Chan <- pi2
		close(pi2Chan)

		// Use a custom way of deriving the commitment from a random coin
		// that is injected by the wizard runtime thereafter.
		commitmentVal := <-randChan
		commitmentVal.BigInt(outs[0])
		return nil
	}
}

// Return whether the current circuit uses api.Commit
func (ctx *compilationCtx) HasCommitment() bool {
	// goes through all the type casting and accesses
	commitmentsInfo := ctx.
		Plonk.SPR.
		CommitmentInfo.(globalCs.PlonkCommitments)

	// Sanity-checks to guard against passing a circuit with more than one commitment
	if len(commitmentsInfo) > 1 {
		utils.Panic("unsupported : cannot wizardize a Plonk circuit with more than one commitment (found %v)", len(commitmentsInfo))
	}

	return len(commitmentsInfo) > 0
}

// Returns the Plonk commitment info of the compiled gnark circuit. This is used
// derive information such as which wires are being committed and how many
// commitments there are.
func (ctx *compilationCtx) CommitmentInfo() globalCs.PlonkCommitment {
	// goes through all the type casting and accesses
	commitmentsInfo := ctx.
		Plonk.SPR.
		CommitmentInfo.(globalCs.PlonkCommitments)
	// Sanity-checks to guard against passing a circuit with more than one commitment
	if len(commitmentsInfo) != 1 {
		utils.Panic("unsupported : cannot wizardize a Plonk circuit with more than one commitment (found %v)", len(commitmentsInfo))
	}

	return commitmentsInfo[0]
}
