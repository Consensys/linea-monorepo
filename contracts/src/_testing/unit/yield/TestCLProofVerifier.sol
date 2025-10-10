// SPDX-License-Identifier: GPL-3.0
pragma solidity 0.8.30;

import { CLProofVerifier } from "../../../yield/libs/CLProofVerifier.sol";
import { GIndex } from "../../../yield/libs/vendor/lido/GIndex.sol";

contract TestCLProofVerifier is CLProofVerifier {
  constructor(
    GIndex _gIFirstValidatorPrev,
    GIndex _gIFirstValidatorCurr,
    uint64 _pivotSlot
  ) CLProofVerifier(_gIFirstValidatorPrev, _gIFirstValidatorCurr, _pivotSlot) {}

  function validateValidatorContainerForPermissionlessUnstake(
    ValidatorWitness memory _witness,
    bytes32 _withdrawalCredentials
  ) external view {
    _validateValidatorContainerForPermissionlessUnstake(_witness, _withdrawalCredentials);
  }

  function verifySlot(ValidatorWitness memory _witness) external view {
    _verifySlot(_witness);
  }

  function validateActivationEpoch(ValidatorWitness memory _witness) external view {
    _validateActivationEpoch(_witness);
  }

  function getValidatorGI(uint256 _offset, uint64 _provenSlot) external view returns (GIndex) {
    return _getValidatorGI(_offset, _provenSlot);
  }

  function getParentBlockRoot(uint64 _childBlockTimestamp) external view returns (bytes32) {
    return _getParentBlockRoot(_childBlockTimestamp);
  }
}
