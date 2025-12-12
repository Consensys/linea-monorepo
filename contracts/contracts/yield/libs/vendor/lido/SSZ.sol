// Copied verbatim from Lido audited contracts - https://github.com/lidofinance/core/blob/7cae7a14192ff094fb0eb089433ac9f6fd70e3c6/contracts/common/lib/SSZ.sol

// SPDX-FileCopyrightText: 2025 Lido <info@lido.fi>
// SPDX-License-Identifier: GPL-3.0

/*
 SSZ library from CSM
 original: https://github.com/lidofinance/community-staking-module/blob/7071c2096983a7780a5f147963aaa5405c0badb1/src/lib/SSZ.sol
*/

// See contracts/COMPILERS.md
// solhint-disable-next-line lido/fixed-compiler-version
pragma solidity ^0.8.25;

import { BeaconBlockHeader, Validator, PendingPartialWithdrawal } from "./BeaconTypes.sol";
import { GIndex } from "./GIndex.sol";
import { Math256 } from "../../../../lib/Math256.sol";

library SSZ {
  error BranchHasMissingItem();
  error BranchHasExtraItem();
  error InvalidProof();
  error OutOfRange();

  uint256 constant MAX_PENDING_PARTIAL_WITHDRAWAL_DEPTH = 27;

  function zeroHash(uint256 i) internal pure returns (bytes32) {
      if (i == 0) return 0x0000000000000000000000000000000000000000000000000000000000000000;
      if (i == 1) return 0xf5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b;
      if (i == 2) return 0xdb56114e00fdd4c1f85c892bf35ac9a89289aaecb1ebd0a96cde606a748b5d71;
      if (i == 3) return 0xc78009fdf07fc56a11f122370658a353aaa542ed63e44c4bc15ff4cd105ab33c;
      if (i == 4) return 0x536d98837f2dd165a55d5eeae91485954472d56f246df256bf3cae19352a123c;
      if (i == 5) return 0x9efde052aa15429fae05bad4d0b1d7c64da64d03d7a1854a588c2cb8430c0d30;
      if (i == 6) return 0xd88ddfeed400a8755596b21942c1497e114c302e6118290f91e6772976041fa1;
      if (i == 7) return 0x87eb0ddba57e35f6d286673802a4af5975e22506c7cf4c64bb6be5ee11527f2c;
      if (i == 8) return 0x26846476fd5fc54a5d43385167c95144f2643f533cc85bb9d16b782f8d7db193;
      if (i == 9) return 0x506d86582d252405b840018792cad2bf1259f1ef5aa5f887e13cb2f0094f51e1;
      if (i == 10) return 0xffff0ad7e659772f9534c195c815efc4014ef1e1daed4404c06385d11192e92b;
      if (i == 11) return 0x6cf04127db05441cd833107a52be852868890e4317e6a02ab47683aa75964220;
      if (i == 12) return 0xb7d05f875f140027ef5118a2247bbb84ce8f2f0f1123623085daf7960c329f5f;
      if (i == 13) return 0xdf6af5f5bbdb6be9ef8aa618e4bf8073960867171e29676f8b284dea6a08a85e;
      if (i == 14) return 0xb58d900f5e182e3c50ef74969ea16c7726c549757cc23523c369587da7293784;
      if (i == 15) return 0xd49a7502ffcfb0340b1d7885688500ca308161a7f96b62df9d083b71fcc8f2bb;
      if (i == 16) return 0x8fe6b1689256c0d385f42f5bbe2027a22c1996e110ba97c171d3e5948de92beb;
      if (i == 17) return 0x8d0d63c39ebade8509e0ae3c9c3876fb5fa112be18f905ecacfecb92057603ab;
      if (i == 18) return 0x95eec8b2e541cad4e91de38385f2e046619f54496c2382cb6cacd5b98c26f5a4;
      if (i == 19) return 0xf893e908917775b62bff23294dbbe3a1cd8e6cc1c35b4801887b646a6f81f17f;
      if (i == 20) return 0xcddba7b592e3133393c16194fac7431abf2f5485ed711db282183c819e08ebaa;
      if (i == 21) return 0x8a8d7fe3af8caa085a7639a832001457dfb9128a8061142ad0335629ff23ff9c;
      if (i == 22) return 0xfeb3c337d7a51a6fbf00b9e34c52e1c9195c969bd4e7a0bfd51d5c5bed9c1167;
      if (i == 23) return 0xe71f0aa83cc32edfbefa9f4d3e0174ca85182eec9f3a09f6a6c0df6377a510d7;
      if (i == 24) return 0x31206fa80a50bb6abe29085058f16212212a60eec8f049fecb92d8c8e0a84bc0;
      if (i == 25) return 0x21352bfecbeddde993839f614c3dac0a3ee37543f9b412b16199dc158e23b544;
      if (i == 26) return 0x619e312724bb6d7c3153ed9de791d764a366b389af13c58bf8a8d90481a46765;
      if (i == 27) return 0x7cdd2986268250628d0c10e385c58c6191e6fbe05191bcc04f133f2cea72c1c4;
      revert OutOfRange();
  }

  function hashTreeRoot(BeaconBlockHeader memory header) internal view returns (bytes32 root) {
    bytes32[8] memory nodes = [
      toLittleEndian(header.slot),
      toLittleEndian(header.proposerIndex),
      header.parentRoot,
      header.stateRoot,
      header.bodyRoot,
      bytes32(0),
      bytes32(0),
      bytes32(0)
    ];

    /// @solidity memory-safe-assembly
    assembly {
      // Count of nodes to hash
      let count := 8

      // Loop over levels
      // prettier-ignore
      for { } 1 { } {
                // Loop over nodes at the given depth

                // Initialize `offset` to the offset of `proof` elements in memory.
                let target := nodes
                let source := nodes
                let end := add(source, shl(5, count))

                // prettier-ignore
                for { } 1 { } {
                    // Read next two hashes to hash
                    mcopy(0x00, source, 0x40)

                    // Call sha256 precompile
                    let result := staticcall(
                        gas(),
                        0x02,
                        0x00,
                        0x40,
                        0x00,
                        0x20
                    )

                    if iszero(result) {
                        // Precompiles returns no data on OutOfGas error.
                        revert(0, 0)
                    }

                    // Store the resulting hash at the target location
                    mstore(target, mload(0x00))

                    // Advance the pointers
                    target := add(target, 0x20)
                    source := add(source, 0x40)

                    if iszero(lt(source, end)) {
                        break
                    }
                }

                count := shr(1, count)
                if eq(count, 1) {
                    root := mload(0x00)
                    break
                }
            }
    }
  }

  function hashTreeRoot(Validator memory validator) internal view returns (bytes32 root) {
    bytes32 pubkeyRoot;

    assembly {
      // Dynamic data types such as bytes are stored at the specified offset.
      let offset := mload(validator)
      // Copy the pubkey to the scratch space.
      mcopy(0x00, add(offset, 32), 48)
      // Clear the last 16 bytes.
      mcopy(48, 0x60, 16)
      // Call sha256 precompile.
      let result := staticcall(gas(), 0x02, 0x00, 0x40, 0x00, 0x20)

      if iszero(result) {
        // Precompiles returns no data on OutOfGas error.
        revert(0, 0)
      }

      pubkeyRoot := mload(0x00)
    }

    bytes32[8] memory nodes = [
      pubkeyRoot,
      validator.withdrawalCredentials,
      toLittleEndian(validator.effectiveBalance),
      toLittleEndian(validator.slashed),
      toLittleEndian(validator.activationEligibilityEpoch),
      toLittleEndian(validator.activationEpoch),
      toLittleEndian(validator.exitEpoch),
      toLittleEndian(validator.withdrawableEpoch)
    ];

    /// @solidity memory-safe-assembly
    assembly {
      // Count of nodes to hash
      let count := 8

      // Loop over levels
      // prettier-ignore
      for { } 1 { } {
                // Loop over nodes at the given depth

                // Initialize `offset` to the offset of `proof` elements in memory.
                let target := nodes
                let source := nodes
                let end := add(source, shl(5, count))

                // prettier-ignore
                for { } 1 { } {
                    // Read next two hashes to hash
                    mcopy(0x00, source, 0x40)

                    // Call sha256 precompile
                    let result := staticcall(
                        gas(),
                        0x02,
                        0x00,
                        0x40,
                        0x00,
                        0x20
                    )

                    if iszero(result) {
                        // Precompiles returns no data on OutOfGas error.
                        revert(0, 0)
                    }

                    // Store the resulting hash at the target location
                    mstore(target, mload(0x00))

                    // Advance the pointers
                    target := add(target, 0x20)
                    source := add(source, 0x40)

                    if iszero(lt(source, end)) {
                        break
                    }
                }

                count := shr(1, count)
                if eq(count, 1) {
                    root := mload(0x00)
                    break
                }
            }
    }
  }

  function hashTreeRoot(PendingPartialWithdrawal memory pendingPartialWithdrawal) internal view returns (bytes32 root) {
    bytes32[4] memory nodes = [
      toLittleEndian(pendingPartialWithdrawal.validatorIndex),
      toLittleEndian(pendingPartialWithdrawal.amount),
      toLittleEndian(pendingPartialWithdrawal.withdrawableEpoch),
      bytes32(0)
    ];

    /// @solidity memory-safe-assembly
    assembly {
      // Count of nodes to hash
      let count := 4

      // Loop over levels
      // prettier-ignore
      for { } 1 { } {
                // Loop over nodes at the given depth

                // Initialize `offset` to the offset of `proof` elements in memory.
                let target := nodes
                let source := nodes
                let end := add(source, shl(5, count))

                // prettier-ignore
                for { } 1 { } {
                    // Read next two hashes to hash
                    mcopy(0x00, source, 0x40)

                    // Call sha256 precompile
                    let result := staticcall(
                        gas(),
                        0x02,
                        0x00,
                        0x40,
                        0x00,
                        0x20
                    )

                    if iszero(result) {
                        // Precompiles returns no data on OutOfGas error.
                        revert(0, 0)
                    }

                    // Store the resulting hash at the target location
                    mstore(target, mload(0x00))

                    // Advance the pointers
                    target := add(target, 0x20)
                    source := add(source, 0x40)

                    if iszero(lt(source, end)) {
                        break
                    }
                }

                count := shr(1, count)
                if eq(count, 1) {
                    root := mload(0x00)
                    break
                }
            }
    }
  }

  // Solidity implementation of the merkleize_chunks spec in https://github.com/ethereum/consensus-specs/blob/5390b77256a9fd6c1ebe0c7e3f8a3da033476ddf/tests/core/pyspec/eth2spec/utils/merkle_minimal.py#L47-L91
  // Algorithm has space complexity of "MAX_PENDING_PARTIAL_WITHDRAWAL_DEPTH+1", despite input array being up to 2**27 items.
  function hashTreeRoot(PendingPartialWithdrawal[] calldata pendingPartialWithdrawal) internal view returns (bytes32 root) {
    uint256 count = pendingPartialWithdrawal.length;
    uint256 depth = count == 0 ? 0 : Math256.bitLength(count - 1);
    bytes32[MAX_PENDING_PARTIAL_WITHDRAWAL_DEPTH + 1] memory tmp;

    for (uint256 i = 0; i < count; i++) {
      mergeSSZChunk(tmp, depth, count, hashTreeRoot(pendingPartialWithdrawal[i]), i);
    }

    if (1 << depth != count) {
      mergeSSZChunk(tmp, depth, count, bytes32(0), count);
    }

    for (uint256 j = depth; j < MAX_PENDING_PARTIAL_WITHDRAWAL_DEPTH; j++) {
      tmp[j + 1] = sha256Pair(tmp[j], zeroHash(j));
    }

    // Do mix_in_length(content_root, actual_length)
    root = sha256Pair(tmp[MAX_PENDING_PARTIAL_WITHDRAWAL_DEPTH], toLittleEndian(count));
  }

  // Mutate `tmp` in-place
  function mergeSSZChunk(bytes32[28] memory tmp, uint256 depth, uint256 count, bytes32 chunk, uint256 chunkIndex) internal view {
      uint256 j = 0;
      bytes32 h = chunk;
      while (true) {
        if (chunkIndex & (1 << j) == 0) {
          if (chunkIndex == count && j < depth) {
            h = sha256Pair(h, zeroHash(j));
          } else {
            break;
          }
        } else {
          h = sha256Pair(tmp[j], h);
        }
        j++;
      }
      tmp[j] = h;
  }

  /// @notice Modified version of `verify` from Solady `MerkleProofLib` to support generalized indices and sha256 precompile.
  /// @dev Reverts if `leaf` doesn't exist in the Merkle tree with `root`, given `proof`.
  function verifyProof(bytes32[] calldata proof, bytes32 root, bytes32 leaf, GIndex gI) internal view {
    uint256 index = gI.index();

    /// @solidity memory-safe-assembly
    assembly {
      // Check if `proof` is empty.
      if iszero(proof.length) {
        // revert InvalidProof()
        mstore(0x00, 0x09bde339)
        revert(0x1c, 0x04)
      }
      // Left shift by 5 is equivalent to multiplying by 0x20.
      let end := add(proof.offset, shl(5, proof.length))
      // Initialize `offset` to the offset of `proof` in the calldata.
      let offset := proof.offset
      // Iterate over proof elements to compute root hash.
      // prettier-ignore
      for { } 1 { } {
                // Slot of `leaf` in scratch space.
                // If the condition is true: 0x20, otherwise: 0x00.
                let scratch := shl(5, and(index, 1))
                index := shr(1, index)
                if iszero(index) {
                    // revert BranchHasExtraItem()
                    mstore(0x00, 0x5849603f)
                    // 0x1c = 28 => offset in 32-byte word of a slot 0x00
                    revert(0x1c, 0x04)
                }
                // Store elements to hash contiguously in scratch space.
                // Scratch space is 64 bytes (0x00 - 0x3f) and both elements are 32 bytes.
                mstore(scratch, leaf)
                // load next proof from calldata to the scratch space at 0x00 or 0x20
                // xor() acts as if
                mstore(xor(scratch, 0x20), calldataload(offset))
                // Call sha256 precompile.
                let result := staticcall(
                    gas(),
                    0x02, // SHA-256 precompile
                    0x00, // input from scratch space from 0x00
                    0x40, // length is 2 leafs of 32 bytes each
                    0x00, // output back to scratch space at 0x00
                    0x20  // length of the output is 32 bytes
                )

                if iszero(result) {
                    // Precompile returns no data on OutOfGas error.
                    revert(0, 0)
                }

                // Reuse `leaf` to store the hash to reduce stack operations.
                leaf := mload(0x00)
                offset := add(offset, 0x20)
                if iszero(lt(offset, end)) {
                    break
                }
            }

      if iszero(eq(index, 1)) {
        // revert BranchHasMissingItem()
        mstore(0x00, 0x1b6661c3)
        revert(0x1c, 0x04)
      }

      if iszero(eq(leaf, root)) {
        // revert InvalidProof()
        mstore(0x00, 0x09bde339)
        revert(0x1c, 0x04)
      }
    }
  }

  // See https://github.com/succinctlabs/telepathy-contracts/blob/5aa4bb7/src/libraries/SimpleSerialize.sol#L17-L28
  function toLittleEndian(uint256 v) internal pure returns (bytes32) {
    v =
      ((v & 0xFF00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF00) >> 8) |
      ((v & 0x00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF00FF) << 8);
    v =
      ((v & 0xFFFF0000FFFF0000FFFF0000FFFF0000FFFF0000FFFF0000FFFF0000FFFF0000) >> 16) |
      ((v & 0x0000FFFF0000FFFF0000FFFF0000FFFF0000FFFF0000FFFF0000FFFF0000FFFF) << 16);
    v =
      ((v & 0xFFFFFFFF00000000FFFFFFFF00000000FFFFFFFF00000000FFFFFFFF00000000) >> 32) |
      ((v & 0x00000000FFFFFFFF00000000FFFFFFFF00000000FFFFFFFF00000000FFFFFFFF) << 32);
    v =
      ((v & 0xFFFFFFFFFFFFFFFF0000000000000000FFFFFFFFFFFFFFFF0000000000000000) >> 64) |
      ((v & 0x0000000000000000FFFFFFFFFFFFFFFF0000000000000000FFFFFFFFFFFFFFFF) << 64);
    v = (v >> 128) | (v << 128);
    return bytes32(v);
  }

  function toLittleEndian(bool v) internal pure returns (bytes32) {
    return bytes32(v ? 1 << 248 : 0);
  }

  function sha256Pair(bytes32 left, bytes32 right) internal view returns (bytes32 result) {
    /// @solidity memory-safe-assembly
    assembly {
      // Store `left` at memory position 0x00
      mstore(0x00, left)
      // Store `right` at memory position 0x20
      mstore(0x20, right)

      // Call SHA-256 precompile (0x02) with 64-byte input at memory 0x00
      let success := staticcall(gas(), 0x02, 0x00, 0x40, 0x00, 0x20)
      if iszero(success) {
        revert(0, 0)
      }

      // Load the resulting hash from memory
      result := mload(0x00)
    }
  }
}
