const std = @import("std");
const merkle = @import("crypto/merkle.zig");
const field = @import("field/koalabear.zig");
const ext = @import("field/koalabear_ext.zig");
const poseidon2 = @import("crypto/poseidon2.zig");
const protocol = @import("protocol/types.zig");
const fri = @import("query/fri.zig");
const pcs = @import("query/pcs.zig");
const verifier = @import("verifier.zig");

comptime {
    // encoded_pointer/encoded_base/mapped_base are all treated as one 64-bit
    // address space below; every supported target (x86_64, aarch64, R5/RV64)
    // has a 64-bit usize, so no width-narrowing cast is needed or checked for.
    if (@bitSizeOf(usize) != 64) @compileError("image_relocation assumes a 64-bit usize");
}

pub const Error = error{
    PointerBeforeImage,
    PointerPastImage,
    AddressOverflow,
};

/// Returns an encoded absolute pointer's byte offset within its image.
pub fn imageOffset(encoded_pointer: u64, encoded_base: usize) Error!usize {
    if (encoded_pointer < encoded_base) return error.PointerBeforeImage;
    return encoded_pointer - encoded_base;
}

/// Relocates an absolute in-image pointer from encoded_base to mapped_base.
pub fn relocatePointer(encoded_pointer: u64, encoded_base: usize, mapped_base: usize) Error!usize {
    const offset = try imageOffset(encoded_pointer, encoded_base);
    return std.math.add(usize, mapped_base, offset) catch error.AddressOverflow;
}

/// Strides and offsets of the image layout, derived from the types so that a
/// field added to any of them moves this walk with it. Nothing downstream would
/// catch a stale number: a wrong offset patches the wrong bytes and leaves real
/// pointers at their encoded addresses, and the image still casts cleanly.
const abi = struct {
    /// {ptr, len}, both usize; only ptr needs adjusting.
    const slice_hdr = @sizeOf([]const u8);
    const slice_len_off = slice_hdr / 2;

    const public_inputs = @offsetOf(verifier.VerifyInput, "public_inputs");
    const rounds = @offsetOf(verifier.Proof, "rounds");
    const module_sizes = @offsetOf(verifier.Proof, "module_sizes");

    const round_message = @sizeOf(protocol.RoundMessage);
    const round_cells = @offsetOf(protocol.RoundMessage, "cells");
    const round_commitment = @offsetOf(protocol.RoundMessage, "commitment");

    /// OpeningProof lives at Proof.pcs_opening + PcsOpening.proof.
    const opening = @offsetOf(verifier.Proof, "pcs_opening") +
        @offsetOf(verifier.PcsOpening, "proof");
    const input_queries = opening + @offsetOf(pcs.OpeningProof, "input_queries");
    const input_caps = opening + @offsetOf(pcs.OpeningProof, "input_caps");

    const input_cap = @sizeOf(pcs.InputCap);
    const input_cap_nodes = @offsetOf(pcs.InputCap, "nodes");
    const input_cap_tables = @offsetOf(pcs.InputCap, "tables");

    const input_cap_table = @sizeOf(pcs.InputCapTable);
    const input_cap_table_rows = @offsetOf(pcs.InputCapTable, "rows");

    const ito = @sizeOf(merkle.InputTreeOpening);
    const ito_siblings = @offsetOf(merkle.InputTreeOpening, "siblings");
    const ito_leaves = @offsetOf(merkle.InputTreeOpening, "leaves");

    const opt_row_pair = @sizeOf(?merkle.RowPair);
    /// An optional's discriminant offset is not derivable (@offsetOf rejects
    /// optionals), so this assumes the flag follows the payload.
    /// test/proof_abi_test.zig pins the real offset against a null and a
    /// non-null value.
    const row_pair_flag = @sizeOf(merkle.RowPair);
    const row_opening = @sizeOf(merkle.RowOpening);
    const row_base = @offsetOf(merkle.RowOpening, "base");
    const row_ext = @offsetOf(merkle.RowOpening, "ext");

    const fri_proof = opening + @offsetOf(pcs.OpeningProof, "fri_proof");
    const round_roots = fri_proof + @offsetOf(fri.Proof, "round_roots");
    const round_caps = fri_proof + @offsetOf(fri.Proof, "round_caps");
    const final_poly = fri_proof + @offsetOf(fri.Proof, "final_poly");
    const running_queries = fri_proof + @offsetOf(fri.Proof, "running_queries");

    const merkle_cap = @sizeOf(merkle.MerkleCap);
    const merkle_cap_nodes = @offsetOf(merkle.MerkleCap, "nodes");
    const merkle_cap_aux = @offsetOf(merkle.MerkleCap, "aux");

    const branch = @sizeOf(merkle.Branch);
    const branch_siblings = @offsetOf(merkle.Branch, "siblings");
};

