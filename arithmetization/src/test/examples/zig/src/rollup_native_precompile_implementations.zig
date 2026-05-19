/// Native local-test precompile implementation table.
///
/// The RISC-V guest target uses `rollup_guest_precompile_implementations.zig`,
/// which is the future interpreter-accelerator boundary. Native tests use this
/// module so the same guest execution path can call ZEVM's real C-backed
/// precompile implementations.
const native = @import("zevm_native_precompile_implementations");

pub const ecrecover = native.ecrecover;

pub const bn254_add_byzantium = native.bn254_add_byzantium;
pub const bn254_mul_byzantium = native.bn254_mul_byzantium;
pub const bn254_pairing_byzantium = native.bn254_pairing_byzantium;

pub const bn254_add_istanbul = native.bn254_add_istanbul;
pub const bn254_mul_istanbul = native.bn254_mul_istanbul;
pub const bn254_pairing_istanbul = native.bn254_pairing_istanbul;

pub const kzg_point_evaluation = native.kzg_point_evaluation;

pub const bls12_g1_add = native.bls12_g1_add;
pub const bls12_g1_msm = native.bls12_g1_msm;
pub const bls12_g2_add = native.bls12_g2_add;
pub const bls12_g2_msm = native.bls12_g2_msm;
pub const bls12_pairing = native.bls12_pairing;
pub const bls12_map_fp_to_g1 = native.bls12_map_fp_to_g1;
pub const bls12_map_fp2_to_g2 = native.bls12_map_fp2_to_g2;

pub const p256verify = native.p256verify;
pub const p256verify_osaka = native.p256verify_osaka;
