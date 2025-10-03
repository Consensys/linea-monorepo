// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.0;

import { IYieldProvider } from "../../../yield/interfaces/IYieldProvider.sol";
import { YieldManagerStorageLayout } from "../../../yield/YieldManagerStorageLayout.sol";

contract MockYieldProvider is IYieldProvider {
  function withdrawableValue(address _yieldProvider) external view returns (uint256 availableBalance) {
    return 0;
  }
  function fundYieldProvider(address _yieldProvider, uint256 _amount) external {}
  function reportYield(address _yieldProvider) external returns (uint256 newReportedYield) {
    return 0;
  }
  function payLSTPrincipal(
    address _yieldProvider,
    uint256 _availableFunds
  ) external returns (uint256 lstPrincipalPaid) {
    return 0;
  }

  function unstake(address _yieldProvider, bytes memory _withdrawalParams) external payable {}

  function unstakePermissionless(
    address _yieldProvider,
    bytes calldata _withdrawalParams,
    bytes calldata _withdrawalParamsProof
  ) external payable returns (uint256 maxUnstakeAmount) {
    return 0;
  }

  function withdrawFromYieldProvider(address _yieldProvider, uint256 _amount) external {}

  function pauseStaking(address _yieldProvider) external {}

  function unpauseStaking(address _yieldProvider) external {}

  function withdrawLST(address _yieldProvider, uint256 _amount, address _recipient) external {}

  function initiateOssification(address _yieldProvider) external {}

  function undoInitiateOssification(address _yieldProvider) external {}

  function processPendingOssification(address _yieldProvider) external returns (bool isOssificationComplete) {
    return true;
  }

  function validateAdditionToYieldManager(
    YieldManagerStorageLayout.YieldProviderRegistration calldata _registration
  ) external view {}
}