/// Patches every slice-pointer in a VerifyInput image from `encoded_base +
/// offset` to `mapped_base + offset`, in place.
///
/// The walk is structural, never scanning for pointer-shaped bytes, so lengths
/// and field values are never touched.
///
/// Call this after loading an image at an address other than the one it was
/// encoded for: the native loader in main.zig and the fixture loader in
/// riscv_proof_image_test.zig both do.
pub fn rebase(img: [*]u8, img_len: usize, encoded_base: usize, mapped_base: usize) void {
    const delta: i64 = @as(i64, @intCast(mapped_base)) - @as(i64, @intCast(encoded_base));

    // Returns the patched pointer's offset within the image, for walking into
    // whatever it points at.
    const patchPtr = struct {
        fn f(image: [*]u8, len: usize, off: usize, enc_base: usize, d: i64) usize {
            if (off + abi.slice_hdr > len) return 0;
            const old_ptr = readU64(image, off);
            if (old_ptr == 0) return 0; // null / empty-slice sentinel
            const new_ptr = @as(u64, @intCast(@as(i64, @intCast(old_ptr)) + d));
            writeU64(image, off, new_ptr);
            return @intCast(@as(i64, @intCast(old_ptr - enc_base)));
        }
    }.f;

    const sliceLen = struct {
        fn f(image: [*]u8, hdr: usize) usize {
            return @intCast(readU64(image, hdr + abi.slice_len_off));
        }
    }.f;

    _ = patchPtr(img, img_len, abi.public_inputs, encoded_base, delta);

    const rounds_ptr = patchPtr(img, img_len, abi.rounds, encoded_base, delta);
    const n_rounds = sliceLen(img, abi.rounds);
    _ = patchPtr(img, img_len, abi.module_sizes, encoded_base, delta);

    for (0..n_rounds) |i| {
        const rm = rounds_ptr + i * abi.round_message;
        _ = patchPtr(img, img_len, rm + abi.round_cells, encoded_base, delta);
    }

    const input_caps_ptr = patchPtr(img, img_len, abi.input_caps, encoded_base, delta);
    const n_input_caps = sliceLen(img, abi.input_caps);
    for (0..n_input_caps) |i| {
        const input_cap = input_caps_ptr + i * abi.input_cap;
        _ = patchPtr(img, img_len, input_cap + abi.input_cap_nodes, encoded_base, delta);

        const tables_hdr = input_cap + abi.input_cap_tables;
        const tables_ptr = patchPtr(img, img_len, tables_hdr, encoded_base, delta);
        const n_tables = sliceLen(img, tables_hdr);
        for (0..n_tables) |j| {
            const table = tables_ptr + j * abi.input_cap_table;
            const rows_hdr = table + abi.input_cap_table_rows;
            const rows_ptr = patchPtr(img, img_len, rows_hdr, encoded_base, delta);
            const n_rows = sliceLen(img, rows_hdr);
            for (0..n_rows) |k| {
                const row = rows_ptr + k * abi.row_opening;
                _ = patchPtr(img, img_len, row + abi.row_base, encoded_base, delta);
                _ = patchPtr(img, img_len, row + abi.row_ext, encoded_base, delta);
            }
        }
    }

    const iq_outer_ptr = patchPtr(img, img_len, abi.input_queries, encoded_base, delta);
    const n_iq = sliceLen(img, abi.input_queries);
    for (0..n_iq) |i| {
        const inner_hdr = iq_outer_ptr + i * abi.slice_hdr;
        const iq_inner_ptr = patchPtr(img, img_len, inner_hdr, encoded_base, delta);
        const n_ito = sliceLen(img, inner_hdr);
        for (0..n_ito) |j| {
            const ito_off = iq_inner_ptr + j * abi.ito;
            _ = patchPtr(img, img_len, ito_off + abi.ito_siblings, encoded_base, delta);

            // leaves holds ?RowPair inline, so elements are walked at that
            // stride rather than dereferenced.
            const leaves_hdr = ito_off + abi.ito_leaves;
            const leaves_ptr = patchPtr(img, img_len, leaves_hdr, encoded_base, delta);
            const n_leaves = sliceLen(img, leaves_hdr);
            for (0..n_leaves) |k| {
                const leaf_off = leaves_ptr + k * abi.opt_row_pair;
                if (leaf_off + abi.row_pair_flag + 1 > img_len) continue;
                if (img[leaf_off + abi.row_pair_flag] == 0) continue; // absent
                for (0..2) |r| {
                    const row = leaf_off + r * abi.row_opening;
                    _ = patchPtr(img, img_len, row + abi.row_base, encoded_base, delta);
                    _ = patchPtr(img, img_len, row + abi.row_ext, encoded_base, delta);
                }
            }
        }
    }

    _ = patchPtr(img, img_len, abi.round_roots, encoded_base, delta);

    const round_caps_ptr = patchPtr(img, img_len, abi.round_caps, encoded_base, delta);
    const n_round_caps = sliceLen(img, abi.round_caps);
    for (0..n_round_caps) |i| {
        const round_cap = round_caps_ptr + i * abi.merkle_cap;
        _ = patchPtr(img, img_len, round_cap + abi.merkle_cap_nodes, encoded_base, delta);
        _ = patchPtr(img, img_len, round_cap + abi.merkle_cap_aux, encoded_base, delta);
    }

    _ = patchPtr(img, img_len, abi.final_poly, encoded_base, delta);

    const rq_outer_ptr = patchPtr(img, img_len, abi.running_queries, encoded_base, delta);
    const n_rq = sliceLen(img, abi.running_queries);
    for (0..n_rq) |i| {
        const inner_hdr = rq_outer_ptr + i * abi.slice_hdr;
        const rq_inner_ptr = patchPtr(img, img_len, inner_hdr, encoded_base, delta);
        const n_br = sliceLen(img, inner_hdr);
        for (0..n_br) |j| {
            const br = rq_inner_ptr + j * abi.branch;
            _ = patchPtr(img, img_len, br + abi.branch_siblings, encoded_base, delta);
        }
    }
}

