package ringsis

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStdParams(t *testing.T) {

	testParams := []struct {
		Params        Params
		ModulusDegree int
		NumLimbs      int
	}{
		{
			Params:        StdParams,
			ModulusDegree: 32,
			NumLimbs:      2,
		},
		{
			Params:        Params{LogTwoBound: 2, LogTwoDegree: 3},
			ModulusDegree: 8,
			NumLimbs:      16,
		},
	}

	for i := range testParams {
		t.Run(fmt.Sprintf("case-%v", i), func(t *testing.T) {
			testCase := testParams[i]
			require.Equal(t, testCase.ModulusDegree, testCase.Params.modulusDegree())
			require.Equal(t, testCase.ModulusDegree, testCase.Params.OutputSize())
			require.Equal(t, testCase.NumLimbs, testCase.Params.NumLimbs())
		})
	}
}
