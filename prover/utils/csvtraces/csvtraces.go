// Package csvtraces provides a way to read and write traces in CSV format.
package csvtraces

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/limbs"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type cfg struct {
	// The number of rows in the trace
	nbRows             int
	skipPrePaddingZero bool
	filterOn           ifaces.Column
	inHex              bool
	renameCols         []string
}

type Option func(*cfg) error

// Octuplet represents an octuplet of columns that can be printed in a frienly way.
type Octuplet struct {
	V    [8]ifaces.Column
	Name string
}

// WithNbRows sets the number of rows in the trace
func WithNbRows(nbRows int) Option {
	return func(c *cfg) error {
		c.nbRows = nbRows
		return nil
	}
}

type CsvTrace struct {
	mapped map[string][]*big.Int
	nbRows int
}

func MustOpenCsvFile(fName string) *CsvTrace {

	f, err := os.Open(fName)
	if err != nil {
		utils.Panic("%v", err.Error())
	}
	defer f.Close()

	ct, err := NewCsvTrace(f)
	if err != nil {
		utils.Panic("could not parse CSV: %v", err.Error())
	}

	return ct
}

func NewCsvTrace(r io.Reader, opts ...Option) (*CsvTrace, error) {
	cfg := &cfg{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	rr := csv.NewReader(r)
	rr.FieldsPerRecord = 0

	data := make(map[string][]*big.Int)
	header, err := rr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header row: %w", err)
	}
	for _, h := range header {
		data[h] = make([]*big.Int, 0)
	}
	var nbRows int
	for row, err := rr.Read(); err != io.EOF; row, err = rr.Read() {
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		for i, h := range header {
			d := new(big.Int)
			if _, ok := d.SetString(row[i], 0); !ok {
				return nil, fmt.Errorf("could not decode hex string: row=%v, header=%v string=%v", i, h, row[i])
			}
			data[h] = append(data[h], d)
		}
		nbRows++
	}
	if cfg.nbRows != 0 {
		if cfg.nbRows < nbRows {
			return nil, fmt.Errorf("invalid number of rows: %d", cfg.nbRows)
		}
		nbRows = cfg.nbRows
	}

	return &CsvTrace{mapped: data, nbRows: nbRows}, nil
}

// GetCommit returns a new column mapped to the current csv trace.
func (c *CsvTrace) GetCommit(b *wizard.Builder, name string) ifaces.Column {
	if _, ok := c.mapped[name]; !ok {
		utils.Panic("column not found %s", name)
	}
	length := utils.NextPowerOfTwo(c.nbRows)
	col := b.RegisterCommit(ifaces.ColID(name), length)
	return col
}

// GetLimbLe a little-endian limb object
func (c *CsvTrace) GetLimbsLe(b *wizard.Builder, name string, numLimbs int) limbs.Limbs[limbs.LittleEndian] {
	return getLimbs[limbs.LittleEndian](c, b, name, numLimbs)
}

// GetLimbBe a big-endian limb object
func (c *CsvTrace) GetLimbsBe(b *wizard.Builder, name string, numLimbs int) limbs.Limbs[limbs.BigEndian] {
	return getLimbs[limbs.BigEndian](c, b, name, numLimbs)
}

// getLimbs returns a limbs object mapped to the provided csv trace.
func getLimbs[E limbs.Endianness](c *CsvTrace, b *wizard.Builder, name string, numLimbs int) limbs.Limbs[E] {
	if _, ok := c.mapped[name]; !ok {
		utils.Panic("column not found %s", name)
	}
	length := utils.NextPowerOfTwo(c.nbRows)
	return limbs.NewLimbs[E](b.CompiledIOP, ifaces.ColID(name), numLimbs, length)

}

// AssignCols assigns a vector of columns
func (c *CsvTrace) AssignCols(run *wizard.ProverRuntime, cols ...ifaces.Column) *CsvTrace {
	for _, col := range cols {
		c.Assign(run, col)
	}
	return c
}

// Assign may assign either a column or a limb. It will panic if provided any
// other type. The function also returns a point to itself to make it chainable.
func (c *CsvTrace) Assign(run *wizard.ProverRuntime, toAssign ...any) *CsvTrace {

	for _, obj := range toAssign {

		if obj, ok := obj.(ifaces.Column); ok {

			name := string(obj.GetColID())
			vBi, ok := c.mapped[name]
			if !ok {
				utils.Panic("column not found %s", name)
			}

			vKoa, err := bigIntsToKoalaStrict(vBi)
			if err != nil {
				utils.Panic("could not convert column assignment for %v into koala, %v", name, err)
			}
			run.AssignColumn(
				obj.GetColID(), smartvectors.RightZeroPadded(vKoa, obj.Size()))

			continue
		}

		if objA, ok := obj.(limbs.Limbed); ok {
			obj := objA.ToBigEndianLimbs()
			name := obj.String()
			vBi, ok := c.mapped[name]
			if !ok {
				utils.Panic("limb not found %s", name)
			}
			obj.AssignAndZeroPadsBigInts(run, vBi)
			continue
		}

		utils.Panic("invalid type %T(%++v)", obj, obj)
	}

	return c
}

