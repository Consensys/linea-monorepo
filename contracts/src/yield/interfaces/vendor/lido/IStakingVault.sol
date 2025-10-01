// SPDX-FileCopyrightText: 2025 Lido <info@lido.fi>
// SPDX-License-Identifier: GPL-3.0

// See contracts/COMPILERS.md
// solhint-disable-next-line lido/fixed-compiler-version
pragma solidity >=0.8.0;

/**
 * @title IStakingVault
 * @author Lido
 * @notice Interface for the `StakingVault` contract
 */
interface IStakingVault {
  function acceptOwnership() external;

  function ossify() external;
}
