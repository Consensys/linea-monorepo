// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";
import { BaseNFTMetadataGenerator } from "../../src/nft-metadata-generators/BaseNFTMetadataGenerator.sol";

contract MockMetadataGenerator is BaseNFTMetadataGenerator {
    string private _baseURI;

    constructor(string memory baseURI) BaseNFTMetadataGenerator() {
        _baseURI = baseURI;
    }

    function generateImageURI(address account, uint256) internal view override returns (string memory) {
        bytes memory uri = abi.encodePacked(_baseURI, Strings.toHexString(account));
        return string(uri);
    }
}
