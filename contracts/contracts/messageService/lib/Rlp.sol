// SPDX-License-Identifier: Apache-2.0

/**
 * @author Hamdi Allam hamdi.allam97@gmail.com
 * @notice Please reach out with any questions or concerns.
 */
pragma solidity 0.8.19;

error NotList();
error WrongBytesLength();
error NoNext();
error MemoryOutOfBounds(uint256 inde);

library RLPReader {
  uint8 internal constant STRING_SHORT_START = 0x80;
  uint8 internal constant STRING_LONG_START = 0xb8;
  uint8 internal constant LIST_SHORT_START = 0xc0;
  uint8 internal constant LIST_LONG_START = 0xf8;
  uint8 internal constant LIST_SHORT_START_MAX = 0xf7;
  uint8 internal constant WORD_SIZE = 32;

  struct RLPItem {
    uint256 len;
    uint256 memPtr;
  }

  struct Iterator {
    RLPItem item; // Item that's being iterated over.
    uint256 nextPtr; // Position of the next item in the list.
  }

  /**
   * @dev Returns the next element in the iteration. Reverts if it has no next element.
   * @param _self The iterator.
   * @return nextItem The next element in the iteration.
   */
  function _next(Iterator memory _self) internal pure returns (RLPItem memory nextItem) {
    if (!_hasNext(_self)) {
      revert NoNext();
    }

    uint256 ptr = _self.nextPtr;
    uint256 itemLength = _itemLength(ptr);
    _self.nextPtr = ptr + itemLength;

    nextItem.len = itemLength;
    nextItem.memPtr = ptr;
  }

  /**
   * @dev Returns the number 'skiptoNum' element in the iteration.
   * @param _self The iterator.
   * @param _skipToNum Element position in the RLP item iterator to return.
   * @return item The number 'skipToNum' element in the iteration.
   */
  function _skipTo(Iterator memory _self, uint256 _skipToNum) internal pure returns (RLPItem memory item) {
    uint256 lenX;
    uint256 memPtrStart = _self.item.memPtr;
    uint256 endPtr;
    uint256 byte0;
    uint256 byteLen;

    assembly {
      // get first byte to know if it is a short/long list
      byte0 := byte(0, mload(memPtrStart))

      // yul has no if/else so if it a short list ( < long list start )
      switch lt(byte0, LIST_LONG_START)
      case 1 {
        // the length is just the difference in bytes
        lenX := sub(byte0, 0xc0)
      }
      case 0 {
        // at this point we care only about lists, so this is the default
        // get how many next bytes indicate the list length
        byteLen := sub(byte0, 0xf7)

        // move one over to the list length start
        memPtrStart := add(memPtrStart, 1)

        // shift over grabbing the bytelen elements
        lenX := div(mload(memPtrStart), exp(256, sub(32, byteLen)))
      }

      // get the end
      endPtr := add(memPtrStart, lenX)
    }

    uint256 ptr = _self.nextPtr;
    uint256 itemLength = _itemLength(ptr);
    _self.nextPtr = ptr + itemLength;

    for (uint256 i; i < _skipToNum - 1; ) {
      ptr = _self.nextPtr;
      if (ptr > endPtr) revert MemoryOutOfBounds(endPtr);
      itemLength = _itemLength(ptr);
      _self.nextPtr = ptr + itemLength;

      unchecked {
        i++;
      }
    }

    item.len = itemLength;
    item.memPtr = ptr;
  }

  /**
   * @dev Returns true if the iteration has more elements.
   * @param _self The iterator.
   * @return True if the iteration has more elements.
   */
  function _hasNext(Iterator memory _self) internal pure returns (bool) {
    RLPItem memory item = _self.item;
    return _self.nextPtr < item.memPtr + item.len;
  }

  /**
   * @param item RLP encoded bytes.
   * @return newItem The RLP item.
   */
  function _toRlpItem(bytes memory item) internal pure returns (RLPItem memory newItem) {
    uint256 memPtr;

    assembly {
      memPtr := add(item, 0x20)
    }

    newItem.len = item.length;
    newItem.memPtr = memPtr;
  }

  /**
   * @dev Creates an iterator. Reverts if item is not a list.
   * @param _self The RLP item.
   * @return iterator 'Iterator' over the item.
   */
  function _iterator(RLPItem memory _self) internal pure returns (Iterator memory iterator) {
    if (!_isList(_self)) {
      revert NotList();
    }

    uint256 ptr = _self.memPtr + _payloadOffset(_self.memPtr);
    iterator.item = _self;
    iterator.nextPtr = ptr;
  }

  /**
   * @param _item The RLP item.
   * @return (memPtr, len) Tuple: Location of the item's payload in memory.
   */
  function _payloadLocation(RLPItem memory _item) internal pure returns (uint256, uint256) {
    uint256 offset = _payloadOffset(_item.memPtr);
    uint256 memPtr = _item.memPtr + offset;
    uint256 len = _item.len - offset; // data length
    return (memPtr, len);
  }

  /**
   * @param _item The RLP item.
   * @return Indicator whether encoded payload is a list.
   */
  function _isList(RLPItem memory _item) internal pure returns (bool) {
    if (_item.len == 0) return false;

    uint8 byte0;
    uint256 memPtr = _item.memPtr;
    assembly {
      byte0 := byte(0, mload(memPtr))
    }

    if (byte0 < LIST_SHORT_START) return false;
    return true;
  }

  /**
   * @param _item The RLP item.
   * @return result Returns the item as an address.
   */
  function _toAddress(RLPItem memory _item) internal pure returns (address) {
    // 1 byte for the length prefix
    if (_item.len != 21) {
      revert WrongBytesLength();
    }

    return address(uint160(_toUint(_item)));
  }

  /**
   * @param _item The RLP item.
   * @return result Returns the item as a uint256.
   */
  function _toUint(RLPItem memory _item) internal pure returns (uint256 result) {
    if (_item.len == 0 || _item.len > 33) {
      revert WrongBytesLength();
    }

    (uint256 memPtr, uint256 len) = _payloadLocation(_item);

    assembly {
      result := mload(memPtr)

      // Shfit to the correct location if neccesary.
      if lt(len, 32) {
        result := div(result, exp(256, sub(32, len)))
      }
    }
  }

  /**
   * @param _item The RLP item.
   * @return result Returns the item as bytes.
   */
  function _toBytes(RLPItem memory _item) internal pure returns (bytes memory result) {
    if (_item.len == 0) {
      revert WrongBytesLength();
    }

    (uint256 memPtr, uint256 len) = _payloadLocation(_item);
    result = new bytes(len);

    uint256 destPtr;
    assembly {
      destPtr := add(0x20, result)
    }

    _copy(memPtr, destPtr, len);
  }

  /*
   * Private Helpers
   */

  /**
   * @param _memPtr Item memory pointer.
   * @return Entire RLP item byte length.
   */
  function _itemLength(uint256 _memPtr) private pure returns (uint256) {
    uint256 itemLen;
    uint256 dataLen;
    uint256 byte0;
    assembly {
      byte0 := byte(0, mload(_memPtr))
    }

    if (byte0 < STRING_SHORT_START) itemLen = 1;
    else if (byte0 < STRING_LONG_START) itemLen = byte0 - STRING_SHORT_START + 1;
    else if (byte0 < LIST_SHORT_START) {
      assembly {
        let byteLen := sub(byte0, 0xb7) // # Of bytes the actual length is.
        _memPtr := add(_memPtr, 1) // Skip over the first byte.

        /* 32 byte word size */
        dataLen := div(mload(_memPtr), exp(256, sub(32, byteLen))) // Right shifting to get the len.
        itemLen := add(dataLen, add(byteLen, 1))
      }
    } else if (byte0 < LIST_LONG_START) {
      itemLen = byte0 - LIST_SHORT_START + 1;
    } else {
      assembly {
        let byteLen := sub(byte0, 0xf7)
        _memPtr := add(_memPtr, 1)

        dataLen := div(mload(_memPtr), exp(256, sub(32, byteLen))) // Right shifting to the correct length.
        itemLen := add(dataLen, add(byteLen, 1))
      }
    }

    return itemLen;
  }

  /**
   * @param _memPtr Item memory pointer.
   * @return Number of bytes until the data.
   */
  function _payloadOffset(uint256 _memPtr) private pure returns (uint256) {
    uint256 byte0;
    assembly {
      byte0 := byte(0, mload(_memPtr))
    }

    if (byte0 < STRING_SHORT_START) return 0;
    else if (byte0 < STRING_LONG_START || (byte0 >= LIST_SHORT_START && byte0 < LIST_LONG_START)) return 1;
    else if (byte0 < LIST_SHORT_START)
      // being explicit
      return byte0 - (STRING_LONG_START - 1) + 1;
    else return byte0 - (LIST_LONG_START - 1) + 1;
  }

  /**
   * @param _src Pointer to source.
   * @param _dest Pointer to destination.
   * @param _len Amount of memory to copy from the source.
   */
  function _copy(uint256 _src, uint256 _dest, uint256 _len) private pure {
    if (_len == 0) return;

    // copy as many word sizes as possible
    for (; _len >= WORD_SIZE; _len -= WORD_SIZE) {
      assembly {
        mstore(_dest, mload(_src))
      }

      _src += WORD_SIZE;
      _dest += WORD_SIZE;
    }

    if (_len > 0) {
      // Left over bytes. Mask is used to remove unwanted bytes from the word.
      uint256 mask = 256 ** (WORD_SIZE - _len) - 1;
      assembly {
        let srcpart := and(mload(_src), not(mask)) // Zero out src.
        let destpart := and(mload(_dest), mask) // Retrieve the bytes.
        mstore(_dest, or(destpart, srcpart))
      }
    }
  }
}
