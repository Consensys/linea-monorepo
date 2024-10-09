// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";
import { BaseNFTMetadataGenerator } from "./BaseNFTMetadataGenerator.sol";

contract NFTMetadataGeneratorURL is BaseNFTMetadataGenerator {
    string public urlPrefix;
    string public urlSuffix;

    constructor(address nft, string memory _urlPrefix, string memory _urlSuffix) BaseNFTMetadataGenerator(nft) {
        urlPrefix = _urlPrefix;
        urlSuffix = _urlSuffix;
    }

    function setURLStrings(string memory _urlPrefix, string memory _urlSuffix) external onlyOwner {
        urlPrefix = _urlPrefix;
        urlSuffix = _urlSuffix;
    }

    function generateImageURI(address account, uint256) internal view override returns (string memory) {
        return string(abi.encodePacked(urlPrefix, Strings.toHexString(account), urlSuffix));
    }
}
