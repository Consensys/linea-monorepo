package main

// #include <stdlib.h>
// #include "shnarf_calculator.h"
import "C"

import (
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/backend/blobsubmission"
)

// Required to make CGO work.
func main() {}

// CalculateShnarf is the bridge between C and Go wrapping the [CraftResponse]
// function.
//
//export CalculateShnarf
func CalculateShnarf(
	eip4844_enabled C.bool,
	compressed_data *C.char,
	parent_state_root_hash *C.char,
	final_state_root_hash *C.char,
	prev_shnarf *C.char,
	conflation_order_starting_block_number C.longlong,
	conflation_order_upper_boundaries_len C.int,
	conflation_order_upper_boundaries *C.longlong,
) *C.response {
	goReq := convertCtoGoRequest(
		eip4844_enabled,
		compressed_data,
		parent_state_root_hash,
		final_state_root_hash,
		prev_shnarf,
		conflation_order_starting_block_number,
		conflation_order_upper_boundaries_len,
		conflation_order_upper_boundaries,
	)
	// fmt.Printf("the request = %++v\n", goReq)

	resp, err := blobsubmission.CraftResponse(goReq)

	if err != nil {
		errMsg := C.CString(err.Error())
		cResp := (*C.response)(C.malloc(C.sizeof_response))
		cResp.commitment = C.CString("")
		cResp.kzg_proof_contract = C.CString("")
		cResp.kzg_proof_sidecar = C.CString("")
		cResp.data_hash = C.CString("")
		cResp.snark_hash = C.CString("")
		cResp.expected_x = C.CString("")
		cResp.expected_y = C.CString("")
		cResp.expected_shnarf = C.CString("")
		cResp.error_message = errMsg
		return cResp
	}

	return convertGoToCResponse(resp)

}

// convertCToGoRequest converts a request struct type from C to its corresponding
// Go Response struct type.
// See prover/backend/blobsubmission/compression.h for the C type declaration.
// See prover/backend/blobsubmission/compression.go for the Go type declaration.
func convertCtoGoRequest(
	eip4844_enabled C.bool,
	compressed_data *C.char,
	parent_state_root_hash *C.char,
	final_state_root_hash *C.char,
	prev_shnarf *C.char,
	conflation_order_starting_block_number C.longlong,
	conflation_order_upper_boundaries_len C.int,
	conflation_order_upper_boundaries *C.longlong,
) *blobsubmission.Request {
	return &blobsubmission.Request{
		Eip4844Enabled:      bool(eip4844_enabled),
		CompressedData:      C.GoString(compressed_data),
		ParentStateRootHash: C.GoString(parent_state_root_hash),
		FinalStateRootHash:  C.GoString(final_state_root_hash),
		PrevShnarf:          C.GoString(prev_shnarf),
		ConflationOrder: blobsubmission.ConflationOrder{
			StartingBlockNumber: int(conflation_order_starting_block_number),
			UpperBoundaries: convertCIntArrayToGo(
				conflation_order_upper_boundaries,
				conflation_order_upper_boundaries_len,
			),
		},
	}
}

// Convert a C array (with an additional length parameter) into a proper go
// array. cPtr is a pointer toward the first element of the list. If the given
// length is smaller than the actual length. All elements coming after are
// ignored and will not be included in the final result. However, if the length
// is larger than the actual size, then this can result in undefined behavior
// (most likely SEGFAULT).
func convertCIntArrayToGo(cPtr *C.longlong, cLength C.int) []int {
	var (
		goLength   = int(cLength)
		goIntArr   = make([]int, cLength)
		goPtr      = unsafe.Pointer(cPtr)
		sizeOfCInt = unsafe.Sizeof(C.longlong(0))
	)

	// Nicer getting this than getting a SIGSEGV
	if goPtr == unsafe.Pointer(nil) {
		panic("was given a null pointer")
	}

	for i := 0; i < goLength; i++ {
		offset := uintptr(i) * sizeOfCInt
		//lint:ignore -- we checked that this was right
		elemPtr := unsafe.Pointer(uintptr(goPtr) + offset)
		elem := *(*C.longlong)(elemPtr)
		goIntArr[i] = int(elem)
	}

	return goIntArr
}

// convertGoToCResponse converts a Response struct type from Go to its
// corresponding C response struct type.
// See prover/backend/blobsubmission/compression.h for the C type declaration.
// See prover/backend/blobsubmission/compression.go for the Go type declaration.
func convertGoToCResponse(resp *blobsubmission.Response) *C.response {
	// fmt.Printf("GoResponse = %++v\n", resp)
	cResp := (*C.response)(C.malloc(C.sizeof_response))
	cResp.commitment = C.CString(resp.Commitment)
	cResp.kzg_proof_contract = C.CString(resp.KzgProofContract)
	cResp.kzg_proof_sidecar = C.CString(resp.KzgProofSidecar)
	cResp.data_hash = C.CString(resp.DataHash)
	cResp.snark_hash = C.CString(resp.SnarkHash)
	cResp.expected_x = C.CString(resp.ExpectedX)
	cResp.expected_y = C.CString(resp.ExpectedY)
	cResp.expected_shnarf = C.CString(resp.ExpectedShnarf)
	cResp.error_message = C.CString("")
	// fmt.Printf("CResponse = %++v\n", cResp)
	return cResp
}
