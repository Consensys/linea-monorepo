const std = @import("std");

const rollup_guest_allocator = @import("rollup_guest_allocator");
const fixture_data = @import("rollup_zevm_guest_fixture_data");
const stateless_rlp = @import("zevm_stateless_rlp");
const zevm = @import("zevm_stateless");

const executor = zevm.executor;
const mpt = zevm.mpt;

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

const KECCAK_EMPTY_LIST: [32]u8 = .{
    0x1d, 0xcc, 0x4d, 0xe8, 0xde, 0xc7, 0x5d, 0x7a,
    0xab, 0x85, 0xb5, 0x67, 0xb6, 0xcc, 0xd4, 0x1a,
    0xd3, 0x12, 0x45, 0x1b, 0x94, 0x8a, 0x74, 0x13,
    0xf0, 0xa1, 0x42, 0xfd, 0x40, 0xd4, 0x93, 0x47,
};

const ZERO_HASH: [32]u8 = [_]u8{0} ** 32;
const ZERO_ADDRESS: [20]u8 = [_]u8{0} ** 20;
const ZERO_BLOOM: [256]u8 = [_]u8{0} ** 256;
const ZERO_NONCE: [8]u8 = [_]u8{0} ** 8;

const PRESTATE_BALANCE: u256 = 1_000_000_000_000_000_000;
const GAS_LIMIT: u64 = 30_000_000;

pub const vectors = struct {
    pub const contract_creation_raw_tx =
        "0xf85d800a83030d408080916005600c60003960056000f3600160005526a0d19f551b39965ae4efb3acb68ffdeba5d9e18bd74387e8c445f3aad84f397a60a06164e46866859ad59589767e73cf3b121cf1da264d1e96f79990cdf27d12bf62";
    pub const contract_creation_sender = "0x7C080bBB0eB9B54e73Ff5e8C270e2a242C3bD5cB";
    pub const contract_creation_init_code = "0x6005600c60003960056000f36001600055";

    pub const ecrecover_raw_tx =
        "0xf8e1800a830186a094000000000000000000000000000000000000000180b8801111111111111111111111111111111111111111111111111111111111111111000000000000000000000000000000000000000000000000000000000000001b458725a29ba10982a3228326a9fe72aaf6700c72716c7adee2be27cbe5864d2f763794bdaece6bfab0d5832a520be1baae37e61efed08a6edc54c2bffa8fec0725a007e35f67a72c9ad312658f229fd4ac3c62752033ee5ce926e213b3626daf092fa03d130989fe009a701bbfa21ac3fd1c582c048024815fa320dd03f3a369d99904";
    pub const ecrecover_sender = "0x920cb04543C6d864aF2841544381d7B44f09cdA2";
    pub const ecrecover_input =
        "0x1111111111111111111111111111111111111111111111111111111111111111000000000000000000000000000000000000000000000000000000000000001b458725a29ba10982a3228326a9fe72aaf6700c72716c7adee2be27cbe5864d2f763794bdaece6bfab0d5832a520be1baae37e61efed08a6edc54c2bffa8fec07";
};

pub const embedded = struct {
    pub const contract_creation_then_ecrecover = fixture_data.contract_creation_then_ecrecover_json;
};

pub const WitnessKind = enum {
    contract_creation,
    ecrecover_precompile,
};

pub const BlockFixture = struct {
    execution_witness: []const u8,
    kind: WitnessKind,
    block_number: u64,
    transaction_count: usize,
    node_count: usize,
    code_count: usize,
    key_count: usize,
    header_count: usize,
    pre_state_root: [32]u8,
    post_state_root: [32]u8,
    receipts_root: [32]u8,
    gas_used: u64,
    raw_transaction: []const u8,
    evm_input: []const u8,
};

pub const PayloadFixture = struct {
    payload: []const u8,
    blocks: []const BlockFixture,
    extension_a: u64,
    extension_b: u64,
};

