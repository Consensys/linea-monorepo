const std = @import("std");

const c = @cImport({
    @cInclude("secp256k1.h");
    @cInclude("secp256k1_recovery.h");
});

var global_ctx_done: std.atomic.Value(bool) = .init(false);
var global_ctx: ?*c.secp256k1_context = null;

fn context() ?*c.secp256k1_context {
    if (!global_ctx_done.load(.acquire)) {
        global_ctx = c.secp256k1_context_create(c.SECP256K1_CONTEXT_VERIFY | c.SECP256K1_CONTEXT_SIGN);
        global_ctx_done.store(true, .release);
    }
    return global_ctx;
}

pub fn keccak256(data: []const u8, output: *[32]u8) void {
    std.crypto.hash.sha3.Keccak256.hash(data, output, .{});
}

pub fn sha256(data: []const u8, output: *[32]u8) void {
    std.crypto.hash.sha2.Sha256.hash(data, output, .{});
}

pub fn secp256k1_verify(msg: *const [32]u8, sig: *const [64]u8, pubkey: *const [64]u8, verified: *bool) void {
    const ctx = context() orelse {
        verified.* = false;
        return;
    };

    var parsed_sig: c.secp256k1_ecdsa_signature = undefined;
    if (c.secp256k1_ecdsa_signature_parse_compact(ctx, &parsed_sig, sig) == 0) {
        verified.* = false;
        return;
    }

    var pubkey_uncompressed: [65]u8 = undefined;
    pubkey_uncompressed[0] = 0x04;
    @memcpy(pubkey_uncompressed[1..], pubkey);

    var parsed_pubkey: c.secp256k1_pubkey = undefined;
    if (c.secp256k1_ec_pubkey_parse(ctx, &parsed_pubkey, &pubkey_uncompressed, 65) == 0) {
        verified.* = false;
        return;
    }

    verified.* = c.secp256k1_ecdsa_verify(ctx, &parsed_sig, msg, &parsed_pubkey) == 1;
}

pub fn ecrecover(msg: *const [32]u8, sig: *const [64]u8, recid: u8, output: *[64]u8) bool {
    const ctx = context() orelse return false;

    var recoverable_sig: c.secp256k1_ecdsa_recoverable_signature = undefined;
    if (c.secp256k1_ecdsa_recoverable_signature_parse_compact(ctx, &recoverable_sig, sig, @intCast(recid)) == 0) {
        return false;
    }

    var pubkey: c.secp256k1_pubkey = undefined;
    if (c.secp256k1_ecdsa_recover(ctx, &pubkey, &recoverable_sig, msg) == 0) {
        return false;
    }

    var pubkey_serialized: [65]u8 = undefined;
    var output_len: usize = 65;
    if (c.secp256k1_ec_pubkey_serialize(ctx, &pubkey_serialized, &output_len, &pubkey, c.SECP256K1_EC_UNCOMPRESSED) == 0) {
        return false;
    }

    @memcpy(output, pubkey_serialized[1..65]);
    return true;
}

pub fn ripemd160(_: []const u8, _: *[32]u8) void {
    @panic("native RIPEMD-160 accelerator is not wired for riscv-guests yet");
}

pub fn modexp(_: []const u8, _: []const u8, _: []const u8, _: []u8) bool {
    @panic("native modexp accelerator is not wired for riscv-guests yet");
}

pub fn bn254_g1_add(_: *const [64]u8, _: *const [64]u8, _: *[64]u8) bool {
    @panic("native BN254 add accelerator is not wired for riscv-guests yet");
}

pub fn bn254_g1_mul(_: *const [64]u8, _: *const [32]u8, _: *[64]u8) bool {
    @panic("native BN254 mul accelerator is not wired for riscv-guests yet");
}

pub fn bn254_pairing(_: anytype, _: *bool) bool {
    @panic("native BN254 pairing accelerator is not wired for riscv-guests yet");
}

pub fn blake2f(_: u32, _: *[64]u8, _: *const [128]u8, _: *const [16]u8, _: u8) bool {
    @panic("native Blake2f accelerator is not wired for riscv-guests yet");
}

pub fn kzg_point_eval(_: *const [48]u8, _: *const [32]u8, _: *const [32]u8, _: *const [48]u8, _: *bool) bool {
    @panic("native KZG point-evaluation accelerator is not wired for riscv-guests yet");
}

pub fn bls12_g1_add(_: *const [96]u8, _: *const [96]u8, _: *[96]u8) bool {
    @panic("native BLS12-381 G1 add accelerator is not wired for riscv-guests yet");
}

pub fn bls12_g1_msm(_: anytype, _: *[96]u8) bool {
    @panic("native BLS12-381 G1 MSM accelerator is not wired for riscv-guests yet");
}

pub fn bls12_g2_add(_: *const [192]u8, _: *const [192]u8, _: *[192]u8) bool {
    @panic("native BLS12-381 G2 add accelerator is not wired for riscv-guests yet");
}

pub fn bls12_g2_msm(_: anytype, _: *[192]u8) bool {
    @panic("native BLS12-381 G2 MSM accelerator is not wired for riscv-guests yet");
}

pub fn bls12_pairing(_: anytype, _: *bool) bool {
    @panic("native BLS12-381 pairing accelerator is not wired for riscv-guests yet");
}

pub fn bls12_map_fp_to_g1(_: *const [48]u8, _: *[96]u8) bool {
    @panic("native BLS12-381 map-fp-to-g1 accelerator is not wired for riscv-guests yet");
}

pub fn bls12_map_fp2_to_g2(_: *const [96]u8, _: *[192]u8) bool {
    @panic("native BLS12-381 map-fp2-to-g2 accelerator is not wired for riscv-guests yet");
}

pub fn secp256r1_verify(_: *const [32]u8, _: *const [64]u8, _: *const [64]u8, verified: *bool) void {
    verified.* = false;
}
