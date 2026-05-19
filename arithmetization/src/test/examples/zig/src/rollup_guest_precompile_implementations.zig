/// RISC-V guest precompile boundary.
///
/// Native tests intentionally do not import this module; they keep ZEVM's native
/// C-backed precompile implementation table. The freestanding guest imports this
/// complete table so the heavy precompiles remain a single injectable boundary
/// for the future Linea interpreter accelerator.
const T = @import("precompile_types");

fn acceleratorNotImplemented(_: []const u8, _: u64) T.PrecompileResult {
    @panic("Linea RISC-V precompile accelerator is not implemented yet");
}

pub const ecrecover: T.PrecompileFn = acceleratorNotImplemented;

pub const bn254_add_byzantium: T.PrecompileFn = acceleratorNotImplemented;
pub const bn254_mul_byzantium: T.PrecompileFn = acceleratorNotImplemented;
pub const bn254_pairing_byzantium: T.PrecompileFn = acceleratorNotImplemented;

pub const bn254_add_istanbul: T.PrecompileFn = acceleratorNotImplemented;
pub const bn254_mul_istanbul: T.PrecompileFn = acceleratorNotImplemented;
pub const bn254_pairing_istanbul: T.PrecompileFn = acceleratorNotImplemented;

pub const kzg_point_evaluation: T.PrecompileFn = acceleratorNotImplemented;

pub const bls12_g1_add: T.PrecompileFn = acceleratorNotImplemented;
pub const bls12_g1_msm: T.PrecompileFn = acceleratorNotImplemented;
pub const bls12_g2_add: T.PrecompileFn = acceleratorNotImplemented;
pub const bls12_g2_msm: T.PrecompileFn = acceleratorNotImplemented;
pub const bls12_pairing: T.PrecompileFn = acceleratorNotImplemented;
pub const bls12_map_fp_to_g1: T.PrecompileFn = acceleratorNotImplemented;
pub const bls12_map_fp2_to_g2: T.PrecompileFn = acceleratorNotImplemented;

pub const p256verify: T.PrecompileFn = acceleratorNotImplemented;
pub const p256verify_osaka: T.PrecompileFn = acceleratorNotImplemented;
