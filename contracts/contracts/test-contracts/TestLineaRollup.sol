// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.33;

import { LineaRollup } from "../LineaRollup.sol";
import { Utils } from "../lib/Utils.sol";

/// @custom:oz-upgrades-unsafe-allow missing-initializer
contract TestLineaRollup is LineaRollup {
  function setFallbackOperatorAddress(address _fallbackOperator) external {
    fallbackOperator = _fallbackOperator;
  }
  
  function addRollingHash(uint256 _messageNumber, bytes32 _messageHash) external {
    _addRollingHash(_messageNumber, _messageHash);
  }

  function setRollingHash(uint256 _messageNumber, bytes32 _rollingHash) external {
    rollingHashes[_messageNumber] = _rollingHash;
  }

  function setLastTimeStamp(uint256 _timestamp) external {
    currentTimestamp = _timestamp;
  }

  function validateL2ComputedRollingHash(uint256 _rollingHashMessageNumber, bytes32 _rollingHash) external view {
    _validateL2ComputedRollingHash(_rollingHashMessageNumber, _rollingHash);
  }

  function calculateY(bytes calldata _data, bytes32 _x) external pure returns (bytes32 y) {
    return _calculateY(_data, _x);
  }

  function setupParentShnarf(bytes32 _shnarf) external {
    blobShnarfExists[_shnarf] = 1;
  }

  function setupParentDataShnarf(bytes32 _parentDataHash, bytes32 _shnarf) external {
    dataShnarfHashes[_parentDataHash] = _shnarf;
  }

  function setLastFinalizedBlock(uint256 _blockNumber) external {
    currentL2BlockNumber = _blockNumber;
  }

  function setupParentFinalizedStateRoot(bytes32 _parentDataHash, bytes32 _blockStateRoot) external {
    dataFinalStateRootHashes[_parentDataHash] = _blockStateRoot;
  }

  function setupStartingBlockForDataHash(bytes32 _dataHash, uint256 _blockNumber) external {
    dataStartingBlock[_dataHash] = _blockNumber;
  }

  function setLastFinalizedShnarf(bytes32 _lastFinalizedShnarf) external {
    currentFinalizedShnarf = _lastFinalizedShnarf;
  }

  function setShnarfFinalBlockNumber(bytes32 _shnarf, uint256 _finalBlockNumber) external {
    blobShnarfExists[_shnarf] = _finalBlockNumber;
  }

  function addL2MerkleRoots(bytes32[] calldata _newRoot, uint256 _treeDepth) external onlyRole(DEFAULT_ADMIN_ROLE) {
    _addL2MerkleRoots(_newRoot, _treeDepth);
  }

  // Adapted from SparseMerkleTreeVerifier._verifyMerkleProof to generate passing proof + root
  function generateMerkleRoot(
    bytes32 _leafHash,
    bytes32[] calldata _proof,
    uint32 _leafIndex
  ) external pure returns (bytes32) {
    require(_leafIndex < uint32((2 ** _proof.length) - 1), "leafIndex out of bounds");

    bytes32 node = _leafHash;

    for (uint256 height; height < _proof.length; ++height) {
      if (((_leafIndex >> height) & 1) == 1) {
        node = Utils._efficientKeccak(_proof[height], node);
      } else {
        node = Utils._efficientKeccak(node, _proof[height]);
      }
    }

    return node;
  }
}
