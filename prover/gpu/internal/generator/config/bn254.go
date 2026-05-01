package config

// BN254 is the curve configuration for BN254.
var BN254 = Curve{
	Name:           "bn254",
	Package:        "bn254",
	FrLimbs:        4,
	FpLimbs:        4,
	ScalarBits:     254,
	GnarkCryptoFr:  "github.com/consensys/gnark-crypto/ecc/bn254/fr",
	GnarkCryptoFFT: "github.com/consensys/gnark-crypto/ecc/bn254/fr/fft",
	GnarkCryptoKZG: "github.com/consensys/gnark-crypto/ecc/bn254/kzg",
	GnarkCryptoIOP: "github.com/consensys/gnark-crypto/ecc/bn254/fr/iop",
	GnarkCryptoHTF: "github.com/consensys/gnark-crypto/ecc/bn254/fr/hash_to_field",
	GnarkCurve:     "github.com/consensys/gnark-crypto/ecc/bn254",
	GnarkCS:        "github.com/consensys/gnark/constraint/bn254",
	GnarkPlonk:     "github.com/consensys/gnark/backend/plonk/bn254",
	CurveIndex:     1, // GNARK_GPU_PLONK2_CURVE_BN254
	EccIDStr:       "BN254",
}