pub const BlockSpec = struct {
    kind: WitnessKind,
    block_number: u64,
    timestamp: u64,
    sender: []const u8,
    raw_transaction: []const u8,
    evm_input: []const u8,
};

pub const Scenario = struct {
    description: []const u8,
    extension_a: u64 = 2,
    extension_b: u64 = 2,
    blocks: []const BlockSpec,
};

pub const scenarios = struct {
    const contract_creation_then_ecrecover_blocks = [_]BlockSpec{
        .{
            .kind = .contract_creation,
            .block_number = 15_537_396,
            .timestamp = 3,
            .sender = vectors.contract_creation_sender,
            .raw_transaction = vectors.contract_creation_raw_tx,
            .evm_input = vectors.contract_creation_init_code,
        },
        .{
            .kind = .ecrecover_precompile,
            .block_number = 15_537_397,
            .timestamp = 4,
            .sender = vectors.ecrecover_sender,
            .raw_transaction = vectors.ecrecover_raw_tx,
            .evm_input = vectors.ecrecover_input,
        },
    };

    pub const contract_creation_then_ecrecover = Scenario{
        .description = "rollup_zevm_guest payload with contract creation and ecrecover precompile execution witnesses",
        .blocks = &contract_creation_then_ecrecover_blocks,
    };
};

const DecodedBlockSpec = struct {
    kind: WitnessKind,
    block_number: u64,
    timestamp: u64,
    sender: [20]u8,
    raw_transaction: []const u8,
    evm_input: []const u8,
};

const PreState = struct {
    root: [32]u8,
    nodes: []const []const u8,
};

const FixtureFile = struct {
    description: ?[]const u8 = null,
    payload: []const u8,
    blocks: []const BlockFixtureFile,
};

const BlockFixtureFile = struct {
    kind: []const u8,
    block_number: u64,
    transaction_count: usize,
    node_count: usize,
    code_count: usize,
    key_count: usize,
    header_count: usize,
    pre_state_root: []const u8,
    post_state_root: []const u8,
    receipts_root: []const u8,
    gas_used: u64,
    raw_transaction: []const u8,
    evm_input: []const u8,
};

const HeaderSpec = struct {
    parent_hash: [32]u8 = ZERO_HASH,
    state_root: [32]u8,
    transactions_root: [32]u8 = EMPTY_TRIE_HASH,
    receipts_root: [32]u8 = EMPTY_TRIE_HASH,
    logs_bloom: [256]u8 = ZERO_BLOOM,
    number: u64,
    gas_limit: u64 = GAS_LIMIT,
    gas_used: u64 = 0,
    timestamp: u64,
    base_fee_per_gas: u64 = 0,
};

const PayloadEnvelope = struct {
    witnesses: []const []const u8,
    extension_a: u64,
    extension_b: u64,
};

pub fn loadPayload(allocator: std.mem.Allocator, fixture_json: []const u8) !PayloadFixture {
    const parsed = try std.json.parseFromSlice(
        FixtureFile,
        allocator,
        fixture_json,
        .{ .ignore_unknown_fields = true },
    );
    defer parsed.deinit();

    const payload = try hexToOwnedBytes(allocator, parsed.value.payload);
    const envelope = try parsePayloadEnvelope(allocator, payload);
    if (envelope.witnesses.len != parsed.value.blocks.len) return error.InvalidFixture;

    const blocks = try allocator.alloc(BlockFixture, parsed.value.blocks.len);
    for (parsed.value.blocks, 0..) |block, i| {
        blocks[i] = .{
            .execution_witness = envelope.witnesses[i],
            .kind = try parseWitnessKind(block.kind),
            .block_number = block.block_number,
            .transaction_count = block.transaction_count,
            .node_count = block.node_count,
            .code_count = block.code_count,
            .key_count = block.key_count,
            .header_count = block.header_count,
            .pre_state_root = try hexToArray(32, block.pre_state_root),
            .post_state_root = try hexToArray(32, block.post_state_root),
            .receipts_root = try hexToArray(32, block.receipts_root),
            .gas_used = block.gas_used,
            .raw_transaction = try hexToOwnedBytes(allocator, block.raw_transaction),
            .evm_input = try hexToOwnedBytes(allocator, block.evm_input),
        };
    }

    return .{
        .payload = payload,
        .blocks = blocks,
        .extension_a = envelope.extension_a,
        .extension_b = envelope.extension_b,
    };
}

