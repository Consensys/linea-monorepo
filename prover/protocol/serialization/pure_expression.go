package serialization

import (
	"fmt"
	"io"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// UnmarshalExprCBOR reads a [symbolic.Expression] from a CBOR-encoded reader.
// It panics on failure as this function is meant to read test-files for
// the symbolic package.
func UnmarshalExprCBOR(r io.Reader) *symbolic.Expression {

	buf, err := io.ReadAll(r)
	if err != nil {
		utils.Panic("error while [io.ReadAll]: %v", err)
	}

	exprVal, err := DeserializeValue(
		buf,
		pureExprMode,
		reflect.TypeOf(&symbolic.Expression{}),
		nil,
	)

	if err != nil {
		utils.Panic("DeserializeValue returned an error for expression: %v", err)
	}

	return exprVal.Interface().(*symbolic.Expression)
}

// MarshalExprCBOR marshals a [symbolic.Expression] using CBOR encoding and
// writes the encoded result into `w`.
func MarshalExprCBOR(w io.Writer, expr *symbolic.Expression) (int64, error) {

	blob, err := SerializeValue(
		reflect.ValueOf(expr),
		pureExprMode,
	)

	if err != nil {
		return 0, fmt.Errorf("could not serialize the expression with SerializeValue: %w", err)
	}

	n, err := w.Write(blob)
	if err != nil {
		return 0, fmt.Errorf("could not write the blob: %w", err)
	}

	return int64(n), nil
}
