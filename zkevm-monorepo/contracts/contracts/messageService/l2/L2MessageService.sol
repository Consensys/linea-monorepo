// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { L2MessageServiceV1 } from "./v1/L2MessageServiceV1.sol";
import { L2MessageManager } from "./L2MessageManager.sol";

/**
 * @title Contract to manage cross-chain messaging on L2.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract L2MessageService is AccessControlUpgradeable, L2MessageServiceV1, L2MessageManager {
  /// @dev Total contract storage is 50 slots with the gap below.
  /// @dev Keep 50 free storage slots for future implementation updates to avoid storage collision.
  uint256[50] private __gap_L2MessageService;
}
