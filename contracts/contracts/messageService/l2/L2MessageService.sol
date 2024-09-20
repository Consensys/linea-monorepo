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

  /**
   * @notice Initializes underlying message service dependencies.
   * @param _rateLimitPeriod The period to rate limit against.
   * @param _rateLimitAmount The limit allowed for withdrawing the period.
   * @param _roleAddresses The list of addresses to grant roles to.
   * @param _pauseTypeRoles The list of pause type roles.
   * @param _unpauseTypeRoles The list of unpause type roles.
   */
  function initialize(
    uint256 _rateLimitPeriod,
    uint256 _rateLimitAmount,
    RoleAddress[] calldata _roleAddresses,
    PauseTypeRole[] calldata _pauseTypeRoles,
    PauseTypeRole[] calldata _unpauseTypeRoles
  ) external initializer {
    __ERC165_init();
    __Context_init();
    __AccessControl_init();
    __RateLimiter_init(_rateLimitPeriod, _rateLimitAmount);

    __ReentrancyGuard_init();
    __PauseManager_init(_pauseTypeRoles, _unpauseTypeRoles);

    _setPermissions(_roleAddresses);

    nextMessageNumber = 1;

    _messageSender = DEFAULT_SENDER_ADDRESS;
    minimumFeeInWei = 0.0001 ether;
  }

  /**
   * @notice Sets permissions for a list of addresses and their roles as well as initialises the PauseManager pauseType:role mappings.
   * @dev This function is a reinitializer and can only be called once per version. Should be called using an upgradeAndCall transaction to the ProxyAdmin.
   * @param _roleAddresses The list of addresses and their roles.
   * @param _pauseTypeRoles The list of pause type roles.
   * @param _unpauseTypeRoles The list of unpause type roles.
   */
  function reinitializePauseTypesAndPermissions(
    RoleAddress[] calldata _roleAddresses,
    PauseTypeRole[] calldata _pauseTypeRoles,
    PauseTypeRole[] calldata _unpauseTypeRoles
  ) external reinitializer(6) {
    _setPermissions(_roleAddresses);
    __PauseManager_init(_pauseTypeRoles, _unpauseTypeRoles);
  }

  /**
   * @notice Sets permissions for a list of addresses and their roles.
   * @dev This function is a reinitializer and can only be called once per version.
   * @param _roleAddresses The list of addresses and their roles.
   */
  function _setPermissions(RoleAddress[] calldata _roleAddresses) internal {
    uint256 roleAddressesLength = _roleAddresses.length;

    for (uint256 i; i < roleAddressesLength; i++) {
      if (_roleAddresses[i].addressWithRole == address(0)) {
        revert ZeroAddressNotAllowed();
      }
      _grantRole(_roleAddresses[i].role, _roleAddresses[i].addressWithRole);
    }
  }
}
