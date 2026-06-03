const std = @import("std");

const Bn254PairingPair = extern struct { g1: [64]u8, g2: [128]u8 };
const Bls12G1MsmPair = extern struct { point: [96]u8, scalar: [32]u8 };
const Bls12G2MsmPair = extern struct { point: [192]u8, scalar: [32]u8 };
const Bls12PairingPair = extern struct { g1: [96]u8, g2: [192]u8 };

// RISC-V guest accel_impl for Zesu's crypto/precompile boundary.
//
// Zesu's precompile table calls src/evm/precompile/default_impls.zig, and those
// wrappers delegate the heavy crypto to the injected "accelerators" module. The
// functions below keep the same single accel_impl boundary for Linea. The
// exported zkvm_* symbols are stable interception points for a future zkVM host
// or circuit-aware interpreter. Until that layer exists, pure hashes run locally
// and unsupported precompile accelerators return failure.
pub fn keccak256(data: []const u8, output: *[32]u8) void {
    _ = @call(.never_inline, zkvm_keccak256, .{ data.ptr, data.len, output });
}

pub fn sha256(data: []const u8, output: *[32]u8) void {
    _ = @call(.never_inline, zkvm_sha256, .{ data.ptr, data.len, output });
}

pub fn secp256k1_verify(msg: *const [32]u8, sig: *const [64]u8, pubkey: *const [64]u8, verified: *bool) void {
    _ = @call(.never_inline, zkvm_secp256k1_verify, .{ msg, sig, pubkey, verified });
}

pub fn ecrecover(msg: *const [32]u8, sig: *const [64]u8, recid: u8, output: *[64]u8) bool {
    return @call(.never_inline, zkvm_secp256k1_ecrecover, .{ msg, sig, recid, output }) == 0;
}

pub fn ripemd160(data: []const u8, output: *[32]u8) void {
    _ = @call(.never_inline, zkvm_ripemd160, .{ data.ptr, data.len, output });
}

pub fn modexp(base: []const u8, exp: []const u8, modulus: []const u8, output: []u8) bool {
    return @call(
        .never_inline,
        zkvm_modexp,
        .{ base.ptr, base.len, exp.ptr, exp.len, modulus.ptr, modulus.len, output.ptr },
    ) == 0;
}

pub fn bn254_g1_add(p1: *const [64]u8, p2: *const [64]u8, result: *[64]u8) bool {
    return @call(.never_inline, zkvm_bn254_g1_add, .{ p1, p2, result }) == 0;
}

pub fn bn254_g1_mul(point: *const [64]u8, scalar: *const [32]u8, result: *[64]u8) bool {
    return @call(.never_inline, zkvm_bn254_g1_mul, .{ point, scalar, result }) == 0;
}

pub fn bn254_pairing(pairs: anytype, verified: *bool) bool {
    const ptr: [*]const Bn254PairingPair = @ptrCast(pairs.ptr);
    return @call(.never_inline, zkvm_bn254_pairing, .{ ptr, pairs.len, verified }) == 0;
}

pub fn blake2f(rounds: u32, h: *[64]u8, m: *const [128]u8, t: *const [16]u8, f: u8) bool {
    return @call(.never_inline, zkvm_blake2f, .{ rounds, h, m, t, f }) == 0;
}

pub fn kzg_point_eval(commitment: *const [48]u8, z: *const [32]u8, y: *const [32]u8, proof: *const [48]u8, verified: *bool) bool {
    return @call(.never_inline, zkvm_kzg_point_eval, .{ commitment, z, y, proof, verified }) == 0;
}

pub fn bls12_g1_add(p1: *const [96]u8, p2: *const [96]u8, result: *[96]u8) bool {
    return @call(.never_inline, zkvm_bls12_g1_add, .{ p1, p2, result }) == 0;
}

pub fn bls12_g1_msm(pairs: anytype, result: *[96]u8) bool {
    const ptr: [*]const Bls12G1MsmPair = @ptrCast(pairs.ptr);
    return @call(.never_inline, zkvm_bls12_g1_msm, .{ ptr, pairs.len, result }) == 0;
}

pub fn bls12_g2_add(p1: *const [192]u8, p2: *const [192]u8, result: *[192]u8) bool {
    return @call(.never_inline, zkvm_bls12_g2_add, .{ p1, p2, result }) == 0;
}

pub fn bls12_g2_msm(pairs: anytype, result: *[192]u8) bool {
    const ptr: [*]const Bls12G2MsmPair = @ptrCast(pairs.ptr);
    return @call(.never_inline, zkvm_bls12_g2_msm, .{ ptr, pairs.len, result }) == 0;
}

pub fn bls12_pairing(pairs: anytype, verified: *bool) bool {
    const ptr: [*]const Bls12PairingPair = @ptrCast(pairs.ptr);
    return @call(.never_inline, zkvm_bls12_pairing, .{ ptr, pairs.len, verified }) == 0;
}

