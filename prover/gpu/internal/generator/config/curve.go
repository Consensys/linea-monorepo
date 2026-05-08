package config

// Curve holds all configuration needed to generate a typed per-curve GPU package.
type Curve struct {
	Name       string // "bn254", "bls12377", "bw6761"
	Package    string // Go package name: "bn254", "bls12377", "bw6761"
	FrLimbs    int    // Fr limb count: 4 or 6
	FpLimbs    int    // Fp limb count: 4, 6, or 12
	ScalarBits int    // scalar bit-width: 254, 253, 377

	// gnark-crypto import paths
	GnarkCryptoFr        string // e.g. "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	GnarkCryptoFFT       string // e.g. "github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	GnarkCryptoKZG       string // e.g. "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	GnarkCryptoIOP       string // e.g. "github.com/consensys/gnark-crypto/ecc/bn254/fr/iop"
	GnarkCryptoHTF       string // e.g. "github.com/consensys/gnark-crypto/ecc/bn254/fr/hash_to_field"
	GnarkCurve           string // e.g. "github.com/consensys/gnark-crypto/ecc/bn254"
	GnarkCS              string // e.g. "github.com/consensys/gnark/constraint/bn254"
	GnarkPlonk           string // e.g. "github.com/consensys/gnark/backend/plonk/bn254"

	// CurveIndex is the integer passed to curve-indexed C API calls (curve ID).
	CurveIndex int

	// EccIDStr is the gnark-crypto ecc.ID string (e.g., "BN254", "BLS12_377", "BW6_761").
	EccIDStr string
}
