// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { IMetadataGenerator } from "./XPNFTMetadataGenerator.sol";

interface IERC20 {
    function balanceOf(address account) external view returns (uint256);
}

contract XPNFTToken is Ownable {
    error XPNFT__TransferNotAllowed();
    error XPNFT__InvalidTokenId();

    IERC20 public xpToken;
    IMetadataGenerator public metadataGenerator;

    string private name = "XPNFT";
    string private symbol = "XPNFT";

    event Transfer(address indexed from, address indexed to, uint256 indexed tokenId);

    modifier onlyValidTokenId(uint256 tokenId) {
        if (tokenId > type(uint160).max) {
            revert XPNFT__InvalidTokenId();
        }
        _;
    }

    constructor(address xpTokenAddress, address _metadataGenerator) Ownable(msg.sender) {
        xpToken = IERC20(xpTokenAddress);
        metadataGenerator = IMetadataGenerator(_metadataGenerator);
    }

    function setMetadataGenerator(address _metadataGenerator) external onlyOwner {
        metadataGenerator = IMetadataGenerator(_metadataGenerator);
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
        address account = address(uint160(tokenId));
        uint256 balance = xpToken.balanceOf(account);
        return metadataGenerator.generate(account, balance);
    }
}
