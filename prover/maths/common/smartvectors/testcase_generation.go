package smartvectors

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
)

// Given a smart-vector as input, generate a string containing a
// hardcoded version of an identical vector. This can be used for
// test case generation and debugging.
func GenerateTestVectorFrom(v SmartVector) string {
	switch vec := v.(type) {
	case *Constant:
		return fmt.Sprintf("smartvectors.NewConstant(field.NewFromString(\"%v\"), %v)", vec.val.String(), vec.Len())
	case *PaddedCircularWindow:
		windowDecl := vector.ExplicitDeclarationStr(vec.window)
		paddindDecl := field.ExplicitDeclarationStr(vec.paddingVal)
		return fmt.Sprintf("smartvectors.NewPaddedCircularWindow(%v, %v, %v, %v)",
			windowDecl, paddindDecl, vec.offset, vec.totLen,
		)
	case *Regular:
		return fmt.Sprintf("smartvectors.NewRegular(%v)", vector.ExplicitDeclarationStr(*vec))
	default:
		panic("unknown case")
	}
}
