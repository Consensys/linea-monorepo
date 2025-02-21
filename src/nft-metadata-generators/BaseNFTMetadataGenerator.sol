// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Base64 } from "@openzeppelin/contracts/utils/Base64.sol";
import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { INFTMetadataGenerator } from "../interfaces/INFTMetadataGenerator.sol";

abstract contract BaseNFTMetadataGenerator is INFTMetadataGenerator, Ownable {
    address public nft;

    constructor(address _nft) Ownable(msg.sender) {
        nft = _nft;
    }

    function generate(address account, uint256 balance) external view returns (string memory) {
        string memory baseName = "KarmaNFT ";
        string memory baseDescription = "This is a KarmaNFT for address ";

        string memory propName = string(abi.encodePacked(baseName, Strings.toHexString(account)));
        string memory propDescription = string(
            abi.encodePacked(baseDescription, Strings.toHexString(account), " with balance ", Strings.toString(balance))
        );

        string memory image = generateImageURI(account, balance);

        bytes memory json = abi.encodePacked(
            "{\"name\":\"", propName, "\",\"description\":\"", propDescription, "\",\"image\":\"", image, "\"}"
        );

        string memory jsonBase64 = Base64.encode(json);
        return string(abi.encodePacked("data:application/json;base64,", jsonBase64));
    }

    function generateImageURI(address account, uint256 balance) internal view virtual returns (string memory);
}
