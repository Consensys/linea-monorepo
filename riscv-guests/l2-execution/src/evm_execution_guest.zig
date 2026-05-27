const std = @import("std");
const builtin = @import("builtin");

const guest_options = @import("guest_options");
const executor = @import("zesu_executor");
const input = @import("zesu_input");
const mpt = @import("zesu_mpt");
const rlp_decode = @import("zesu_rlp_decode");
const evm_guest_allocator = @import("evm_guest_allocator");

const GUEST_INPUT_OFFSET: usize = guest_options.input_offset;
const GUEST_HEAP_OFFSET: usize = 0x50000000;
const GUEST_HEAP_SIZE: usize = 256 * 1024 * 1024;

/// Guest input format:
///   u64 execution_witness_count
///   repeated execution_witness_count times:
///     u64 execution_witness_len
///     u8[execution_witness_len] execution_witness
pub const RunResult = struct {
    execution_witness_count: usize,
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

    fn readSliceArray(self: *Reader, allocator: std.mem.Allocator) ![]const []const u8 {
        const count: usize = @intCast(try self.readU64());
        const items = try allocator.alloc([]const u8, count);
        for (items) |*item| {
            item.* = try self.readLenPrefixedBytes();
        }
        return items;
    }
};

fn guestMain() callconv(.c) noreturn {
    const heap = @as([*]u8, @ptrFromInt(GUEST_HEAP_OFFSET))[0..GUEST_HEAP_SIZE];
    var fba = std.heap.FixedBufferAllocator.init(heap);
    const allocator = fba.allocator();
    evm_guest_allocator.init(allocator);

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
    evm_guest_allocator.init(allocator);
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

    return .{
        .execution_witness_count = execution_witness_count,
    };
}

fn executeWitnessedBlock(allocator: std.mem.Allocator, block_input: []const u8) !BlockResult {
    const si = try decodeStatelessInput(allocator, block_input);

    const proof_out = try executor.executeStatelessInput(allocator, si, si.chain_config.fork_name);
    const ep = si.new_payload_request.execution_payload;

    if (!std.mem.eql(u8, &proof_out.post_state_root, &ep.state_root)) return error.PostStateRootMismatch;
    if (!std.mem.eql(u8, &proof_out.receipts_root, &ep.receipts_root)) {
        return error.ReceiptsRootMismatch;
    }

    return .{
        .number = ep.block_number,
        .pre_state_root = proof_out.pre_state_root,
        .post_state_root = proof_out.post_state_root,
        .receipts_root = proof_out.receipts_root,
    };
}

const ParsedBlock = struct {
    header: input.BlockHeader,
    transactions: []const input.Transaction,
    withdrawals: []const input.Withdrawal,
};

fn decodeStatelessInput(allocator: std.mem.Allocator, block_input: []const u8) !input.StatelessInput {
    var reader = Reader.initSlice(block_input);
    const block_rlp = try reader.readLenPrefixedBytes();
    const nodes = try reader.readSliceArray(allocator);
    const codes = try reader.readSliceArray(allocator);
    _ = try reader.readSliceArray(allocator);
    const headers = try reader.readSliceArray(allocator);
    if (reader.pos != block_input.len) return error.TrailingWitnessInput;

    const block = try decodeBlockRlp(allocator, block_rlp);
    return .{
        .new_payload_request = .{
            .execution_payload = input.payloadFromBlock(block.header, block.transactions, block.withdrawals),
            .parent_beacon_block_root = block.header.parent_beacon_block_root orelse @splat(0),
        },
        .witness = .{
            .nodes = nodes,
            .codes = codes,
            .headers = headers,
        },
        .chain_config = .{ .fork_name = "Cancun" },
    };
}

fn decodeBlockRlp(allocator: std.mem.Allocator, raw: []const u8) !ParsedBlock {
    const outer = mpt.rlp.decodeItem(raw) catch return error.InvalidBlock;
    const block_payload = switch (outer.item) {
        .list => |payload| payload,
        .bytes => return error.InvalidBlock,
    };

    const header_item = mpt.rlp.decodeItem(block_payload) catch return error.InvalidBlock;
    const header_payload = switch (header_item.item) {
        .list => |payload| payload,
        .bytes => return error.InvalidBlock,
    };
    const header = try rlp_decode.decodeBlockHeader(allocator, header_payload);

    const after_header = block_payload[header_item.consumed..];
    const txs_item = mpt.rlp.decodeItem(after_header) catch return error.InvalidBlock;
    const txs_payload = switch (txs_item.item) {
        .list => |payload| payload,
        .bytes => return error.InvalidBlock,
    };
    const transactions = try rlp_decode.decodeTxList(allocator, txs_payload);

    var withdrawals: []const input.Withdrawal = &.{};
    const after_txs = after_header[txs_item.consumed..];
    if (after_txs.len > 0) {
        const ommers_item = mpt.rlp.decodeItem(after_txs) catch return .{
            .header = header,
            .transactions = transactions,
            .withdrawals = &.{},
        };
        const after_ommers = after_txs[ommers_item.consumed..];
        if (after_ommers.len > 0) {
            const withdrawals_item = mpt.rlp.decodeItem(after_ommers) catch return .{
                .header = header,
                .transactions = transactions,
                .withdrawals = &.{},
            };
            const withdrawals_payload = switch (withdrawals_item.item) {
                .list => |payload| payload,
                .bytes => &.{},
            };
            withdrawals = rlp_decode.decodeWithdrawals(allocator, withdrawals_payload) catch &.{};
        }
    }

    return .{
        .header = header,
        .transactions = transactions,
        .withdrawals = withdrawals,
    };
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
