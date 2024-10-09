// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Base64 } from "@openzeppelin/contracts/utils/Base64.sol";
import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { BaseNFTMetadataGenerator } from "./BaseNFTMetadataGenerator.sol";

contract NFTMetadataGeneratorSVG is BaseNFTMetadataGenerator {
    string public imagePrefix = "";
    string public imageSuffix = "";

    constructor(address nft, string memory _imagePrefix, string memory _imageSuffix) BaseNFTMetadataGenerator(nft) {
        imagePrefix = _imagePrefix;
        imageSuffix = _imageSuffix;
    }

    function setImageStrings(string memory _imagePrefix, string memory _imageSuffix) external onlyOwner {
        imagePrefix = _imagePrefix;
        imageSuffix = _imageSuffix;
    }

    function generateImageURI(address, uint256 balance) internal view override returns (string memory) {
        string memory text = Strings.toString(balance / 1e18);
        bytes memory svg = abi.encodePacked(imagePrefix, text, imageSuffix);

        return string(abi.encodePacked("data:image/svg+xml;base64,", Base64.encode(svg)));
    }
}
