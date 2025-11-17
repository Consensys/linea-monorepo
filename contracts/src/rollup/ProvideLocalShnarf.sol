// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { LineaRollupBase } from "./LineaRollupBase.sol";

abstract contract ProvideLocalShnarf is LineaRollupBase {
  function blobShnarfExists(bytes32 _shnarf) public view returns (uint256 shnarfExists) {
    shnarfExists = _blobShnarfExists[_shnarf];
  }
}
