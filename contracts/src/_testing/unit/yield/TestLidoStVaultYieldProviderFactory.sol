// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { LidoStVaultYieldProviderFactory } from "../../../yield/LidoStVaultYieldProviderFactory.sol";
import { TestLidoStVaultYieldProvider } from "./TestLidoStVaultYieldProvider.sol";

import { GIndex } from "../../../yield/libs/vendor/lido/GIndex.sol";

/// @custom:oz-upgrades-unsafe-allow missing-initializer
contract TestLidoStVaultYieldProviderFactory is LidoStVaultYieldProviderFactory {
  constructor(
    address _l1MessageService,
    address _yieldManager,
    address _vaultHub,
    address _steth,
    GIndex _gIFirstValidator,
    GIndex _gIFirstValidatorAfterChange,
    uint64 _changeSlot
  )
    LidoStVaultYieldProviderFactory(
      _l1MessageService,
      _yieldManager,
      _vaultHub,
      _steth,
      _gIFirstValidator,
      _gIFirstValidatorAfterChange,
      _changeSlot
    )
  {}

  function createTestLidoStVaultYieldProvider() external returns (address yieldProviderAddress) {
    yieldProviderAddress = address(
      new TestLidoStVaultYieldProvider(
        L1_MESSAGE_SERVICE,
        YIELD_MANAGER,
        VAULT_HUB,
        STETH,
        GI_FIRST_VALIDATOR,
        GI_FIRST_VALIDATOR_AFTER_CHANGE,
        CHANGE_SLOT
      )
    );
    emit LidoStVaultYieldProviderCreated(yieldProviderAddress);
  }
}
