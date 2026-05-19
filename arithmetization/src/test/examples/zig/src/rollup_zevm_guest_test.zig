const std = @import("std");

const fixtures = @import("rollup_zevm_guest_fixtures");
const guest = @import("rollup_zevm_guest");
const precompile = @import("zevm_precompile");
const rollup_guest_allocator = @import("rollup_guest_allocator");

test "guest program executes two meaningful execution witnesses in one payload" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const allocator = arena.allocator();

    const bundle = try fixtures.loadPayload(allocator, fixtures.embedded.contract_creation_then_ecrecover);
    try std.testing.expectEqual(@as(usize, 2), bundle.blocks.len);

    try std.testing.expectEqual(fixtures.WitnessKind.contract_creation, bundle.blocks[0].kind);
    try std.testing.expectEqual(fixtures.WitnessKind.ecrecover_precompile, bundle.blocks[1].kind);
    try std.testing.expectEqual(@as(u64, 15_537_396), bundle.blocks[0].block_number);
    try std.testing.expectEqual(@as(u64, 15_537_397), bundle.blocks[1].block_number);

    for (bundle.blocks) |block| {
        try std.testing.expectEqual(@as(usize, 1), block.transaction_count);
        try std.testing.expectEqual(@as(usize, 1), block.node_count);
        try std.testing.expectEqual(@as(usize, 1), block.key_count);
        try std.testing.expectEqual(@as(usize, 1), block.header_count);
        try std.testing.expect(block.raw_transaction.len > 0);
        try std.testing.expect(block.evm_input.len > 0);
        try std.testing.expect(block.gas_used > 21_000);
        try std.testing.expect(!std.mem.eql(u8, &block.pre_state_root, &block.post_state_root));
    }

    const result = try guest.runPayload(allocator, bundle.payload);
    try std.testing.expectEqual(@as(usize, 2), result.execution_witness_count);
    try std.testing.expectEqual(bundle.extension_a + bundle.extension_b, result.extension_sum);
}

test "guest parser rejects truncated witness list entries" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const allocator = arena.allocator();

    const bundle = try fixtures.loadPayload(allocator, fixtures.embedded.contract_creation_then_ecrecover);

    var payload = std.ArrayListUnmanaged(u8).empty;
    try appendU64(&payload, allocator, 2);
    try appendLenPrefixedBytes(&payload, allocator, bundle.blocks[0].execution_witness);
    try appendU64(&payload, allocator, 8);

    try std.testing.expectError(
        error.UnexpectedEndOfInput,
        guest.runPayload(allocator, try payload.toOwnedSlice(allocator)),
    );
}

test "guest parser rejects bad extension arithmetic after valid witnesses" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const allocator = arena.allocator();

    const bundle = try fixtures.loadPayload(allocator, fixtures.embedded.contract_creation_then_ecrecover);
    const payload = try allocator.dupe(u8, bundle.payload);
    payload[payload.len - 1] = 1;

    try std.testing.expectError(error.ExtensionCheckFailed, guest.runPayload(allocator, payload));
}

test "guest parser rejects trailing bytes after payload" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const allocator = arena.allocator();

    const bundle = try fixtures.loadPayload(allocator, fixtures.embedded.contract_creation_then_ecrecover);

    var payload = std.ArrayListUnmanaged(u8).empty;
    try payload.appendSlice(allocator, bundle.payload);
    try payload.append(allocator, 0);

    try std.testing.expectError(error.TrailingInput, guest.runPayload(allocator, try payload.toOwnedSlice(allocator)));
}

test "native ecrecover precompile returns the expected signer" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const allocator = arena.allocator();
    rollup_guest_allocator.init(allocator);

    const ec_rec = (precompile.PrecompileId{ .EcRec = {} }).precompile(.Berlin).?;
    const result = ec_rec.execute(try hexToOwnedBytes(allocator, fixtures.vectors.ecrecover_input), 3_000);
    const output = switch (result) {
        .success => |out| out,
        .err => return error.NativeEcrecoverFailed,
    };

    const expected = try hexToArray(20, "0x7C080bBB0eB9B54e73Ff5e8C270e2a242C3bD5cB");
    try std.testing.expect(!output.reverted);
    try std.testing.expectEqual(@as(usize, 32), output.bytes.len);
    try std.testing.expectEqualSlices(u8, &expected, output.bytes[12..32]);
}

fn appendLenPrefixedBytes(
    out: *std.ArrayListUnmanaged(u8),
    allocator: std.mem.Allocator,
    bytes: []const u8,
) !void {
    try appendU64(out, allocator, bytes.len);
    try out.appendSlice(allocator, bytes);
}

fn appendU64(out: *std.ArrayListUnmanaged(u8), allocator: std.mem.Allocator, value: u64) !void {
    var bytes: [8]u8 = undefined;
    std.mem.writeInt(u64, &bytes, value, .big);
    try out.appendSlice(allocator, &bytes);
}

fn hexToOwnedBytes(allocator: std.mem.Allocator, hex: []const u8) ![]const u8 {
    const stripped = stripHexPrefix(hex);
    if (stripped.len % 2 != 0) return error.OddHexLength;
    const out = try allocator.alloc(u8, stripped.len / 2);
    _ = try std.fmt.hexToBytes(out, stripped);
    return out;
}

fn hexToArray(comptime len: usize, hex: []const u8) ![len]u8 {
    const stripped = stripHexPrefix(hex);
    if (stripped.len != len * 2) return error.InvalidHexLength;
    var out: [len]u8 = undefined;
    _ = try std.fmt.hexToBytes(&out, stripped);
    return out;
}

fn stripHexPrefix(hex: []const u8) []const u8 {
    if (hex.len >= 2 and hex[0] == '0' and (hex[1] == 'x' or hex[1] == 'X')) return hex[2..];
    return hex;
}
