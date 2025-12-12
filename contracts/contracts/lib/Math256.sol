// SPDX-FileCopyrightText: 2023 Lido <info@lido.fi>
// SPDX-License-Identifier: MIT

// Copied from: https://github.com/OpenZeppelin/openzeppelin-contracts/blob/0457042d93d9dfd760dbaa06a4d2f1216fdbe297/contracts/utils/math/Math.sol

// See contracts/COMPILERS.md
// solhint-disable-next-line
pragma solidity >=0.4.24 <0.9.0;

/**
 * @title Library for 256-bit arithmetic operations.
 * @author Consensys Software Inc.
 * @notice Provides utility functions for safe arithmetic operations on uint256 values.
 * @custom:security-contact security-report@linea.build
 */
library Math256 {
  /// @dev Returns the smallest of two numbers.
  function min(uint256 a, uint256 b) internal pure returns (uint256) {
    return a < b ? a : b;
  }

  /// @dev Returns the larger of two numbers.
  function max(uint256 a, uint256 b) internal pure returns (uint256) {
    return a > b ? a : b;
  }

  /// @dev Returns a - b, or 0 if b > a.
  /// This is a saturating subtraction: it never reverts or goes negative.
  function safeSub(uint256 a, uint256 b) internal pure returns (uint256) {
    return a > b ? a - b : 0;
  }

  /// @dev Returns the number of bits required to represent the value.
  /// @param x The input value.
  /// @return r The bit length of x. Returns 0 if x is 0, otherwise returns the position of the highest set bit + 1.
  /// @notice Uses binary search to efficiently find the highest set bit. For non-zero values, this is equivalent to ceil(log2(x + 1)).
  function bitLength(uint256 x) internal pure returns (uint256 r) {
      if (x == 0) return 0; // special case

      if (x >= 1 << 128) { x >>= 128; r += 128; }
      if (x >= 1 << 64)  { x >>= 64;  r += 64; }
      if (x >= 1 << 32)  { x >>= 32;  r += 32; }
      if (x >= 1 << 16)  { x >>= 16;  r += 16; }
      if (x >= 1 << 8)   { x >>= 8;   r += 8; }
      if (x >= 1 << 4)   { x >>= 4;   r += 4; }
      if (x >= 1 << 2)   { x >>= 2;   r += 2; }
      if (x >= 1 << 1)   {             r += 1; }

      // r is now index of highest set bit (0-based), so add 1 for bit length
      return r + 1;
  }
}
