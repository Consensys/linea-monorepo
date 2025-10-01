// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { YieldManagerStorageLayout } from "./YieldManagerStorageLayout.sol";
import { IYieldProvider } from "./interfaces/IYieldProvider.sol";
import { ErrorUtils } from "../libraries/ErrorUtils.sol";

/**
 * @title Contract to handle native yield operations with Lido Staking Vault.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract YieldProviderBase is YieldManagerStorageLayout, IYieldProvider {
  address immutable YIELD_MANAGER;

  constructor(address _l1MessageService, address _yieldManager) {
      ErrorUtils.revertIfZeroAddress(_l1MessageService);
      ErrorUtils.revertIfZeroAddress(_yieldManager);
      L1_MESSAGE_SERVICE = _l1MessageService;
      YIELD_MANAGER = _yieldManager;
  }

  // @dev Because we place most the bulk of input validation in YieldManager, we need to ensure the YieldManager is the only caller for YieldProvider mutator logic
  modifier onlyDelegateCall() {
    if (address(this) != YIELD_MANAGER) {
      revert ContextIsNotYieldManager();
    }
    _;
  }
}
