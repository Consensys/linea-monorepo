const std = @import("std");
const verifier_ray = @import("verifier_ray");

const verifier = verifier_ray.verifier;
const merkle = verifier_ray.crypto.merkle;
const fri = verifier_ray.query.fri;
const pcs = verifier_ray.query.pcs;

const slice_ptr_offset = 0;
const slice_len_offset = @sizeOf(usize);

fn putSlice(image: []u8, header: usize, pointer: usize, len: usize) void {
    std.mem.writeInt(u64, image[header + slice_ptr_offset ..][0..8], @intCast(pointer), .little);
    std.mem.writeInt(u64, image[header + slice_len_offset ..][0..8], @intCast(len), .little);
}

fn expectPointer(image: []const u8, header: usize, expected: usize) !void {
    try std.testing.expectEqual(
        @as(u64, @intCast(expected)),
        std.mem.readInt(u64, image[header + slice_ptr_offset ..][0..8], .little),
    );
}

test "image relocation rebases every Merkle cap slice" {
    const encoded_base: usize = 0x08800000;

    const opening = @offsetOf(verifier.Proof, "pcs_opening") +
        @offsetOf(verifier.PcsOpening, "proof");
    const input_caps_header = opening + @offsetOf(pcs.OpeningProof, "input_caps");
    const fri_proof = opening + @offsetOf(pcs.OpeningProof, "fri_proof");
    const round_caps_header = fri_proof + @offsetOf(fri.Proof, "round_caps");

    const input_caps = 144;
    const input_cap_0_nodes = 208;
    const input_cap_0_tables = 240;
    const input_cap_1_nodes = 288;
    const input_cap_1_tables = 320;
    const input_cap_0_table_0_rows = 344;
    const input_cap_0_table_1_rows = 408;
    const input_cap_1_table_0_rows = 440;
    const row_payloads = [_][2]usize{
        .{ 472, 476 },
        .{ 500, 504 },
        .{ 528, 532 },
        .{ 556, 560 },
    };
    const round_caps = 584;
    const round_cap_0_nodes = 648;
    const round_cap_0_aux = 680;
    const round_cap_1_nodes = 716;

    var image: [760]u8 = @splat(0);
    const mapped_base = @intFromPtr(&image);

    putSlice(&image, input_caps_header, encoded_base + input_caps, 2);

    const input_cap_0 = input_caps;
    putSlice(&image, input_cap_0 + @offsetOf(pcs.InputCap, "nodes"), encoded_base + input_cap_0_nodes, 1);
    putSlice(&image, input_cap_0 + @offsetOf(pcs.InputCap, "tables"), encoded_base + input_cap_0_tables, 2);

    const input_cap_1 = input_caps + @sizeOf(pcs.InputCap);
    putSlice(&image, input_cap_1 + @offsetOf(pcs.InputCap, "nodes"), encoded_base + input_cap_1_nodes, 1);
    putSlice(&image, input_cap_1 + @offsetOf(pcs.InputCap, "tables"), encoded_base + input_cap_1_tables, 1);

    const input_cap_0_table_0 = input_cap_0_tables;
    const input_cap_0_table_1 = input_cap_0_tables + @sizeOf(pcs.InputCapTable);
    const input_cap_1_table_0 = input_cap_1_tables;
    putSlice(
        &image,
        input_cap_0_table_0 + @offsetOf(pcs.InputCapTable, "rows"),
        encoded_base + input_cap_0_table_0_rows,
        2,
    );
    putSlice(
        &image,
        input_cap_0_table_1 + @offsetOf(pcs.InputCapTable, "rows"),
        encoded_base + input_cap_0_table_1_rows,
        1,
    );
    putSlice(
        &image,
        input_cap_1_table_0 + @offsetOf(pcs.InputCapTable, "rows"),
        encoded_base + input_cap_1_table_0_rows,
        1,
    );

    const row_offsets = [_]usize{
        input_cap_0_table_0_rows,
        input_cap_0_table_0_rows + @sizeOf(merkle.RowOpening),
        input_cap_0_table_1_rows,
        input_cap_1_table_0_rows,
    };
    for (row_offsets, row_payloads) |row, payloads| {
        putSlice(&image, row + @offsetOf(merkle.RowOpening, "base"), encoded_base + payloads[0], 1);
        putSlice(&image, row + @offsetOf(merkle.RowOpening, "ext"), encoded_base + payloads[1], 1);
    }

    putSlice(&image, round_caps_header, encoded_base + round_caps, 2);
    const round_cap_0 = round_caps;
    putSlice(&image, round_cap_0 + @offsetOf(merkle.MerkleCap, "nodes"), encoded_base + round_cap_0_nodes, 1);
    putSlice(&image, round_cap_0 + @offsetOf(merkle.MerkleCap, "aux"), encoded_base + round_cap_0_aux, 1);

    const round_cap_1 = round_caps + @sizeOf(merkle.MerkleCap);
    putSlice(&image, round_cap_1 + @offsetOf(merkle.MerkleCap, "nodes"), encoded_base + round_cap_1_nodes, 1);
    // Go encodes empty slices as a non-null pointer to the image base.
    putSlice(&image, round_cap_1 + @offsetOf(merkle.MerkleCap, "aux"), encoded_base, 0);

    verifier_ray.image_relocation.rebase(&image, image.len, encoded_base, mapped_base);

    try expectPointer(&image, input_caps_header, mapped_base + input_caps);
    try expectPointer(&image, input_cap_0 + @offsetOf(pcs.InputCap, "nodes"), mapped_base + input_cap_0_nodes);
    try expectPointer(&image, input_cap_0 + @offsetOf(pcs.InputCap, "tables"), mapped_base + input_cap_0_tables);
    try expectPointer(&image, input_cap_1 + @offsetOf(pcs.InputCap, "nodes"), mapped_base + input_cap_1_nodes);
    try expectPointer(&image, input_cap_1 + @offsetOf(pcs.InputCap, "tables"), mapped_base + input_cap_1_tables);
    try expectPointer(
        &image,
        input_cap_0_table_0 + @offsetOf(pcs.InputCapTable, "rows"),
        mapped_base + input_cap_0_table_0_rows,
    );
    try expectPointer(
        &image,
        input_cap_0_table_1 + @offsetOf(pcs.InputCapTable, "rows"),
        mapped_base + input_cap_0_table_1_rows,
    );
    try expectPointer(
        &image,
        input_cap_1_table_0 + @offsetOf(pcs.InputCapTable, "rows"),
        mapped_base + input_cap_1_table_0_rows,
    );
    for (row_offsets, row_payloads) |row, payloads| {
        try expectPointer(&image, row + @offsetOf(merkle.RowOpening, "base"), mapped_base + payloads[0]);
        try expectPointer(&image, row + @offsetOf(merkle.RowOpening, "ext"), mapped_base + payloads[1]);
    }

    try expectPointer(&image, round_caps_header, mapped_base + round_caps);
    try expectPointer(&image, round_cap_0 + @offsetOf(merkle.MerkleCap, "nodes"), mapped_base + round_cap_0_nodes);
    try expectPointer(&image, round_cap_0 + @offsetOf(merkle.MerkleCap, "aux"), mapped_base + round_cap_0_aux);
    try expectPointer(&image, round_cap_1 + @offsetOf(merkle.MerkleCap, "nodes"), mapped_base + round_cap_1_nodes);
    try expectPointer(&image, round_cap_1 + @offsetOf(merkle.MerkleCap, "aux"), mapped_base);
}

