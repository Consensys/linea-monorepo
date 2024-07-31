// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.24;

import { SparseMerkleTreeVerifier } from "../messageService/lib/SparseMerkleTreeVerifier.sol";

contract TestSparseMerkleTreeVerifier {
  using SparseMerkleTreeVerifier for *;

  function verifyMerkleProof(
    bytes32 _leafHash,
    bytes32[] calldata _proof,
    uint32 _leafIndex,
    bytes32 _root
  ) external pure returns (bool) {
    return SparseMerkleTreeVerifier._verifyMerkleProof(_leafHash, _proof, _leafIndex, _root);
  }

  function efficientKeccak(bytes32 _left, bytes32 _right) external pure returns (bytes32 value) {
    return SparseMerkleTreeVerifier._efficientKeccak(_left, _right);
  }

  function getLeafHash(
    address _from,
    address _to,
    uint256 _fee,
    uint256 _value,
    uint256 _messageNumber,
    bytes calldata _calldata
  ) external pure returns (bytes32) {
    return keccak256(abi.encodePacked(_from, _to, _fee, _value, _messageNumber, _calldata));
  }
}
