#pragma once

// ─────────────────────────────────────────────────────────────────────────────
// Field type definitions and SoA storage for BLS12-377 Fp and Fr
//
// Two field types are used throughout the GPU library:
//
//   Fp (base field, 377 bits, 6 × 64-bit limbs):
//     Used for elliptic curve point coordinates (x, y, t, z).
//     Arithmetic in fp.cuh.
//
//   Fr (scalar field, 253 bits, 4 × 64-bit limbs):
//     Used for polynomial coefficients, NTT elements, and MSM scalars.
//     Arithmetic in fr_arith.cuh.
//
// GPU memory layout — Structure of Arrays (SoA):
//
//   AoS (Array of Structs) — how CPU stores field elements:
//     [a₀[0] a₀[1] a₀[2] a₀[3]] [a₁[0] a₁[1] a₁[2] a₁[3]] ...
//     └──── element 0 ────────┘ └──── element 1 ────────┘
//
//   SoA (Structure of Arrays) — how GPU stores for coalesced access:
//     limb0: [a₀[0], a₁[0], a₂[0], ..., aₙ₋₁[0]]  ← one 256-bit load
//     limb1: [a₀[1], a₁[1], a₂[1], ..., aₙ₋₁[1]]     per warp covers
//     limb2: [a₀[2], a₁[2], a₂[2], ..., aₙ₋₁[2]]     32 consecutive
//     limb3: [a₀[3], a₁[3], a₂[3], ..., aₙ₋₁[3]]     elements
//
//   When a warp of 32 threads accesses consecutive elements, SoA ensures
//   each limb array is accessed contiguously → coalesced 256-byte transactions
//   instead of strided access with 4× the memory transactions.
//
// Exception: MSM points use AoS (G1EdXY, 96 bytes per point) because the
// accumulate kernel accesses points by random index from radix sort output.
// SoA would require 2 separate random accesses and double TLB misses.
// ─────────────────────────────────────────────────────────────────────────────

#include <array>
#include <cstdint>
#include <cstdio>
#include <memory>

#ifdef __CUDACC__
#include <cuda_runtime.h>
#else
#define __host__
#define __device__
#define __forceinline__
#endif

