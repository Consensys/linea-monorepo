// SPDX-License-Identifier: MIT

// See contracts/COMPILERS.md
// solhint-disable-next-line lido/fixed-compiler-version
pragma solidity ^0.8.25;

/**
 * @notice Modified & stripped BLS Lib to support ETH beacon spec for validator deposit message verification.
 * @author Lido
 * @author Solady (https://github.com/Vectorized/solady/blob/dcdfab80f4e6cb9ac35c91610b2a2ec42689ec79/src/utils/ext/ithaca/BLS.sol)
 * @author Ithaca (https://github.com/ithacaxyz/odyssey-examples/blob/main/chapter1/contracts/src/libraries/BLS.sol)
 */
// solhint-disable contract-name-capwords
library BLS12_381 {
  /// @notice Extracted part from `SSZ.verifyProof` for hashing two leaves
  /// @dev Combines 2 bytes32 in 64 bytes input for sha256 precompile
  /// @dev Copied verbatim from Lido audited codebase - https://github.com/lidofinance/core/blob/7cae7a14192ff094fb0eb089433ac9f6fd70e3c6/contracts/common/lib/BLS.sol#L393-L410
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