pub fn buildPayload(allocator: std.mem.Allocator, scenario: Scenario) !PayloadFixture {
    rollup_guest_allocator.init(allocator);

    const blocks = try allocator.alloc(BlockFixture, scenario.blocks.len);
    const witnesses = try allocator.alloc([]const u8, scenario.blocks.len);
    for (scenario.blocks, 0..) |block, i| {
        blocks[i] = try buildBlockFixture(allocator, try decodeBlockSpec(allocator, block));
        witnesses[i] = blocks[i].execution_witness;
    }

    return .{
        .payload = try buildGuestPayload(allocator, witnesses, scenario.extension_a, scenario.extension_b),
        .blocks = blocks,
        .extension_a = scenario.extension_a,
        .extension_b = scenario.extension_b,
    };
}

pub fn writePayloadFixtureJson(writer: anytype, scenario: Scenario, bundle: PayloadFixture) !void {
    try writer.writeAll("{\n  \"description\": ");
    try writeJsonString(writer, scenario.description);
    try writer.writeAll(",\n  \"payload\": \"0x");
    try writeHex(writer, bundle.payload);
    try writer.writeAll(
        \\",
        \\  "blocks": [
        \\
    );

    for (bundle.blocks, 0..) |block, i| {
        if (i > 0) try writer.writeAll(",\n");
        try writer.print(
            \\    {{
            \\      "kind": "{s}",
            \\      "block_number": {d},
            \\      "transaction_count": {d},
            \\      "node_count": {d},
            \\      "code_count": {d},
            \\      "key_count": {d},
            \\      "header_count": {d},
            \\      "pre_state_root": "0x
        , .{
            witnessKindName(block.kind),
            block.block_number,
            block.transaction_count,
            block.node_count,
            block.code_count,
            block.key_count,
            block.header_count,
        });
        try writeHex(writer, &block.pre_state_root);
        try writer.writeAll(
            \\",
            \\      "post_state_root": "0x
        );
        try writeHex(writer, &block.post_state_root);
        try writer.writeAll(
            \\",
            \\      "receipts_root": "0x
        );
        try writeHex(writer, &block.receipts_root);
        try writer.print(
            \\",
            \\      "gas_used": {d},
            \\      "raw_transaction": "0x
        , .{block.gas_used});
        try writeHex(writer, block.raw_transaction);
        try writer.writeAll(
            \\",
            \\      "evm_input": "0x
        );
        try writeHex(writer, block.evm_input);
        try writer.writeAll(
            \\"
            \\    }
        );
    }

    try writer.writeAll(
        \\
        \\  ]
        \\}
        \\
    );
}

