// SPDX-License-Identifier: GPL-3.0
pragma solidity 0.8.30;

import { ValidatorContainerProofVerifier } from "../../../yield/libs/ValidatorContainerProofVerifier.sol";
import { GIndex, pack } from "../../../yield/libs/vendor/lido/GIndex.sol";
import { IValidatorContainerProofVerifier } from "../../../yield/interfaces/IValidatorContainerProofVerifier.sol";

contract TestValidatorContainerProofVerifier is ValidatorContainerProofVerifier {
  constructor(
    address _admin,
    GIndex _gIFirstValidator,
    GIndex _gIPendingPartialWithdrawalsRoot
  ) ValidatorContainerProofVerifier(_admin, _gIFirstValidator, _gIPendingPartialWithdrawalsRoot) {}

  function verifySlot(
    uint64 _slot,
    uint64 _proposerIndex,
    bytes32[] calldata _proof
  ) external view {
    _verifySlot(_slot, _proposerIndex, _proof);
  }

  function validateActivationEpoch(
    uint64 _slot,
    uint64 _activationEpoch
  ) external view {
    _validateActivationEpoch(_slot, _activationEpoch);
  }

  function getValidatorGI(uint256 _offset) external view returns (GIndex) {
    return _getValidatorGI(_offset);
  }

  function getValidatorGIInLeftSubtree(uint256 _offset) external view returns (GIndex) {
    uint256 gIFirstValidatorIndex = GI_FIRST_VALIDATOR.index();
    uint8 pow = GI_FIRST_VALIDATOR.pow();
    // Halve GI_FIRST_VALIDATOR
    uint256 gIFirstValidatorIndexInLeftSubtree = gIFirstValidatorIndex / 2;
    GIndex gIFirstValidatorInLeftSubtree = pack(gIFirstValidatorIndexInLeftSubtree, pow - 1);
    return gIFirstValidatorInLeftSubtree.shr(_offset);
  }

  function getParentBlockRoot(uint64 _childBlockTimestamp) external view returns (bytes32) {
    return _getParentBlockRoot(_childBlockTimestamp);
  }
}
