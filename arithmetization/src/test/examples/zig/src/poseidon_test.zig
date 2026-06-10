const wrappers = @import("wrappers");
const custom_std = wrappers.custom_std;
const poseidon = wrappers.poseidon;

export fn main() noreturn {
    // buf_* variables represent all-zeros inputs
    const buf_1   = [_]u8{0} ** 1;
    const buf_15  = [_]u8{0} ** 15;
    const buf_16  = [_]u8{0} ** 16;
    const buf_256 = [_]u8{0} ** 256;

    // extract pointers to the inputs
    const data_1:   [*c]const u8 = &buf_1;
    const data_15:  [*c]const u8 = &buf_15;
    const data_16:  [*c]const u8 = &buf_16;
    const data_256: [*c]const u8 = &buf_256;

    // pointer for writing output
    const output: [*c]poseidon.zkvm_poseison_hash = @ptrFromInt(0x08000000);

    _ = poseidon.zkvmPoseidon(data_1   , 1   , output); // "00".repeat(1)
    _ = poseidon.zkvmPoseidon(data_15  , 15  , output); // poseidon of "00".repeat(15)
    _ = poseidon.zkvmPoseidon(data_16  , 16  , output); // poseidon of "00".repeat(16)
    _ = poseidon.zkvmPoseidon(data_256 , 256 , output); // poseidon of "00".repeat(256)

    // inputs:
    // 1: 0000000000000000000000000000000000000000000000000000000000000000
    // 15: 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
    // 16: 000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
    // 256: 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
    //
    // hashes are discarded; for reference:
    // keccak( "00".repeat(1)   ) = <TBD>
    // keccak( "00".repeat(15)  ) = <TBD>
    // keccak( "00".repeat(16)  ) = <TBD>
    // keccak( "00".repeat(256) ) = <TBD>

    custom_std.exit(0);
}
