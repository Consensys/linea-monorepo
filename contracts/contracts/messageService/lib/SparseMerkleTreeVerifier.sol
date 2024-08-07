// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.24;

import { Utils } from "../../lib/Utils.sol";

/**
 * @title Library to verify sparse merkle proofs and to get the leaf hash value
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
library SparseMerkleTreeVerifier {
  using Utils for *;

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
        node = Utils._efficientKeccak(_proof[height], node);
      } else {
        node = Utils._efficientKeccak(node, _proof[height]);
      }
    }
    return node == _root;
  }
}
