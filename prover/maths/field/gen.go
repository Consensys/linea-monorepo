package field

import (
	"fmt"
	"runtime"

	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
)

func ParBatchInvert(a []Element, numcpu int) []Element {

	if numcpu == 0 {
		numcpu = runtime.GOMAXPROCS(0)
	}

	res := make([]Element, len(a))

	parallel.Execute(len(a), func(start, stop int) {
		subRes := BatchInvert(a[start:stop])
		copy(res[start:stop], subRes)
	}, numcpu)

	return res
}

// Returns a string declaring in hard-code the field element value
func ExplicitDeclarationStr(x Element) string {
	return fmt.Sprintf("field.NewFromString(\"%v\")", x.String())
}
