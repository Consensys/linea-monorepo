//! Per-precompile accelerator toggle for the ZkC guest.
//!
//! Zesu's freestanding build routes every precompile through `extern fn zkvm_*` symbols
//! (src/crypto/extern_bridge.zig). This module decides, per symbol, who provides it in the final
//! guest — and the in-guest implementations are **delegated to zesu-zkvm's `stdlibs_accel`**
//! (imported as `zesu_zkvm_accel`) so we don't maintain our own crypto:
//!
//!   .intercept — leave `zkvm_X` UNDEFINED in the guest object; the prover/ZkC resolves it at link
//!                time (e.g. arithmetized keccak). Use once ZkC implements that precompile.
//!   .native    — DEFINE `zkvm_X` in-guest as a thin C-ABI wrapper over zesu-zkvm `stdlibs_accel`
//!                (real for keccak/sha256/ecrecover/secp256k1_verify; a failing stub for the rest
//!                until upstream implements them). ZkC interprets it like any other rv64im code.
//!
//! INVARIANT: every symbol is `.intercept` XOR defined here (`.native`); a symbol that is neither
//! is an unresolved link error. Keep this table in lockstep with what the prover intercepts.
//!
//! Only the freestanding RISC-V guest references these (see evm_execution_guest.zig, pulled in for
//! `builtin.cpu.arch == .riscv64`); the native host build uses Zesu's `default.zig` (C libs).

const zesu_accel = @import("zesu_zkvm_accel"); // zesu-zkvm linea/src/runtime/stdlibs_accel.zig
const linea_accel = @import("linea_zkvm_accel"); // TODO: comment

pub const Mode = enum { intercept, native };

/// THE TOGGLE TABLE — flip `.native` → `.intercept` once ZkC arithmetizes that precompile.
/// keccak is intercepted by ZkC; everything else runs in-guest via zesu-zkvm's stdlibs_accel.
pub const policy = struct {
    pub const keccak256: Mode = .native;
    pub const sha256: Mode = .native;
    pub const secp256k1_verify: Mode = .native;
    pub const secp256k1_ecrecover: Mode = .native;
    pub const ripemd160: Mode = .native;
    pub const modexp: Mode = .native;
    pub const bn254_g1_add: Mode = .native;
    pub const bn254_g1_mul: Mode = .native;
    pub const bn254_pairing: Mode = .native;
    pub const blake2f: Mode = .native;
    pub const kzg_point_eval: Mode = .native;
    pub const bls12_g1_add: Mode = .native;
    pub const bls12_g1_msm: Mode = .native;
    pub const bls12_g2_add: Mode = .native;
    pub const bls12_g2_msm: Mode = .native;
    pub const bls12_pairing: Mode = .native;
    pub const bls12_map_fp_to_g1: Mode = .native;
    pub const bls12_map_fp2_to_g2: Mode = .native;
    pub const secp256r1_verify: Mode = .native;
};

// Pairing/MSM pair layouts — must byte-match zesu's extern_bridge.zig; passed straight to
// stdlibs_accel's `anytype` parameters.
const Bn254PairingPair = extern struct { g1: [64]u8, g2: [128]u8 };
const Bls12G1MsmPair = extern struct { point: [96]u8, scalar: [32]u8 };
const Bls12G2MsmPair = extern struct { point: [192]u8, scalar: [32]u8 };
const Bls12PairingPair = extern struct { g1: [96]u8, g2: [192]u8 };

const OK: i32 = 0;
const ERR: i32 = 1;

// ── C-ABI wrappers: extern zkvm_* (ptr+len) → stdlibs_accel's slice/array API ────────────────────
// A `.native` symbol's wrapper is exported (below); an `.intercept` symbol's wrapper is never
// referenced, so its `accel.*` call is not compiled in (the prover supplies that symbol instead).

