package arithmetization

import (
	_ "embed"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/sirupsen/logrus"
)

const TraceOverflowExitCode = 77

// Embed the whole constraint system at compile time, so no
// more need to keep it in sync
//
//go:embed zkevm.bin
var zkevmStr string

func ReadZkevmBin(optConfig *mir.OptimisationConfig) (schema *air.Schema, metadata map[string]string, err error) {
	var (
		binf binfile.BinaryFile
		buf  []byte = []byte(zkevmStr)
	)

	fmt.Println("Starting ReadZkevmBin...")

	// TODO: why is only this one needed??
	gob.Register(binfile.Attribute(&corset.SourceMap{}))

	// Parse zkbinary file
	err = binf.UnmarshalBinary(buf)
	if err != nil {
		fmt.Printf("Error during UnmarshalBinary: %v\n", err)
		return nil, nil, fmt.Errorf("could not parse the read bytes of the 'zkevm.bin' file into an hir.Schema: %w", err)
	}
	fmt.Println("UnmarshalBinary successful")

	// Extract schema
	if reflect.DeepEqual(binf.Schema, hir.Schema{}) {
		fmt.Println("binf.Schema is empty after unmarshaling zkevm.bin")
		return nil, nil, fmt.Errorf("binf.Schema is empty after unmarshaling zkevm.bin")
	}
	fmt.Println("binf.Schema extraction successful")

	hirSchema := &binf.Schema
	metadata, err = binf.Header.GetMetaData()
	if err != nil {
		fmt.Printf("Error extracting metadata: %v\n", err)
		return nil, nil, fmt.Errorf("failed to extract metadata: %w", err)
	}
	fmt.Println("Metadata extraction successful")

	// Ensure LowerToMir() does not return a zero-value struct
	mirSchema := hirSchema.LowerToMir()
	if reflect.DeepEqual(mirSchema, mir.Schema{}) {
		fmt.Println("LowerToMir() returned an empty struct")
		return nil, nil, fmt.Errorf("LowerToMir() returned an empty struct")
	}
	fmt.Println("LowerToMir() successful")

	// Ensure LowerToAir() does not return a zero-value struct
	airSchema := mirSchema.LowerToAir(*optConfig)
	if reflect.DeepEqual(airSchema, air.Schema{}) {
		fmt.Println("LowerToAir() returned an empty struct")
		return nil, nil, fmt.Errorf("LowerToAir() returned an empty struct")
	}
	fmt.Println("LowerToAir() successful")

	fmt.Println("ReadZkevmBin completed successfully")
	return airSchema, metadata, nil
}

func ReadLtTraces(f io.ReadCloser, sch *air.Schema) (trace.Trace, error) {

	defer f.Close()

	readBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed reading the file: %w", err)
	}

	rawTraces, err := lt.FromBytes(readBytes)
	if err != nil {
		return nil, fmt.Errorf("failed parsing the bytes of the raw trace '.lt' file: %w", err)
	}

	expTraces, errs := schema.NewTraceBuilder(sch).Build(rawTraces)
	if len(errs) > 0 {
		logrus.Warnf("corset expansion gave the following errors: %v", errors.Join(errs...).Error())
	}

	return expTraces, nil
}