namespace gnark_gpu {

// =============================================================================
// Field parameters for BLS12-377
// =============================================================================

// BLS12-377 base field Fp (377 bits, 6 limbs)
struct Fp_params {
	static constexpr size_t LIMBS = 6;
	static constexpr uint64_t MODULUS[6] = {
		0x8508c00000000001ULL, 0x170b5d4430000000ULL, 0x1ef3622fba094800ULL,
		0x1a22d9f300f5138fULL, 0xc63b05c06ca1493bULL, 0x01ae3a4617c510eaULL,
	};
	static constexpr uint64_t INV = 0x8508bfffffffffffULL; // -p^{-1} mod 2^64
};

// BLS12-377 scalar field Fr (253 bits, 4 limbs)
struct Fr_params {
	static constexpr size_t LIMBS = 4;
	static constexpr uint64_t MODULUS[4] = {
		0x0a11800000000001ULL,
		0x59aa76fed0000001ULL,
		0x60b44d1e5c37b001ULL,
		0x12ab655e9a2ca556ULL,
	};
	static constexpr uint64_t INV = 0x0a117fffffffffffULL;
	// Montgomery R = 2^256 mod q (i.e., "one" in Montgomery form)
	static constexpr uint64_t ONE[4] = {
		0x7d1c7ffffffffff3ULL,
		0x7257f50f6ffffff2ULL,
		0x16d81575512c0feeULL,
		0x0d4bda322bbb9a9dULL,
	};
};

// =============================================================================
// Field element (single element, for host-side or AoS usage)
// =============================================================================

template <typename Params> struct Field {
	uint64_t limbs[Params::LIMBS];

	__host__ __device__ constexpr Field() : limbs{} {}

	__host__ __device__ constexpr Field(uint64_t v) : limbs{} { limbs[0] = v; }

	__host__ __device__ bool operator==(const Field &other) const {
		for(size_t i = 0; i < Params::LIMBS; ++i) {
			if(limbs[i] != other.limbs[i]) return false;
		}
		return true;
	}

	__host__ __device__ bool operator!=(const Field &other) const { return !(*this == other); }
};

using Fr = Field<Fr_params>;
using Fp = Field<Fp_params>;

// =============================================================================
// HostFieldVector: Host-side SoA storage (mirrors FieldVector layout)
// =============================================================================

template <typename Params> class HostFieldVector {
	size_t count_ = 0;
	std::array<std::unique_ptr<uint64_t[]>, Params::LIMBS> data_ = {};

  public:
	HostFieldVector() = default;

	explicit HostFieldVector(size_t n) : count_(n) {
		for(size_t i = 0; i < Params::LIMBS; ++i) {
			data_[i] = std::make_unique<uint64_t[]>(n);
		}
	}

	size_t size() const { return count_; }
	static constexpr size_t limbs() { return Params::LIMBS; }

	// Access limb array
	uint64_t *limb(size_t i) { return data_[i].get(); }
	const uint64_t *limb(size_t i) const { return data_[i].get(); }

	// Get raw pointers for copy operations
	auto raw_ptrs() {
		std::array<uint64_t *, Params::LIMBS> ptrs;
		for(size_t i = 0; i < Params::LIMBS; ++i) {
			ptrs[i] = data_[i].get();
		}
		return ptrs;
	}

	auto raw_ptrs() const {
		std::array<uint64_t *, Params::LIMBS> ptrs;
		for(size_t i = 0; i < Params::LIMBS; ++i) {
			ptrs[i] = const_cast<uint64_t *>(data_[i].get());
		}
		return ptrs;
	}

	// Set element at index (from a Field)
	void set(size_t idx, const Field<Params> &f) {
		for(size_t i = 0; i < Params::LIMBS; ++i) {
			data_[i][idx] = f.limbs[i];
		}
	}

	// Get element at index (as a Field)
	Field<Params> get(size_t idx) const {
		Field<Params> f;
		for(size_t i = 0; i < Params::LIMBS; ++i) {
			f.limbs[i] = data_[i][idx];
		}
		return f;
	}
};

using HostFrVector = HostFieldVector<Fr_params>;
using HostFpVector = HostFieldVector<Fp_params>;

// =============================================================================
// CUDA-only: FieldVector and PTX intrinsics
// =============================================================================

#ifdef __CUDACC__

// =============================================================================
// FieldVector: GPU-friendly SoA (Structure of Arrays) for field elements
// Stores N field elements as LIMBS separate arrays for coalesced memory access
// =============================================================================

template <typename Params> class FieldVector {
	size_t count_ = 0;
	std::array<uint64_t *, Params::LIMBS> device_ = {};

  public:
	FieldVector() = default;

	explicit FieldVector(size_t n) : count_(n) { allocate(); }

	~FieldVector() { free(); }

	// Non-copyable
	FieldVector(const FieldVector &) = delete;
	FieldVector &operator=(const FieldVector &) = delete;

	// Movable
	FieldVector(FieldVector &&other) noexcept : count_(other.count_), device_(other.device_) {
		other.count_ = 0;
		other.device_ = {};
	}

	FieldVector &operator=(FieldVector &&other) noexcept {
		if(this != &other) {
			free();
			count_ = other.count_;
			device_ = other.device_;
			other.count_ = 0;
			other.device_ = {};
		}
		return *this;
	}

	void resize(size_t n) {
		if(n == count_) return;
		free();
		count_ = n;
		allocate();
	}

	// Copy from host arrays (one per limb)
	void copy_host_to_device(const std::array<uint64_t *, Params::LIMBS> &host) {
		for(size_t i = 0; i < Params::LIMBS; ++i) {
			cudaMemcpy(device_[i], host[i], count_ * sizeof(uint64_t), cudaMemcpyHostToDevice);
		}
	}

	// Copy to host arrays (one per limb)
	void copy_device_to_host(const std::array<uint64_t *, Params::LIMBS> &host) const {
		for(size_t i = 0; i < Params::LIMBS; ++i) {
			cudaMemcpy(host[i], device_[i], count_ * sizeof(uint64_t), cudaMemcpyDeviceToHost);
		}
	}

	// Accessors
	size_t size() const { return count_; }
	size_t bytes() const { return count_ * Params::LIMBS * sizeof(uint64_t); }
	static constexpr size_t limbs() { return Params::LIMBS; }

	// Get device pointer for limb i
	uint64_t *limb(size_t i) { return device_[i]; }
	const uint64_t *limb(size_t i) const { return device_[i]; }

	// Get all device pointers (for kernel launches)
	auto &device_ptrs() { return device_; }
	const auto &device_ptrs() const { return device_; }

  private:
	void allocate() {
		for(size_t i = 0; i < Params::LIMBS; ++i) {
			cudaMalloc(&device_[i], count_ * sizeof(uint64_t));
		}
	}

	void free() {
		for(size_t i = 0; i < Params::LIMBS; ++i) {
			if(device_[i]) {
				cudaFree(device_[i]);
				device_[i] = nullptr;
			}
		}
	}
};

using FrVector = FieldVector<Fr_params>;
using FpVector = FieldVector<Fp_params>;

// =============================================================================
// PTX intrinsics for Montgomery arithmetic
// =============================================================================

// PTX multiply-add with carry: lo = (a*b + c + carry_in) mod 2^64
__device__ __forceinline__ void madc_u64(uint64_t &lo, uint64_t &carry, uint64_t a, uint64_t b, uint64_t c,
										 uint64_t carry_in) {
	asm volatile("add.cc.u64 %0, %4, %5;\n\t"
				 "madc.lo.cc.u64 %0, %2, %3, %0;\n\t"
				 "madc.hi.u64 %1, %2, %3, 0;"
				 : "=&l"(lo), "=l"(carry)
				 : "l"(a), "l"(b), "l"(c), "l"(carry_in));
}

#endif // __CUDACC__

} // namespace gnark_gpu