fn keccak256(data: [*]const u8, len: usize, output: *[32]u8) callconv(.c) i32 {
    // Our 'zkvm_keccak256' writes into a '*zkvm_keccak256_hash' — an extern struct wrapping
    // '[32]u8 align(8)'. The 'output' param here is '*[32]u8', a different pointee type with
    // weaker alignment (implicitly align(1)), so it can't be passed directly. 'hash' gives a
    // correctly typed, 8-aligned destination; we copy its bytes out to 'output' afterwards.
    var hash: linea_accel.keccak.zkvm_keccak256_hash = undefined;
    _ = linea_accel.keccak.zkvm_keccak256(data, len, &hash);
    output.* = hash.data;
    return OK;
}
fn sha256(data: [*]const u8, len: usize, output: *[32]u8) callconv(.c) i32 {
    zesu_accel.sha256(data[0..len], output);
    return OK;
}
fn ripemd160(data: [*]const u8, len: usize, output: *[32]u8) callconv(.c) i32 {
    zesu_accel.ripemd160(data[0..len], output);
    return OK;
}
fn secp256k1_ecrecover(msg: *const [32]u8, sig: *const [64]u8, recid: u8, output: *[64]u8) callconv(.c) i32 {
    return if (zesu_accel.ecrecover(msg, sig, recid, output)) OK else ERR;
}
fn secp256k1_verify(msg: *const [32]u8, sig: *const [64]u8, pubkey: *const [64]u8, verified: *bool) callconv(.c) i32 {
    zesu_accel.secp256k1_verify(msg, sig, pubkey, verified);
    return OK;
}
fn secp256r1_verify(msg: *const [32]u8, sig: *const [64]u8, pubkey: *const [64]u8, verified: *bool) callconv(.c) i32 {
    zesu_accel.secp256r1_verify(msg, sig, pubkey, verified);
    return OK;
}
fn modexp(base: [*]const u8, base_len: usize, exp: [*]const u8, exp_len: usize, modulus: [*]const u8, mod_len: usize, output: [*]u8) callconv(.c) i32 {
    return if (zesu_accel.modexp(base[0..base_len], exp[0..exp_len], modulus[0..mod_len], output[0..mod_len])) OK else ERR;
}
fn bn254_g1_add(p1: *const [64]u8, p2: *const [64]u8, result: *[64]u8) callconv(.c) i32 {
    return if (zesu_accel.bn254_g1_add(p1, p2, result)) OK else ERR;
}
fn bn254_g1_mul(point: *const [64]u8, scalar: *const [32]u8, result: *[64]u8) callconv(.c) i32 {
    return if (zesu_accel.bn254_g1_mul(point, scalar, result)) OK else ERR;
}
fn bn254_pairing(pairs: [*]const Bn254PairingPair, num_pairs: usize, verified: *bool) callconv(.c) i32 {
    return if (zesu_accel.bn254_pairing(pairs[0..num_pairs], verified)) OK else ERR;
}
fn blake2f(rounds: u32, h: *[64]u8, m: *const [128]u8, t: *const [16]u8, f: u8) callconv(.c) i32 {
    return if (zesu_accel.blake2f(rounds, h, m, t, f)) OK else ERR;
}
fn kzg_point_eval(commitment: *const [48]u8, z: *const [32]u8, y: *const [32]u8, proof: *const [48]u8, verified: *bool) callconv(.c) i32 {
    return if (zesu_accel.kzg_point_eval(commitment, z, y, proof, verified)) OK else ERR;
}
fn bls12_g1_add(p1: *const [96]u8, p2: *const [96]u8, result: *[96]u8) callconv(.c) i32 {
    return if (zesu_accel.bls12_g1_add(p1, p2, result)) OK else ERR;
}
fn bls12_g1_msm(pairs: [*]const Bls12G1MsmPair, num_pairs: usize, result: *[96]u8) callconv(.c) i32 {
    return if (zesu_accel.bls12_g1_msm(pairs[0..num_pairs], result)) OK else ERR;
}
fn bls12_g2_add(p1: *const [192]u8, p2: *const [192]u8, result: *[192]u8) callconv(.c) i32 {
    return if (zesu_accel.bls12_g2_add(p1, p2, result)) OK else ERR;
}
fn bls12_g2_msm(pairs: [*]const Bls12G2MsmPair, num_pairs: usize, result: *[192]u8) callconv(.c) i32 {
    return if (zesu_accel.bls12_g2_msm(pairs[0..num_pairs], result)) OK else ERR;
}
fn bls12_pairing(pairs: [*]const Bls12PairingPair, num_pairs: usize, verified: *bool) callconv(.c) i32 {
    return if (zesu_accel.bls12_pairing(pairs[0..num_pairs], verified)) OK else ERR;
}
fn bls12_map_fp_to_g1(field_element: *const [48]u8, result: *[96]u8) callconv(.c) i32 {
    return if (zesu_accel.bls12_map_fp_to_g1(field_element, result)) OK else ERR;
}
fn bls12_map_fp2_to_g2(field_element: *const [96]u8, result: *[192]u8) callconv(.c) i32 {
    return if (zesu_accel.bls12_map_fp2_to_g2(field_element, result)) OK else ERR;
}

