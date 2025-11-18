// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.30;

import { IAcceptShnarfData } from "./interfaces/IAcceptShnarfData.sol";
import { LineaRollupBase } from "./LineaRollupBase.sol";

/**
 * @title Contract to manage cross-chain messaging on L1, L2 data submission, and rollup proof verification.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract ShnarfDataAcceptor is IAcceptShnarfData, LineaRollupBase {
  /**
   * @notice Accepts and stores that a shnarf exists.
   * @dev OPERATOR_ROLE is required to execute.
   * @param _parentShnarf The parent shnarf.
   * @param _shnarf The shnarf to indicate exists.
   * @param _finalStateRootHash The final state root hash in the data.
   */
  function acceptDataInfo(
    bytes32 _parentShnarf,
    bytes32 _shnarf,
    bytes32 _finalStateRootHash
  ) external virtual whenTypeAndGeneralNotPaused(PauseType.BLOB_SUBMISSION) onlyRole(OPERATOR_ROLE) {
    if (_blobShnarfExists[_parentShnarf] == 0) {
      revert ParentBlobNotSubmitted(_parentShnarf);
    }

    if (_blobShnarfExists[_shnarf] != 0) {
      revert DataAlreadySubmitted(_shnarf);
    }

    _blobShnarfExists[_shnarf] = SHNARF_EXISTS_DEFAULT_VALUE;

    emit DataSubmittedV3(_parentShnarf, _shnarf, _finalStateRootHash);
  }
}