fn buildBlockFixture(allocator: std.mem.Allocator, request: DecodedBlockSpec) !BlockFixture {
    const pre_state = try buildSingleAccountPreState(allocator, request.sender);

    const parent_header = try blockHeaderRlp(allocator, .{
        .state_root = pre_state.root,
        .number = request.block_number - 1,
        .timestamp = request.timestamp - 1,
    });
    const parent_hash = mpt.keccak256(parent_header);

    const tx_key = try rlpU64(allocator, 0);
    const transactions_root = try singleLeafTrieRoot(allocator, tx_key, request.raw_transaction);

    const placeholder_header = try blockHeaderRlp(allocator, .{
        .parent_hash = parent_hash,
        .state_root = ZERO_HASH,
        .transactions_root = transactions_root,
        .receipts_root = ZERO_HASH,
        .number = request.block_number,
        .timestamp = request.timestamp,
    });
    const placeholder_block = try blockRlp(allocator, placeholder_header, request.raw_transaction);
    const parsed = try stateless_rlp.json.parseBlockFromRlp(allocator, placeholder_block);

    var node_index = try mpt.buildNodeIndex(allocator, pre_state.nodes);
    defer node_index.deinit();

    var block_hashes = std.ArrayListUnmanaged(executor.BlockHashEntry).empty;
    defer block_hashes.deinit(allocator);
    try block_hashes.append(allocator, .{
        .number = request.block_number - 1,
        .hash = parent_hash,
    });

    const proof = try executor.executeBlock(
        allocator,
        pre_state.root,
        try singlePreAlloc(allocator, request.sender),
        &node_index,
        parsed.header,
        parsed.transactions,
        parsed.withdrawals,
        block_hashes.items,
        null,
    );

    const gas_used = if (proof.receipts.len == 0)
        0
    else
        proof.receipts[proof.receipts.len - 1].cumulative_gas_used;
    const logs_bloom = if (proof.receipts.len == 0)
        ZERO_BLOOM
    else
        proof.receipts[proof.receipts.len - 1].logs_bloom;

    const final_header = try blockHeaderRlp(allocator, .{
        .parent_hash = parent_hash,
        .state_root = proof.post_state_root,
        .transactions_root = transactions_root,
        .receipts_root = proof.receipts_root,
        .logs_bloom = logs_bloom,
        .number = request.block_number,
        .gas_used = gas_used,
        .timestamp = request.timestamp,
    });
    const final_block = try blockRlp(allocator, final_header, request.raw_transaction);

    const keys = try allocator.alloc([]const u8, 1);
    keys[0] = try allocator.dupe(u8, &request.sender);

    const headers = try allocator.alloc([]const u8, 1);
    headers[0] = parent_header;

    return .{
        .execution_witness = try executionWitnessBytes(
            allocator,
            final_block,
            pre_state.nodes,
            &.{},
            keys,
            headers,
        ),
        .kind = request.kind,
        .block_number = request.block_number,
        .transaction_count = 1,
        .node_count = pre_state.nodes.len,
        .code_count = 0,
        .key_count = keys.len,
        .header_count = headers.len,
        .pre_state_root = pre_state.root,
        .post_state_root = proof.post_state_root,
        .receipts_root = proof.receipts_root,
        .gas_used = gas_used,
        .raw_transaction = request.raw_transaction,
        .evm_input = request.evm_input,
    };
}

fn decodeBlockSpec(allocator: std.mem.Allocator, spec: BlockSpec) !DecodedBlockSpec {
    return .{
        .kind = spec.kind,
        .block_number = spec.block_number,
        .timestamp = spec.timestamp,
        .sender = try hexToArray(20, spec.sender),
        .raw_transaction = try hexToOwnedBytes(allocator, spec.raw_transaction),
        .evm_input = try hexToOwnedBytes(allocator, spec.evm_input),
    };
}

fn buildSingleAccountPreState(allocator: std.mem.Allocator, address: [20]u8) !PreState {
    const account_rlp = try accountRlp(allocator, 0, PRESTATE_BALANCE, EMPTY_TRIE_HASH, KECCAK_EMPTY);
    const address_hash = mpt.keccak256(&address);
    const nibbles = try bytesToNibbles(allocator, &address_hash);
    const leaf = try trieLeaf(allocator, nibbles, account_rlp);

    const nodes = try allocator.alloc([]const u8, 1);
    nodes[0] = leaf;
    return .{
        .root = mpt.keccak256(leaf),
        .nodes = nodes,
    };
}

