package plonk2

// Option configures a prepared plonk2 prover.
type Option func(*proverConfig)

type proverConfig struct {
	enabled         bool
	cpuFallback     bool
	strict          bool
	legacyBLSGPU    bool
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

// WithCPUFallback controls whether Prove delegates to gnark's CPU prover when
// the GPU path is disabled or returns an error.
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

// WithLegacyBLS12377Backend enables the historical gpu/plonk BLS12-377 bridge.
//
// This is kept only as an explicit comparison path; the default enabled backend
// is the curve-generic GPU prover.
func WithLegacyBLS12377Backend(enabled bool) Option {
	return func(c *proverConfig) {
		c.legacyBLSGPU = enabled
	}
}

// WithMemoryLimit records a hard VRAM budget in bytes.
func WithMemoryLimit(bytes uint64) Option {
	return func(c *proverConfig) {
		c.memoryLimit = bytes
	}
}

// WithPinnedHostLimit records a pinned host memory budget in bytes.
func WithPinnedHostLimit(bytes uint64) Option {
	return func(c *proverConfig) {
		c.pinnedHostLimit = bytes
	}
}

// WithTrace records a phase trace output path.
func WithTrace(path string) Option {
	return func(c *proverConfig) {
		c.tracePath = path
	}
}
