/// RISC-V guest secp256k1 boundary for transaction sender recovery.
///
/// Native tests intentionally keep ZEVM's C-backed `secp256k1_impl.zig`.
/// The freestanding guest imports this module so sender recovery has a single
/// injectable boundary for the future Linea interpreter accelerator.
pub const Secp256k1 = struct {
    pub fn sign(_: Secp256k1, _: [32]u8, _: [32]u8) ?struct { sig: [64]u8, recid: u8 } {
        @panic("Linea RISC-V secp256k1 accelerator is not implemented yet");
    }

    pub fn ecrecover(_: Secp256k1, _: [32]u8, _: [64]u8, _: u8) ?[20]u8 {
        @panic("Linea RISC-V secp256k1 accelerator is not implemented yet");
    }
};

pub fn getContext() ?Secp256k1 {
    return .{};
}
