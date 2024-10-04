// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Base64 } from "@openzeppelin/contracts/utils/Base64.sol";
import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";

interface IMetadataGenerator {
    function generate(address account, uint256 balance) external view returns (string memory);
}

contract XPNFTMetadataGenerator is IMetadataGenerator, Ownable {
    address public nft;

    string private _imagePrefix = "";
    string private _imageSuffix = "";

    constructor(address _nft, string memory imagePrefix, string memory imageSuffix) Ownable(msg.sender) {
        nft = _nft;
        _imagePrefix = imagePrefix;
        _imageSuffix = imageSuffix;
    }

    function setImageStrings(string memory imagePrefix, string memory imageSuffix) external onlyOwner {
        _imagePrefix = imagePrefix;
        _imageSuffix = imageSuffix;
    }

    function generate(address account, uint256 balance) external view returns (string memory) {
        string memory baseName = "XPNFT Token ";
        string memory baseDescription = "This is a XPNFT token for address ";

        string memory propName = string(abi.encodePacked(baseName, Strings.toHexString(account)));
        string memory propDescription = string(
            abi.encodePacked(baseDescription, Strings.toHexString(account), " with balance ", Strings.toString(balance))
        );
        string memory image = _generateImage(balance);

        bytes memory json = abi.encodePacked(
            "{\"name\":\"",
            propName,
            "\",\"description\":\"",
            propDescription,
            "\",\"image\":\"data:image/svg+xml;base64,",
            image,
            "\"}"
        );

        string memory jsonBase64 = Base64.encode(json);
        return string(abi.encodePacked("data:application/json;base64,", jsonBase64));
    }

    function _generateImage(uint256 balance) internal view returns (string memory) {
        string memory text = Strings.toString(balance / 10e18);
        bytes memory svg = abi.encodePacked(_imagePrefix, text, _imageSuffix);

        return Base64.encode(svg);
    }
}
