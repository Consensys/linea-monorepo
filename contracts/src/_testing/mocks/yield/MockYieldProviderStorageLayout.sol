// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.0;

import { YieldManagerStorageLayout } from "../../../yield/YieldManagerStorageLayout.sol";

abstract contract MockYieldProviderStorageLayout {
  /// @dev keccak256(abi.encode(uint256(keccak256("linea.storage.MockYieldProviderStorage")) - 1)) & ~bytes32(uint256(0xff))
  bytes32 private constant MockYieldProviderStorageLocation =
    0x6872667b28c0553451b98c9cc79a262d1e6603e3af2375dec0299e2db861a700;

  struct TestYieldManagerStorage {
    mapping (address => MockYieldProviderStorage) _mockYieldProviderStorage;
  }

  struct MockYieldProviderStorage {
    uint256 withdrawableValueReturnVal;
    uint256 reportYieldReturnVal;
    uint256 payLSTPrincipalReturnVal;
    uint256 unstakePermissionlessReturnVal;
  }

  function _getMockYieldProviderStorage(address _yieldProvider) internal view returns (MockYieldProviderStorage storage) {
    TestYieldManagerStorage storage $;
    assembly {
      $.slot := MockYieldProviderStorageLocation
    }
    return $._mockYieldProviderStorage[_yieldProvider];
  }

  /*//////////////////////////////////////////////////////////////
                            MOCK HELPERS
  //////////////////////////////////////////////////////////////*/

  function setWithdrawableValueReturnVal(address _yieldProvider, uint256 _val) external {
    _getMockYieldProviderStorage(_yieldProvider).withdrawableValueReturnVal = _val;
  }

  function setReportYieldReturnVal(address _yieldProvider, uint256 _val) external {
    _getMockYieldProviderStorage(_yieldProvider).reportYieldReturnVal = _val;
  }

  function setPayLSTPrincipalReturnVal(address _yieldProvider, uint256 _val) external {
    _getMockYieldProviderStorage(_yieldProvider).payLSTPrincipalReturnVal = _val;
  }

  function setUnstakePermissionlessReturnVal(address _yieldProvider, uint256 _val) external {
    _getMockYieldProviderStorage(_yieldProvider).unstakePermissionlessReturnVal = _val;
  }
}