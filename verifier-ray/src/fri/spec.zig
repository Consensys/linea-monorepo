const std = @import("std");
const layout_types = @import("../layout.zig");

pub const Slot = layout_types.Slot;

/// NamedSlot pairs a column or air-chunk name with its tree slot. Used in
/// comptime Layout arrays in place of a runtime StringHashMap.
pub const NamedSlot = struct {
    name: []const u8,
    slot: Slot,
};

/// ColRef identifies a column by its prover-ray source name and protocol key.
pub const ColRef = struct {
    name: []const u8,
    key: []const u8,
};

/// DQLevel holds the DEEP-quotient structure for one domain size. Evaluation
/// points are encoded as shift exponents: the actual evaluation point for
/// shift k is ω_N^k · ζ, where ζ is the out-of-domain challenge derived at
/// transcript-replay time.
pub const DQLevel = struct {
    size: u32,
    shifts: []const u32,
    col_groups: []const []const ColRef,
    air_chunks: []const []const u8,
};

/// DQLayout holds the DEEP-quotient structure for all distinct domain sizes.
pub const DQLayout = struct {
    levels: []const DQLevel,
};

/// Layout holds the program-determined tree layout as comptime-consumable
/// arrays. It mirrors the runtime layout.Layout but uses []const NamedSlot
/// instead of StringHashMap so the value can be a comptime constant.
pub const Layout = struct {
    num_trees: u32,
    setup_begin: u32,
    setup_end: u32,
    trace_begin: []const u32,
    trace_end: []const u32,
    air_begin: u32,
    air_end: u32,
    tree_size: []const u32,
    col_slots: []const NamedSlot,
    air_chunk_slots: []const NamedSlot,

    /// colSlot returns the Slot for the column with the given name. It is
    /// evaluated at compile time; an unknown name is a compile error.
    pub fn colSlot(comptime self: Layout, comptime name: []const u8) Slot {
        inline for (self.col_slots) |ns| {
            if (std.mem.eql(u8, ns.name, name)) return ns.slot;
        }
        @compileError("unknown column name: " ++ name);
    }

    /// airChunkSlot returns the Slot for the air chunk with the given name.
    /// It is evaluated at compile time; an unknown name is a compile error.
    pub fn airChunkSlot(comptime self: Layout, comptime name: []const u8) Slot {
        inline for (self.air_chunk_slots) |ns| {
            if (std.mem.eql(u8, ns.name, name)) return ns.slot;
        }
        @compileError("unknown air chunk name: " ++ name);
    }
};
