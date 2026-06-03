const std = @import("std");

const options = @import("fixture_writer_options");
const fixtures = @import("evm_execution_fixtures");

pub fn main() !void {
    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();
    const allocator = arena.allocator();

    const scenario = fixtures.scenarios.contract_creation_then_ecrecover;
    const bundle = try fixtures.buildPayload(allocator, scenario);

    const io = std.Options.debug_io;
    const out_file = try std.Io.Dir.createFileAbsolute(io, options.output_path, .{});
    defer out_file.close(io);
    var buffer: [4096]u8 = undefined;
    var file_writer = out_file.writerStreaming(io, &buffer);
    const writer = &file_writer.interface;

    try fixtures.writePayloadFixtureJson(writer, scenario, bundle);
    try writer.flush();
}
