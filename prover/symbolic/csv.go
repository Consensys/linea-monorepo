package symbolic

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	isConstantStr = "isConstant"
	isZeroStr     = "isZero"
	isOneStr      = "isOne"
)

func WriteConstantHoodAsCsv(w io.Writer, inputs []smartvectors.SmartVector) {

	fmt.Fprintf(w, "cnt, %v, %v, %v\n", isConstantStr, isZeroStr, isOneStr)

	for i := range inputs {

		var (
			c, isC = inputs[i].(*smartvectors.Constant)
			isZero = isC && c.Val() == field.Zero()
			isOne  = isC && c.Val() == field.One()
		)

		fmt.Fprintf(w, "%v, %v, %v, %v\n", i, isC, isZero, isOne)
	}
}

func ReadConstanthoodFromCsv(r io.ReadCloser) [][3]bool {

	res := make([][3]bool, 0, 1<<15)
	rr := csv.NewReader(r)
	rr.FieldsPerRecord = 0

	header, err := rr.Read()
	if err != nil {
		utils.Panic("read header row: %v", err)
	}

	header[1] = strings.TrimSpace(header[1])
	if header[1] != isConstantStr {
		utils.Panic("unexpected field name: %v", header[1])
	}

	for row, err := rr.Read(); err != io.EOF; row, err = rr.Read() {

		if err != nil {
			utils.Panic("read row: %v", err)
		}

		g := [3]bool{}

		for k := 1; k < 4; k++ {
			var (
				x       = strings.TrimSpace(row[k])
				boolVar bool
			)

			switch {
			case x == "true":
				boolVar = true
			case x == "false":
				boolVar = false
			default:
				utils.Panic("invalid isConstant value = %v", x)
			}

			g[k-1] = boolVar
		}

		res = append(res, g)
	}

	return res
}

func (b *ExpressionBoard) WriteStatsToCSV(w io.Writer) {

	fmt.Fprintf(w, "nodeCount, level, numInLevel, numChildren, numParent, operation, numCoeff1, numCoeff-1, numCoeff0, numCoeff2, numCoeff-2\n")

	for i, node := range b.Nodes {

		var (
			cntCoeff1, cntCoeffMin1 int
			cntCoeff0               int
			cntCoeff2, cntCoeffMin2 int
			operation               string
			coeffs                  = []int{}
		)

		switch op := node.Operator.(type) {
		case LinComb:
			operation = "lin-comb"
			coeffs = op.Coeffs
		case Product:
			operation = "product"
			coeffs = op.Exponents
		case PolyEval:
			operation = "polyeval"
		}

		for _, c := range coeffs {
			switch {
			case c == 0:
				cntCoeff0++
			case c == 1:
				cntCoeff1++
			case c == -1:
				cntCoeffMin1++
			case c == 2:
				cntCoeff2++
			case c == -2:
				cntCoeffMin2++
			}
		}

		fmt.Fprintf(
			w, "%v, %v, %v, %v, %v, %v, %v, %v, %v, %v\n",
			i, 0, i, len(node.Children),
			operation, cntCoeff1, cntCoeffMin1, cntCoeff0, cntCoeff2,
			cntCoeffMin2,
		)
	}

}
