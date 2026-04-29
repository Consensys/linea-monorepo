package plonk2

// Option configures a prepared plonk2 prover.
type Option func(*proverConfig)

type proverConfig struct {
	enabled         bool
	cpuFallback     bool
	strict          bool
	memoryLimit     uint64
	pinnedHostLimit uint64
	tracePath       string
}

func defaultProverConfig() proverConfig {
	return proverConfig{
		cpuFallback: true,
	}
}

// WithEnabled controls whether the plonk2 GPU path is attempted.
func WithEnabled(enabled bool) Option {
	return func(c *proverConfig) {
		c.enabled = enabled
	}
}

// WithCPUFallback controls whether Prove delegates to gnark's CPU prover while
// the curve-generic GPU prover is still being wired.
func WithCPUFallback(enabled bool) Option {
	return func(c *proverConfig) {
		c.cpuFallback = enabled
	}
}

// WithStrictMode disables fallback on GPU-disabled or GPU-error paths.
func WithStrictMode(strict bool) Option {
	return func(c *proverConfig) {
		c.strict = strict
	}
}

// WithMemoryLimit records a future hard VRAM budget in bytes.
func WithMemoryLimit(bytes uint64) Option {
	return func(c *proverConfig) {
		c.memoryLimit = bytes
	}
}

// WithPinnedHostLimit records a future pinned host memory budget in bytes.
func WithPinnedHostLimit(bytes uint64) Option {
	return func(c *proverConfig) {
		c.pinnedHostLimit = bytes
	}
}

// WithTrace records a future phase trace output path.
func WithTrace(path string) Option {
	return func(c *proverConfig) {
		c.tracePath = path
	}
}
