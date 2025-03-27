// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

interface INFTMetadataGenerator {
    function generate(address account, uint256 balance) external view returns (string memory);
}
