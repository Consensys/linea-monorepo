// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { YieldManagerStorageLayout } from "./YieldManagerStorageLayout.sol";
import { IYieldProvider } from "./interfaces/IYieldProvider.sol";

/**
 * @title Contract to handle native yield operations with Lido Staking Vault.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract YieldProviderBase is YieldManagerStorageLayout, IYieldProvider {
  address immutable YIELD_MANAGER;

  // @dev Because we place most the bulk of input validation in YieldManager, we need to ensure the YieldManager is the only caller for YieldProvider mutator logic
  modifier onlyDelegateCall() {
    if (address(this) != YIELD_MANAGER) {
      revert ContextIsNotYieldManager();
    }
    _;
  }
}