fn singlePreAlloc(
    allocator: std.mem.Allocator,
    address: [20]u8,
) !std.AutoHashMapUnmanaged([20]u8, executor.AllocAccount) {
    var pre_alloc = std.AutoHashMapUnmanaged([20]u8, executor.AllocAccount){};
    try pre_alloc.put(allocator, address, .{
        .balance = PRESTATE_BALANCE,
        .nonce = 0,
        .code = &.{},
        .pre_storage_root = EMPTY_TRIE_HASH,
    });
    return pre_alloc;
}

fn accountRlp(
    allocator: std.mem.Allocator,
    nonce: u64,
    balance: u256,
    storage_root: [32]u8,
    code_hash: [32]u8,
) ![]const u8 {
    const parts = [_][]const u8{
        try rlpU64(allocator, nonce),
        try rlpU256(allocator, balance),
        try rlpBytes(allocator, &storage_root),
        try rlpBytes(allocator, &code_hash),
    };
    return rlpList(allocator, &parts);
}

fn singleLeafTrieRoot(allocator: std.mem.Allocator, key: []const u8, value: []const u8) ![32]u8 {
    const nibbles = try bytesToNibbles(allocator, key);
    const leaf = try trieLeaf(allocator, nibbles, value);
    return mpt.keccak256(leaf);
}

fn trieLeaf(allocator: std.mem.Allocator, nibbles: []const u8, value: []const u8) ![]const u8 {
    const path = try hpEncode(allocator, nibbles, true);
    const parts = [_][]const u8{
        try rlpBytes(allocator, path),
        try rlpBytes(allocator, value),
    };
    return rlpList(allocator, &parts);
}

fn blockHeaderRlp(allocator: std.mem.Allocator, spec: HeaderSpec) ![]const u8 {
    const parts = [_][]const u8{
        try rlpBytes(allocator, &spec.parent_hash),
        try rlpBytes(allocator, &KECCAK_EMPTY_LIST),
        try rlpBytes(allocator, &ZERO_ADDRESS),
        try rlpBytes(allocator, &spec.state_root),
        try rlpBytes(allocator, &spec.transactions_root),
        try rlpBytes(allocator, &spec.receipts_root),
        try rlpBytes(allocator, &spec.logs_bloom),
        try rlpU256(allocator, 0),
        try rlpU64(allocator, spec.number),
        try rlpU64(allocator, spec.gas_limit),
        try rlpU64(allocator, spec.gas_used),
        try rlpU64(allocator, spec.timestamp),
        try rlpBytes(allocator, &.{}),
        try rlpBytes(allocator, &ZERO_HASH),
        try rlpBytes(allocator, &ZERO_NONCE),
        try rlpU64(allocator, spec.base_fee_per_gas),
    };
    return rlpList(allocator, &parts);
}

fn blockRlp(allocator: std.mem.Allocator, header: []const u8, raw_tx: []const u8) ![]const u8 {
    const txs = [_][]const u8{raw_tx};
    const body = [_][]const u8{
        header,
        try rlpList(allocator, &txs),
        try rlpList(allocator, &.{}),
    };
    return rlpList(allocator, &body);
}

fn executionWitnessBytes(
    allocator: std.mem.Allocator,
    block: []const u8,
    nodes: []const []const u8,
    codes: []const []const u8,
    keys: []const []const u8,
    headers: []const []const u8,
) ![]const u8 {
    var out = std.ArrayListUnmanaged(u8).empty;
    try appendLenPrefixedBytes(&out, allocator, block);
    try appendSliceArray(&out, allocator, nodes);
    try appendSliceArray(&out, allocator, codes);
    try appendSliceArray(&out, allocator, keys);
    try appendSliceArray(&out, allocator, headers);
    return out.toOwnedSlice(allocator);
}

