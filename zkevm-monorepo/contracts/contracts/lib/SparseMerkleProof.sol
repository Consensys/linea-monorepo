// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.24;

import { Mimc } from "./Mimc.sol";

library SparseMerkleProof {
  using Mimc for *;

  struct Account {
    uint64 nonce;
    uint256 balance;
    bytes32 storageRoot;
    bytes32 mimcCodeHash;
    bytes32 keccakCodeHash;
    uint64 codeSize;
  }

  struct Leaf {
    uint256 prev;
    uint256 next;
    bytes32 hKey;
    bytes32 hValue;
  }

  error WrongBytesLength(uint256 expectedLength, uint256 bytesLength);

  uint256 internal constant TREE_DEPTH = 40;
  bytes32 internal constant ZERO_HASH = 0x0;

  /**
   * @notice Format input and verify sparse merkle proof
   * @param _rawProof Raw sparse merkle tree proof
   * @param _leafIndex Index of the leaf
   * @param _root Sparse merkle root
   */
  function verifyProof(bytes[] calldata _rawProof, uint256 _leafIndex, bytes32 _root) external pure returns (bool) {
    (bytes32 nextFreeNode, bytes32 leafHash, bytes32[] memory proof) = _formatProof(_rawProof);
    return _verify(proof, leafHash, _leafIndex, _root, nextFreeNode);
  }

  /**
   * @notice Hash a value using MIMC hash
   * @param _input Value to hash
   * @return {bytes32} Mimc hash
   */
  function mimcHash(bytes calldata _input) external pure returns (bytes32) {
    return Mimc.hash(_input);
  }

  /**
   * @notice Get leaf
   * @param _encodedLeaf Encoded leaf bytes (prev, next, hKey, hValue)
   * @return Leaf Formatted leaf struct
   */
  function getLeaf(bytes calldata _encodedLeaf) external pure returns (Leaf memory) {
    return _parseLeaf(_encodedLeaf);
  }

  /**
   * @notice Get account
   * @param _encodedAccountValue Encoded account value bytes (nonce, balance, storageRoot, mimcCodeHash, keccakCodeHash, codeSize)
   * @return Account Formatted account struct
   */
  function getAccount(bytes calldata _encodedAccountValue) external pure returns (Account memory) {
    return _parseAccount(_encodedAccountValue);
  }

  /**
   * @notice Hash account value
   * @param _value Encoded account value bytes (nonce, balance, storageRoot, mimcCodeHash, keccakCodeHash, codeSize)
   * @return {bytes32} Account value hash
   */
  function hashAccountValue(bytes calldata _value) external pure returns (bytes32) {
    Account memory account = _parseAccount(_value);
    (bytes32 msb, bytes32 lsb) = _splitBytes32(account.keccakCodeHash);
    return
      Mimc.hash(
        abi.encode(
          account.nonce,
          account.balance,
          account.storageRoot,
          account.mimcCodeHash,
          lsb,
          msb,
          account.codeSize
        )
      );
  }

  /**
   * @notice Hash storage value
   * @param _value Encoded storage value bytes
   * @return {bytes32} Storage value hash
   */
  function hashStorageValue(bytes32 _value) external pure returns (bytes32) {
    (bytes32 msb, bytes32 lsb) = _splitBytes32(_value);
    return Mimc.hash(abi.encodePacked(lsb, msb));
  }

  /**
   * @notice Parse leaf value
   * @param _encodedLeaf Encoded leaf bytes (prev, next, hKey, hValue)
   * @return {Leaf} Formatted leaf struct
   */
  function _parseLeaf(bytes calldata _encodedLeaf) private pure returns (Leaf memory) {
    if (_encodedLeaf.length < 128) {
      revert WrongBytesLength(128, _encodedLeaf.length);
    }
    return abi.decode(_encodedLeaf, (Leaf));
  }

  /**
   * @notice Parse account value
   * @param _value Encoded account value bytes (nonce, balance, storageRoot, mimcCodeHash, keccakCodeHash, codeSize)
   * @return {Account} Formatted account struct
   */
  function _parseAccount(bytes calldata _value) private pure returns (Account memory) {
    if (_value.length < 192) {
      revert WrongBytesLength(192, _value.length);
    }
    return abi.decode(_value, (Account));
  }

  /**
   * @notice Split bytes32 into two bytes32 taking most significant bits and least significant bits
   * @param _b bytes to split
   * @return msb Most significant bits
   * @return lsb Least significant bits
   */
  function _splitBytes32(bytes32 _b) private pure returns (bytes32 msb, bytes32 lsb) {
    assembly {
      msb := shr(128, _b)
      lsb := and(_b, 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF)
    }
  }

  /**
   * @notice Format proof
   * @param _rawProof Non formatted proof array
   * @return (bytes32, bytes32, bytes32[]) NextFreeNode, leafHash and formatted proof array
   */
  function _formatProof(bytes[] calldata _rawProof) private pure returns (bytes32, bytes32, bytes32[] memory) {
    uint256 rawProofLength = _rawProof.length;
    uint256 formattedProofLength = rawProofLength - 2;

    bytes32[] memory proof = new bytes32[](formattedProofLength);
    bytes32 nextFreeNode = bytes32(_rawProof[0][:32]);
    bytes32 leafHash = Mimc.hash(_rawProof[rawProofLength - 1]);

    for (uint256 i = 1; i < formattedProofLength; ) {
      proof[formattedProofLength - i] = Mimc.hash(_rawProof[i]);
      unchecked {
        ++i;
      }
    }

    // If the sibling leaf (_rawProof[formattedProofLength]) is equal to zero bytes we don't hash it
    if (_isZeroBytes(_rawProof[formattedProofLength])) {
      proof[0] = ZERO_HASH;
    } else {
      proof[0] = Mimc.hash(_rawProof[formattedProofLength]);
    }

    return (nextFreeNode, leafHash, proof);
  }

  /**
   * @notice Check if bytes contain only zero byte elements
   * @param _data Bytes to be checked
   * @return isZeroBytes true if bytes contain only zero byte elements, false otherwise
   */
  function _isZeroBytes(bytes calldata _data) private pure returns (bool isZeroBytes) {
    isZeroBytes = true;
    assembly {
      let dataStart := _data.offset

      for {
        let currentPtr := dataStart
      } lt(currentPtr, add(dataStart, _data.length)) {
        currentPtr := add(currentPtr, 0x20)
      } {
        let dataWord := calldataload(currentPtr)

        if eq(iszero(dataWord), 0) {
          isZeroBytes := 0
          break
        }
      }
    }
  }

  /**
   * @notice Verify sparse merkle proof
   * @param _proof Sparse merkle tree proof
   * @param _leafHash Leaf hash
   * @param _leafIndex Index of the leaf
   * @param _root Sparse merkle root
   * @param _nextFreeNode Next free node
   */
  function _verify(
    bytes32[] memory _proof,
    bytes32 _leafHash,
    uint256 _leafIndex,
    bytes32 _root,
    bytes32 _nextFreeNode
  ) private pure returns (bool) {
    bytes32 computedHash = _leafHash;
    uint256 currentIndex = _leafIndex;

    for (uint256 height; height < TREE_DEPTH; ++height) {
      if ((currentIndex >> height) & 1 == 1) computedHash = Mimc.hash(abi.encodePacked(_proof[height], computedHash));
      else computedHash = Mimc.hash(abi.encodePacked(computedHash, _proof[height]));
    }

    return Mimc.hash(abi.encodePacked(_nextFreeNode, computedHash)) == _root;
  }
}