// Define (export) exactly the `.native` symbols; `.intercept` ones stay undefined for the prover.
comptime {
    if (policy.keccak256 == .native) @export(&keccak256, .{ .name = "zkvm_keccak256" });
    if (policy.sha256 == .native) @export(&sha256, .{ .name = "zkvm_sha256" });
    if (policy.secp256k1_verify == .native) @export(&secp256k1_verify, .{ .name = "zkvm_secp256k1_verify" });
    if (policy.secp256k1_ecrecover == .native) @export(&secp256k1_ecrecover, .{ .name = "zkvm_secp256k1_ecrecover" });
    if (policy.ripemd160 == .native) @export(&ripemd160, .{ .name = "zkvm_ripemd160" });
    if (policy.modexp == .native) @export(&modexp, .{ .name = "zkvm_modexp" });
    if (policy.bn254_g1_add == .native) @export(&bn254_g1_add, .{ .name = "zkvm_bn254_g1_add" });
    if (policy.bn254_g1_mul == .native) @export(&bn254_g1_mul, .{ .name = "zkvm_bn254_g1_mul" });
    if (policy.bn254_pairing == .native) @export(&bn254_pairing, .{ .name = "zkvm_bn254_pairing" });
    if (policy.blake2f == .native) @export(&blake2f, .{ .name = "zkvm_blake2f" });
    if (policy.kzg_point_eval == .native) @export(&kzg_point_eval, .{ .name = "zkvm_kzg_point_eval" });
    if (policy.bls12_g1_add == .native) @export(&bls12_g1_add, .{ .name = "zkvm_bls12_g1_add" });
    if (policy.bls12_g1_msm == .native) @export(&bls12_g1_msm, .{ .name = "zkvm_bls12_g1_msm" });
    if (policy.bls12_g2_add == .native) @export(&bls12_g2_add, .{ .name = "zkvm_bls12_g2_add" });
    if (policy.bls12_g2_msm == .native) @export(&bls12_g2_msm, .{ .name = "zkvm_bls12_g2_msm" });
    if (policy.bls12_pairing == .native) @export(&bls12_pairing, .{ .name = "zkvm_bls12_pairing" });
    if (policy.bls12_map_fp_to_g1 == .native) @export(&bls12_map_fp_to_g1, .{ .name = "zkvm_bls12_map_fp_to_g1" });
    if (policy.bls12_map_fp2_to_g2 == .native) @export(&bls12_map_fp2_to_g2, .{ .name = "zkvm_bls12_map_fp2_to_g2" });
    if (policy.secp256r1_verify == .native) @export(&secp256r1_verify, .{ .name = "zkvm_secp256r1_verify" });
}
