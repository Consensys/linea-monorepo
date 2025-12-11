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

  /// @dev Returns the smallest power of 2 that is greater than or equal to x.
  /// @param x The input value.
  /// @return r The smallest power of 2 >= x. Returns 1 if x is 0.
  /// @notice If x is already a power of 2, returns x. For MaxUint256, the result wraps to 0 due to overflow.
  function nextPow2(uint256 x) internal pure returns (uint256 r) {
      assembly {
          // If x is 0 → next pow2 is 1
          r := x
          if iszero(r) {
              r := 1
          }
          // Decrement first (standard trick)
          r := sub(r, 1)
          // Spread highest bit rightwards
          r := or(r, shr(1, r))
          r := or(r, shr(2, r))
          r := or(r, shr(4, r))
          r := or(r, shr(8, r))
          r := or(r, shr(16, r))
          r := or(r, shr(32, r))
          r := or(r, shr(64, r))
          r := or(r, shr(128, r))
          // Add 1 → next power of 2
          r := add(r, 1)
      }
  }
}
