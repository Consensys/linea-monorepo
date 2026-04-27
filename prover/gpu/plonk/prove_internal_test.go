//go:build cuda

package plonk

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunGPUProveStepConvertsPanicToError(t *testing.T) {
	err := runGPUProveStep("test step", func() error {
		panic("boom")
	})

	require.ErrorContains(t, err, "test step panic: boom")
}

func TestRunGPUProveStepReturnsErrors(t *testing.T) {
	expected := errors.New("expected error")
	err := runGPUProveStep("test step", func() error {
		return expected
	})

	require.ErrorIs(t, err, expected)
}
