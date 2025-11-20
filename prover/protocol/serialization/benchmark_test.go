package serialization_test

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// runSerdeBenchmark is updated to properly benchmark serialization and deserialization separately.
// For serialization benchmark, it serializes b.N times.
// For deserialization benchmark, it serializes once outside the loop, then deserializes b.N times.
func runSerdeBenchmark(b *testing.B, input any, name string, onlySerialize bool) {
	// In case the test panics, log the error but do not let the panic
	// interrupt the test.
	defer func() {
		if r := recover(); r != nil {
			b.Errorf("Panic during serialization/deserialization of %s: %v", name, r)
			debug.PrintStack()
		}
	}()

	if input == nil {
		b.Fatal("test input is nil")
	}

	var output = reflect.New(reflect.TypeOf(input)).Interface()
	var bBytes []byte
	var err error

	// Serialize once to get the bytes for deserialization benchmark
	if !onlySerialize {
		bBytes, err = serialization.Serialize(input)
		if err != nil {
			b.Fatalf("Error during initial serialization of %s: %v", name, err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if onlySerialize {
			// Benchmark only serialization
			_, err = serialization.Serialize(input)
			if err != nil {
				b.Fatalf("Error during serialization of %s: %v", name, err)
			}
		} else {
			// Benchmark only deserialization (using pre-serialized bytes)
			err = serialization.Deserialize(bBytes, output)
			if err != nil {
				b.Fatalf("Error during deserialization of %s: %v", name, err)
			}
		}
	}
}

// BenchmarkSerZkEVM benchmarks serialization of ZkEVM separately.
func BenchmarkSerZkEVM(b *testing.B) {
	runSerdeBenchmark(b, z, "ZkEVM-Serialize", true)
}

// BenchmarkDeserZkEVM benchmarks deserialization of ZkEVM separately.
func BenchmarkDeserZkEVM(b *testing.B) {
	runSerdeBenchmark(b, z, "ZkEVM-Deserialize", false)
}

func getTestExpressions() []*symbolic.Expression {
	// First set from the struct list
	comp := wizard.NewCompiledIOP()
	a := comp.InsertColumn(0, "a", 16, column.Committed)
	aNext := column.Shift(a, 2)
	b := comp.InsertColumn(0, "b", 16, column.Committed)
	c := comp.InsertColumn(0, "c", 16, column.Committed)
	d := comp.InsertColumn(0, "d", 16, column.Committed)

	exprs := []*symbolic.Expression{
		symbolic.Add(a, aNext),
		symbolic.Add(a, b),
		symbolic.Mul(a, b),
		symbolic.NewVariable(a),
		symbolic.NewConstant(0),
		symbolic.NewConstant(1),
		symbolic.NewConstant(-1),
		symbolic.Add(
			symbolic.Mul(
				symbolic.Add(a, b),
				symbolic.Add(c, d),
			),
			symbolic.NewConstant(1),
		),
	}

	// Second set
	aVar := symbolic.NewDummyVar("a")
	bVar := symbolic.NewDummyVar("b")
	cVar := symbolic.NewDummyVar("c")
	xVar := symbolic.NewDummyVar("x")

	exprs = append(exprs, []*symbolic.Expression{
		aVar.Add(bVar),
		aVar.Add(aVar).Mul(bVar),
		aVar.Neg().Add(bVar).Neg().Mul(cVar).Add(aVar),
		aVar.Sub(bVar).Mul(cVar),
		aVar.Mul(aVar).Mul(bVar).Mul(aVar).Mul(cVar).Mul(cVar),
		aVar.Mul(
			symbolic.NewPolyEval(xVar, []*symbolic.Expression{
				aVar, bVar, aVar, aVar.Add(cVar),
			}),
		),
	}...)

	return exprs
}

// Each expression gets its own benchmark name: expr_0, expr_1, ...
// so you can run subsets via: go test -bench="expr_3" -benchmem

func BenchmarkSerExpr(b *testing.B) {
	exprs := getTestExpressions()

	for i, expr := range exprs {
		b.Run(
			// Label each sub-benchmark clearly
			// (feel free to replace expr_%d with something more descriptive)
			fmt.Sprintf("expr_%02d", i),
			func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := serialization.Serialize(expr)
					if err != nil {
						b.Fatal(err)
					}
				}
			},
		)
	}
}

func BenchmarkDeserExpr(b *testing.B) {
	exprs := getTestExpressions()

	// Pre-serialize all
	serialized := make([][]byte, len(exprs))
	for i, expr := range exprs {
		ser, err := serialization.Serialize(expr)
		if err != nil {
			b.Fatalf("serialize expr_%02d: %v", i, err)
		}
		serialized[i] = ser
	}

	for i, ser := range serialized {
		b.Run(fmt.Sprintf("expr_%02d", i), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var deserialized *symbolic.Expression
				err := serialization.Deserialize(ser, &deserialized)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
