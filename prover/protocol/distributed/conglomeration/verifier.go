package conglomeration

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	fiatShamirHistoryStr    = "fiat-shamir-history"
	fiatShamirTranscriptStr = "fiat-shamir-transcript"
)

var _ wizard.Runtime = &RuntimeWithReplacedFS{}

// PreVortexVerifierStep is a step replicating the verifier of the tmpl at round
// `Round` before the Vortex compilation step.
type PreVortexVerifierStep struct {
	Ctxs      []*recursionCtx
	Round     int
	isSkipped bool
}

// RuntimeWithReplacedFS is a runtime that wraps another runtime and replaces
// the returns of [Fs] and [FsHistory] with the ones provided in the struct.
type RuntimeWithReplacedFS struct {
	wizard.Runtime
	FS                *fiatshamir.State
	FiatShamirHistory [][2][]field.Element
}

// GnarkRuntimeWithReplacedFS is a GnarkRuntime that wraps another runtime and
// replaces the returns of [Fs] and [FsHistory] with the ones provided in the struct.
type GnarkRuntimeWithReplacedFS struct {
	wizard.GnarkRuntime
	FS                *fiatshamir.GnarkFiatShamir
	FiatShamirHistory [][2][]frontend.Variable
}

// FsJoinHook is a fiat-shamir hook whose purpose is to join the interal
// fiat-shamir states of the segment proofs into the main one. It works
// by setting the main fs-state as the hash of the internal states.
type FsJoinHook struct {
	Ctxs      []*recursionCtx
	isSkipped bool
}

// SubFsInitialize is a fiat-shamir hook whose purpose is to initialize the
// internal fiat-shamir states of the segment proofs. It is set as a FS hook
// as it is guaranteed to be run before any prover/verifier step, ensuring the
// fs states are available at the beginning.
type SubFsInitialize struct {
	Ctxs      []*recursionCtx
	isSkipped bool
}

func (pa PreVortexVerifierStep) Run(run wizard.Runtime) error {

	var err error

	for _, ctx := range pa.Ctxs {

		generateRandomCoins(run, ctx, pa.Round)

		// Wraps the runtime into a translation adapter
		var (
			wrappedRun = &runtimeTranslator{Prefix: ctx.Translator.Prefix, Rt: run}
		)

		// Copy the verifier actions of the template into the target
		for _, va := range ctx.VerifierActions[pa.Round] {
			err = errors.Join(err, va.Run(wrappedRun))
		}
	}

	return err
}

// generateRandomCoins generates all the coins for the current round
// so that they are made available to the forthcoming verifier actions.
func generateRandomCoins(run wizard.Runtime, ctx *recursionCtx, currRound int) {

	var (
		spec = run.GetSpec()
		// Wraps the runtime into a translation adapter, first to get the FS state
		// and history.
		wrappedRun      wizard.Runtime = &runtimeTranslator{Prefix: ctx.Translator.Prefix, Rt: run}
		fsAny, _                       = wrappedRun.GetState(fiatShamirTranscriptStr)
		fsHistoryAny, _                = wrappedRun.GetState(fiatShamirHistoryStr)
		fs                             = fsAny.(*fiatshamir.State)
		fsHistory                      = fsHistoryAny.([][2][]field.Element)
		initialState                   = fs.State()
	)

	// Wraps it a second time
	wrappedRun = &RuntimeWithReplacedFS{
		Runtime:           wrappedRun,
		FS:                fs,
		FiatShamirHistory: fsHistory,
	}

	if currRound > 0 {

		cols := ctx.Columns[currRound-1]
		for _, col := range cols {

			name := unprefix(ctx.Translator.Prefix, col.GetColID())
			if ctx.Tmpl.Columns.IsExplicitlyExcludedFromProverFS(name) {
				continue
			}

			instance := run.GetColumn(col.GetColID())
			fs.UpdateSV(instance)
		}

		queries := ctx.QueryParams[currRound-1]
		for _, q := range queries {
			params := run.GetParams(q.Name())
			params.UpdateFS(fs)
		}
	}

	toCompute := ctx.Coins[currRound]
	for _, coin := range toCompute {
		info := spec.Coins.Data(coin.Name)
		value := info.Sample(fs)
		run.InsertCoin(coin.Name, value)
	}

	for _, fsHook := range ctx.FsHooks[currRound] {
		fsHook.Run(wrappedRun)
	}

	fsHistory[currRound] = [2][]field.Element{
		initialState,
		fs.State(),
	}

	wrappedRun.SetState(fiatShamirHistoryStr, fsHistory)
	wrappedRun.SetState(fiatShamirTranscriptStr, fs)
}

