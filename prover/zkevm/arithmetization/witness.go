package arithmetization

import "io"

// TraceGetter specifies the content of the zkEVM columns and is used as the input
// of the prover for all that relates to the zkEVM arithmetization. It consists
// of a Reader to be passed to corset to derive the assignment. In practice, it
// will be a function opening a file and returning a ReadCloser from the file.
type TraceGetter func() io.ReadCloser
