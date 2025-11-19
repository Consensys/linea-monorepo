package p256verify

// Limits defines limits for P256Verify module.
type Limits struct {
	// NbInputInstances is the number of P256 input instances per a single
	// verification circuit. gnark circuit size approximately 709k constraints
	// and in Plonk-in-Wizard (with externalized range checks) approximately
	// 183k rows per instance.
	NbInputInstances int
}