// Fs returns the Fiat-Shamir state
func (run *RuntimeWithReplacedFS) Fs() *fiatshamir.State {
	return run.FS
}

// FsHistory returns the Fiat-Shamir state history
func (run *RuntimeWithReplacedFS) FsHistory() [][2][]field.Element {
	return run.FiatShamirHistory
}

func (pa PreVortexVerifierStep) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	for _, ctx := range pa.Ctxs {

		pa.generateRandomCoinsGnark(api, run, ctx, pa.Round)

		// Wraps the runtime into a translation adapter
		var wrappedRun = &gnarkRuntimeTranslator{Prefix: ctx.Translator.Prefix, Rt: run}

		// Copy the verifier actions of the template into the target
		for _, va := range ctx.VerifierActions[pa.Round] {
			va.RunGnark(api, wrappedRun)
		}
	}
}

// generateRandomCoinsGnark generates all the coins for the current round
// so that they are made available to the forthcoming verifier actions.
func (pa PreVortexVerifierStep) generateRandomCoinsGnark(api frontend.API, run wizard.GnarkRuntime, ctx *recursionCtx, currRound int) {

	const (
		fiatShamirHistoryStr    = "fiat-shamir-history"
		fiatShamirTranscriptStr = "fiat-shamir-transcript"
	)

	var (
		spec = run.GetSpec()
		// Wraps the runtime into a translation adapter, first to get the FS state
		// and history.
		wrappedRun      wizard.GnarkRuntime = &gnarkRuntimeTranslator{Prefix: ctx.Translator.Prefix, Rt: run}
		fsAny, _                            = wrappedRun.GetState(fiatShamirTranscriptStr)
		fsHistoryAny, _                     = wrappedRun.GetState(fiatShamirHistoryStr)
		fs                                  = fsAny.(*fiatshamir.GnarkFiatShamir)
		fsHistory                           = fsHistoryAny.([][2][]frontend.Variable)
		initialState                        = fs.State()
	)

	// Wraps it a second time
	wrappedRun = &GnarkRuntimeWithReplacedFS{
		GnarkRuntime:      wrappedRun,
		FS:                fs,
		FiatShamirHistory: fsHistory,
	}

	if currRound > 0 {

		cols := ctx.Columns[currRound-1]
		for _, col := range cols {

			name := unprefix(ctx.Translator.Prefix, col.GetColID())
			if ctx.Tmpl.Columns.IsExplicitlyExcludedFromProverFS(name) {
				continue
			}

			instance := run.GetColumn(col.GetColID())
			fs.UpdateVec(instance)
		}

		queries := ctx.QueryParams[currRound-1]
		for _, q := range queries {
			params := run.GetParams(q.Name())
			params.UpdateFS(fs)
		}
	}

	toCompute := ctx.Coins[currRound]
	for _, c := range toCompute {
		info := spec.Coins.Data(c.Name)
		switch info.Type {
		case coin.Field:
			value := fs.RandomField()
			run.InsertCoin(c.Name, value)
		case coin.IntegerVec:
			value := fs.RandomManyIntegers(info.Size, info.UpperBound)
			run.InsertCoin(c.Name, value)
		}
	}

	for _, fsHook := range ctx.FsHooks[currRound] {
		fsHook.RunGnark(api, wrappedRun)
	}

	fsHistory[currRound] = [2][]frontend.Variable{
		initialState,
		fs.State(),
	}

	wrappedRun.SetState(fiatShamirHistoryStr, fsHistory)
	wrappedRun.SetState(fiatShamirTranscriptStr, fs)
}

