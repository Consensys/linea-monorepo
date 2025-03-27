// SPDX-License-Identifier: Apache-2.0

/**
 * @title Library for RLP Encoding data.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
pragma solidity ^0.8.28;

/// @custom:attribution https://github.com/bakaoh/solidity-rlp-encode
/// @title RLPWriter
/// @author RLPWriter is a library for encoding Solidity types to RLP bytes. Adapted from Bakaoh's
///         RLPEncode library (https://github.com/bakaoh/solidity-rlp-encode) with
///         modifications to improve legibility and gas consumption.
library RlpWriter {
  function _encodeBytes(bytes memory _bytesIn) internal pure returns (bytes memory bytesOut) {
    if (_bytesIn.length == 1 && uint8(_bytesIn[0]) < 0x80) {
      return _bytesIn; // Return as-is for single-byte < 0x80
    }

    bytes memory lengthPrefix = _writeLength(_bytesIn.length, 128);
    uint256 prefixLen = lengthPrefix.length;
    uint256 dataLen = _bytesIn.length;
    uint256 totalLen = prefixLen + dataLen;

    assembly {
      // Allocate output buffer
      bytesOut := mload(0x40)
      mstore(bytesOut, totalLen)

      // Copy prefix
      let dest := add(bytesOut, 0x20)
      let src := add(lengthPrefix, 0x20)
      mcopy(dest, src, prefixLen)

      // Copy input bytes
      dest := add(dest, prefixLen)
      src := add(_bytesIn, 0x20)
      mcopy(dest, src, dataLen)

      // Move free memory pointer
      mstore(0x40, add(add(bytesOut, 0x20), totalLen))
    }
  }

  function _encodeUint(uint256 _uintIn) internal pure returns (bytes memory uintBytes) {
    uintBytes = _encodeBytes(_toBinary(_uintIn));
  }

  function _encodeAddress(address _addressIn) internal pure returns (bytes memory addressBytes) {
    bytes memory addressRaw = new bytes(20);
    assembly {
      mstore(add(addressRaw, 0x20), shl(96, _addressIn)) // store address left-aligned
    }
    addressBytes = _encodeBytes(addressRaw);
  }

  function _encodeString(string memory _stringIn) internal pure returns (bytes memory stringBytes) {
    stringBytes = _encodeBytes(bytes(_stringIn));
  }

  function _encodeBool(bool _boolIn) internal pure returns (bytes memory boolBytes) {
    boolBytes = new bytes(1);
    boolBytes[0] = (_boolIn ? bytes1(0x01) : bytes1(0x80));
  }

  function _encodeList(bytes[] memory _bytesToEncode) internal pure returns (bytes memory listBytes) {
    listBytes = _flatten(_bytesToEncode);
    listBytes = abi.encodePacked(_writeLength(listBytes.length, 192), listBytes);
  }

  /// @notice Encode integer in big endian binary form with no leading zeroes.
  /// @param _uintValue The integer to encode.
  /// @return binaryBytes RLP encoded bytes.
  function _toBinary(uint256 _uintValue) private pure returns (bytes memory binaryBytes) {
    assembly {
      let ptr := mload(0x40) // Get free memory pointer

      let i := 0
      // Scan for first non-zero byte from MSB (big-endian)
      for {} lt(i, 32) {
        i := add(i, 1)
      } {
        if iszero(and(shr(sub(248, mul(i, 8)), _uintValue), 0xff)) {
          continue
        }
        break
      }

      let length := sub(32, i) // Number of non-zero bytes
      binaryBytes := ptr
      mstore(binaryBytes, length)

      // Write stripped bytes
      for {
        let j := 0
      } lt(j, length) {
        j := add(j, 1)
      } {
        let shift := mul(sub(length, add(j, 1)), 8)
        let b := and(shr(shift, _uintValue), 0xff)
        mstore8(add(add(binaryBytes, 0x20), j), b)
      }

      // Move free memory pointer
      mstore(0x40, add(add(ptr, 0x20), length))
    }
  }

  function _writeLength(uint256 _itemLength, uint256 _offset) private pure returns (bytes memory lengthBytes) {
    assembly {
      // Start from free memory pointer
      lengthBytes := mload(0x40)

      switch lt(_itemLength, 56)
      case 1 {
        // Case: short length
        mstore8(add(lengthBytes, 0x20), add(_itemLength, _offset))
        mstore(lengthBytes, 1) // Set bytes length to 1
        mstore(0x40, add(lengthBytes, 0x21)) // Advance free memory pointer
      }
      default {
        // Case: long length
        let temp := _itemLength
        let lengthOfLength := 0

        for {} gt(temp, 0) {} {
          lengthOfLength := add(lengthOfLength, 1)
          temp := shr(8, temp)
        }

        // First byte: offset + 55 + lengthOfLength
        mstore8(add(lengthBytes, 0x20), add(add(lengthOfLength, _offset), 55))

        // Write big-endian bytes of _itemLength
        for {
          let i := 0
        } lt(i, lengthOfLength) {
          i := add(i, 1)
        } {
          let shift := mul(8, sub(lengthOfLength, add(i, 1)))
          let b := and(shr(shift, _itemLength), 0xff)
          mstore8(add(add(lengthBytes, 0x21), i), b)
        }

        let totalLen := add(lengthOfLength, 1)
        mstore(lengthBytes, totalLen) // Set bytes length
        mstore(0x40, add(add(lengthBytes, 0x20), totalLen)) // Advance free memory pointer
      }
    }
  }

  // @custom:attribution https://github.com/sammayo/solidity-rlp-encoder
  /// @notice Flattens a list of byte strings into one byte string.
  /// @dev mcopy is used for the Cancun EVM fork. See original for other forks.
  /// @param _bytesList List of byte strings to flatten.
  /// @return flattenedBytes The flattened byte string.
  function _flatten(bytes[] memory _bytesList) private pure returns (bytes memory flattenedBytes) {
    uint256 bytesListLength = _bytesList.length;
    if (bytesListLength == 0) {
      return new bytes(0);
    }

    uint256 flattenedBytesLength;
    uint256 reusableCounter;
    for (; reusableCounter < bytesListLength; reusableCounter++) {
      unchecked {
        flattenedBytesLength += _bytesList[reusableCounter].length;
      }
    }

    flattenedBytes = new bytes(flattenedBytesLength);

    uint256 flattenedPtr;
    assembly {
      flattenedPtr := add(flattenedBytes, 0x20)
    }

    bytes memory item;
    uint256 itemLength;

    for (reusableCounter = 0; reusableCounter < bytesListLength; reusableCounter++) {
      item = _bytesList[reusableCounter];
      itemLength = item.length;
      assembly {
        mcopy(flattenedPtr, add(item, 0x20), itemLength)
        flattenedPtr := add(flattenedPtr, itemLength)
      }
    }
  }
}
