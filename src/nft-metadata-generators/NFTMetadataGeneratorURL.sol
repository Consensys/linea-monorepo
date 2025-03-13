// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";
import { BaseNFTMetadataGenerator } from "./BaseNFTMetadataGenerator.sol";

/**
 * @title NFTMetadataGeneratorURL
 * @notice NFT metadata generator contract that generates image URI based on account address.
 */
contract NFTMetadataGeneratorURL is BaseNFTMetadataGenerator {
    /// @notice URL prefix
    string public urlPrefix;
    /// @notice URL suffix
    string public urlSuffix;

    constructor(string memory _urlPrefix, string memory _urlSuffix) BaseNFTMetadataGenerator() {
        urlPrefix = _urlPrefix;
        urlSuffix = _urlSuffix;
    }

    /**
     * @notice Sets the URL prefix and suffix
     * @dev Only the owner can call this function
     * @param _urlPrefix The URL prefix
     * @param _urlSuffix The URL suffix
     */
    function setURLStrings(string memory _urlPrefix, string memory _urlSuffix) external onlyOwner {
        urlPrefix = _urlPrefix;
        urlSuffix = _urlSuffix;
    }

    /**
     * @notice Generates the image URI for the NFT based on the owner's address and balance
     * @param account The address of the NFT owner
     */
    function generateImageURI(address account, uint256) internal view override returns (string memory, string memory) {
        return ("image", string(abi.encodePacked(urlPrefix, Strings.toHexString(account), urlSuffix)));
    }
}
