// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Base64 } from "@openzeppelin/contracts/utils/Base64.sol";
import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";

interface IERC20 {
    function balanceOf(address account) external view returns (uint256);
}

contract XPNFTToken is Ownable {
    error XPNFT__TransferNotAllowed();
    error XPNFT__InvalidTokenId();

    string private _name = "XPNFT";
    string private _symbol = "XPNFT";
    string private _imagePrefix = "";
    string private _imageSuffix = "";

    IERC20 private _xpToken;

    event Transfer(address indexed from, address indexed to, uint256 indexed tokenId);

    modifier onlyValidTokenId(uint256 tokenId) {
        if (tokenId > type(uint160).max) {
            revert XPNFT__InvalidTokenId();
        }
        _;
    }

    constructor(address xpTokenAddress, string memory imagePrefix, string memory imageSuffix) Ownable(msg.sender) {
        _xpToken = IERC20(xpTokenAddress);
        _imagePrefix = imagePrefix;
        _imageSuffix = imageSuffix;
    }

    function setImageStrings(string memory imagePrefix, string memory imageSuffix) external onlyOwner {
        _imagePrefix = imagePrefix;
        _imageSuffix = imageSuffix;
    }

    function name() external view returns (string memory) {
        return _name;
    }

    function symbol() external view returns (string memory) {
        return _symbol;
    }

    function mint() external {
        emit Transfer(msg.sender, msg.sender, uint256(uint160(msg.sender)));
    }

    function balanceOf(address) external pure returns (uint256) {
        return 1;
    }

    function ownerOf(uint256 tokenId) external pure onlyValidTokenId(tokenId) returns (address) {
        address owner = address(uint160(tokenId));
        return owner;
    }

    function safeTransferFrom(address, address, uint256, bytes calldata) external pure {
        revert XPNFT__TransferNotAllowed();
    }

    function safeTransferFrom(address, address, uint256) external pure {
        revert XPNFT__TransferNotAllowed();
    }

    function transferFrom(address, address, uint256) external pure {
        revert XPNFT__TransferNotAllowed();
    }

    function approve(address, uint256) external pure {
        revert XPNFT__TransferNotAllowed();
    }

    function setApprovalForAll(address, bool) external pure {
        revert XPNFT__TransferNotAllowed();
    }

    function getApproved(uint256) external pure returns (address) {
        return address(0);
    }

    function isApprovedForAll(address, address) external pure returns (bool) {
        return false;
    }

    function tokenURI(uint256 tokenId) external view onlyValidTokenId(tokenId) returns (string memory) {
        address owner = address(uint160(tokenId));
        return _createTokenURI(owner);
    }

    function _createTokenURI(address owner) internal view returns (string memory) {
        string memory baseName = "XPNFT Token ";
        string memory baseDescription = "This is a XPNFT token for address ";
        uint256 balance = _xpToken.balanceOf(owner) / 1e18;

        string memory propName = string(abi.encodePacked(baseName, Strings.toHexString(owner)));
        string memory propDescription = string(
            abi.encodePacked(baseDescription, Strings.toHexString(owner), " with balance ", Strings.toString(balance))
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
        string memory text = Strings.toString(balance);
        bytes memory svg = abi.encodePacked(_imagePrefix, text, _imageSuffix);

        return Base64.encode(svg);
    }
}
