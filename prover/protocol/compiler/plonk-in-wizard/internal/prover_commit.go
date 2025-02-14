package plonk

import (
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/backend/witness"
	globalCs "github.com/consensys/gnark/constraint"
	cs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/gnark/constraint/solver"
	fcs "github.com/consensys/gnark/frontend/cs"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
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

type (
	// initialCommitProverAction is a wrapper-type for [compilationCtx] implementing
	// the interface [wizard.ProverAction]. It is responsible, when using the
	// BBS22 commitment feature, to assign the Cp and PI polynomials so that the
	// BBS22 randomness can be derived.
	initialBBSProverAction struct {
		compilationCtx
		proverStateLock *sync.Mutex
	}
	// lrCommitProverAction is a wrapper-type for [compilationCtx] implementing the
	// interface [wizard.ProverAction]. It is responsible, when using the BBS22
	// commitment feature, to assign the LRO polynomials once the BBS22
	// randomness has been derived.
	lroCommitProverAction struct {
		compilationCtx
		proverStateLock *sync.Mutex
	}
)

// Run implements the [wizard.ProverAction] interface.
func (pa initialBBSProverAction) Run(run *wizard.ProverRuntime, wa WitnessAssigner) {

	var (
		ctx             = compilationCtx(pa.compilationCtx)
		numEffInstances = wa.NumEffWitnesses(run)
	)

	// Store the information

	parallel.Execute(pa.maxNbInstances, func(start, stop int) {
		for i := start; i < stop; i++ {

			if i >= numEffInstances {
				run.AssignColumn(ctx.Columns.TinyPI[i].GetColID(), smartvectors.NewConstant(field.Zero(), ctx.Columns.TinyPI[i].Size()))
				run.AssignColumn(ctx.Columns.Cp[i].GetColID(), smartvectors.NewConstant(field.Zero(), ctx.Columns.Cp[i].Size()))
				run.AssignColumn(ctx.Columns.Activators[i].GetColID(), smartvectors.NewConstant(field.Zero(), 1))
				continue
			}

			// Initialize the channels
			solSync := solverSync{
				comChan:  make(chan []field.Element, 1),
				randChan: make(chan field.Element, 1),
				solChan:  make(chan *cs.SparseR1CSSolution, 1),
			}

			// Store the channels in the runtime so that we can
			// access them in later rounds
			pa.proverStateLock.Lock()
			run.State.InsertNew(ctx.Sprintf("SOLSYNC_%v", i), solSync)
			pa.proverStateLock.Unlock()

			// Create the witness assignment
			witness, pubWitness, err := wa.Assign(run, i)
			if err != nil {
				utils.Panic("Could not create the witness: %v", err)
			}

			if ctx.TinyPISize() > 0 {
				// Convert public witness to smart-vector
				pubWitSV := smartvectors.RightZeroPadded(
					[]field.Element(pubWitness.Vector().(fr.Vector)),
					ctx.TinyPISize(),
				)

				// Assign the public witness
				run.AssignColumn(ctx.Columns.TinyPI[i].GetColID(), pubWitSV)
			}

			// This starts the gnark solver in the background. The current
			// function does not wait for it to terminate as it execution will
			// span over the next round. The current function will however wait
			// for Cp to be available before returning as we need it to derive
			// the randomness.
			go ctx.runGnarkPlonkProver(witness, &solSync)

			// Get the commitment from the chan once ready
			com := <-solSync.comChan

			// And assign it in the runtime
			run.AssignColumn(ctx.Columns.Cp[i].GetColID(), smartvectors.NewRegular(com))
			run.AssignColumn(ctx.Columns.Activators[i].GetColID(), smartvectors.NewConstant(field.One(), 1))
		}
	})
}

// Run implements the [wizard.ProverAction] interface
func (pa lroCommitProverAction) Run(run *wizard.ProverRuntime) {

	ctx := compilationCtx(pa.compilationCtx)

	parallel.Execute(ctx.maxNbInstances, func(start, stop int) {
		for i := start; i < stop; i++ {

			// Retrieve the solsync. Not finding it means the instance is not
			// used.
			pa.proverStateLock.Lock()
			solsync_, foundSolSync := run.State.TryGet(ctx.Sprintf("SOLSYNC_%v", i))
			run.State.TryDel(ctx.Sprintf("SOLSYNC_%v", i))
			pa.proverStateLock.Unlock()

			if !foundSolSync {
				zeroCol := smartvectors.NewConstant(field.Zero(), ctx.Columns.L[i].Size())
				run.AssignColumn(ctx.Columns.L[i].GetColID(), zeroCol)
				run.AssignColumn(ctx.Columns.R[i].GetColID(), zeroCol)
				run.AssignColumn(ctx.Columns.O[i].GetColID(), zeroCol)
				continue
			}

			// Inject the coin which will be assigned to the randomness
			solsync := solsync_.(solverSync)
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

	})

	if ctx.RangeCheck.Enabled && !ctx.RangeCheck.wasCancelled {
		ctx.assignRangeChecked(run)
	}
}

// Run the gnark solver and put the result in solSync.solChan
func (ctx compilationCtx) runGnarkPlonkProver(
	witness witness.Witness,
	solSync *solverSync,
) {

	// This is the hint used to derive the BBS22 randomness
	commitHintID := solver.GetHintID(fcs.Bsb22CommitmentComputePlaceholder)

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
		utils.Panic("Error in the solver: circ=%v err=%v", ctx.name, err)
	}

	// Once the solver has finished, return the solution
	// in the dedicated channel and terminate the solver task
	solSync.solChan <- sol_.(*cs.SparseR1CSSolution)
	close(solSync.solChan)
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
