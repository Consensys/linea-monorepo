package plonk2

// Option configures a plonk2 Prover.
type Option func(*proverConfig)

type proverConfig struct {
	enabled     bool
	cpuFallback bool
	strict      bool
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
func WithStrictMode(strict bool) Option {
	return func(c *proverConfig) {
		c.strict = strict
		if strict {
			c.cpuFallback = false
		}
	}
}
