/// Native local-test secp256k1 implementation.
///
/// Freestanding guests import `rollup_guest_secp256k1.zig` as the accelerator
/// boundary. Native tests import this wrapper so transaction sender recovery
/// uses ZEVM's real libsecp256k1-backed implementation.
const native = @import("zevm_native_secp256k1");

pub const Secp256k1 = native.Secp256k1;
pub const getContext = native.getContext;
