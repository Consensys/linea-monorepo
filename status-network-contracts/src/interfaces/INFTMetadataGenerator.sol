// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

/**
 * @title INFTMetadataGenerator
 * @notice Interface for a contract that generates NFT metadata.
 */
interface INFTMetadataGenerator {
    /**
     * @notice Generates metadata for the NFT based on the owner's address and balance
     * @param account The address of the NFT owner
     * @param balance The balance of the NFT owner
     * @return A string representing the metadata URI for the NFT
     */
    function generate(address account, uint256 balance) external view returns (string memory);
}
