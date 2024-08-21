// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.19;

import { IPauseManager } from "../../interfaces/IPauseManager.sol";

/**
 * @title Interface declaring Canonical Token Bridge functions, events and errors.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface ITokenBridge {
  /**
   * @dev Contract will be used as proxy implementation.
   * @param messageService The address of the MessageService contract.
   * @param tokenBeacon The address of the tokenBeacon.
   * @param sourceChainId The source chain id of the current layer
   * @param targetChainId The target chaind id of the targeted layer
   * @param reservedTokens The list of reserved tokens to be set
   * @param roleAddresses The list of role addresses.
   * @param pauseTypeRoles The list of pause type roles.
   * @param unpauseTypeRoles The list of unpause type roles.
   */
  struct InitializationData {
    address messageService;
    address tokenBeacon;
    uint256 sourceChainId;
    uint256 targetChainId;
    address[] reservedTokens;
    IPauseManager.RoleAddress[] roleAddresses;
    IPauseManager.PauseTypeRole[] pauseTypeRoles;
    IPauseManager.PauseTypeRole[] unpauseTypeRoles;
  }

  event TokenReserved(address indexed token);
  event ReservationRemoved(address indexed token);
  event CustomContractSet(address indexed nativeToken, address indexed customContract, address indexed setBy);
  /// @dev DEPRECATED in favor of BridgingInitiatedV2.
  event BridgingInitiated(address indexed sender, address recipient, address indexed token, uint256 indexed amount);
  event BridgingInitiatedV2(address indexed sender, address indexed recipient, address indexed token, uint256 amount);
  /// @dev DEPRECATED in favor of BridgingFinalizedV2.
  event BridgingFinalized(
    address indexed nativeToken,
    address indexed bridgedToken,
    uint256 indexed amount,
    address recipient
  );
  event BridgingFinalizedV2(
    address indexed nativeToken,
    address indexed bridgedToken,
    uint256 amount,
    address indexed recipient
  );
  event NewToken(address indexed token);
  event NewTokenDeployed(address indexed bridgedToken, address indexed nativeToken);
  event RemoteTokenBridgeSet(address indexed remoteTokenBridge, address indexed setBy);
  event TokenDeployed(address indexed token);
  event DeploymentConfirmed(address[] tokens, address indexed confirmedBy);
  event MessageServiceUpdated(
    address indexed newMessageService,
    address indexed oldMessageService,
    address indexed setBy
  );

  error ReservedToken(address token);
  error RemoteTokenBridgeAlreadySet(address remoteTokenBridge);
  error AlreadyBridgedToken(address token);
  error InvalidPermitData(bytes4 permitData, bytes4 permitSelector);
  error PermitNotFromSender(address owner);
  error PermitNotAllowingBridge(address spender);
  error ZeroAmountNotAllowed(uint256 amount);
  error NotReserved(address token);
  error TokenNotDeployed(address token);
  error AlreadyBrigedToNativeTokenSet(address token);
  error NativeToBridgedTokenAlreadySet(address token);
  error StatusAddressNotAllowed(address token);
  error DecimalsAreUnknown(address token);

  /**
   * @notice Similar to `bridgeToken` function but allows to pass additional
   *   permit data to do the ERC20 approval in a single transaction.
   * @param _token The address of the token to be bridged.
   * @param _amount The amount of the token to be bridged.
   * @param _recipient The address that will receive the tokens on the other chain.
   * @param _permitData The permit data for the token, if applicable.
   */
  function bridgeTokenWithPermit(
    address _token,
    uint256 _amount,
    address _recipient,
    bytes calldata _permitData
  ) external payable;

  /**
   * @dev It can only be called from the Message Service. To finalize the bridging
   *   process, a user or postmen needs to use the `claimMessage` function of the
   *   Message Service to trigger the transaction.
   * @param _nativeToken The address of the token on its native chain.
   * @param _amount The amount of the token to be received.
   * @param _recipient The address that will receive the tokens.
   * @param _chainId The source chainId or target chaindId for this token
   * @param _tokenMetadata Additional data used to deploy the bridged token if it
   *   doesn't exist already.
   */
  function completeBridging(
    address _nativeToken,
    uint256 _amount,
    address _recipient,
    uint256 _chainId,
    bytes calldata _tokenMetadata
  ) external;

  /**
   * @dev Change the address of the Message Service.
   * @param _messageService The address of the new Message Service.
   */
  function setMessageService(address _messageService) external;

  /**
   * @dev It can only be called from the Message Service. To change the status of
   *   the native tokens to DEPLOYED meaning they have been deployed on the other chain
   *   a user or postman needs to use the `claimMessage` function of the
   *   Message Service to trigger the transaction.
   * @param _nativeTokens The addresses of the native tokens.
   */
  function setDeployed(address[] memory _nativeTokens) external;

  /**
   * @dev Sets the address of the remote token bridge. Can only be called once.
   * @param _remoteTokenBridge The address of the remote token bridge to be set.
   */
  function setRemoteTokenBridge(address _remoteTokenBridge) external;

  /**
   * @dev Removes a token from the reserved list.
   * @param _token The address of the token to be removed from the reserved list.
   */
  function removeReserved(address _token) external;

  /**
   * @dev Linea can set a custom ERC20 contract for specific ERC20.
   *   For security purpose, Linea can only call this function if the token has
   *   not been bridged yet.
   * @param _nativeToken address of the token on the source chain.
   * @param _targetContract address of the custom contract.
   */
  function setCustomContract(address _nativeToken, address _targetContract) external;
}
