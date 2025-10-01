// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Ownable2Step } from "@openzeppelin/contracts/access/Ownable2Step.sol";
import { Pausable } from "@openzeppelin/contracts/security/Pausable.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { MerkleProof } from "@openzeppelin/contracts/utils/cryptography/MerkleProof.sol";
import { IVotes } from "@openzeppelin/contracts/governance/utils/IVotes.sol";

/**
 * @title KarmaAirdrop
 * @notice A contract for distributing Karma tokens via a Merkle tree airdrop.
 * Users can claim their tokens by providing a valid Merkle proof.
 * The contract owner can set the Merkle root only once.
 */
contract KarmaAirdrop is Ownable2Step, Pausable {
    /// @notice Emitted when merkleroot is already set
    error KarmaAirdrop__MerkleRootAlreadySet();
    /// @notice Emitted when merkleroot is not set
    error KarmaAirdrop__MerkleRootNotSet();
    /// @notice Emitted when trying to update merkle root while contract is not paused
    error KarmaAirdrop__MustBePausedToUpdate();
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
    /// @notice Whether the merkle root can be updated more than once
    bool public immutable allowMerkleRootUpdate;
    /// @notice The default delegatee address for new claimers
    address public immutable defaultDelegatee;
    /// @notice The Merkle root of the airdrop
    bytes32 public merkleRoot;
    /// @notice Current epoch - incremented with each merkle root update
    uint256 public epoch;
    /// @notice A bitmap to track claimed indices per epoch
    mapping(uint256 => mapping(uint256 => uint256)) private claimedBitMap;

    constructor(address _token, address _owner, bool _allowMerkleRootUpdate, address _defaultDelegatee) {
        token = _token;
        allowMerkleRootUpdate = _allowMerkleRootUpdate;
        defaultDelegatee = _defaultDelegatee;
        _transferOwnership(_owner);
    }

    /**
     * @notice Sets the Merkle root for the airdrop. Can only be called by the owner.
     * If allowMerkleRootUpdate is false, can only be called once.
     * When updating an existing merkle root, the contract must be paused to prevent front-running.
     * When the merkle root is updated, the epoch is incremented, creating a new bitmap.
     * @param _merkleRoot The Merkle root to set
     */
    function setMerkleRoot(bytes32 _merkleRoot) external onlyOwner {
        if (!allowMerkleRootUpdate && merkleRoot != bytes32(0)) {
            revert KarmaAirdrop__MerkleRootAlreadySet();
        }

        // When updating an existing merkle root (not the first time), contract must be paused
        if (allowMerkleRootUpdate && merkleRoot != bytes32(0) && !paused()) {
            revert KarmaAirdrop__MustBePausedToUpdate();
        }

        // Increment epoch to create a new bitmap
        if (merkleRoot != bytes32(0)) {
            epoch++;
        }

        merkleRoot = _merkleRoot;
        emit MerkleRootSet(merkleRoot);
    }

    /**
     * @notice Checks if a claim has been made for a given index in the current epoch
     * @param index The index to check
     * @return True if the index has been claimed, false otherwise
     */
    function isClaimed(uint256 index) public view returns (bool) {
        uint256 claimedWordIndex = index / 256;
        uint256 claimedBitIndex = index % 256;
        uint256 claimedWord = claimedBitMap[epoch][claimedWordIndex];
        uint256 mask = (1 << claimedBitIndex);
        return claimedWord & mask == mask;
    }

    function _setClaimed(uint256 index) private {
        uint256 claimedWordIndex = index / 256;
        uint256 claimedBitIndex = index % 256;
        claimedBitMap[epoch][claimedWordIndex] = claimedBitMap[epoch][claimedWordIndex] | (1 << claimedBitIndex);
    }

    /**
     * @notice Claims tokens for a given index, account, and amount, if the provided Merkle proof is valid
     * @param index The index of the claim
     * @param account The address of the account to claim tokens for
     * @param amount The amount of tokens to claim
     * @param merkleProof The Merkle proof to validate the claim
     * @param nonce The nonce for the delegation signature
     * @param expiry The expiry timestamp for the delegation signature
     * @param v The v component of the delegation signature
     * @param r The r component of the delegation signature
     * @param s The s component of the delegation signature
     */
    function claim(
        uint256 index,
        address account,
        uint256 amount,
        bytes32[] calldata merkleProof,
        uint256 nonce,
        uint256 expiry,
        uint8 v,
        bytes32 r,
        bytes32 s
    )
        external
    {
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

        // If the account has no karma balance before this claim, delegate to the default delegatee
        if (IERC20(token).balanceOf(account) == amount) {
            IVotes(token).delegateBySig(defaultDelegatee, nonce, expiry, v, r, s);
        }

        emit Claimed(index, account, amount);
    }

    /**
     * @notice Pauses the contract, preventing claims. Can only be called by the owner.
     */
    function pause() external onlyOwner {
        _pause();
    }

    /**
     * @notice Unpauses the contract, allowing claims. Can only be called by the owner.
     */
    function unpause() external onlyOwner {
        _unpause();
    }
}
