// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Base64 } from "@openzeppelin/contracts/utils/Base64.sol";
import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";
import { BaseNFTMetadataGenerator } from "./BaseNFTMetadataGenerator.sol";

/**
 * @title NFTMetadataGeneratorSVG
 * @notice NFT metadata generator contract that generates SVG image URIs.
 * @dev Generates SVG image URIs with a balance number.
 */
contract NFTMetadataGeneratorSVG is BaseNFTMetadataGenerator {
    /// @notice SVG image prefix
    string public imagePrefix = "";
    /// @notice SVG image suffix
    string public imageSuffix = "";

    constructor(string memory _imagePrefix, string memory _imageSuffix) BaseNFTMetadataGenerator() {
        imagePrefix = _imagePrefix;
        imageSuffix = _imageSuffix;
    }

    /**
     * @notice Sets the SVG image prefix and suffix
     * @dev Only the owner can call this function
     * @param _imagePrefix The SVG image prefix
     * @param _imageSuffix The SVG image suffix
     */
    function setImageStrings(string memory _imagePrefix, string memory _imageSuffix) external onlyOwner {
        imagePrefix = _imagePrefix;
        imageSuffix = _imageSuffix;
    }

    /**
     * @notice Generates the image URI for the NFT based on the owner's address and balance
     * @param balance The balance of the NFT owner
     */
    function generateImageURI(address, uint256 balance) internal view override returns (string memory) {
        string memory text = Strings.toString(balance / 1e18);
        bytes memory svg = abi.encodePacked(imagePrefix, text, imageSuffix);

        return string(abi.encodePacked("data:image/svg+xml;base64,", Base64.encode(svg)));
    }
}
