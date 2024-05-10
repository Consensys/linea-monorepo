// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.24;

/**
 * @title Library to verify sparse merkle proofs and to get the leaf hash value
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
library SparseMerkleTreeVerifier {
  /**
   * @notice Verify merkle proof
   * @param _leafHash Leaf hash.
   * @param _proof Sparse merkle tree proof.
   * @param _leafIndex Index of the leaf.
   * @param _root Merkle root.
   */
  function _verifyMerkleProof(
    bytes32 _leafHash,
    bytes32[] calldata _proof,
    uint32 _leafIndex,
    bytes32 _root
  ) internal pure returns (bool) {
    bytes32 node = _leafHash;

    for (uint256 height; height < _proof.length; ++height) {
      if (((_leafIndex >> height) & 1) == 1) {
        node = _efficientKeccak(_proof[height], node);
      } else {
        node = _efficientKeccak(node, _proof[height]);
      }
    }
    return node == _root;
  }

  /**
   * @notice Performs a gas optimized keccak hash
   * @param _left Left value.
   * @param _right Right value.
   */
  function _efficientKeccak(bytes32 _left, bytes32 _right) internal pure returns (bytes32 value) {
    assembly {
      mstore(0x00, _left)
      mstore(0x20, _right)
      value := keccak256(0x00, 0x40)
    }
  }
}
