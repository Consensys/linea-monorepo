const std = @import("std");

/// One zkevm stateless-block fixture (execution-spec-tests `tests-zkevm` format), reduced to the
/// fields the guest smoke test needs.
pub const StatelessBlockFixture = struct {
    network: []const u8,
    /// SSZ-encoded SszStatelessInput — decode with zesu's `ssz_decode`.
    input: []const u8,
    /// Expected 105-byte SSZ SszStatelessValidationResult — compare with zesu's `ssz_output`.
    expected_output: []const u8,
};

pub const embedded = struct {
    /// Embedded straight from the execution_spec_tests_zkevm dependency (see build.zig).
    pub const zkevm_stateless_block = @embedFile("zkevm_stateless_block.json");
};

/// Parse an execution-spec-tests zkevm blockchain_test fixture and return the first block's SSZ
/// input/output. Fixture shape: { "<test_name>": { "network": "...", "blocks": [
///   { "statelessInputBytes": "0x...", "statelessOutputBytes": "0x...", ... } ] } }
pub fn loadStatelessBlock(allocator: std.mem.Allocator, fixture_json: []const u8) !StatelessBlockFixture {
    const parsed = try std.json.parseFromSlice(std.json.Value, allocator, fixture_json, .{});
    defer parsed.deinit();

    if (parsed.value != .object) return error.InvalidFixture;
    var it = parsed.value.object.iterator();
    const entry = it.next() orelse return error.EmptyFixture;
    const test_case = entry.value_ptr.*;
    if (test_case != .object) return error.InvalidFixture;

    const network = switch (test_case.object.get("network") orelse return error.InvalidFixture) {
        .string => |s| s,
        else => return error.InvalidFixture,
    };

    const blocks = switch (test_case.object.get("blocks") orelse return error.InvalidFixture) {
        .array => |a| a,
        else => return error.InvalidFixture,
    };
    if (blocks.items.len == 0) return error.InvalidFixture;
    const block = blocks.items[0];
    if (block != .object) return error.InvalidFixture;

    const in_hex = switch (block.object.get("statelessInputBytes") orelse return error.InvalidFixture) {
        .string => |s| s,
        else => return error.InvalidFixture,
    };
    const out_hex = switch (block.object.get("statelessOutputBytes") orelse return error.InvalidFixture) {
        .string => |s| s,
        else => return error.InvalidFixture,
    };

    return .{
        .network = try allocator.dupe(u8, network),
        .input = try hexToOwnedBytes(allocator, in_hex),
        .expected_output = try hexToOwnedBytes(allocator, out_hex),
    };
}

fn hexToOwnedBytes(allocator: std.mem.Allocator, hex: []const u8) ![]const u8 {
    const stripped = if (hex.len >= 2 and hex[0] == '0' and (hex[1] == 'x' or hex[1] == 'X')) hex[2..] else hex;
    if (stripped.len % 2 != 0) return error.OddHexLength;
    const out = try allocator.alloc(u8, stripped.len / 2);
    _ = try std.fmt.hexToBytes(out, stripped);
    return out;
}