/// Byte-for-byte layout of the aggregator pair image header: two absolute
/// little-endian u64 pointers, matching `EncodeAggregatorPair`'s header write
/// and `verifier.AggregatorInput`.
pub const pair_header_len = 2 * @sizeOf(u64);

/// Rebases an aggregator pair image (see `verifier.AggregatorInput`): patches
/// the two header pointers, then rebases each `VerifyInput` sub-image they
/// point to. `EncodeAggregatorPair` encodes each sub-image with `Encode` at its
/// own absolute start address, so each is walked relative to that start rather
/// than to the pair image's base.
pub fn rebasePair(img: [*]u8, img_len: usize, encoded_base: usize, mapped_base: usize) Error!void {
    const root_len = @sizeOf(verifier.VerifyInput);
    if (img_len < pair_header_len) return error.PointerPastImage;
    const a_ptr = readU64(img, 0);
    const b_ptr = readU64(img, @sizeOf(u64));
    const a_off = try imageOffset(a_ptr, encoded_base);
    const b_off = try imageOffset(b_ptr, encoded_base);
    if (img_len < pair_header_len + 2 * root_len or
        a_off != pair_header_len or b_off < a_off + root_len or
        b_off % @alignOf(verifier.VerifyInput) != 0 or b_off + root_len > img_len)
    {
        return error.PointerPastImage;
    }

    // The pair encoder validates each member structurally before writing it.
    // Keep native relocation bounded to each member so malformed roots cannot
    // make one walk alias the other.
    writeU64(img, 0, @intCast(try relocatePointer(a_ptr, encoded_base, mapped_base)));
    writeU64(img, @sizeOf(u64), @intCast(try relocatePointer(b_ptr, encoded_base, mapped_base)));
    rebase(img + a_off, b_off - a_off, encoded_base + a_off, mapped_base + a_off);
    rebase(img + b_off, img_len - b_off, encoded_base + b_off, mapped_base + b_off);
}

const Region = struct {
    offset: usize,
    len: usize,
};

