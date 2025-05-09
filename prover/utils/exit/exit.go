package exit

import "os"

const (
	limitOverflowExitCode          = 77
	unsatisfiedConstraintsExitCode = 78
)

func OnLimitOverflow() {
	os.Exit(limitOverflowExitCode)
}

func OnSatisfiedConstraints() {
	os.Exit(unsatisfiedConstraintsExitCode)
}
