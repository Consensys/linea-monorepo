const ext = @import("field/koalabear_ext.zig");

pub const DQLayout = struct {
    sizes: []const u32 = &.{},
    eval_points: []const []const ext.Ext = &.{},
    column_names: []const []const []const []const u8 = &.{},
    column_keys: []const []const []const []const u8 = &.{},
    air_chunks: []const []const []const u8 = &.{},
};
