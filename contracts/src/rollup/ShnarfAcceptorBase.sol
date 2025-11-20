// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.30;

import { IShnarfAcceptorBase } from "./interfaces/IShnarfAcceptorBase.sol";
import { LineaRollupBase } from "./LineaRollupBase.sol";

abstract contract ShnarfAcceptorBase is LineaRollupBase, IShnarfAcceptorBase {
  /// @dev Value indicating a shnarf exists.
  uint256 internal constant SHNARF_EXISTS_DEFAULT_VALUE = 1;

  /**
   * @notice Accepts and stores that a shnarf exists.
   * @dev OPERATOR_ROLE is required to execute.
   * @param _parentShnarf The parent shnarf.
   * @param _shnarf The shnarf to indicate exists.
   * @param _finalStateRootHash The final state root hash in the data.
   */
  function _acceptShnarfInfo(bytes32 _parentShnarf, bytes32 _shnarf, bytes32 _finalStateRootHash) internal virtual {
    if (_blobShnarfExists[_parentShnarf] == 0) {
      revert ParentShnarfNotSubmitted(_parentShnarf);
    }

    if (_blobShnarfExists[_shnarf] != 0) {
      revert ShnarfAlreadySubmitted(_shnarf);
    }

    _blobShnarfExists[_shnarf] = SHNARF_EXISTS_DEFAULT_VALUE;

    emit DataSubmittedV3(_parentShnarf, _shnarf, _finalStateRootHash);
  }
}
