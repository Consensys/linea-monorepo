//go:build cuda && iciclebench

package plonk2

/*
#include <stdint.h>

extern int bn254_get_root_of_unity(uint64_t max_size, void* rou);
extern int bls12_377_get_root_of_unity(uint64_t max_size, void* rou);
extern int bw6_761_get_root_of_unity(uint64_t max_size, void* rou);

extern int bn254_ntt_init_domain(void* primitive_root, void* ctx);
extern int bls12_377_ntt_init_domain(void* primitive_root, void* ctx);
extern int bw6_761_ntt_init_domain(void* primitive_root, void* ctx);
*/
import "C"

import "unsafe"

func setICICLEBN254RootRaw(root unsafe.Pointer, n int) int {
	return int(C.bn254_get_root_of_unity(C.uint64_t(n), root))
}

func setICICLEBLS12377RootRaw(root unsafe.Pointer, n int) int {
	return int(C.bls12_377_get_root_of_unity(C.uint64_t(n), root))
}

func setICICLEBW6761RootRaw(root unsafe.Pointer, n int) int {
	return int(C.bw6_761_get_root_of_unity(C.uint64_t(n), root))
}

func initICICLEBN254DomainRaw(root, cfg unsafe.Pointer) int {
	return int(C.bn254_ntt_init_domain(root, cfg))
}

func initICICLEBLS12377DomainRaw(root, cfg unsafe.Pointer) int {
	return int(C.bls12_377_ntt_init_domain(root, cfg))
}

func initICICLEBW6761DomainRaw(root, cfg unsafe.Pointer) int {
	return int(C.bw6_761_ntt_init_domain(root, cfg))
}