/// Validates an encoded VerifyInput without constructing Zig slices or tagged
/// unions from untrusted bytes.
pub fn validateImage(img: [*]const u8, img_len: usize, encoded_base: usize) Error!void {
    try requireRange(0, @sizeOf(verifier.VerifyInput), img_len);
    try validateScalars(img, img_len, try resolveSlice(img, img_len, abi.public_inputs, @sizeOf(protocol.Scalar), @alignOf(protocol.Scalar), encoded_base));
    _ = try resolveSlice(img, img_len, abi.module_sizes, @sizeOf(usize), @alignOf(usize), encoded_base);

    const rounds = try resolveSlice(img, img_len, abi.rounds, abi.round_message, @alignOf(protocol.RoundMessage), encoded_base);
    for (0..rounds.len) |i| {
        const round = rounds.offset + i * abi.round_message;
        try validateOptionalDigest(img, img_len, round + abi.round_commitment);
        try validateScalars(img, img_len, try resolveSlice(img, img_len, round + abi.round_cells, @sizeOf(protocol.Scalar), @alignOf(protocol.Scalar), encoded_base));
    }

    const input_caps = try resolveSlice(img, img_len, abi.input_caps, abi.input_cap, @alignOf(pcs.InputCap), encoded_base);
    for (0..input_caps.len) |i| {
        const cap = input_caps.offset + i * abi.input_cap;
        try validateElements(img, img_len, try resolveSlice(img, img_len, cap + abi.input_cap_nodes, @sizeOf(poseidon2.Digest), @alignOf(poseidon2.Digest), encoded_base), @sizeOf(poseidon2.Digest));
        const tables = try resolveSlice(img, img_len, cap + abi.input_cap_tables, abi.input_cap_table, @alignOf(pcs.InputCapTable), encoded_base);
        for (0..tables.len) |j| {
            const table = tables.offset + j * abi.input_cap_table;
            try validateRows(img, img_len, try resolveSlice(img, img_len, table + abi.input_cap_table_rows, abi.row_opening, @alignOf(merkle.RowOpening), encoded_base), encoded_base);
        }
    }

    const queries = try resolveSlice(img, img_len, abi.input_queries, abi.slice_hdr, @alignOf([]const merkle.InputTreeOpening), encoded_base);
    for (0..queries.len) |i| {
        const openings = try resolveSlice(img, img_len, queries.offset + i * abi.slice_hdr, abi.ito, @alignOf(merkle.InputTreeOpening), encoded_base);
        for (0..openings.len) |j| {
            const opening = openings.offset + j * abi.ito;
            try validateElements(img, img_len, try resolveSlice(img, img_len, opening + abi.ito_siblings, @sizeOf(poseidon2.Digest), @alignOf(poseidon2.Digest), encoded_base), @sizeOf(poseidon2.Digest));
            const leaves = try resolveSlice(img, img_len, opening + abi.ito_leaves, abi.opt_row_pair, @alignOf(?merkle.RowPair), encoded_base);
            for (0..leaves.len) |k| {
                const leaf = leaves.offset + k * abi.opt_row_pair;
                const flag = img[leaf + abi.row_pair_flag];
                if (flag > 1) return error.PointerPastImage;
                if (flag == 1) try validateRows(img, img_len, .{ .offset = leaf, .len = 2 }, encoded_base);
            }
        }
    }

    try validateElements(img, img_len, try resolveSlice(img, img_len, abi.round_roots, @sizeOf(poseidon2.Digest), @alignOf(poseidon2.Digest), encoded_base), @sizeOf(poseidon2.Digest));
    const caps = try resolveSlice(img, img_len, abi.round_caps, abi.merkle_cap, @alignOf(merkle.MerkleCap), encoded_base);
    for (0..caps.len) |i| {
        const cap = caps.offset + i * abi.merkle_cap;
        try validateElements(img, img_len, try resolveSlice(img, img_len, cap + abi.merkle_cap_nodes, @sizeOf(poseidon2.Digest), @alignOf(poseidon2.Digest), encoded_base), @sizeOf(poseidon2.Digest));
        const aux = try resolveSlice(img, img_len, cap + abi.merkle_cap_aux, @sizeOf(?poseidon2.Digest), @alignOf(?poseidon2.Digest), encoded_base);
        for (0..aux.len) |j| try validateOptionalDigest(img, img_len, aux.offset + j * @sizeOf(?poseidon2.Digest));
    }
    try validateElements(img, img_len, try resolveSlice(img, img_len, abi.final_poly, @sizeOf(ext.Ext), @alignOf(ext.Ext), encoded_base), @sizeOf(ext.Ext));

    const running = try resolveSlice(img, img_len, abi.running_queries, abi.slice_hdr, @alignOf([]const merkle.Branch), encoded_base);
    for (0..running.len) |i| {
        const branches = try resolveSlice(img, img_len, running.offset + i * abi.slice_hdr, abi.branch, @alignOf(merkle.Branch), encoded_base);
        for (0..branches.len) |j| {
            const branch = branches.offset + j * abi.branch;
            try validateElements(img, img_len, try resolveSlice(img, img_len, branch + abi.branch_siblings, @sizeOf(poseidon2.Digest), @alignOf(poseidon2.Digest), encoded_base), @sizeOf(poseidon2.Digest));
            try validateElements(img, img_len, .{ .offset = branch + @offsetOf(merkle.Branch, "leaf"), .len = 1 }, @sizeOf(poseidon2.Digest));
        }
    }
}

