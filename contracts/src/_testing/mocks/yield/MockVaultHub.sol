// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.0;

import { IVaultHub } from "../../../yield/interfaces/vendor/lido/IVaultHub.sol";
import { ICommonVaultOperations } from "../../../yield/interfaces/vendor/lido/ICommonVaultOperations.sol";

contract MockVaultHub is IVaultHub {
  bool isVaultConnectedReturn;
  bool isSettleLidoFeesWithdrawingFromVault;
  uint256 settleVaultObligationAmount;

  function setIsVaultConnectedReturn(bool _val) external {
    isVaultConnectedReturn = _val;
  }

  function setIsSettleLidoFeesWithdrawingFromVault(bool _value) external {
    isSettleLidoFeesWithdrawingFromVault = _value;
  }

  function setSettleVaultObligationAmount(uint256 _val) external {
    settleVaultObligationAmount = _val;
  }

  function settleLidoFees(address _vault) external override {
    if (isSettleLidoFeesWithdrawingFromVault) {
      ICommonVaultOperations vault = ICommonVaultOperations(_vault);
      vault.withdraw(address(0), settleVaultObligationAmount);
    }
  }

  function isVaultConnected(address _vault) external view override returns (bool) {
    return isVaultConnectedReturn;
  }
}
