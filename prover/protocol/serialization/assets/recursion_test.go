package assets

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/sirupsen/logrus"
)

var testRec = recurSegComp.Recursion

func TestSerdeRecursion(t *testing.T) {

	if testRec == nil {
		t.Skip("Skipping TestSerdeRecursion due to nil recursion struct")
	}

	// Serialize
	serialized, err := SerializeRecursion(testRec)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	logrus.Println("Serialization recursion successful")

	// Deserialize
	deserialized, err := DeserializeRecursion(serialized)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	logrus.Println("Deserialization recursion successful")

	// Compare
	if !test_utils.CompareExportedFields(testRec, deserialized) {
		t.Errorf("Original and deserialized Recursion structs don't match")
	}

}

var testComp = testRec.InputCompiledIOP

func TestSerdeRecurIOP(t *testing.T) {

	// fmt.Println("********Recur. Comp. IOP********************************************")
	// pcsCtx := testComp.PcsCtxs
	// fmt.Printf("reflec type(name):%s type(string):%s kind:%s of PcsCtx \n", reflect.TypeOf(pcsCtx).Name(), reflect.TypeOf(pcsCtx).String(), reflect.TypeOf(pcsCtx).Kind())
	// fmt.Println("************FIN Recur. Comp. IOP******************************************")

	serComp, err := serialization.SerializeCompiledIOP(testComp)
	if err != nil {
		t.Fatalf("error during ser. recursion input compiled-iop:%s\n", err.Error())
	}

	deSerComp, err := serialization.DeserializeCompiledIOP(serComp)
	if err != nil {
		t.Fatalf("error during deser. recursion input compiled-iop:%s\n", err.Error())
	}

	logrus.Printf("Column exists in original iop:%v\n", testComp.Columns.Exists(ifaces.ColID(CHECK_COLUMN_NAME)))
	logrus.Printf("Column exists in deseriop:%v\n", deSerComp.Columns.Exists(ifaces.ColID(CHECK_COLUMN_NAME)))
	logrus.Printf("Column exists in plonk-iop:%v\n", testRec.PlonkCtx.GetPlonkInternalIOP().Columns.Exists(ifaces.ColID(CHECK_COLUMN_NAME)))

	if !test_utils.CompareExportedFields(testComp, deSerComp) {
		t.Errorf("Mismatch in exported fields after RecursedCompiledIOP serde")
	}
}
