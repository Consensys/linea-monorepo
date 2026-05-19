const std = @import("std");
const builtin = @import("builtin");

const zevm = @import("zevm_stateless");
const stateless_rlp = @import("zevm_stateless_rlp");
const rollup_guest_allocator = @import("rollup_guest_allocator");

const executor = zevm.executor;
const input = zevm.input;
const mpt = zevm.mpt;

const GUEST_INPUT_OFFSET: usize = 0x08000000;
const GUEST_HEAP_OFFSET: usize = 0x50000000;
const GUEST_HEAP_SIZE: usize = 256 * 1024 * 1024;

const EMPTY_TRIE_HASH: [32]u8 = .{
    0x56, 0xe8, 0x1f, 0x17, 0x1b, 0xcc, 0x55, 0xa6,
    0xff, 0x83, 0x45, 0xe6, 0x92, 0xc0, 0xf8, 0x6e,
    0x5b, 0x48, 0xe0, 0x1b, 0x99, 0x6c, 0xad, 0xc0,
    0x01, 0x62, 0x2f, 0xb5, 0xe3, 0x63, 0xb4, 0x21,
};

const KECCAK_EMPTY: [32]u8 = .{
    0xc5, 0xd2, 0x46, 0x01, 0x86, 0xf7, 0x23, 0x3c,
    0x92, 0x7e, 0x7d, 0xb2, 0xdc, 0xc7, 0x03, 0xc0,
    0xe5, 0x00, 0xb6, 0x53, 0xca, 0x82, 0x27, 0x3b,
    0x7b, 0xfa, 0xd8, 0x04, 0x5d, 0x85, 0xa4, 0x70,
};

/// Guest input format:
///   u64 execution_witness_count
///   repeated execution_witness_count times:
///     u64 execution_witness_len
///     u8[execution_witness_len] execution_witness
///   u64 extension_a
///   u64 extension_b
pub const RunResult = struct {
    execution_witness_count: usize,
    extension_sum: u64,
};

const BlockResult = struct {
    number: u64,
    pre_state_root: [32]u8,
    post_state_root: [32]u8,
    receipts_root: [32]u8,
};

const Reader = struct {
    bytes: ?[]const u8,
    base: [*]const u8,
    pos: usize = 0,

    fn initSlice(bytes: []const u8) Reader {
        return .{
            .bytes = bytes,
            .base = undefined,
        };
    }

    fn initMemory(address: usize) Reader {
        return .{
            .bytes = null,
            .base = @as([*]const u8, @ptrFromInt(address)),
        };
    }

    fn readBytes(self: *Reader, len: usize) ![]const u8 {
        if (self.bytes) |bytes| {
            if (self.pos > bytes.len or len > bytes.len - self.pos) return error.UnexpectedEndOfInput;
            const out = bytes[self.pos..][0..len];
            self.pos += len;
            return out;
        }

        const start = self.pos;
        self.pos += len;
        return self.base[start..][0..len];
    }

    fn readU64(self: *Reader) !u64 {
        const bytes = try self.readBytes(8);
        return std.mem.readInt(u64, bytes[0..8], .big);
    }

    fn readLenPrefixedBytes(self: *Reader) ![]const u8 {
        const len: usize = @intCast(try self.readU64());
        return self.readBytes(len);
    }
};

fn guestMain() callconv(.c) noreturn {
    const heap = @as([*]u8, @ptrFromInt(GUEST_HEAP_OFFSET))[0..GUEST_HEAP_SIZE];
    var fba = std.heap.FixedBufferAllocator.init(heap);
    const allocator = fba.allocator();
    rollup_guest_allocator.init(allocator);

    var reader = Reader.initMemory(GUEST_INPUT_OFFSET);
    _ = runFromReader(allocator, &reader) catch exit(1);
    exit(0);
}

comptime {
    if (!builtin.is_test) {
        @export(&guestMain, .{ .name = "main" });
    }
}

/// Execute the same payload format consumed by the RISC-V guest entry point.
///
/// The RISC-V interpreter calls `guestMain`, which reads bytes from
/// `GUEST_INPUT_OFFSET`. Native callers can use this bounded slice entry point
/// to execute the identical parser and block-transition logic locally.
pub fn runPayload(allocator: std.mem.Allocator, payload: []const u8) !RunResult {
    rollup_guest_allocator.init(allocator);
    var reader = Reader.initSlice(payload);
    const result = try runFromReader(allocator, &reader);
    if (reader.pos != payload.len) return error.TrailingInput;
    return result;
}

fn runFromReader(allocator: std.mem.Allocator, reader: *Reader) !RunResult {
    const execution_witness_count: usize = @intCast(try reader.readU64());
    var execution_witness_index: usize = 0;
    while (execution_witness_index < execution_witness_count) : (execution_witness_index += 1) {
        const execution_witness = try reader.readLenPrefixedBytes();
        _ = try executeWitnessedBlock(allocator, execution_witness);
    }

    const a = try reader.readU64();
    const b = try reader.readU64();
    const sum = std.math.add(u64, a, b) catch return error.ExtensionCheckFailed;
    if (sum != 4) return error.ExtensionCheckFailed;

    return .{
        .execution_witness_count = execution_witness_count,
        .extension_sum = sum,
    };
}