test "pair members relocate relative to their own encoded and mapped bases" {
    const guest_base: usize = 0x08800000;
    const mapped_base: usize = 0x50000000;
    const member_offsets = [_]usize{ 16, 0x1238 };
    const payload_offsets = [_]usize{ 112, 0x4560 };

    for (member_offsets, payload_offsets) |member_offset, payload_offset| {
        const encoded_member_base = guest_base + member_offset;
        const mapped_member_base = mapped_base + member_offset;
        const encoded_pointer = encoded_member_base + payload_offset;

        try std.testing.expectEqual(
            payload_offset,
            try verifier_ray.image_relocation.imageOffset(encoded_pointer, encoded_member_base),
        );
        try std.testing.expectEqual(
            mapped_member_base + payload_offset,
            try verifier_ray.image_relocation.relocatePointer(encoded_pointer, encoded_member_base, mapped_member_base),
        );
    }
}

test "single image relocation preserves its existing base-relative behavior" {
    const guest_base: usize = 0x08800000;
    const mapped_base: usize = 0x70000000;
    const payload_offset: usize = 0x3210;

    try std.testing.expectEqual(
        mapped_base + payload_offset,
        try verifier_ray.image_relocation.relocatePointer(guest_base + payload_offset, guest_base, mapped_base),
    );
}

test "relocation rejects pointers outside the encoded image" {
    try std.testing.expectError(
        error.PointerBeforeImage,
        verifier_ray.image_relocation.relocatePointer(0x1000, 0x2000, 0x3000),
    );
    try std.testing.expectError(
        error.AddressOverflow,
        verifier_ray.image_relocation.relocatePointer(0x2001, 0x2000, std.math.maxInt(usize)),
    );
}

test "pair relocation rebases the header and each sub-image against its own start" {
    const relocation = verifier_ray.image_relocation;
    const encoded_base: usize = 0x08800000;
    const hdr = relocation.pair_header_len;
    // Each sub-image is a zeroed VerifyInput (every other slice null) whose
    // public_inputs points just past it.
    const payload = @sizeOf(verifier.VerifyInput);
    const sub_len = payload + 32;
    const public_inputs_header = @offsetOf(verifier.VerifyInput, "public_inputs");

    var image: [hdr + 2 * sub_len]u8 align(8) = @splat(0);
    const mapped_base = @intFromPtr(&image);
    const subs = [_]usize{ hdr, hdr + sub_len };
    for (subs, 0..) |sub, i| {
        std.mem.writeInt(u64, image[i * 8 ..][0..8], encoded_base + sub, .little);
        putSlice(image[sub..], public_inputs_header, encoded_base + sub + payload, 1);
    }

    try relocation.rebasePair(&image, image.len, encoded_base, mapped_base);

    for (subs, 0..) |sub, i| {
        try std.testing.expectEqual(
            @as(u64, mapped_base + sub),
            std.mem.readInt(u64, image[i * 8 ..][0..8], .little),
        );
        try expectPointer(image[sub..], public_inputs_header, mapped_base + sub + payload);
    }
}

test "pair relocation rejects header pointers outside the image" {
    const relocation = verifier_ray.image_relocation;
    const encoded_base: usize = 0x08800000;
    var image: [64]u8 align(8) = @splat(0);

    std.mem.writeInt(u64, image[0..8], encoded_base + 0x1000, .little);
    std.mem.writeInt(u64, image[8..16], encoded_base + 16, .little);
    try std.testing.expectError(
        error.PointerPastImage,
        relocation.rebasePair(&image, image.len, encoded_base, @intFromPtr(&image)),
    );

    std.mem.writeInt(u64, image[0..8], encoded_base - 8, .little);
    try std.testing.expectError(
        error.PointerBeforeImage,
        relocation.rebasePair(&image, image.len, encoded_base, @intFromPtr(&image)),
    );
}
