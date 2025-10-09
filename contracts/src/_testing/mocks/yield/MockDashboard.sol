// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.0;

import { IDashboard } from "../../../yield/interfaces/vendor/lido/IDashboard.sol";
import { ICommonVaultOperations } from "../../../yield/interfaces/vendor/lido/ICommonVaultOperations.sol";
import { IStakingVault } from "../../../yield/interfaces/vendor/lido/IStakingVault.sol";

contract MockDashboard is IDashboard {
  IStakingVault private stakingVaultReturn;
  uint256 private totalValueReturn;
  uint256 private liabilitySharesReturn;
  uint256 private withdrawableValueReturn;
  uint256 private nodeOperatorDisbursableFeeReturn;
  bool isRebalanceVaultWithSharesWithdrawingFromVault;

  function setRebalanceVaultWithSharesWithdrawingFromVault(bool _value) external {
    isRebalanceVaultWithSharesWithdrawingFromVault = _value;
  }

  function setStakingVaultReturn(IStakingVault _value) external {
    stakingVaultReturn = _value;
  }

  function setTotalValueReturn(uint256 _value) external {
    totalValueReturn = _value;
  }

  function setLiabilitySharesReturn(uint256 _value) external {
    liabilitySharesReturn = _value;
  }

  function setWithdrawableValueReturn(uint256 _value) external {
    withdrawableValueReturn = _value;
  }

  function setNodeOperatorDisbursableFeeReturn(uint256 _value) external {
    nodeOperatorDisbursableFeeReturn = _value;
  }

  function stakingVault() external view override returns (IStakingVault) {
    return stakingVaultReturn;
  }

  function totalValue() external view override returns (uint256) {
    return totalValueReturn;
  }

  function liabilityShares() external view override returns (uint256) {
    return liabilitySharesReturn;
  }

  function withdrawableValue() external view override returns (uint256) {
    return withdrawableValueReturn;
  }

  function voluntaryDisconnect() external override {}

  function abandonDashboard(address) external override {}

  function mintStETH(address, uint256) external payable override {}

  function rebalanceVaultWithShares(uint256 _amount) external override {
    if (isRebalanceVaultWithSharesWithdrawingFromVault) {
      ICommonVaultOperations stakingVault = ICommonVaultOperations(stakingVaultReturn);
      stakingVault.withdraw(address(0), _amount);
    }
  }

  function rebalanceVaultWithEther(uint256) external payable override {}

  function nodeOperatorDisbursableFee() external view override returns (uint256) {
    return nodeOperatorDisbursableFeeReturn;
  }

  function disburseNodeOperatorFee() external override {}

  function reconnectToVaultHub() external override {}

  function fund() external payable override {
    ICommonVaultOperations stakingVault = ICommonVaultOperations(stakingVaultReturn);
    stakingVault.fund{ value: msg.value }();
  }

  error MockWithdrawFailed();

  function withdraw(address _recipient, uint256 _amount) external override {
    ICommonVaultOperations stakingVault = ICommonVaultOperations(stakingVaultReturn);
    stakingVault.withdraw(_recipient, _amount);
  }

  receive() external payable {}

  function pauseBeaconChainDeposits() external override {}

  function resumeBeaconChainDeposits() external override {}

  function triggerValidatorWithdrawals(bytes calldata, uint64[] calldata, address) external payable override {}
}