// Fs returns the Fiat-Shamir state
func (run *GnarkRuntimeWithReplacedFS) Fs() *fiatshamir.GnarkFiatShamir {
	return run.FS
}

// FsHistory returns the Fiat-Shamir state history
func (run *GnarkRuntimeWithReplacedFS) FsHistory() [][2][]frontend.Variable {
	return run.FiatShamirHistory
}

func (pa *PreVortexVerifierStep) IsSkipped() bool {
	return pa.isSkipped
}

func (pa *PreVortexVerifierStep) Skip() {
	pa.isSkipped = true
}

func (fs *FsJoinHook) Run(run wizard.Runtime) error {

	mainState := field.NewElement(0)

	for _, ctx := range fs.Ctxs {

		var (
			wrappedRun wizard.Runtime = &runtimeTranslator{Prefix: ctx.Translator.Prefix, Rt: run}
			fsAny, _                  = wrappedRun.GetState(fiatShamirTranscriptStr)
			fs                        = fsAny.(*fiatshamir.State)
		)

		mainState = mimc.BlockCompression(mainState, fs.State()[0])
	}

	fmt.Printf("[join-fs-hook] mainState: %v\n", mainState.String())

	run.Fs().SetState([]field.Element{mainState})

	return nil
}

func (fs *FsJoinHook) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	mainState := frontend.Variable(0)

	for _, ctx := range fs.Ctxs {

		var (
			wrappedRun wizard.GnarkRuntime = &gnarkRuntimeTranslator{Prefix: ctx.Translator.Prefix, Rt: run}
			fsAny, _                       = wrappedRun.GetState(fiatShamirTranscriptStr)
			fs                             = fsAny.(*fiatshamir.GnarkFiatShamir)
		)

		mainState = mimc.GnarkBlockCompression(api, mainState, fs.State()[0])
	}

	run.Fs().SetState([]frontend.Variable{mainState})
}

func (fs *FsJoinHook) Skip() {
	fs.isSkipped = true
}

func (fs *FsJoinHook) IsSkipped() bool {
	return fs.isSkipped
}

func (fs *SubFsInitialize) Run(run wizard.Runtime) error {

	for _, ctx := range fs.Ctxs {

		var (
			wrappedRun wizard.Runtime = &runtimeTranslator{Prefix: ctx.Translator.Prefix, Rt: run}
			fs                        = fiatshamir.NewMiMCFiatShamir()
			fsHistory                 = make([][2][]field.Element, ctx.LastRound+1)
		)

		fs.Update(ctx.Tmpl.FiatShamirSetup)

		wrappedRun.SetState(fiatShamirHistoryStr, fsHistory)
		wrappedRun.SetState(fiatShamirTranscriptStr, fs)
	}

	return nil
}

func (fs *SubFsInitialize) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	for _, ctx := range fs.Ctxs {

		var (
			wrappedRun wizard.GnarkRuntime = &gnarkRuntimeTranslator{Prefix: ctx.Translator.Prefix, Rt: run}
			fs                             = fiatshamir.NewGnarkFiatShamir(api, run.GetHasherFactory())
			fsHistory                      = make([][2][]frontend.Variable, ctx.LastRound+1)
		)

		fs.Update(api, ctx.Tmpl.FiatShamirSetup)

		wrappedRun.SetState(fiatShamirHistoryStr, fsHistory)
		wrappedRun.SetState(fiatShamirTranscriptStr, fs)
	}
}

func (fs *SubFsInitialize) Skip() {
	fs.isSkipped = true
}

func (fs *SubFsInitialize) IsSkipped() bool {
	return fs.isSkipped
}
