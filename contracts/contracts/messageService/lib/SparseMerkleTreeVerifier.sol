// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.26;

import { Utils } from "../../lib/Utils.sol";

/**
 * @title Library to verify sparse merkle proofs and to get the leaf hash value
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
library SparseMerkleTreeVerifier {
  using Utils for *;

  /**
   * @dev Custom error for when the leaf index is out of bounds.
   */
  error LeafIndexOutOfBounds(uint32 leafIndex, uint32 maxAllowedIndex);

  /**
   * @notice Verify merkle proof
   * @param _leafHash Leaf hash.
   * @param _proof Sparse merkle tree proof.
   * @param _leafIndex Index of the leaf.
   * @param _root Merkle root.
   * @dev The depth of the tree is expected to be validated elsewhere beforehand.
   * @return proofIsValid Returns if the proof is valid or not.
   */
  function _verifyMerkleProof(
    bytes32 _leafHash,
    bytes32[] calldata _proof,
    uint32 _leafIndex,
    bytes32 _root
  ) internal pure returns (bool proofIsValid) {
    uint32 maxAllowedIndex = uint32((2 ** _proof.length) - 1);
    if (_leafIndex > maxAllowedIndex) {
      revert LeafIndexOutOfBounds(_leafIndex, maxAllowedIndex);
    }

    bytes32 node = _leafHash;

    for (uint256 height; height < _proof.length; ++height) {
      if (((_leafIndex >> height) & 1) == 1) {
        node = Utils._efficientKeccak(_proof[height], node);
      } else {
        node = Utils._efficientKeccak(node, _proof[height]);
      }
    }
    proofIsValid = node == _root;
  }
}
