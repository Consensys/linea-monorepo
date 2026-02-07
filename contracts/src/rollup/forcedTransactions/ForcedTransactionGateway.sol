// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.33;
import { IGenericErrors } from "../../interfaces/IGenericErrors.sol";
import { IAcceptForcedTransactions } from "./interfaces/IAcceptForcedTransactions.sol";
import { IForcedTransactionGateway } from "./interfaces/IForcedTransactionGateway.sol";
import { IAddressFilter } from "./interfaces/IAddressFilter.sol";
import { Mimc } from "../../libraries/Mimc.sol";
import { FinalizedStateHashing } from "../../libraries/FinalizedStateHashing.sol";
import { AccessControl } from "@openzeppelin/contracts/access/AccessControl.sol";

import { LibRLP } from "solady/src/utils/LibRLP.sol";

import { ECDSA } from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
/**
 * @title Contract to manage forced transactions on L1.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract ForcedTransactionGateway is AccessControl, IForcedTransactionGateway {
  using Mimc for *;
  using LibRLP for *;
  using FinalizedStateHashing for *;

  /// @notice Contains the minimum gas allowed for a forced transaction.
  uint256 private constant MIN_GAS_LIMIT = 21000;

  /// @notice Contains the destination address to store the forced transactions on.
  IAcceptForcedTransactions public immutable LINEA_ROLLUP;

  /// @notice Contains the destination chain ID used in the RLP encoding.
  uint256 public immutable DESTINATION_CHAIN_ID;

  /// @notice Contains the buffer for computing the L2 block the transaction will be processed by.
  uint256 public immutable L2_BLOCK_BUFFER;

  /// @notice Contains the maximum gas allowed for a forced transaction.
  uint256 public immutable MAX_GAS_LIMIT;

  /// @notice Contains the maximum calldata length allowed for a forced transaction.
  uint256 public immutable MAX_INPUT_LENGTH_LIMIT;

  /// @notice Contains the address for the transaction address filter.
  IAddressFilter public immutable ADDRESS_FILTER;

  /// @notice Toggles the feature switch for using the address filter.
  bool public useAddressFilter = true;

  constructor(
    address _lineaRollup,
    uint256 _destinationChainId,
    uint256 _l2BlockBuffer,
    uint256 _maxGasLimit,
    uint256 _maxInputLengthBuffer,
    address _defaultAdmin,
    address _addressFilter
  ) {
    require(_lineaRollup != address(0), IGenericErrors.ZeroAddressNotAllowed());
    require(_destinationChainId != 0, IGenericErrors.ZeroValueNotAllowed());
    require(_l2BlockBuffer != 0, IGenericErrors.ZeroValueNotAllowed());
    require(_maxGasLimit != 0, IGenericErrors.ZeroValueNotAllowed());
    require(_maxInputLengthBuffer != 0, IGenericErrors.ZeroValueNotAllowed());
    require(_defaultAdmin != address(0), IGenericErrors.ZeroAddressNotAllowed());
    require(_addressFilter != address(0), IGenericErrors.ZeroAddressNotAllowed());

    LINEA_ROLLUP = IAcceptForcedTransactions(_lineaRollup);
    DESTINATION_CHAIN_ID = _destinationChainId;
    L2_BLOCK_BUFFER = _l2BlockBuffer;
    MAX_GAS_LIMIT = _maxGasLimit;
    MAX_INPUT_LENGTH_LIMIT = _maxInputLengthBuffer;
    ADDRESS_FILTER = IAddressFilter(_addressFilter);
    _grantRole(DEFAULT_ADMIN_ROLE, _defaultAdmin);
  }

  /**
   * @notice Function to submit forced transactions.
   * @param _forcedTransaction The fields required for the transaction excluding chainId.
   * @param _lastFinalizedState The last finalized state validated to use the timestamp in block number calculation.
   */
  function submitForcedTransaction(
    Eip1559Transaction memory _forcedTransaction,
    LastFinalizedState memory _lastFinalizedState
  ) external payable {
    require(_forcedTransaction.gasLimit >= MIN_GAS_LIMIT, GasLimitTooLow());
    require(_forcedTransaction.gasLimit <= MAX_GAS_LIMIT, MaxGasLimitExceeded());
    require(_forcedTransaction.input.length <= MAX_INPUT_LENGTH_LIMIT, CalldataInputLengthLimitExceeded());
    require(
      _forcedTransaction.maxPriorityFeePerGas > 0 && _forcedTransaction.maxFeePerGas > 0,
      GasFeeParametersContainZero(_forcedTransaction.maxFeePerGas, _forcedTransaction.maxPriorityFeePerGas)
    );

    require(
      _forcedTransaction.maxPriorityFeePerGas <= _forcedTransaction.maxFeePerGas,
      MaxPriorityFeePerGasHigherThanMaxFee(_forcedTransaction.maxFeePerGas, _forcedTransaction.maxPriorityFeePerGas)
    );

    require(_forcedTransaction.yParity <= 1, YParityGreaterThanOne(_forcedTransaction.yParity));

    (
      bytes32 currentFinalizedState,
      bytes32 previousForcedTransactionRollingHash,
      uint256 currentFinalizedL2BlockNumber,
      uint256 forcedTransactionFeeAmount
    ) = LINEA_ROLLUP.getRequiredForcedTransactionFields();

    require(msg.value == forcedTransactionFeeAmount, ForcedTransactionFeeNotMet(forcedTransactionFeeAmount, msg.value));

    if (
      currentFinalizedState !=
      FinalizedStateHashing._computeLastFinalizedState(
        _lastFinalizedState.messageNumber,
        _lastFinalizedState.messageRollingHash,
        _lastFinalizedState.forcedTransactionNumber,
        _lastFinalizedState.forcedTransactionRollingHash,
        _lastFinalizedState.timestamp
      )
    ) {
      revert FinalizationStateIncorrect(
        currentFinalizedState,
        FinalizedStateHashing._computeLastFinalizedState(
          _lastFinalizedState.messageNumber,
          _lastFinalizedState.messageRollingHash,
          _lastFinalizedState.forcedTransactionNumber,
          _lastFinalizedState.forcedTransactionRollingHash,
          _lastFinalizedState.timestamp
        )
      );
    }

    LibRLP.List memory accessList = _buildAccessList(_forcedTransaction.accessList);
    LibRLP.List memory transactionFieldList = LibRLP.p();
    transactionFieldList = LibRLP.p(transactionFieldList, DESTINATION_CHAIN_ID);
    transactionFieldList = LibRLP.p(transactionFieldList, _forcedTransaction.nonce);
    transactionFieldList = LibRLP.p(transactionFieldList, _forcedTransaction.maxPriorityFeePerGas);
    transactionFieldList = LibRLP.p(transactionFieldList, _forcedTransaction.maxFeePerGas);
    transactionFieldList = LibRLP.p(transactionFieldList, _forcedTransaction.gasLimit);

    if (_forcedTransaction.to == address(0)) {
      transactionFieldList = LibRLP.p(transactionFieldList, bytes(""));
    } else {
      transactionFieldList = LibRLP.p(transactionFieldList, _forcedTransaction.to);
    }
    transactionFieldList = LibRLP.p(transactionFieldList, _forcedTransaction.value);
    transactionFieldList = LibRLP.p(transactionFieldList, _forcedTransaction.input);
    transactionFieldList = LibRLP.p(transactionFieldList, accessList);

    bytes32 hashedPayload = keccak256(abi.encodePacked(hex"02", LibRLP.encode(transactionFieldList)));

    address signer;
    unchecked {
      (signer, ) = ECDSA.tryRecover(
        hashedPayload,
        _forcedTransaction.yParity + 27,
        bytes32(_forcedTransaction.r),
        bytes32(_forcedTransaction.s)
      );
    }
    require(signer != address(0), SignerAddressZero());

    if (useAddressFilter) {
      require(!ADDRESS_FILTER.addressIsFiltered(signer), AddressIsFiltered());
      require(!ADDRESS_FILTER.addressIsFiltered(_forcedTransaction.to), AddressIsFiltered());
    }

    uint256 blockNumberDeadline;
    unchecked {
      /// @dev The computation uses 1s block time making block number and seconds interchangeable,
      ///      while the chain might currently differ at >1s, this gives additional inclusion time.
      blockNumberDeadline =
        currentFinalizedL2BlockNumber +
        block.timestamp -
        _lastFinalizedState.timestamp +
        L2_BLOCK_BUFFER;
    }

    bytes32 hashedPayloadMsb;
    bytes32 hashedPayloadLsb;
    assembly {
      hashedPayloadMsb := shr(128, hashedPayload)
      hashedPayloadLsb := and(hashedPayload, 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF)
    }

    transactionFieldList = LibRLP.p(transactionFieldList, _forcedTransaction.yParity);
    transactionFieldList = LibRLP.p(transactionFieldList, _forcedTransaction.r);
    transactionFieldList = LibRLP.p(transactionFieldList, _forcedTransaction.s);

    LINEA_ROLLUP.storeForcedTransaction{ value: msg.value }(
      Mimc.hash(
        abi.encode(
          previousForcedTransactionRollingHash,
          hashedPayloadMsb,
          hashedPayloadLsb,
          blockNumberDeadline,
          signer
        )
      ),
      signer,
      blockNumberDeadline,
      abi.encodePacked(hex"02", LibRLP.encode(transactionFieldList))
    );
  }

  /**
   * @notice Function to toggle the usage of the address filter.
   * @dev Only callable by an account with the DEFAULT_ADMIN_ROLE.
   * @param _useAddressFilter Bool indicating whether or not to use the address filter.
   */
  function toggleUseAddressFilter(bool _useAddressFilter) external onlyRole(DEFAULT_ADMIN_ROLE) {
    require(useAddressFilter != _useAddressFilter, AddressFilterAlreadySet(_useAddressFilter));
    useAddressFilter = _useAddressFilter;
    emit AddressFilterSet(_useAddressFilter);
  }

  /**
   * @notice Function to convert the transaction access list to a LibRLP.List.
   * @param _accessList The transaction access list to convert.
   * @return list The List object.
   */
  function _buildAccessList(AccessList[] memory _accessList) internal pure returns (LibRLP.List memory list) {
    unchecked {
      list = LibRLP.p();
      for (uint256 i; i < _accessList.length; ++i) {
        LibRLP.List memory keys = LibRLP.p();
        bytes32[] memory ks = _accessList[i].storageKeys;
        for (uint256 j; j < ks.length; ++j) {
          bytes memory b = new bytes(32);
          assembly {
            mstore(add(b, 0x20), mload(add(ks, add(0x20, shl(5, j)))))
          }
          keys = LibRLP.p(keys, b);
        }
        LibRLP.List memory acct = LibRLP.p();
        acct = LibRLP.p(acct, _accessList[i].contractAddress);
        acct = LibRLP.p(acct, keys);
        list = LibRLP.p(list, acct);
      }
    }
  }
}