// AssignLimbsBE assigns a limb object and returns the receiver
func (c *CsvTrace) AssignLimbsBE(run *wizard.ProverRuntime, name ifaces.ColID, column []ifaces.Column) *CsvTrace {
	l := limbs.NewLimbsFromRawUnsafe[limbs.BigEndian](name, column)
	return c.Assign(run, l)
}

func (c *CsvTrace) CheckAssignment(run *wizard.ProverRuntime, objects ...any) *CsvTrace {
	for _, obj := range objects {
		c.checkAssignment(run, obj)
	}
	return c
}

// CheckAssignmentCols is the same as [CheckAssignment] but specifically when
// the input is a list of columns. This allows using slice to variadic implicit
// conversion.
func (c *CsvTrace) CheckAssignmentCols(run *wizard.ProverRuntime, objects ...ifaces.Column) *CsvTrace {
	for _, obj := range objects {
		c.checkAssignment(run, obj)
	}
	return c
}

func (c *CsvTrace) checkAssignment(run *wizard.ProverRuntime, obj any) {

	var (
		wizBi []*big.Int
		csvBi []*big.Int
		name  string
		ok    bool
	)

	switch obj := obj.(type) {
	case ifaces.Column:
		name = string(obj.GetColID())
		vKoala := run.GetColumn(obj.GetColID()).IntoRegVecSaveAlloc()
		csvBi, ok = c.mapped[name]
		if !ok {
			utils.Panic("column not found in csv, %s, %v", name, utils.StringKeysOfMap(c.mapped))
		}
		wizBi = koalaVecToBigInt(vKoala)

	case limbs.Limbed:
		name = obj.String()
		wizBi = obj.ToBigEndianLimbs().GetAssignmentAsBigInt(run)
		csvBi, ok = c.mapped[name]
		if !ok {
			utils.Panic("limb not found in csv, %s", name)
		}

	default:
		utils.Panic("invalid type %T(%++v)", obj, obj)
	}

	if len(wizBi) < c.nbRows {
		utils.Panic("assignment for %s has not been assigned with the expected length, found %v in CSV and %v in wizard", name, c.nbRows, len(wizBi))
	}

	for i := 0; i < c.nbRows; i++ {
		if wizBi[i].Cmp(csvBi[i]) != 0 {
			utils.Panic("assignment for %s has not been assigned correctly: row %d CSV=%s got Wizard=%s", name, i, csvBi[i].String(), wizBi[i].String())
		}
	}

	for i := c.nbRows; i < len(wizBi); i++ {
		if !isZeroBigInt(wizBi[i]) {
			utils.Panic("assignment for %s has not been zero-padded correctly: row %d, got Wizard=%s", name, i, wizBi[i].String())
		}
	}
}

func (c *CsvTrace) LenPadded() int {
	return utils.NextPowerOfTwo(c.nbRows)
}

func WriteExplicitFromKoala(w io.Writer, names []string, cols [][]field.Element, inHex bool) {

	fmt.Fprintf(w, "%v\n", strings.Join(names, ","))
	for i := range cols[0] {
		row := []string{}
		for j := range cols {
			var bi big.Int
			cols[j][i].BigInt(&bi)
			row = append(row, fmtBigInt(inHex, &bi))
		}
		fmt.Fprintf(w, "%v\n", strings.Join(row, ","))
	}
}

func fmtBigInt(inHex bool, x *big.Int) string {
	if !inHex || x.Uint64() < 1<<10 {
		return x.String()
	}
	return "0x" + x.Text(16)
}

func bigIntsToKoalaStrict(vBi []*big.Int) ([]field.Element, error) {
	vKoala := make([]field.Element, len(vBi))
	for i := range vBi {
		// Without this check, the [SetBigInt] method will happily
		// modulo reduce the provided value which would make debugging
		// more difficult.
		if vBi[i].Cmp(field.Modulus()) >= 0 {
			return nil, fmt.Errorf("value #%d, is %v which is greater than koalabear modulus (%v)", i, vBi[i], field.Modulus())
		}
		vKoala[i].SetBigInt(vBi[i])
	}
	return vKoala, nil
}

func koalaVecToBigInt(vKoala []field.Element) []*big.Int {
	vBi := make([]*big.Int, len(vKoala))
	for i := range vKoala {
		vBi[i] = new(big.Int)
		vBi[i] = vKoala[i].BigInt(vBi[i])
	}
	return vBi
}

func isZeroBigInt(bi *big.Int) bool {
	return bi.Cmp(big.NewInt(0)) == 0
}
