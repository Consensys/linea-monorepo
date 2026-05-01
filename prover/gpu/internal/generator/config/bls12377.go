package config

// BLS12377 is the curve configuration for BLS12-377.
var BLS12377 = Curve{
	Name:           "bls12377",
	Package:        "bls12377",
	FrLimbs:        4,
	FpLimbs:        6,
	ScalarBits:     253,
	GnarkCryptoFr:  "github.com/consensys/gnark-crypto/ecc/bls12-377/fr",
	GnarkCryptoFFT: "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft",
	GnarkCryptoKZG: "github.com/consensys/gnark-crypto/ecc/bls12-377/kzg",
	GnarkCryptoIOP: "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/iop",
	GnarkCryptoHTF: "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/hash_to_field",
	GnarkCurve:     "github.com/consensys/gnark-crypto/ecc/bls12-377",
	GnarkCS:        "github.com/consensys/gnark/constraint/bls12-377",
	GnarkPlonk:     "github.com/consensys/gnark/backend/plonk/bls12-377",
	CurveIndex:     2, // GNARK_GPU_PLONK2_CURVE_BLS12_377
	EccIDStr:       "BLS12_377",
}
