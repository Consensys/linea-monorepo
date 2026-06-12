const ext = @import("../field/koalabear_ext.zig");
const protocol = @import("../protocol/root.zig");
const types = @import("types.zig");
const verify = @import("verify.zig");

pub const Error = error{
    BadDimensions,
    CoinIndexOutOfBounds,
};

pub const CoinIndexMap = struct {
    gamma: ?usize = null,
    fold_alphas: []const usize,
    query_coins: []const usize,
};

pub fn resolve(
    comptime map: CoinIndexMap,
    all_coins: []const protocol.Coin,
    params: types.Params,
    fold_alpha_storage: []ext.Ext,
    query_position_storage: []u32,
) Error!verify.FriChallenges {
    if (fold_alpha_storage.len < map.fold_alphas.len) return Error.BadDimensions;
    if (query_position_storage.len < map.query_coins.len) return Error.BadDimensions;
    if (map.fold_alphas.len != params.num_rounds) return Error.BadDimensions;
    if (map.query_coins.len != params.num_queries) return Error.BadDimensions;

    const gamma = if (map.gamma) |index| try coinAt(all_coins, index) else null;

    inline for (map.fold_alphas, 0..) |coin_index, i| {
        fold_alpha_storage[i] = try coinAt(all_coins, coin_index);
    }

    inline for (map.query_coins, 0..) |coin_index, i| {
        query_position_storage[i] = queryIndex(try coinAt(all_coins, coin_index), params.n / 2);
    }

    return .{
        .gamma = gamma,
        .fold_alphas = fold_alpha_storage[0..map.fold_alphas.len],
        .query_positions = query_position_storage[0..map.query_coins.len],
    };
}

fn coinAt(all_coins: []const protocol.Coin, index: usize) Error!ext.Ext {
    if (index >= all_coins.len) return Error.CoinIndexOutOfBounds;
    return all_coins[index];
}

pub fn queryIndex(coin: protocol.Coin, modulus: u32) u32 {
    if (modulus == 0) return 0;
    const wide = (@as(u64, coin.B0.a0.value) << 31) ^ @as(u64, coin.B0.a1.value);
    return @intCast(wide % modulus);
}
