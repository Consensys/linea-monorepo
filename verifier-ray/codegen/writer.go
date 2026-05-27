package codegen

import (
	"fmt"
	"io"
	"strings"
)

// writer is a small helper for accumulating indented Zig source lines.
type writer struct {
	buf    strings.Builder
	indent int
}

func (w *writer) line(format string, args ...any) {
	fmt.Fprintf(&w.buf, "%s%s\n", strings.Repeat("    ", w.indent), fmt.Sprintf(format, args...))
}

func (w *writer) blank() { w.buf.WriteByte('\n') }
func (w *writer) in()    { w.indent++ }
func (w *writer) out() {
	if w.indent == 0 {
		panic("writer: unbalanced out() call")
	}
	w.indent--
}

func (w *writer) writeTo(dst io.Writer) error {
	_, err := io.WriteString(dst, w.buf.String())
	return err
}
