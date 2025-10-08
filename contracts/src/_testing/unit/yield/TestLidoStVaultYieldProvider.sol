// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { LidoStVaultYieldProvider } from "../../../yield/LidoStVaultYieldProvider.sol";
import { GIndex } from "../../../yield/libs/vendor/lido/GIndex.sol";

/// @custom:oz-upgrades-unsafe-allow missing-initializer
contract TestLidoStVaultYieldProvider is LidoStVaultYieldProvider {
  constructor(
    address _l1MessageService,
    address _yieldManager,
    address _vaultHub,
    address _steth,
    GIndex _gIFirstValidator,
    GIndex _gIFirstValidatorAfterChange,
    uint64 _changeSlot
  )
    LidoStVaultYieldProvider(
      _l1MessageService,
      _yieldManager,
      _vaultHub,
      _steth,
      _gIFirstValidator,
      _gIFirstValidatorAfterChange,
      _changeSlot
    )
  {}
}
