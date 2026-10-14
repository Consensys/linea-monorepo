//! Reads the generated aggregator pair image as a real
//! `verifier.AggregatorInput` and runs the two-proof aggregation end-to-end.
//!
//! The fixture is `testdata/riscv_proof_pair_image.bin`, written by
//! codegen/generate-aggregator-image: the SAME honest riscv proof that backs
//! `testdata/riscv_proof_image.bin`, encoded twice into one pair image (the
//! prover is deterministic, so a second proving run would yield this identical
//! proof anyway). This is the aggregator's cross-language end-to-end check: Go
//! writes the pair layout (16-byte pointer header + two relocated sub-images),
//! Zig maps and rebases it, and `verifier.verifyPair` accepts both proofs plus
//! their public-input consistency against the real compiled system.
//!
//! The fixture carries the production RISC-V system's registered public input.
//! Differing-statement coverage is pinned by verify_pair_test.zig using two
//! independently valid synthetic proofs.

const std = @import("std");
const verifier_ray = @import("verifier_ray");
const riscv_system = @import("riscv_system");

const verifier = verifier_ray.verifier;

// The base address at which the pair image was encoded.
// Must match GuestBase in prover-ray/wiop/proofserialization/layout.go.
const encoded_base: usize = 0x08800000;

const image_path = "testdata/riscv_proof_pair_image.bin";

const o_rdonly: c_int = 0;
const prot_read: c_int = 1;
const prot_write: c_int = 2;
const map_private: c_int = 2;
const seek_end: c_int = 2;
const map_failed = ~@as(usize, 0);

extern fn open(path: [*:0]const u8, flags: c_int) c_int;
extern fn mmap(address: ?*anyopaque, length: usize, prot: c_int, flags: c_int, fd: c_int, offset: i64) *anyopaque;
extern fn close(fd: c_int) c_int;
extern fn lseek(fd: c_int, offset: i64, whence: c_int) i64;

fn loadPairImage() !struct { input: *const verifier.AggregatorInput, base: usize } {
    const fd = open(image_path, o_rdonly);
    if (fd < 0) return error.ImageMissing;
    defer _ = close(fd);

    const image_len = lseek(fd, 0, seek_end);
    if (image_len <= 0) return error.ImageMissing;
    const img_len: usize = @intCast(image_len);

    // Map the file privately at any address the OS picks. Pointer rewrites are
    // copy-on-write and cannot modify the stored pair image.
    const p = mmap(null, img_len, prot_read | prot_write, map_private, fd, 0);
    if (@intFromPtr(p) == map_failed) return error.MmapFailed;

    try verifier_ray.image_relocation.rebasePair(@ptrCast(p), img_len, encoded_base, @intFromPtr(p));
    return .{ .input = @ptrCast(@alignCast(p)), .base = @intFromPtr(p) };
}

test "a Go-encoded aggregator pair image verifies both proofs and their consistency" {
    const loaded = loadPairImage() catch |err| switch (err) {
        // The fixture is generated (not committed — it is above Git hosting
        // size limits); `make generate-testdata` produces it before test runs.
        error.ImageMissing => return error.SkipZigTest,
        else => return err,
    };
    const input = loaded.input;

    // The header must point at two distinct in-image roots, laid out
    // header-then-A-then-B.
    try std.testing.expectEqual(loaded.base + verifier_ray.image_relocation.pair_header_len, @intFromPtr(input.a));
    try std.testing.expect(@intFromPtr(input.b) > @intFromPtr(input.a));
    try std.testing.expect(input.a.proof.rounds.len > 0);
    try std.testing.expect(input.b.proof.rounds.len > 0);

    try verifier.verifyPair(
        riscv_system.system_0_spec,
        riscv_system.system_0_systems,
        input.a.*,
        input.b.*,
    );
}