fn executeWitnessedBlock(allocator: std.mem.Allocator, block_input: []const u8) !BlockResult {
    const si = try stateless_rlp.decode(allocator, block_input);

    var node_index = try mpt.buildNodeIndex(allocator, si.witness.nodes);
    defer node_index.deinit();

    const pre_state_root = si.witness.state_root;
    const pre_alloc = try buildPreAlloc(allocator, si, &node_index);

    var block_hashes = std.ArrayListUnmanaged(executor.BlockHashEntry).empty;
    defer block_hashes.deinit(allocator);
    try decodeBlockHashes(allocator, si.witness.headers, &block_hashes);

    const proof_out = try executor.executeBlock(
        allocator,
        pre_state_root,
        pre_alloc,
        &node_index,
        si.block,
        si.transactions,
        si.withdrawals,
        block_hashes.items,
        null,
    );

    if (!std.mem.eql(u8, &proof_out.post_state_root, &si.block.state_root)) return error.PostStateRootMismatch;
    if (!std.mem.eql(u8, &proof_out.receipts_root, &si.block.receipts_root)) {
        return error.ReceiptsRootMismatch;
    }

    return .{
        .number = si.block.number,
        .pre_state_root = proof_out.pre_state_root,
        .post_state_root = proof_out.post_state_root,
        .receipts_root = proof_out.receipts_root,
    };
}

fn buildPreAlloc(
    allocator: std.mem.Allocator,
    si: input.StatelessInput,
    node_index: *mpt.NodeIndex,
) !std.AutoHashMapUnmanaged([20]u8, executor.AllocAccount) {
    var pre_alloc = std.AutoHashMapUnmanaged([20]u8, executor.AllocAccount){};
    var current_addr: ?[20]u8 = null;

    for (si.witness.keys) |key| {
        if (key.len == 20) {
            var addr: [20]u8 = undefined;
            @memcpy(&addr, key[0..20]);
            current_addr = addr;

            const account_state = (try mpt.verifyAccountIndexed(
                si.witness.state_root,
                addr,
                node_index,
            )) orelse continue;

            const code: []const u8 = blk: {
                if (std.mem.eql(u8, &account_state.code_hash, &KECCAK_EMPTY)) break :blk &.{};
                for (si.witness.codes) |code_bytes| {
                    if (std.mem.eql(u8, &mpt.keccak256(code_bytes), &account_state.code_hash)) {
                        break :blk code_bytes;
                    }
                }
                break :blk &.{};
            };

            const entry = try pre_alloc.getOrPut(allocator, addr);
            if (!entry.found_existing) {
                entry.value_ptr.* = .{
                    .balance = account_state.balance,
                    .nonce = account_state.nonce,
                    .code = code,
                    .pre_storage_root = account_state.storage_root,
                };
            }
        } else if (key.len == 52) {
            var addr: [20]u8 = undefined;
            @memcpy(&addr, key[0..20]);
            current_addr = addr;

            var raw_slot: [32]u8 = undefined;
            @memcpy(&raw_slot, key[20..52]);
            try putStorageSlot(allocator, &pre_alloc, node_index, si.witness.state_root, addr, raw_slot);
        } else if (key.len == 32) {
            if (current_addr) |addr| {
                var raw_slot: [32]u8 = undefined;
                @memcpy(&raw_slot, key[0..32]);
                try putStorageSlot(allocator, &pre_alloc, node_index, si.witness.state_root, addr, raw_slot);
            }
        }
    }

    return pre_alloc;
}

fn putStorageSlot(
    allocator: std.mem.Allocator,
    pre_alloc: *std.AutoHashMapUnmanaged([20]u8, executor.AllocAccount),
    node_index: *mpt.NodeIndex,
    state_root: [32]u8,
    addr: [20]u8,
    raw_slot: [32]u8,
) !void {
    const acct_state = (try mpt.verifyAccountIndexed(state_root, addr, node_index)) orelse return;
    const value = try mpt.verifyStorageIndexed(acct_state.storage_root, raw_slot, node_index);
    if (value != 0) {
        const entry = try pre_alloc.getOrPut(allocator, addr);
        if (!entry.found_existing) entry.value_ptr.* = .{};
        try entry.value_ptr.*.storage.put(allocator, hashToU256(raw_slot), value);
    }
}

fn hashToU256(hash: [32]u8) u256 {
    var result: u256 = 0;
    for (hash) |byte| {
        result = (result << 8) | byte;
    }
    return result;
}

fn decodeBlockHashes(
    allocator: std.mem.Allocator,
    headers: []const []const u8,
    out: *std.ArrayListUnmanaged(executor.BlockHashEntry),
) !void {
    for (headers) |hdr_rlp| {
        const hash = mpt.keccak256(hdr_rlp);
        const outer = mpt.rlp.decodeItem(hdr_rlp) catch continue;
        var rest = switch (outer.item) {
            .list => |payload| payload,
            .bytes => continue,
        };

        var skip: usize = 0;
        while (skip < 8 and rest.len > 0) : (skip += 1) {
            const field = mpt.rlp.decodeItem(rest) catch break;
            rest = rest[field.consumed..];
        }
        if (rest.len == 0) continue;

        const number_item = mpt.rlp.decodeItem(rest) catch continue;
        const number_bytes = switch (number_item.item) {
            .bytes => |bytes| bytes,
            .list => continue,
        };
        if (number_bytes.len > 8) continue;

        var number: u64 = 0;
        for (number_bytes) |byte| {
            number = (number << 8) | byte;
        }
        try out.append(allocator, .{ .number = number, .hash = hash });
    }
}

fn exit(code: u64) noreturn {
    if (builtin.cpu.arch == .riscv64) {
        asm volatile (
            \\mv a0, %[code]
            \\li a7, 93
            \\ecall
            :
            : [code] "r" (code),
            : .{ .x10 = true, .x17 = true });
        unreachable;
    }

    std.debug.panic("guest exit({d})", .{code});
}
