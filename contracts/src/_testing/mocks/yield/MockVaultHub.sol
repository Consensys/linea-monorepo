// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.0;

import { IVaultHub } from "../../../yield/interfaces/vendor/lido/IVaultHub.sol";

contract MockVaultHub is IVaultHub {
  bool isVaultConnectedReturn;

  function setIsVaultConnectedReturn(bool _val) external {
    isVaultConnectedReturn = _val;
  }

  function settleVaultObligations(address _vault) external override {}

  function isVaultConnected(address _vault) external view override returns (bool) {
    return isVaultConnectedReturn;
  }
}
