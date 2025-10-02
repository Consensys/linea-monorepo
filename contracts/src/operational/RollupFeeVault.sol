// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.30;

import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { L2MessageService } from "../messaging/l2/L2MessageService.sol";
import { TokenBridge } from "../bridging/token/TokenBridge.sol";
import { IRollupFeeVault } from "./interfaces/IRollupFeeVault.sol";
import { IV3DexSwap } from "./interfaces/IV3DexSwap.sol";

/**
 * @title Upgradeable Fee Vault Contract.
 * @notice Accepts ETH for later economic functions.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract RollupFeeVault is AccessControlUpgradeable, IRollupFeeVault {
  bytes32 public constant INVOICE_SETTER_ROLE = keccak256("INVOICE_SETTER_ROLE");
  bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");

  /// @notice Decentralized exchange contract for swapping ETH to LINEA tokens.
  IV3DexSwap public v3Dex;
  /// @notice Address to receive operating costs.
  address public operatingCostsReceiver;
  /// @notice Amount of operating costs owed.
  uint256 public operatingCosts;
  /// @notice Timestamp of the last operating costs update.
  uint256 public lastOperatingCostsUpdate;
  /// @notice Address of the token bridge contract.
  TokenBridge public tokenBridge;
  /// @notice Address of the L2 message service contract.
  L2MessageService public messageService;
  /// @notice Address of the L1 burner contract to which LINEA tokens are bridged for burning.
  address public l1BurnerContract;
  /// @notice Address of the LINEA token contract.
  address public lineaToken;

  /// @custom:oz-upgrades-unsafe-allow constructor
  constructor() {
    _disableInitializers();
  }

  /**
   * @notice Initializes the contract state.
   * @param _defaultAdmin Address to be granted the default admin role.
   * @param _invoiceSetter Address to be granted the invoice role.
   * @param _burner Address to be granted the burner role.
   * @param _operatingCostsReceiver Address to receive operating costs.
   * @param _tokenBridge Address of the token bridge contract.
   * @param _messageService Address of the L2 message service contract.
   * @param _l1BurnerContract Address of the L1 burner contract.
   * @param _lineaToken Address of the LINEA token contract.
   * @param _dex Address of the DEX contract.
   */
  function initialize(
    address _defaultAdmin,
    address _invoiceSetter,
    address _burner,
    address _operatingCostsReceiver,
    address _tokenBridge,
    address _messageService,
    address _l1BurnerContract,
    address _lineaToken,
    address _dex
  ) external initializer {
    __AccessControl_init();
    __RollupFeeVault_init(
      _defaultAdmin,
      _invoiceSetter,
      _burner,
      _operatingCostsReceiver,
      _tokenBridge,
      _messageService,
      _l1BurnerContract,
      _lineaToken,
      _dex
    );
  }

  /**
   * @notice Initializes the contract state for upgrade.
   * @param _defaultAdmin Address to be granted the default admin role.
   * @param _invoiceSetter Address to be granted the invoice role.
   * @param _burner Address to be granted the burner role.
   * @param _operatingCostsReceiver Address to receive operating costs.
   * @param _tokenBridge Address of the token bridge contract.
   * @param _messageService Address of the L2 message service contract.
   * @param _l1BurnerContract Address of the L1 burner contract.
   * @param _lineaToken Address of the LINEA token contract.
   * @param _dex Address of the DEX contract.
   */
  function initializeRolesAndStorageVariables(
    address _defaultAdmin,
    address _invoiceSetter,
    address _burner,
    address _operatingCostsReceiver,
    address _tokenBridge,
    address _messageService,
    address _l1BurnerContract,
    address _lineaToken,
    address _dex
  ) external reinitializer(2) {
    __AccessControl_init();
    __RollupFeeVault_init(
      _defaultAdmin,
      _invoiceSetter,
      _burner,
      _operatingCostsReceiver,
      _tokenBridge,
      _messageService,
      _l1BurnerContract,
      _lineaToken,
      _dex
    );
  }

  function __RollupFeeVault_init(
    address _defaultAdmin,
    address _invoiceSetter,
    address _burner,
    address _operatingCostsReceiver,
    address _tokenBridge,
    address _messageService,
    address _l1BurnerContract,
    address _lineaToken,
    address _dex
  ) internal onlyInitializing {
    require(_defaultAdmin != address(0), ZeroAddressNotAllowed());
    require(_invoiceSetter != address(0), ZeroAddressNotAllowed());
    require(_burner != address(0), ZeroAddressNotAllowed());
    require(_operatingCostsReceiver != address(0), ZeroAddressNotAllowed());
    require(_tokenBridge != address(0), ZeroAddressNotAllowed());
    require(_messageService != address(0), ZeroAddressNotAllowed());
    require(_l1BurnerContract != address(0), ZeroAddressNotAllowed());
    require(_lineaToken != address(0), ZeroAddressNotAllowed());
    require(_dex != address(0), ZeroAddressNotAllowed());

    _grantRole(DEFAULT_ADMIN_ROLE, _defaultAdmin);
    _grantRole(INVOICE_SETTER_ROLE, _invoiceSetter);
    _grantRole(BURNER_ROLE, _burner);

    lastOperatingCostsUpdate = block.timestamp;

    operatingCostsReceiver = _operatingCostsReceiver;
    tokenBridge = TokenBridge(_tokenBridge);
    messageService = L2MessageService(_messageService);
    l1BurnerContract = _l1BurnerContract;
    lineaToken = _lineaToken;
    v3Dex = IV3DexSwap(_dex);
  }

  /**
   * @notice Send operating costs to the designated receiver.
   * @param _startTimestamp Start of the period the costs are covering.
   * @param _endTimestamp End of the period the costs are covering.
   * @param _amount New operating costs value.
   */
  function sendOperatingCosts(
    uint256 _startTimestamp,
    uint256 _endTimestamp,
    uint256 _amount
  ) external payable onlyRole(INVOICE_SETTER_ROLE) {
    require(_startTimestamp == lastOperatingCostsUpdate + 1, TimestampsNotInSequence());
    require(_endTimestamp > _startTimestamp, EndTimestampMustBeGreaterThanStartTimestamp());
    require(_amount != 0, ZeroOperatingCosts());

    uint256 totalAmountOwing = operatingCosts + _amount;
    lastOperatingCostsUpdate = _endTimestamp;

    address payable receiver = payable(operatingCostsReceiver);
    uint256 balanceAvailable = address(this).balance;
    uint256 amountToPay;

    if (balanceAvailable == 0) {
      operatingCosts = totalAmountOwing;
      amountToPay = 0;
    } else if (balanceAvailable < totalAmountOwing) {
      operatingCosts = totalAmountOwing - balanceAvailable;
      amountToPay = balanceAvailable;
    } else {
      operatingCosts = 0;
      amountToPay = totalAmountOwing;
    }

    if (amountToPay > 0) {
      (bool success, ) = receiver.call{ value: amountToPay }("");
      require(success, OperatingCostsTransferFailed());
    }

    emit InvoiceProcessed(receiver, _startTimestamp, _endTimestamp, amountToPay, _amount);
  }

  /**
   * @notice Update the operating costs.
   * @param _newOperatingCosts New operating costs value.
   */
  function updateOperatingCosts(uint256 _newOperatingCosts) external onlyRole(DEFAULT_ADMIN_ROLE) {
    operatingCosts = _newOperatingCosts;
    lastOperatingCostsUpdate = block.timestamp;
    emit OperatingCostsUpdated(_newOperatingCosts);
  }

  /**
   * @notice Updates the address of the L1 burner contract.
   * @param _newL1BurnerContract Address of the new L1 burner contract.
   */
  function updateL1BurnerContract(address _newL1BurnerContract) external onlyRole(DEFAULT_ADMIN_ROLE) {
    require(_newL1BurnerContract != address(0), ZeroAddressNotAllowed());
    l1BurnerContract = _newL1BurnerContract;
    emit L1BurnerContractUpdated(_newL1BurnerContract);
  }

  /**
   * @notice Updates the address of the DEX contract.
   * @param _newDex Address of the new DEX contract.
   */
  function updateDex(address _newDex) external onlyRole(DEFAULT_ADMIN_ROLE) {
    require(_newDex != address(0), ZeroAddressNotAllowed());
    v3Dex = IV3DexSwap(_newDex);
    emit DexUpdated(_newDex);
  }

  /**
   * @notice Updates the address of the operating costs receiver.
   * @param _newOperatingCostsReceiver Address of the new operating costs receiver.
   */
  function updateOperatingCostsReceiver(address _newOperatingCostsReceiver) external onlyRole(DEFAULT_ADMIN_ROLE) {
    require(_newOperatingCostsReceiver != address(0), ZeroAddressNotAllowed());
    operatingCostsReceiver = _newOperatingCostsReceiver;
    emit OperatingCostsReceiverUpdated(_newOperatingCostsReceiver);
  }

  /**
   * @notice Burns 20% of the ETH balance and uses the rest to buy LINEA tokens which are then bridged to L1 to be burned.
   * @param _minLineaOut Number of LINEA tokens to receive as a minimum (slippage protection).
   * @param _deadline Time after which the transaction will revert if not yet processed.
   * @param _sqrtPriceLimitX96 Price limit of the swap as a Q64.96 value.
   */
  function burnAndBridge(
    uint256 _minLineaOut,
    uint256 _deadline,
    uint160 _sqrtPriceLimitX96
  ) public onlyRole(BURNER_ROLE) {
    require(operatingCosts == 0, ZeroOperatingCosts());

    uint256 minimumFee = messageService.minimumFeeInWei();

    require(address(this).balance > minimumFee, InsufficientBalance());

    uint256 balanceAvailable = address(this).balance - minimumFee;

    uint256 ethToBurn = (balanceAvailable * 20) / 100;
    (bool success, ) = address(0).call{ value: ethToBurn }("");
    require(success, EthBurnFailed());

    uint256 nbLineaTokens = v3Dex.swap{ value: balanceAvailable - ethToBurn }(
      _minLineaOut,
      _deadline,
      _sqrtPriceLimitX96
    );

    IERC20(lineaToken).approve(address(tokenBridge), nbLineaTokens);

    tokenBridge.bridgeToken{ value: minimumFee }(lineaToken, nbLineaTokens, l1BurnerContract);

    emit EthBurntSwappedAndBridged(ethToBurn, nbLineaTokens);
  }

  /**
   * @notice Fallback function - Receives Funds.
   */
  fallback() external payable {
    emit EthReceived(msg.value);
  }

  /**
   * @notice Receive function - Receives Funds.
   */
  receive() external payable {
    emit EthReceived(msg.value);
  }
}
