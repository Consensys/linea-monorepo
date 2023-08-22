// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

/**
 * @title Decoding functions for message service anchoring and bytes slicing.
 * @author ConsenSys Software Inc.
 * @notice You can use this to slice bytes and extract anchoring hashes from calldata.
 **/
library CodecV2 {
  /**
   * @notice Decodes a collection of bytes32 (hashes) from the calldata of a transaction.
   * @dev Extracts and decodes skipping the function selector (selector is expected in the input).
   * @dev A check beforehand must be performed to confirm this is the correct type of transaction.
   * @param _calldataWithSelector The calldata for the transaction.
   * @return bytes32[] - array of message hashes.
   **/
  function _extractXDomainAddHashes(bytes memory _calldataWithSelector) internal pure returns (bytes32[] memory) {
    assembly {
      let len := sub(mload(_calldataWithSelector), 4)
      _calldataWithSelector := add(_calldataWithSelector, 0x4)
      mstore(_calldataWithSelector, len)
    }

    return abi.decode(_calldataWithSelector, (bytes32[]));
  }
}
