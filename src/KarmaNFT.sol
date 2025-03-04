// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { INFTMetadataGenerator } from "./interfaces/INFTMetadataGenerator.sol";

interface IERC20 {
    function balanceOf(address account) external view returns (uint256);
}

contract KarmaNFT is Ownable {
    error KarmaNFT__TransferNotAllowed();
    error KarmaNFT__InvalidTokenId();

    IERC20 public karmaToken;
    INFTMetadataGenerator public metadataGenerator;

    string private name = "KarmaNFT";
    string private symbol = "KARMANFT";

    event Transfer(address indexed from, address indexed to, uint256 indexed tokenId);

    modifier onlyValidTokenId(uint256 tokenId) {
        if (tokenId > type(uint160).max) {
            revert KarmaNFT__InvalidTokenId();
        }
        _;
    }

    constructor(address karmaTokenAddress, address _metadataGenerator) Ownable() {
        karmaToken = IERC20(karmaTokenAddress);
        metadataGenerator = INFTMetadataGenerator(_metadataGenerator);
        _transferOwnership(msg.sender);
    }

    function setMetadataGenerator(address _metadataGenerator) external onlyOwner {
        metadataGenerator = INFTMetadataGenerator(_metadataGenerator);
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
        revert KarmaNFT__TransferNotAllowed();
    }

    function safeTransferFrom(address, address, uint256) external pure {
        revert KarmaNFT__TransferNotAllowed();
    }

    function transferFrom(address, address, uint256) external pure {
        revert KarmaNFT__TransferNotAllowed();
    }

    function approve(address, uint256) external pure {
        revert KarmaNFT__TransferNotAllowed();
    }

    function setApprovalForAll(address, bool) external pure {
        revert KarmaNFT__TransferNotAllowed();
    }

    function getApproved(uint256) external pure returns (address) {
        return address(0);
    }

    function isApprovedForAll(address, address) external pure returns (bool) {
        return false;
    }

    function tokenURI(uint256 tokenId) external view onlyValidTokenId(tokenId) returns (string memory) {
        address account = address(uint160(tokenId));
        uint256 balance = karmaToken.balanceOf(account);
        return metadataGenerator.generate(account, balance);
    }
}
