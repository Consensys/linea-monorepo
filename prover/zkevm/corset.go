//go:build !nocorset

package zkevm

import (
	_ "embed"
	"io"
	"math/big"
	"os"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/sirupsen/logrus"
)

//#include <corset.h>
import "C"

// Embed the whole constraint system at compile time, so no
// more need to keep it in sync
//
//go:embed zkevm.bin
var zkevmStr string

// Error code that we return when encountering a trace that do not
// fit in the prover's expected traces.
const tracesTooLargeCode = 333

func corsetErrToString(err error) string {
	if errNo, ok := err.(syscall.Errno); ok {
		return C.GoString(C.corset_err_to_string(C.int(int(errNo))))
	} else {
		return "not an Errno error"
	}
}

func getColumnNames(trace *C.Trace) (r []string) {
	ncols, err := C.trace_column_count(trace)
	if err != nil {
		logrus.Panicf("In getColumnNames/trace_column_count: %v", err)
	}

	names, err := C.trace_column_names(trace)
	if err != nil {
		logrus.Panicf("In getColumnNames/trace_column_names: %v", err)
	}

	for _, name_c := range unsafe.Slice(names, ncols) {
		name := C.GoString(name_c)
		r = append(r, strings.Clone(name))
	}

	return
}

func corsetUint256ToFr(inBytes [32]byte, to *fr.Element, xi *big.Int) {
	// Putting endianness-conversion in Corset spares 20% time
	xi.SetBytes(inBytes[:])
	to.SetInterface(xi)
	return
}

type workElement struct {
	filled bool
	name   ifaces.ColID
	values smartvectors.SmartVector
}

func columnToAssignment(r chan workElement,
	run *wizard.ProverRuntime,
	nameStr string,
	trace *C.Trace,
) {
	logrus.Tracef("Importing column %v", nameStr)
	col, err := C.trace_column_by_name(trace, C.CString(nameStr))
	if err != nil {
		logrus.Panicf("Failed to retrieve column %v", nameStr)
	}

	// First, fetch the raw memory, of type [][4]_ctype_ulong
	values_raw := [][4]C.ulong{}
	if col.values != nil {
		values_raw = unsafe.Slice(col.values, col.values_len)
	}
	witness_ := make([]fr.Element, len(values_raw))

	// Check that _ctype_ulong are indeed 64bit long.
	// libcorset defines them as uint64_t from stdint.h, but Go is
	// too precious to adhere to such lowly standards; guaranteed-length
	// integers are for people dumber than Rob Pike.
	// XXX now disabled for performances reason
	// XXX re-enable on new architectures
	// if col.values_len > 0 {
	//	int_len := uint64(reflect.TypeOf(values_raw[0][0]).Bits())
	//	if int_len != 64 {
	//		logrus.Panicf("Cgo _ctype_ulong are not 64b long: %v", int_len)
	//	}
	// }

	var witness smartvectors.SmartVector
	name := ifaces.ColID(nameStr)
	if !run.Spec.Columns.Exists(name) {
		logrus.Debugf("Got an undeclared column %v - skipping\n", name)
		r <- workElement{false, name, witness}
		return
	}

	var xi big.Int
	for i, raw_i := range values_raw {
		b := *(*[32]byte)(unsafe.Pointer(&raw_i))
		corsetUint256ToFr(b, &witness_[i], &xi)
	}
	/*
		We pad the inputs according the the padding strategy
	*/
	h := run.Spec.Columns.GetHandle(name)

	/*
		Resize the witness if necessary
	*/
	switch {
	case len(witness_) == h.Size():
		witness = smartvectors.NewRegular(witness_)
	case len(witness_) == 0:
		logrus.Debugf("Empty assignment found for %v", name)
		var padding fr.Element
		corsetUint256ToFr(*(*[32]byte)(unsafe.Pointer(&col.padding_value)), &padding, &xi)
		witness = smartvectors.NewConstant(padding, h.Size())
	case len(witness_) < h.Size():
		var padding fr.Element
		corsetUint256ToFr(*(*[32]byte)(unsafe.Pointer(&col.padding_value)), &padding, &xi)
		witness = smartvectors.LeftPadded(witness_, padding, h.Size())
	case len(witness_) > h.Size():
		logrus.Errorf(
			"ERROR : assignment %v has size %v but expected %v",
			h.GetColID(), len(witness_), h.Size(),
		)
		os.Exit(tracesTooLargeCode)
	default:
		logrus.Panicf("Unreachable - length of the assignment %v - length of the commitment %v", len(witness_), h.Size())
	}

	r <- workElement{true, name, witness}
}

func AssignFromCorset(reader io.Reader, run *wizard.ProverRuntime) {
	if len(zkevmStr) < 2 {
		utils.Panic("Prover container was not properly built: zkevm.bin is empty!")
	}

	logrus.Info("Loading zkEVM...")
	corset, err := C.corset_from_string(C.CString(zkevmStr))
	if err != nil {
		utils.Panic("Error while reading constraints, Corset says `%v`", corsetErrToString(err))
	}
	logrus.Info("Done.")

	logrus.Info("Parsing JSON...")
	traceBytes, err := io.ReadAll(reader)
	if err != nil {
		utils.Panic("Could not read trace file, Corset says: `%v`", corsetErrToString(err))
	}
	logrus.Info("Done.")

	numberOfThreads := runtime.NumCPU() / 2

	logrus.Infof("Expanding trace... using %v threads", numberOfThreads)
	trace, err := C.trace_compute_from_string(
		corset,
		C.CString(string(traceBytes)), // trace file content
		C.uint(numberOfThreads),       // # threads
		true,                          // convert Fr to big endian
		false,                         // crash on missing columns in the trace?
	)
	if trace == nil {
		utils.Panic("Error while computing trace, Corset says: `%v`", corsetErrToString(err))
	}
	logrus.Info("Done.")

	logrus.Info("Converting columns from uint64_t to fr.Element...")
	columnNames := getColumnNames(trace)
	queue := make(chan workElement)
	for i := 0; i < len(columnNames); i++ {
		go columnToAssignment(queue, run, columnNames[i], trace)
	}
	for i := 0; i < len(columnNames); i++ {
		r := <-queue
		if r.filled {
			run.AssignColumn(r.name, ifaces.ColAssignment(r.values))
		}
	}
	logrus.Info("Done.")

	// SmartVectors constructors copy their arguments, so no need
	// to keep this around.
	C.trace_free(trace)
}