pub fn bls12_map_fp_to_g1(field_element: *const [48]u8, result: *[96]u8) bool {
    return @call(.never_inline, zkvm_bls12_map_fp_to_g1, .{ field_element, result }) == 0;
}

pub fn bls12_map_fp2_to_g2(field_element: *const [96]u8, result: *[192]u8) bool {
    return @call(.never_inline, zkvm_bls12_map_fp2_to_g2, .{ field_element, result }) == 0;
}

pub fn secp256r1_verify(msg: *const [32]u8, sig: *const [64]u8, pubkey: *const [64]u8, verified: *bool) void {
    _ = @call(.never_inline, zkvm_secp256r1_verify, .{ msg, sig, pubkey, verified });
}

export fn zkvm_keccak256(data: [*]const u8, len: usize, output: *[32]u8) callconv(.c) i32 {
    std.crypto.hash.sha3.Keccak256.hash(data[0..len], output, .{});
    return 0;
}

export fn zkvm_sha256(data: [*]const u8, len: usize, output: *[32]u8) callconv(.c) i32 {
    std.crypto.hash.sha2.Sha256.hash(data[0..len], output, .{});
    return 0;
}

export fn zkvm_secp256k1_verify(_: *const [32]u8, _: *const [64]u8, _: *const [64]u8, verified: *bool) callconv(.c) i32 {
    verified.* = false;
    return 1;
}

export fn zkvm_secp256k1_ecrecover(_: *const [32]u8, _: *const [64]u8, _: u8, output: *[64]u8) callconv(.c) i32 {
    output.* = [_]u8{0} ** 64;
    return 1;
}

export fn zkvm_ripemd160(_: [*]const u8, _: usize, output: *[32]u8) callconv(.c) i32 {
    output.* = [_]u8{0} ** 32;
    return 1;
}

export fn zkvm_modexp(
    _: [*]const u8,
    _: usize,
    _: [*]const u8,
    _: usize,
    _: [*]const u8,
    mod_len: usize,
    output: [*]u8,
) callconv(.c) i32 {
    @memset(output[0..mod_len], 0);
    return 1;
}

export fn zkvm_bn254_g1_add(_: *const [64]u8, _: *const [64]u8, result: *[64]u8) callconv(.c) i32 {
    result.* = [_]u8{0} ** 64;
    return 1;
}

export fn zkvm_bn254_g1_mul(_: *const [64]u8, _: *const [32]u8, result: *[64]u8) callconv(.c) i32 {
    result.* = [_]u8{0} ** 64;
    return 1;
}

export fn zkvm_bn254_pairing(_: [*]const Bn254PairingPair, _: usize, verified: *bool) callconv(.c) i32 {
    verified.* = false;
    return 1;
}

export fn zkvm_blake2f(_: u32, _: *[64]u8, _: *const [128]u8, _: *const [16]u8, _: u8) callconv(.c) i32 {
    return 1;
}

export fn zkvm_kzg_point_eval(_: *const [48]u8, _: *const [32]u8, _: *const [32]u8, _: *const [48]u8, verified: *bool) callconv(.c) i32 {
    verified.* = false;
    return 1;
}

export fn zkvm_bls12_g1_add(_: *const [96]u8, _: *const [96]u8, result: *[96]u8) callconv(.c) i32 {
    result.* = [_]u8{0} ** 96;
    return 1;
}

export fn zkvm_bls12_g1_msm(_: [*]const Bls12G1MsmPair, _: usize, result: *[96]u8) callconv(.c) i32 {
    result.* = [_]u8{0} ** 96;
    return 1;
}

export fn zkvm_bls12_g2_add(_: *const [192]u8, _: *const [192]u8, result: *[192]u8) callconv(.c) i32 {
    result.* = [_]u8{0} ** 192;
    return 1;
}

export fn zkvm_bls12_g2_msm(_: [*]const Bls12G2MsmPair, _: usize, result: *[192]u8) callconv(.c) i32 {
    result.* = [_]u8{0} ** 192;
    return 1;
}

export fn zkvm_bls12_pairing(_: [*]const Bls12PairingPair, _: usize, verified: *bool) callconv(.c) i32 {
    verified.* = false;
    return 1;
}

export fn zkvm_bls12_map_fp_to_g1(_: *const [48]u8, result: *[96]u8) callconv(.c) i32 {
    result.* = [_]u8{0} ** 96;
    return 1;
}

export fn zkvm_bls12_map_fp2_to_g2(_: *const [96]u8, result: *[192]u8) callconv(.c) i32 {
    result.* = [_]u8{0} ** 192;
    return 1;
}

export fn zkvm_secp256r1_verify(_: *const [32]u8, _: *const [64]u8, _: *const [64]u8, verified: *bool) callconv(.c) i32 {
    verified.* = false;
    return 1;
}
