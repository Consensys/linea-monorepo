package plonk2

// Option configures a plonk2 Prover.
type Option func(*proverConfig)

type proverConfig struct {
	// enabled is the master kill-switch for the GPU path. When false, Prove
	// returns gnark's CPU prover output without ever touching the device.
	enabled bool
	// cpuFallback controls whether a GPU-side error falls back to the CPU
	// prover. Default true. Disabled by WithStrictMode for tests that must
	// fail loudly when the GPU disagrees with the reference.
	cpuFallback bool
}

func defaultProverConfig() proverConfig {
	return proverConfig{
		cpuFallback: true,
	}
}

// WithEnabled controls whether the GPU path is attempted.
func WithEnabled(enabled bool) Option {
	return func(c *proverConfig) { c.enabled = enabled }
}

// WithCPUFallback controls whether Prove falls back to gnark's CPU prover
// when the GPU path is disabled or returns an error. Default: true.
func WithCPUFallback(enabled bool) Option {
	return func(c *proverConfig) { c.cpuFallback = enabled }
}

// WithStrictMode disables the CPU fallback and returns errors instead.
// Used by tests that must fail when the GPU path errors, rather than
// silently falling through to the CPU and masking the bug.
func WithStrictMode(strict bool) Option {
	return func(c *proverConfig) {
		if strict {
			c.cpuFallback = false
		}
	}
}
