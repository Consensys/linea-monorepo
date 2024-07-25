package utils

import "io"

var _ io.Writer = NoWriter{}

// NoWriter implements the [io.Write] implementation and does nothing. It can
// be supplied to instantiate a logger which we don't want to use. When calling
// [Write] it will pretend it wrote everything.
//
// It is necessary because the zero value of slog.Logger panics when called.
// Supplying a NoWriter to the slog.New function allows by passing that issue
// without actually instantiating an actual logger.
type NoWriter struct{}

func (w NoWriter) Write(b []byte) (int, error) {
	return len(b), nil
}
