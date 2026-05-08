package gpu

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfiguredDeviceID(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		t.Setenv(EnvDeviceID, "")

		id, configured, err := ConfiguredDeviceID()
		require.NoError(t, err)
		require.False(t, configured, "unset device env should preserve default routing")
		require.Zero(t, id, "unset device env should report the default id")
	})

	t.Run("valid", func(t *testing.T) {
		t.Setenv(EnvDeviceID, "1")

		id, configured, err := ConfiguredDeviceID()
		require.NoError(t, err)
		require.True(t, configured, "set device env should enable explicit routing")
		require.Equal(t, 1, id, "configured device id should match the env var")
	})

	t.Run("invalid", func(t *testing.T) {
		t.Setenv(EnvDeviceID, "gpu1")

		_, configured, err := ConfiguredDeviceID()
		require.Error(t, err, "non-integer device id should be rejected")
		require.True(t, configured, "invalid env still counts as explicit configuration")
	})

	t.Run("negative", func(t *testing.T) {
		t.Setenv(EnvDeviceID, "-1")

		_, configured, err := ConfiguredDeviceID()
		require.Error(t, err, "negative device id should be rejected")
		require.True(t, configured, "negative env still counts as explicit configuration")
	})
}
