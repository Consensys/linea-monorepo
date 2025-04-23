// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.19;

/**
 * @title Library to hash cross-chain messages.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
library MessageHashing {
  /**
   * @notice Hashes messages using assembly for efficiency.
   * @dev Adding 0xc0 is to indicate the calldata offset relative to the memory being added to.
   * @dev If the calldata is not modulus 32, the extra bit needs to be added on at the end else the hash is wrong.
   * @param _from The from address.
   * @param _to The to address.
   * @param _fee The fee paid for delivery.
   * @param _valueSent The value to be sent when delivering.
   * @param _messageNumber The unique message number.
   * @param _calldata The calldata to be passed to the destination address.
   */
  function _hashMessage(
    address _from,
    address _to,
    uint256 _fee,
    uint256 _valueSent,
    uint256 _messageNumber,
    bytes memory _calldata
  ) internal pure returns (bytes32 messageHash) {
    assembly {
      let mPtr := mload(0x40)
      mstore(mPtr, _from)
      mstore(add(mPtr, 0x20), _to)
      mstore(add(mPtr, 0x40), _fee)
      mstore(add(mPtr, 0x60), _valueSent)
      mstore(add(mPtr, 0x80), _messageNumber)
      mstore(add(mPtr, 0xa0), 0xc0)
      let dataLen := mload(_calldata)
      mstore(add(mPtr, 0xc0), dataLen)

      let rem := mod(dataLen, 0x20)
      let extra := 0
      if iszero(iszero(rem)) {
        extra := sub(0x20, rem)
      }

      // Copy the actual bytes from _calldata (skipping the 32-byte length prefix)
      let dataPtr := add(_calldata, 0x20)
      let destPtr := add(mPtr, 0xe0)
      for {
        let i := 0
      } lt(i, dataLen) {
        i := add(i, 0x20)
      } {
        mstore(add(destPtr, i), mload(add(dataPtr, i)))
      }

      messageHash := keccak256(mPtr, add(0xe0, add(dataLen, extra)))
    }
  }
}
