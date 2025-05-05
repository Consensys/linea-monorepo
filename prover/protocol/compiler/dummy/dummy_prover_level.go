package dummy

import (
	"errors"
	"fmt"
	"sync"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/sirupsen/logrus"
)

// Option is an option for the compiler
type Option func(*optionSet)

// WithMsg tells the [CompileAtProverLvl] to use the given message
// in the panic messages in case it fails. This is handy to differentiate
// which invokation of the compiler failed.
func WithMsg(msg string) Option {
	return func(o *optionSet) {
		o.msg = msg
	}
}

// optionSet is a set of options that can be passed to the compiler.
type optionSet struct {
	// msg is an identifier shown in the panic message of the [CompileAtProverLvl]
	// to help identifying which invokation of the compiler failed.
	msg string
}

// CompileAtProverLvl instantiate the oracle as the prover. Meaning that the
// prover is responsible for checking all the queries and the verifier does not
// see any compiled IOP.
//
// This is useful for quick "manual" testing and debugging. One perk is that
// unlike [CompileAtVerifierLvl] the FS is not pressured as we don't push many
// column in plain-text to the verifier. The drawback is that since it happens
// at prover level, the "errors" result in panics. This makes it not very
// suitable for established unit-tests where we want to analyze the errors.
func CompileAtProverLvl(opts ...Option) func(*wizard.CompiledIOP) {
	os := &optionSet{}
	for _, opt := range opts {
		opt(os)
	}
	return func(comp *wizard.CompiledIOP) {
		compileAtProverLvl(comp, os)
	}
}

func compileAtProverLvl(comp *wizard.CompiledIOP, os *optionSet) {

	/*
		Registers all declared commitments and query parameters
		as messages in the same round. This steps is only relevant
		for the compiledIOP. We elaborate on how to update the provers
		and verifiers to account for this. Additionally, we take the queries
		as we compile them from the `CompiledIOP`.
	*/
	numRounds := comp.NumRounds()

	/*
		The filter returns true, as long as the query has not been marked as
		already compiled. This is to avoid them being compiled a second time.
	*/
	queriesParamsToCompile := comp.QueriesParams.AllUnignoredKeys()
	queriesNoParamsToCompile := comp.QueriesNoParams.AllUnignoredKeys()

	/*
		One step to be run at the end, by verifying every constraint
		"a la mano"
	*/
	comp.RegisterProverAction(numRounds-1, &dummyVerifierProverAction{
		comp:                     comp,
		queriesParamsToCompile:   queriesParamsToCompile,
		queriesNoParamsToCompile: queriesNoParamsToCompile,
		os:                       os,
	})
}

// dummyVerifierProverAction is the action to verify queries at the prover level.
// It implements the [wizard.ProverAction] interface.
type dummyVerifierProverAction struct {
	comp                     *wizard.CompiledIOP
	queriesParamsToCompile   []ifaces.QueryID
	queriesNoParamsToCompile []ifaces.QueryID
	os                       *optionSet
}

// Run executes the dummy verification by checking all queries.
func (a *dummyVerifierProverAction) Run(run *wizard.ProverRuntime) {
	logrus.Infof("started to run the dummy verifier")

	var finalErr error
	lock := sync.Mutex{}

	/*
		Test all the query with parameters
	*/
	parallel.Execute(len(a.queriesParamsToCompile), func(start, stop int) {
		for i := start; i < stop; i++ {
			name := a.queriesParamsToCompile[i]
			lock.Lock()
			q := a.comp.QueriesParams.Data(name)
			lock.Unlock()
			if err := q.Check(run); err != nil {
				err = fmt.Errorf("%v\nfailed %v - %v", finalErr, name, err)
				lock.Lock()
				finalErr = errors.Join(finalErr, err)
				lock.Unlock()
				logrus.Debugf("query %v failed\n", name)
			} else {
				logrus.Debugf("query %v passed\n", name)
			}
		}
	})

	/*
		Test the queries without parameters
	*/
	parallel.Execute(len(a.queriesNoParamsToCompile), func(start, stop int) {
		for i := start; i < stop; i++ {
			name := a.queriesNoParamsToCompile[i]
			lock.Lock()
			q := a.comp.QueriesNoParams.Data(name)
			lock.Unlock()
			if err := q.Check(run); err != nil {
				err = fmt.Errorf("%v\nfailed %v - %v", finalErr, name, err)
				lock.Lock()
				finalErr = errors.Join(finalErr, err)
				lock.Unlock()
			} else {
				logrus.Debugf("query %v passed\n", name)
			}
		}
	})

	if finalErr != nil {
		utils.Panic("dummy.Compile brought errors: msg=%v: err=%v", a.os.msg, finalErr.Error())
	}
}
