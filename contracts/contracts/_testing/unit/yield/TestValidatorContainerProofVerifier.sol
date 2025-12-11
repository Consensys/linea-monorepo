// SPDX-License-Identifier: GPL-3.0
pragma solidity 0.8.30;

import { ValidatorContainerProofVerifier } from "../../../yield/libs/ValidatorContainerProofVerifier.sol";
import { GIndex } from "../../../yield/libs/vendor/lido/GIndex.sol";
import { IValidatorContainerProofVerifier } from "../../../yield/interfaces/IValidatorContainerProofVerifier.sol";

contract TestValidatorContainerProofVerifier is ValidatorContainerProofVerifier {
  constructor(
    GIndex _gIFirstValidatorPrev,
    GIndex _gIFirstValidatorCurr,
    uint64 _pivotSlot,
    GIndex _gIPendingPartialWithdrawalsRoot
  ) ValidatorContainerProofVerifier(_gIFirstValidatorPrev, _gIFirstValidatorCurr, _pivotSlot, _gIPendingPartialWithdrawalsRoot) {}

  function verifySlot(IValidatorContainerProofVerifier.ValidatorContainerWitness calldata _witness) external view {
    _verifySlot(_witness.slot, _witness.proposerIndex, _witness.proof);
  }

  function validateActivationEpoch(
    IValidatorContainerProofVerifier.ValidatorContainerWitness calldata _witness
  ) external view {
    _validateActivationEpoch(_witness);
  }

  function getValidatorGI(uint256 _offset, uint64 _provenSlot) external view returns (GIndex) {
    return _getValidatorGI(_offset, _provenSlot);
  }

  function getParentBlockRoot(uint64 _childBlockTimestamp) external view returns (bytes32) {
    return _getParentBlockRoot(_childBlockTimestamp);
  }
}
