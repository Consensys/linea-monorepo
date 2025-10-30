// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.8.25;

import { SSZ } from "../../../../../yield/libs/vendor/lido/SSZ.sol";
import { Validator, BeaconBlockHeader } from "../../../../../yield/libs/vendor/lido/BeaconTypes.sol";
import { GIndex } from "../../../../../yield/libs/vendor/lido/GIndex.sol";

contract TestSSZ {
  error BranchHasMissingItem();
  error BranchHasExtraItem();
  error InvalidProof();

  function hashTreeRoot_BeaconBlockHeader(BeaconBlockHeader memory header) external view returns (bytes32) {
    return SSZ.hashTreeRoot(header);
  }

  function hashTreeRoot_Validator(Validator memory validator) external view returns (bytes32) {
    return SSZ.hashTreeRoot(validator);
  }

  function verifyProof(bytes32[] calldata proof, bytes32 root, bytes32 leaf, GIndex gI) external view {
    SSZ.verifyProof(proof, root, leaf, gI);
  }

  function toLittleEndianUint(uint256 v) external pure returns (bytes32) {
    return SSZ.toLittleEndian(v);
  }

  function toLittleEndianBool(bool v) external pure returns (bytes32) {
    return SSZ.toLittleEndian(v);
  }
}