fn buildGuestPayload(
    allocator: std.mem.Allocator,
    witnesses: []const []const u8,
    extension_a: u64,
    extension_b: u64,
) ![]const u8 {
    var out = std.ArrayListUnmanaged(u8).empty;
    try appendU64(&out, allocator, witnesses.len);
    for (witnesses) |witness| {
        try appendLenPrefixedBytes(&out, allocator, witness);
    }
    try appendU64(&out, allocator, extension_a);
    try appendU64(&out, allocator, extension_b);
    return out.toOwnedSlice(allocator);
}

fn appendSliceArray(
    out: *std.ArrayListUnmanaged(u8),
    allocator: std.mem.Allocator,
    items: []const []const u8,
) !void {
    try appendU64(out, allocator, items.len);
    for (items) |item| {
        try appendLenPrefixedBytes(out, allocator, item);
    }
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

fn parsePayloadEnvelope(allocator: std.mem.Allocator, payload: []const u8) !PayloadEnvelope {
    var pos: usize = 0;
    const count: usize = @intCast(try readU64(payload, &pos));
    const witnesses = try allocator.alloc([]const u8, count);
    for (witnesses) |*witness| {
        witness.* = try readLenPrefixedBytes(payload, &pos);
    }
    const extension_a = try readU64(payload, &pos);
    const extension_b = try readU64(payload, &pos);
    if (pos != payload.len) return error.InvalidFixture;

    return .{
        .witnesses = witnesses,
        .extension_a = extension_a,
        .extension_b = extension_b,
    };
}

fn readLenPrefixedBytes(bytes: []const u8, pos: *usize) ![]const u8 {
    const len: usize = @intCast(try readU64(bytes, pos));
    if (pos.* > bytes.len or len > bytes.len - pos.*) return error.UnexpectedEndOfInput;
    const out = bytes[pos.*..][0..len];
    pos.* += len;
    return out;
}

fn readU64(bytes: []const u8, pos: *usize) !u64 {
    if (pos.* > bytes.len or 8 > bytes.len - pos.*) return error.UnexpectedEndOfInput;
    const out = std.mem.readInt(u64, bytes[pos.*..][0..8], .big);
    pos.* += 8;
    return out;
}

fn rlpBytes(allocator: std.mem.Allocator, data: []const u8) ![]const u8 {
    if (data.len == 1 and data[0] < 0x80) return allocator.dupe(u8, data);
    if (data.len == 0) return allocator.dupe(u8, &.{0x80});
    if (data.len <= 55) {
        const out = try allocator.alloc(u8, data.len + 1);
        out[0] = @intCast(0x80 + data.len);
        @memcpy(out[1..], data);
        return out;
    }

    var len_buf: [8]u8 = undefined;
    const len_bytes = bigEndianLength(data.len, &len_buf);
    const out = try allocator.alloc(u8, 1 + len_bytes.len + data.len);
    out[0] = @intCast(0xb7 + len_bytes.len);
    @memcpy(out[1..][0..len_bytes.len], len_bytes);
    @memcpy(out[1 + len_bytes.len ..], data);
    return out;
}

fn rlpList(allocator: std.mem.Allocator, items: []const []const u8) ![]const u8 {
    var payload_len: usize = 0;
    for (items) |item| payload_len += item.len;

    if (payload_len <= 55) {
        const out = try allocator.alloc(u8, payload_len + 1);
        out[0] = @intCast(0xc0 + payload_len);
        var pos: usize = 1;
        for (items) |item| {
            @memcpy(out[pos..][0..item.len], item);
            pos += item.len;
        }
        return out;
    }

    var len_buf: [8]u8 = undefined;
    const len_bytes = bigEndianLength(payload_len, &len_buf);
    const out = try allocator.alloc(u8, 1 + len_bytes.len + payload_len);
    out[0] = @intCast(0xf7 + len_bytes.len);
    @memcpy(out[1..][0..len_bytes.len], len_bytes);
    var pos: usize = 1 + len_bytes.len;
    for (items) |item| {
        @memcpy(out[pos..][0..item.len], item);
        pos += item.len;
    }
    return out;
}

fn rlpU64(allocator: std.mem.Allocator, value: u64) ![]const u8 {
    if (value == 0) return rlpBytes(allocator, &.{});
    var bytes: [8]u8 = undefined;
    std.mem.writeInt(u64, &bytes, value, .big);
    return rlpBytes(allocator, trimLeadingZeroes(&bytes));
}

fn rlpU256(allocator: std.mem.Allocator, value: u256) ![]const u8 {
    if (value == 0) return rlpBytes(allocator, &.{});
    var bytes: [32]u8 = undefined;
    std.mem.writeInt(u256, &bytes, value, .big);
    return rlpBytes(allocator, trimLeadingZeroes(&bytes));
}

fn bigEndianLength(value: usize, buf: *[8]u8) []const u8 {
    var v = value;
    var len: usize = 0;
    var i: usize = 8;
    while (true) {
        i -= 1;
        buf[i] = @intCast(v & 0xff);
        len += 1;
        v >>= 8;
        if (v == 0) break;
    }
    return buf[i..8];
}

fn trimLeadingZeroes(bytes: []const u8) []const u8 {
    var i: usize = 0;
    while (i < bytes.len and bytes[i] == 0) : (i += 1) {}
    return bytes[i..];
}

fn bytesToNibbles(allocator: std.mem.Allocator, bytes: []const u8) ![]const u8 {
    const out = try allocator.alloc(u8, bytes.len * 2);
    for (bytes, 0..) |byte, i| {
        out[i * 2] = byte >> 4;
        out[i * 2 + 1] = byte & 0x0f;
    }
    return out;
}

fn hpEncode(allocator: std.mem.Allocator, path: []const u8, is_leaf: bool) ![]const u8 {
    const out = try allocator.alloc(u8, 1 + path.len / 2);
    const flag: u8 = if (is_leaf) 2 else 0;

    if (path.len % 2 != 0) {
        out[0] = ((flag | 1) << 4) | path[0];
        var i: usize = 1;
        var j: usize = 1;
        while (i + 1 < path.len) : ({
            i += 2;
            j += 1;
        }) {
            out[j] = (path[i] << 4) | path[i + 1];
        }
    } else {
        out[0] = flag << 4;
        var i: usize = 0;
        var j: usize = 1;
        while (i + 1 < path.len) : ({
            i += 2;
            j += 1;
        }) {
            out[j] = (path[i] << 4) | path[i + 1];
        }
    }

    return out;
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

fn parseWitnessKind(name: []const u8) !WitnessKind {
    if (std.mem.eql(u8, name, witnessKindName(.contract_creation))) return .contract_creation;
    if (std.mem.eql(u8, name, witnessKindName(.ecrecover_precompile))) return .ecrecover_precompile;
    return error.InvalidFixture;
}

fn witnessKindName(kind: WitnessKind) []const u8 {
    return switch (kind) {
        .contract_creation => "contract_creation",
        .ecrecover_precompile => "ecrecover_precompile",
    };
}

fn writeHex(writer: anytype, bytes: []const u8) !void {
    const alphabet = "0123456789abcdef";
    for (bytes) |byte| {
        const pair = [_]u8{
            alphabet[byte >> 4],
            alphabet[byte & 0x0f],
        };
        try writer.writeAll(&pair);
    }
}

fn writeJsonString(writer: anytype, value: []const u8) !void {
    try writer.writeAll("\"");
    for (value) |byte| {
        switch (byte) {
            '"' => try writer.writeAll("\\\""),
            '\\' => try writer.writeAll("\\\\"),
            '\n' => try writer.writeAll("\\n"),
            '\r' => try writer.writeAll("\\r"),
            '\t' => try writer.writeAll("\\t"),
            else => try writer.writeAll(&[_]u8{byte}),
        }
    }
    try writer.writeAll("\"");
}
