//go:build cuda

package gpu

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeviceFromEnvOrCurrent_CUDA(t *testing.T) {
	t.Setenv(EnvDeviceID, "0")

	dev, id, err := DeviceFromEnvOrCurrent()
	if err != nil {
		t.Skipf("CUDA device 0 unavailable: %v", err)
	}
	require.NotNil(t, dev, "configured CUDA device should be available")
	require.Equal(t, 0, id, "configured device id should be returned")
	require.Equal(t, 0, dev.DeviceID(), "selected device should match the env var")
}

func TestDeviceFromEnvOrCurrent_CUDADevice1(t *testing.T) {
	t.Setenv(EnvDeviceID, "1")

	dev, id, err := DeviceFromEnvOrCurrent()
	if err != nil {
		t.Skipf("CUDA device 1 unavailable: %v", err)
	}
	require.NotNil(t, dev, "configured CUDA device should be available")
	require.Equal(t, 1, id, "configured device id should be returned")
	require.Equal(t, 1, dev.DeviceID(), "selected device should match the env var")
}
