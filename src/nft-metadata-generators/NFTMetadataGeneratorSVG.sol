// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Base64 } from "@openzeppelin/contracts/utils/Base64.sol";
import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { BaseNFTMetadataGenerator } from "./BaseNFTMetadataGenerator.sol";

contract NFTMetadataGeneratorSVG is BaseNFTMetadataGenerator {
    string private _imagePrefix = "";
    string private _imageSuffix = "";

    constructor(address nft, string memory imagePrefix, string memory imageSuffix) BaseNFTMetadataGenerator(nft) {
        _imagePrefix = imagePrefix;
        _imageSuffix = imageSuffix;
    }

    function setImageStrings(string memory imagePrefix, string memory imageSuffix) external onlyOwner {
        _imagePrefix = imagePrefix;
        _imageSuffix = imageSuffix;
    }

    function generateImageURI(address, uint256 balance) internal view override returns (string memory) {
        string memory text = Strings.toString(balance / 10e18);
        bytes memory svg = abi.encodePacked(_imagePrefix, text, _imageSuffix);

        return Base64.encode(svg);
    }
}
