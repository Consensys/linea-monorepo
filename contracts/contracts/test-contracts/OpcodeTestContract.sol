// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

contract OpcodeTestContract {
  // bytes32(uint256(keccak256("opcodeTestContract.gasLimit")) - 1)
  bytes32 private constant GAS_LIMIT_SLOT = 0x8a6969ba29f186c962469e088de2cadf70bf152ef5985279682dcca1927a9240;

  function getGasLimit() external view returns (uint256) {
    uint256 gasLimit;
    assembly {
      gasLimit := sload(GAS_LIMIT_SLOT)
    }
    return gasLimit;
  }

  function setGasLimit() external {
    assembly {
      let gasLimit := gaslimit()
      sstore(GAS_LIMIT_SLOT, gasLimit)
    }
  }
}
