// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { IYieldManager } from "./interfaces/IYieldManager.sol";

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
    uint256 _minimumWithdrawalReservePercentageBps;
    uint256 _minimumWithdrawalReserveAmount;
    address[] _yieldProviders;
    mapping(address yieldProvider => IYieldManager.YieldProviderData) _yieldProviderData;
  }

  // keccak256(abi.encode(uint256(keccak256("linea.storage.YieldManagerStorage")) - 1)) & ~bytes32(uint256(0xff))
  bytes32 private constant YieldManagerStorageLocation = 0xdc1272075efdca0b85fb2d76cbb5f26d954dc18e040d6d0b67071bd5cbd04300;

  function _getYieldManagerStorage() internal pure returns (YieldManagerStorage storage $) {
      assembly {
          $.slot := YieldManagerStorageLocation
      }
  }
}