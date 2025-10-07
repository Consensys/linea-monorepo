// SPDX-FileCopyrightText: 2023 Lido <info@lido.fi>
// SPDX-License-Identifier: MIT

// Copied from: https://github.com/OpenZeppelin/openzeppelin-contracts/blob/0457042d93d9dfd760dbaa06a4d2f1216fdbe297/contracts/utils/math/Math.sol

// See contracts/COMPILERS.md
// solhint-disable-next-line
pragma solidity >=0.4.24 <0.9.0;

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
}
