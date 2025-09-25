// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { INFTMetadataGenerator } from "./interfaces/INFTMetadataGenerator.sol";

interface IERC20 {
    function balanceOf(address account) external view returns (uint256);
}

/**
 * @title KarmaNFT
 * @notice A non-transferable NFT that represents a user's karma based on their ERC20 token balance.
 * Each address can mint one NFT, and the NFT's metadata is generated dynamically based on the user's token balance.
 */
contract KarmaNFT is Ownable {
    /// @notice Error thrown when a transfer is attempted
    error KarmaNFT__TransferNotAllowed();
    /// @notice Error thrown when an invalid token ID is used
    error KarmaNFT__InvalidTokenId();

    /// @notice The ERC20 token used to determine karma
    IERC20 public immutable karmaToken;
    /// @notice The metadata generator contract
    INFTMetadataGenerator public metadataGenerator;

    /// @notice The name of the NFT
    string private name = "KarmaNFT";
    /// @notice The symbol of the NFT
    string private symbol = "KARMANFT";

    /// @notice Emitted when an NFT is transferred (minted in this case)
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

    /**
     * @notice Sets a new metadata generator contract
     * @dev Only the owner can call this function
     * @param _metadataGenerator The address of the new metadata generator contract
     */
    function setMetadataGenerator(address _metadataGenerator) external onlyOwner {
        metadataGenerator = INFTMetadataGenerator(_metadataGenerator);
    }

    /**
     * @notice Emits transfer event to simulate minting an NFT to the caller's address.
     */
    function mint() external {
        emit Transfer(msg.sender, msg.sender, uint256(uint160(msg.sender)));
    }

    /**
     * @notice Returns balance of NFTs for an address (always 1).
     * @return The balance of NFTs (always 1).
     */
    function balanceOf(address) external pure returns (uint256) {
        return 1;
    }

    /**
     * @notice Returns the owner of a given token ID.
     * @param tokenId The token ID to query.
     * @return address The owner address corresponding to the token ID.
     */
    function ownerOf(uint256 tokenId) external pure onlyValidTokenId(tokenId) returns (address) {
        address owner = address(uint160(tokenId));
        return owner;
    }

    /**
     * @notice Not allowed as the NFT is non-transferable.
     */
    function safeTransferFrom(address, address, uint256, bytes calldata) external pure {
        revert KarmaNFT__TransferNotAllowed();
    }

    /**
     * @notice Not allowed as the NFT is non-transferable.
     */
    function safeTransferFrom(address, address, uint256) external pure {
        revert KarmaNFT__TransferNotAllowed();
    }

    /**
     * @notice Not allowed as the NFT is non-transferable.
     */
    function transferFrom(address, address, uint256) external pure {
        revert KarmaNFT__TransferNotAllowed();
    }

    /**
     * @notice Not allowed as the NFT is non-transferable.
     */
    function approve(address, uint256) external pure {
        revert KarmaNFT__TransferNotAllowed();
    }

    /**
     * @notice Not allowed as the NFT is non-transferable.
     */
    function setApprovalForAll(address, bool) external pure {
        revert KarmaNFT__TransferNotAllowed();
    }

    /**
     * @notice Returns approved address for a token ID (always zero address).
     */
    function getApproved(uint256) external pure returns (address) {
        return address(0);
    }

    /**
     * @notice Returns if an operator is approved for all (always false).
     */
    function isApprovedForAll(address, address) external pure returns (bool) {
        return false;
    }

    /**
     * @notice Returns the token uRI for a given token ID.
     * @param tokenId The token ID to query.
     * @return string The token URI containing metadata.
     */
    function tokenURI(uint256 tokenId) external view onlyValidTokenId(tokenId) returns (string memory) {
        address account = address(uint160(tokenId));
        uint256 balance = karmaToken.balanceOf(account);
        return metadataGenerator.generate(account, balance);
    }
}
