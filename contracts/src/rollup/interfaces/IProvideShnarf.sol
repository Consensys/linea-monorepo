// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.30;
interface IProvideShnarf {
  function blobShnarfExists(bytes32 _shnarf) external view returns (uint256 shnarfExists);
}
