// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { RLPReader } from "../messageService/lib/Rlp.sol";

using RLPReader for RLPReader.RLPItem;
using RLPReader for RLPReader.Iterator;
using RLPReader for bytes;

contract TestRlp {
  uint8 internal constant STRING_SHORT_START = 0x80;
  uint8 internal constant STRING_LONG_START = 0xb8;
  uint8 internal constant LIST_SHORT_START = 0xc0;
  uint8 internal constant LIST_LONG_START = 0xf8;
  uint8 internal constant WORD_SIZE = 32;

  function next(bytes memory _self) external pure returns (RLPReader.RLPItem memory nextItem, uint256 itemNextMemPtr) {
    RLPReader.RLPItem memory rlpItem = _self._toRlpItem();
    RLPReader.Iterator memory it = rlpItem._iterator();
    uint256 itemNextPtr = it.nextPtr;
    return (it._next(), itemNextPtr);
  }

  function hasNext(bytes memory _self) external pure returns (bool) {
    RLPReader.RLPItem memory rlpItem = _self._toRlpItem();
    RLPReader.Iterator memory it = rlpItem._iterator();
    return it._hasNext();
  }

  function skipTo(bytes memory _self, uint256 _skipToNum) external pure returns (RLPReader.RLPItem memory item) {
    RLPReader.RLPItem memory rlpItem = _self._toRlpItem();
    RLPReader.Iterator memory it = rlpItem._iterator();
    return it._skipTo(_skipToNum);
  }

  function skipToReturnBytes(bytes memory _self, uint256 _skipToNum) external pure returns (bytes memory item) {
    RLPReader.RLPItem memory rlpItem = _self._toRlpItem();
    RLPReader.Iterator memory it = rlpItem._iterator();
    return it._skipTo(_skipToNum)._toBytes();
  }

  function fromRlpItemToAddress(RLPReader.RLPItem memory _item) external pure returns (address) {
    return _item._toAddress();
  }

  function iterator(bytes memory _item) external pure {
    _item._toRlpItem()._iterator();
  }

  function payloadLocation(
    bytes memory _item
  ) external pure returns (uint256 ptr, uint256 itemlen, uint256 rlpItemPtr) {
    RLPReader.RLPItem memory rlpItem = _item._toRlpItem();
    (uint memPtr, uint len) = rlpItem._payloadLocation();
    return (memPtr, len, rlpItem.memPtr);
  }

  function isList(bytes memory _item) external pure returns (bool) {
    RLPReader.RLPItem memory rlpItem = _item._toRlpItem();
    return rlpItem._isList();
  }

  function toAddress(bytes memory _item) external pure returns (address) {
    RLPReader.RLPItem memory rlpItem = _item._toRlpItem();
    return rlpItem._toAddress();
  }

  function toUint(bytes memory _item) external pure returns (uint256 result) {
    RLPReader.RLPItem memory rlpItem = _item._toRlpItem();
    return rlpItem._toUint();
  }

  function toBytes(bytes memory _item) external pure returns (bytes memory result) {
    RLPReader.RLPItem memory rlpItem = _item._toRlpItem();
    return rlpItem._toBytes();
  }

  function itemLength(bytes memory item) external pure returns (uint) {
    uint memPtr;
    assembly {
      memPtr := add(item, 0x20)
    }

    return _itemLength(memPtr);
  }

  function _itemLength(uint256 _memPtr) private pure returns (uint256) {
    uint256 itemLen;
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
        let dataLen := div(mload(_memPtr), exp(256, sub(32, byteLen))) // Right shifting to get the len.
        itemLen := add(dataLen, add(byteLen, 1))
      }
    } else if (byte0 < LIST_LONG_START) {
      itemLen = byte0 - LIST_SHORT_START + 1;
    } else {
      assembly {
        let byteLen := sub(byte0, 0xf7)
        _memPtr := add(_memPtr, 1)

        let dataLen := div(mload(_memPtr), exp(256, sub(32, byteLen))) // Right shifting to the correct length.
        itemLen := add(dataLen, add(byteLen, 1))
      }
    }

    return itemLen;
  }
}
