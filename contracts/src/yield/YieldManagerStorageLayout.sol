// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

/**
 * @title Contract to handle shared storage of YieldManager and YieldProvider contracts
 * @author ConsenSys Software Inc.
 * @dev Generally we use the pattern that YieldManager is the single storage writer. However there are certain storage value updates that are intricately intertwined with the yield provider logic. We mark these out.
 * @custom:security-contact security-report@linea.build
 */
abstract contract YieldManagerStorageLayout {
  /// @custom:storage-location erc7201:linea.storage.YieldManager
  struct YieldManagerStorage {
    address _l1MessageService;
    uint16 _minimumWithdrawalReservePercentageBps;
    uint16 _targetWithdrawalReservePercentageBps;
    uint256 _minimumWithdrawalReserveAmount;
    uint256 _targetWithdrawalReserveAmount;
    uint256 _userFundsInYieldProvidersTotal;
    uint256 _pendingPermissionlessUnstake;
    address[] _yieldProviders;
    mapping(address l2YieldRecipient => bool) _isL2YieldRecipientKnown;
    mapping(address yieldProvider => YieldProviderStorage) _yieldProviderStorage;
  }

  struct YieldProviderStorage {
    YieldProviderVendor yieldProviderVendor;
    bool isStakingPaused;
    bool isOssificationInitiated;
    bool isOssified;
    address primaryEntrypoint;
    address ossifiedEntrypoint;
    address receiveCaller;
    uint96 yieldProviderIndex;
    uint256 userFunds; // Represents user funds in YieldProvider. Must only be decremented by withdraw operations. Any other reduction of vault totalValue must be reported as negativeYield.
    uint256 yieldReportedCumulative; // Must increment 1:1 with userFunds
    // Timing of below operations is highly specific to the yield provider, so we will mutate their state in the YieldProvider contract
    uint256 currentNegativeYield; // Required to socialize losses if permanent
    uint256 lstLiabilityPrincipal;
  }

  enum YieldProviderVendor {
    LIDO_STVAULT
  }

  struct YieldProviderRegistration {
    YieldProviderVendor yieldProviderVendor;
    address primaryEntrypoint;
    address ossifiedEntrypoint;
    address receiveCaller;
  }

  // keccak256(abi.encode(uint256(keccak256("linea.storage.YieldManagerStorage")) - 1)) & ~bytes32(uint256(0xff))
  bytes32 private constant YieldManagerStorageLocation =
    0xdc1272075efdca0b85fb2d76cbb5f26d954dc18e040d6d0b67071bd5cbd04300;

  function _getYieldManagerStorage() internal pure returns (YieldManagerStorage storage $) {
    assembly {
      $.slot := YieldManagerStorageLocation
    }
  }

  function _getYieldProviderStorage(address _yieldProvider) internal view returns (YieldProviderStorage storage $$) {
    $$ = _getYieldManagerStorage()._yieldProviderStorage[_yieldProvider];
  }
}
