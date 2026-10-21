//! Per-chain EIP-6110/7002/7251 system-contract addresses for Linea networks.
//!
//! zesu's `input.SystemContracts.forChainId` only knows Ethereum mainnet/Sepolia. Linea genesis
//! configs override the deposit, withdrawal-request and consolidation-request contracts, and zesu
//! silently skips a system call to an undeployed address, so executing a Linea block against the
//! Ethereum defaults would yield an empty `requests_hash` instead of the one Besu produced. The
//! EIP-4788 beacon-roots and EIP-2935 history-storage contracts sit at their canonical addresses on
//! every Linea chain, so those keep zesu's defaults.
//!
//! Sources (keep in sync with them):
//!   - 59144 (Linea mainnet): docs/getting-started/linea-mainnet/geth/geth-genesis.json
//!   - 59141 (Linea Sepolia): docs/getting-started/linea-sepolia/geth/geth-genesis.json
//!   - 1337  (local stack):   docker/config/l2-genesis-initialization/genesis-besu.json.template,
//!                            plus the EIP-8282 builder overrides init.sh adds for the RISC-V stack
//! Any other chain id falls back to zesu's `forChainId`.

const std = @import("std");
const input = @import("zesu_input");
const primitives = @import("zesu_primitives");

pub const SystemContracts = input.SystemContracts;

pub const LINEA_MAINNET_CHAIN_ID: u64 = 59144;
pub const LINEA_SEPOLIA_CHAIN_ID: u64 = 59141;
pub const LOCAL_STACK_CHAIN_ID: u64 = 1337;

pub fn forChainId(chain_id: u64) SystemContracts {
    return switch (chain_id) {
        LINEA_MAINNET_CHAIN_ID => .{
            .deposit = addr("94d4a7449B968b839dcbc64A61F195CA300986ae"),
            .withdrawal_request = addr("66355689a9f067eeb9dc9d899E4192676988279C"),
            .consolidation_request = addr("F0e003F0dE2d583Ae28FA8cBF66aa096CdAce3ff"),
        },
        LINEA_SEPOLIA_CHAIN_ID => .{
            .deposit = addr("0bdae9550159eb73559e74d27fbc989641b24c8e"),
            .withdrawal_request = addr("0e97F666F1d99A60590F7Ce7fA5A5a7C754Ff714"),
            .consolidation_request = addr("94d94fac95855ed81294ed76f6b3da899e7063a8"),
        },
        LOCAL_STACK_CHAIN_ID => .{
            .deposit = addr("3a69eD7FA6f6CA1dB2649327c4A2E666130823bE"),
            .withdrawal_request = addr("3a69eD7FA6f6CA1dB2649327c4A2E666130823bE"),
            .consolidation_request = addr("9AFC00F4CEaadb97A3822FCA2225FaD78839FBde"),
            // Set by the RISC-V local stack's Amsterdam genesis (l2-genesis-initialization/init.sh);
            // unused before Amsterdam, so harmless for the Osaka zkEVM stack on the same chain id.
            .builder_deposit_request = addr("0000000000000000000000000000000000009999"),
            .builder_exit_request = addr("0000000000000000000000000000000000009999"),
        },
        else => SystemContracts.forChainId(chain_id),
    };
}

fn addr(comptime hex: *const [40]u8) primitives.Address {
    return comptime blk: {
        var out: primitives.Address = undefined;
        _ = std.fmt.hexToBytes(&out, hex) catch unreachable;
        break :blk out;
    };
}
