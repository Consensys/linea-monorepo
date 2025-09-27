// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

/**
 * @title Contract to handle shared storage of YieldManager and YieldProvider contracts
 * @author ConsenSys Software Inc.
 * @dev Pattern we abide by that YieldManager is single writer, and YieldProviders have read-only access. Unfortunately we don't have a succinct Solidity syntax to enforce this.
 * @custom:security-contact security-report@linea.build
 */
abstract contract YieldManagerStorageLayout {
  /// @custom:storage-location erc7201:linea.storage.YieldManager
  struct YieldManagerStorage {
    // Should we struct pack this?
    address _l1MessageService;
    uint256 _targetWithdrawalReservePercentageBps;
    uint256 _targetWithdrawalReserveAmount;
    uint256 _minimumWithdrawalReservePercentageBps; // Can be uint16, because max value of 10000
    uint256 _minimumWithdrawalReserveAmount;
    address[] _yieldProviders;
    uint256 _userFundsInYieldProvidersTotal;
    uint256 _pendingPermissionlessUnstake;
    mapping(address yieldProvider => YieldProviderData) _yieldProviderData;
  }

  enum YieldProviderType {
      LIDO_STVAULT
  }

  /**
   * @notice Supporting data for compressed calldata submission including compressed data.
   * @dev finalStateRootHash is used to set state root at the end of the data.
   */
  struct YieldProviderRegistration {
    YieldProviderType yieldProviderType;
    address yieldProviderEntrypoint;
    address yieldProviderOssificationEntrypoint;
  }

  struct YieldProviderData {
    YieldProviderRegistration registration;
    uint96 yieldProviderIndex;
    bool isStakingPaused;
    bool isOssificationInitiated;
    bool isOssified;
    // Incremented 1:1 with yieldReportedCumulative, because yieldReported becomes user funds
    // Is only allowed to be decremented by withdraw operations. Any other reduction of vault totalValue must be reported as negativeYield.
    uint256 userFunds;
    uint256 yieldReportedCumulative;
    // Required to socialize losses if permanent
    uint256 currentNegativeYield;
    uint256 lstLiabilityPrincipal;
    uint256 lstLiabilityShares;
  }

  // keccak256(abi.encode(uint256(keccak256("linea.storage.YieldManagerStorage")) - 1)) & ~bytes32(uint256(0xff))
  bytes32 private constant YieldManagerStorageLocation = 0xdc1272075efdca0b85fb2d76cbb5f26d954dc18e040d6d0b67071bd5cbd04300;

  function _getYieldManagerStorage() internal pure returns (YieldManagerStorage storage $) {
      assembly {
          $.slot := YieldManagerStorageLocation
      }
  }

  function _getYieldProviderDataStorage(address _yieldProvider) internal view returns (YieldProviderData storage) {
    YieldManagerStorage storage $$ = _getYieldManagerStorage();
    return $$._yieldProviderData[_yieldProvider];
  }
}