package define

// Registers the columns for the phoney RLP. The phoney RLP is not included in
// the constraints, add it. This is a fix since theses columns are in the traces
// but are not in the define of the arithmetization
func definePhoneyRlp(builder *Builder) {
	builder.RegisterCommit("phoneyRLP.TX_NUM", builder.Traces.PhoneyRlp)
	builder.RegisterCommit("phoneyRLP.INDEX", builder.Traces.PhoneyRlp)
	builder.RegisterCommit("phoneyRLP.LIMB", builder.Traces.PhoneyRlp)
	builder.RegisterCommit("phoneyRLP.nBYTES", builder.Traces.PhoneyRlp)
}
