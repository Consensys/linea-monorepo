package config

// BW6761 is the curve configuration for BW6-761.
var BW6761 = Curve{
	Name:           "bw6761",
	Package:        "bw6761",
	FrLimbs:        6,
	FpLimbs:        12,
	ScalarBits:     377,
	GnarkCryptoFr:  "github.com/consensys/gnark-crypto/ecc/bw6-761/fr",
	GnarkCryptoFFT: "github.com/consensys/gnark-crypto/ecc/bw6-761/fr/fft",
	GnarkCryptoKZG: "github.com/consensys/gnark-crypto/ecc/bw6-761/kzg",
	GnarkCryptoIOP: "github.com/consensys/gnark-crypto/ecc/bw6-761/fr/iop",
	GnarkCryptoHTF: "github.com/consensys/gnark-crypto/ecc/bw6-761/fr/hash_to_field",
	GnarkCurve:     "github.com/consensys/gnark-crypto/ecc/bw6-761",
	GnarkCS:        "github.com/consensys/gnark/constraint/bw6-761",
	GnarkPlonk:     "github.com/consensys/gnark/backend/plonk/bw6-761",
	CurveIndex:     3, // GNARK_GPU_PLONK2_CURVE_BW6_761
	EccIDStr:       "BW6_761",
}