fn resolveSlice(img: [*]const u8, img_len: usize, header: usize, elem_size: usize, alignment: usize, encoded_base: usize) Error!Region {
    try requireRange(header, abi.slice_hdr, img_len);
    const pointer = readU64(img, header);
    if (pointer < encoded_base) return error.PointerBeforeImage;
    const offset: usize = @intCast(pointer - encoded_base);
    const len: usize = @intCast(readU64(img, header + abi.slice_len_off));
    if (offset % alignment != 0) return error.PointerPastImage;
    const extent = std.math.mul(usize, len, elem_size) catch return error.AddressOverflow;
    try requireRange(offset, extent, img_len);
    return .{ .offset = offset, .len = len };
}

fn validateRows(img: [*]const u8, img_len: usize, rows: Region, encoded_base: usize) Error!void {
    for (0..rows.len) |i| {
        const row = rows.offset + i * abi.row_opening;
        try validateElements(img, img_len, try resolveSlice(img, img_len, row + abi.row_base, @sizeOf(field.Element), @alignOf(field.Element), encoded_base), @sizeOf(field.Element));
        try validateElements(img, img_len, try resolveSlice(img, img_len, row + abi.row_ext, @sizeOf(ext.Ext), @alignOf(ext.Ext), encoded_base), @sizeOf(ext.Ext));
    }
}

fn validateScalars(img: [*]const u8, img_len: usize, scalars: Region) Error!void {
    for (0..scalars.len) |i| {
        const scalar = scalars.offset + i * @sizeOf(protocol.Scalar);
        if (img[scalar + @sizeOf(ext.Ext)] > 1) return error.PointerPastImage;
        try validateElements(img, img_len, .{ .offset = scalar, .len = 1 }, @sizeOf(ext.Ext));
    }
}

fn validateOptionalDigest(img: [*]const u8, img_len: usize, offset: usize) Error!void {
    try requireRange(offset, @sizeOf(?poseidon2.Digest), img_len);
    const flag = img[offset + @sizeOf(poseidon2.Digest)];
    if (flag > 1) return error.PointerPastImage;
    if (flag == 1) try validateElements(img, img_len, .{ .offset = offset, .len = 1 }, @sizeOf(poseidon2.Digest));
}

fn validateElements(img: [*]const u8, img_len: usize, values: Region, stride: usize) Error!void {
    try requireRange(values.offset, values.len * stride, img_len);
    for (0..values.len) |i| {
        for (0..stride / @sizeOf(field.Element)) |j| {
            if (readU32(img, values.offset + i * stride + j * @sizeOf(field.Element)) >= field.modulus)
                return error.PointerPastImage;
        }
    }
}

fn requireRange(offset: usize, len: usize, image_len: usize) Error!void {
    const end = std.math.add(usize, offset, len) catch return error.AddressOverflow;
    if (end > image_len) return error.PointerPastImage;
}

fn readU32(img: [*]const u8, off: usize) u32 {
    return @as(u32, img[off]) |
        (@as(u32, img[off + 1]) << 8) |
        (@as(u32, img[off + 2]) << 16) |
        (@as(u32, img[off + 3]) << 24);
}

fn readU64(img: [*]const u8, off: usize) u64 {
    return @as(u64, img[off]) |
        (@as(u64, img[off + 1]) << 8) |
        (@as(u64, img[off + 2]) << 16) |
        (@as(u64, img[off + 3]) << 24) |
        (@as(u64, img[off + 4]) << 32) |
        (@as(u64, img[off + 5]) << 40) |
        (@as(u64, img[off + 6]) << 48) |
        (@as(u64, img[off + 7]) << 56);
}

fn writeU64(img: [*]u8, off: usize, v: u64) void {
    img[off + 0] = @truncate(v);
    img[off + 1] = @truncate(v >> 8);
    img[off + 2] = @truncate(v >> 16);
    img[off + 3] = @truncate(v >> 24);
    img[off + 4] = @truncate(v >> 32);
    img[off + 5] = @truncate(v >> 40);
    img[off + 6] = @truncate(v >> 48);
    img[off + 7] = @truncate(v >> 56);
}
