// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Ownable2Step } from "@openzeppelin/contracts/access/Ownable2Step.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { MerkleProof } from "@openzeppelin/contracts/utils/cryptography/MerkleProof.sol";

/**
 * @title KarmaAirdrop
 * @notice A contract for distributing Karma tokens via a Merkle tree airdrop.
 * Users can claim their tokens by providing a valid Merkle proof.
 * The contract owner can set the Merkle root only once.
 */
contract KarmaAirdrop is Ownable2Step {
    /// @notice Emitted when merkleroot is already set
    error KarmaAirdrop__MerkleRootAlreadySet();
    /// @notice Emitted when merkleroot is not set
    error KarmaAirdrop__MerkleRootNotSet();
    /// @notice Emitted when a claim is already made
    error KarmaAirdrop__AlreadyClaimed();
    /// @notice Emitted when a proof is invalid
    error KarmaAirdrop__InvalidProof();
    /// @notice Emitted when token transfer fails
    error KarmaAirdrop__TransferFailed();

    /// @notice Emitted when a claim is made
    event Claimed(uint256 index, address account, uint256 amount);
    /// @notice Emitted when merkleroot is set
    event MerkleRootSet(bytes32 merkleRoot);

    /// @notice The address of the Karma token contract
    address public immutable token;
    /// @notice The Merkle root of the airdrop
    bytes32 public merkleRoot;
    /// @notice A bitmap to track claimed indices
    mapping(uint256 => uint256) private claimedBitMap;

    constructor(address _token, address _owner) {
        token = _token;
        _transferOwnership(_owner);
    }

    /**
     * @notice Sets the Merkle root for the airdrop. Can only be called once by the owner.
     * @param _merkleRoot The Merkle root to set
     */
    function setMerkleRoot(bytes32 _merkleRoot) external onlyOwner {
        if (merkleRoot != bytes32(0)) {
            revert KarmaAirdrop__MerkleRootAlreadySet();
        }
        merkleRoot = _merkleRoot;
        emit MerkleRootSet(merkleRoot);
    }

    /**
     * @notice Checks if a claim has been made for a given index
     * @param index The index to check
     * @return True if the index has been claimed, false otherwise
     */
    function isClaimed(uint256 index) public view returns (bool) {
        uint256 claimedWordIndex = index / 256;
        uint256 claimedBitIndex = index % 256;
        uint256 claimedWord = claimedBitMap[claimedWordIndex];
        uint256 mask = (1 << claimedBitIndex);
        return claimedWord & mask == mask;
    }

    function _setClaimed(uint256 index) private {
        uint256 claimedWordIndex = index / 256;
        uint256 claimedBitIndex = index % 256;
        claimedBitMap[claimedWordIndex] = claimedBitMap[claimedWordIndex] | (1 << claimedBitIndex);
    }

    /**
     * @notice Claims tokens for a given index, account, and amount, if the provided Merkle proof is valid
     * @param index The index of the claim
     * @param account The address of the account to claim tokens for
     * @param amount The amount of tokens to claim
     * @param merkleProof The Merkle proof to validate the claim
     */
    function claim(uint256 index, address account, uint256 amount, bytes32[] calldata merkleProof) external {
        if (merkleRoot == bytes32(0)) {
            revert KarmaAirdrop__MerkleRootNotSet();
        }
        if (isClaimed(index)) {
            revert KarmaAirdrop__AlreadyClaimed();
        }

        // Verify the merkle proof.
        bytes32 node = keccak256(abi.encodePacked(index, account, amount));
        if (!MerkleProof.verify(merkleProof, merkleRoot, node)) {
            revert KarmaAirdrop__InvalidProof();
        }

        // Mark it claimed and send the token.
        _setClaimed(index);
        if (!IERC20(token).transfer(account, amount)) {
            revert KarmaAirdrop__TransferFailed();
        }

        emit Claimed(index, account, amount);
    }
}
