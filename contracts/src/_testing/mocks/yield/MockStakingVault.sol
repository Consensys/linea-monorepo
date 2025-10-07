// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.0;

import { IStakingVault } from "../../../yield/interfaces/vendor/lido/IStakingVault.sol";

contract MockStakingVault is IStakingVault {
  function acceptOwnership() external override {}

  function ossify() external override {}

  function fund() external payable override {}

  function withdraw(address, uint256) external override {}

  function pauseBeaconChainDeposits() external override {}

  function resumeBeaconChainDeposits() external override {}

  function triggerValidatorWithdrawals(bytes calldata, uint64[] calldata, address) external payable override {}
}
