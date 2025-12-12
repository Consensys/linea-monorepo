// SPDX-License-Identifier: GPL-3.0
pragma solidity 0.8.30;

import { ValidatorContainerProofVerifier } from "../../../yield/libs/ValidatorContainerProofVerifier.sol";
import { GIndex } from "../../../yield/libs/vendor/lido/GIndex.sol";
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

  function getParentBlockRoot(uint64 _childBlockTimestamp) external view returns (bytes32) {
    return _getParentBlockRoot(_childBlockTimestamp);
  }
}
