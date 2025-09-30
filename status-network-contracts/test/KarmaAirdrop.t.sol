// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";

import { DeploymentConfig } from "../script/DeploymentConfig.s.sol";
import { DeployKarmaAirdropScript } from "../script/DeployKarmaAirdrop.s.sol";
import { MockToken } from "./mocks/MockToken.sol";

import { KarmaAirdrop } from "../src/KarmaAirdrop.sol";

contract KarmaAirdropTest is Test {
    KarmaAirdrop internal airdrop;
    MockToken internal rewardToken;

    address internal owner = makeAddr("owner");

    function setUp() public virtual {
        rewardToken = new MockToken("Karma Token", "KT");

        DeployKarmaAirdropScript deployScript = new DeployKarmaAirdropScript();
        (airdrop,) = deployScript.runForTest(address(rewardToken), owner);
    }

    function test_Owner() public view {
        assertEq(airdrop.owner(), owner);
    }
}

contract SetMerkleRootTest is KarmaAirdropTest {
    bytes32 internal merkleRoot = 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa;

    function setUp() public override {
        super.setUp();
    }

    function test_RevertWhen_NotOwner_SetMerkleRoot() public {
        vm.prank(address(0x1234));
        vm.expectRevert("Ownable: caller is not the owner");
        airdrop.setMerkleRoot(merkleRoot);
    }

    function test_Revert_When_UpdateMerkleRoot_WhenNotAllowed() public {
        bytes32 newMerkleRoot = 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabb;
        vm.prank(owner);
        airdrop.setMerkleRoot(merkleRoot);
        vm.prank(owner);
        vm.expectRevert(KarmaAirdrop.KarmaAirdrop__MerkleRootAlreadySet.selector);
        airdrop.setMerkleRoot(newMerkleRoot);
    }

    function test_Revert_When_UpdateMerkleRoot_WhileNotPaused() public {
        KarmaAirdrop updatableAirdrop = new KarmaAirdrop(address(rewardToken), owner, true);

        bytes32 newMerkleRoot = 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabb;

        // Set initial merkle root (first time, no pause required)
        vm.prank(owner);
        updatableAirdrop.setMerkleRoot(merkleRoot);
        assertEq(updatableAirdrop.merkleRoot(), merkleRoot);

        // Try to update merkle root without pausing (should fail)
        vm.prank(owner);
        vm.expectRevert(KarmaAirdrop.KarmaAirdrop__MustBePausedToUpdate.selector);
        updatableAirdrop.setMerkleRoot(newMerkleRoot);
    }

    function test_Success_When_UpdateMerkleRoot_WhenAllowed() public {
        KarmaAirdrop updatableAirdrop = new KarmaAirdrop(address(rewardToken), owner, true);

        bytes32 newMerkleRoot = 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabb;

        // Set initial merkle root (first time, no pause required)
        vm.prank(owner);
        updatableAirdrop.setMerkleRoot(merkleRoot);
        assertEq(updatableAirdrop.merkleRoot(), merkleRoot);

        // Pause the contract before updating
        vm.prank(owner);
        updatableAirdrop.pause();

        // Update merkle root (should succeed when paused)
        vm.prank(owner);
        updatableAirdrop.setMerkleRoot(newMerkleRoot);
        assertEq(updatableAirdrop.merkleRoot(), newMerkleRoot);
    }

    function test_Success_When_UpdateMerkleRoot_IncreasesEpoch() public {
        KarmaAirdrop updatableAirdrop = new KarmaAirdrop(address(rewardToken), owner, true);

        bytes32 newMerkleRoot = 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabb;
        bytes32 thirdMerkleRoot = 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaacc;

        // Initial epoch should be 0
        assertEq(updatableAirdrop.epoch(), 0);

        // Set initial merkle root (first time) - epoch should remain 0
        vm.prank(owner);
        updatableAirdrop.setMerkleRoot(merkleRoot);
        assertEq(updatableAirdrop.epoch(), 0);

        // Pause and update merkle root - epoch should increase to 1
        vm.startPrank(owner);
        updatableAirdrop.pause();
        updatableAirdrop.setMerkleRoot(newMerkleRoot);
        assertEq(updatableAirdrop.epoch(), 1);

        // Update again - epoch should increase to 2
        updatableAirdrop.setMerkleRoot(thirdMerkleRoot);
        assertEq(updatableAirdrop.epoch(), 2);
        vm.stopPrank();
    }

    function test_Success_When_UpdateMerkleRoot_ResetsClaimedBitmap() public {
        KarmaAirdrop updatableAirdrop = new KarmaAirdrop(address(rewardToken), owner, true);

        // Set up first merkle tree
        uint256 index = 0;
        address account = makeAddr("alice");
        uint256 amount = 100e18;
        bytes32 leaf = keccak256(abi.encodePacked(index, account, amount));
        bytes32[] memory merkleProof = new bytes32[](0);

        // Fund the airdrop contract
        rewardToken.mint(address(updatableAirdrop), amount * 2);

        // Set initial merkle root and claim
        vm.prank(owner);
        updatableAirdrop.setMerkleRoot(leaf);
        updatableAirdrop.claim(index, account, amount, merkleProof);
        assertTrue(updatableAirdrop.isClaimed(index));

        // Pause before updating merkle root
        vm.prank(owner);
        updatableAirdrop.pause();

        // Update merkle root - this should reset the bitmap
        bytes32 newMerkleRoot = keccak256(abi.encodePacked(index, account, amount));
        vm.prank(owner);
        updatableAirdrop.setMerkleRoot(newMerkleRoot);

        // Unpause to allow claims
        vm.prank(owner);
        updatableAirdrop.unpause();

        // Verify the claim was reset
        assertFalse(updatableAirdrop.isClaimed(index));

        // Should be able to claim again with new merkle tree
        updatableAirdrop.claim(index, account, amount, merkleProof);
        assertTrue(updatableAirdrop.isClaimed(index));
        assertEq(rewardToken.balanceOf(account), amount * 2);
    }

    function test_Success_When_SetMerkleRoot() public {
        vm.prank(owner);
        airdrop.setMerkleRoot(merkleRoot);
        assertEq(airdrop.merkleRoot(), merkleRoot);
    }
}

contract ClaimTest is KarmaAirdropTest {
    function setUp() public override {
        super.setUp();
    }

    function _hashPair(bytes32 a, bytes32 b) public pure returns (bytes32) {
        return a < b ? keccak256(abi.encodePacked(a, b)) : keccak256(abi.encodePacked(b, a));
    }

    function test_Revert_When_ClaimBeforeMerkleRootSet() public {
        uint256 index = 0;
        address account = makeAddr("alice");
        uint256 amount = 100e18;
        bytes32[] memory merkleProof = new bytes32[](0);

        vm.expectRevert(KarmaAirdrop.KarmaAirdrop__MerkleRootNotSet.selector);
        airdrop.claim(index, account, amount, merkleProof);
    }

    function test_Success_When_ClaimWithValidProof() public {
        // Set up test data
        uint256 index = 0;
        address account = makeAddr("alice");
        uint256 amount = 100e18;

        // Create a simple merkle tree with one leaf
        // Leaf: keccak256(abi.encodePacked(index, account, amount))
        bytes32 leaf = keccak256(abi.encodePacked(index, account, amount));
        bytes32 merkleRoot = leaf; // Single leaf tree - root equals leaf
        bytes32[] memory merkleProof = new bytes32[](0); // Empty proof for single leaf

        // Fund the airdrop contract with tokens
        rewardToken.mint(address(airdrop), amount);

        // Set merkle root
        vm.prank(owner);
        airdrop.setMerkleRoot(merkleRoot);

        // Verify initial state
        assertFalse(airdrop.isClaimed(index));
        assertEq(rewardToken.balanceOf(account), 0);
        assertEq(rewardToken.balanceOf(address(airdrop)), amount);

        // Claim tokens
        vm.expectEmit(true, true, true, true);
        emit KarmaAirdrop.Claimed(index, account, amount);
        airdrop.claim(index, account, amount, merkleProof);

        // Verify final state
        assertTrue(airdrop.isClaimed(index));
        assertEq(rewardToken.balanceOf(account), amount);
        assertEq(rewardToken.balanceOf(address(airdrop)), 0);
    }

    function test_Success_When_ClaimFromComplexMerkleTree() public {
        //          root
        //         /    \
        //      node01  node23
        //      /  \    /  \
        //   leaf0 leaf1 leaf2 leaf3
        //   (alice)(bob)(charlie)(david)
        //
        //   For Bob's claim (index 1), the proof consists of:
        //   1. leaf0 (Alice's leaf) - Bob's sibling
        //   2. node23 (Charlie+David's combined node) - The uncle node

        bytes32 leaf0 = keccak256(abi.encodePacked(uint256(0), makeAddr("alice"), uint256(100e18)));
        bytes32 leaf1 = keccak256(abi.encodePacked(uint256(1), makeAddr("bob"), uint256(200e18)));
        bytes32 leaf2 = keccak256(abi.encodePacked(uint256(2), makeAddr("charlie"), uint256(300e18)));
        bytes32 leaf3 = keccak256(abi.encodePacked(uint256(3), makeAddr("david"), uint256(400e18)));

        bytes32 node01 = _hashPair(leaf0, leaf1);
        bytes32 node23 = _hashPair(leaf2, leaf3);
        bytes32 merkleRoot = _hashPair(node01, node23);

        rewardToken.mint(address(airdrop), 1000e18);

        vm.prank(owner);
        airdrop.setMerkleRoot(merkleRoot);

        bytes32[] memory merkleProofBob = new bytes32[](2);
        merkleProofBob[0] = leaf0; // Sibling of leaf1
        merkleProofBob[1] = node23; // Uncle node

        bytes32[] memory merkleProofCharlie = new bytes32[](2);
        merkleProofCharlie[0] = leaf3;
        merkleProofCharlie[1] = node01; // Uncle node

        // Verify initial state
        assertFalse(airdrop.isClaimed(1));
        assertEq(rewardToken.balanceOf(makeAddr("bob")), 0);

        // Claim tokens
        vm.expectEmit(true, true, true, true);
        emit KarmaAirdrop.Claimed(1, makeAddr("bob"), 200e18);
        airdrop.claim(1, makeAddr("bob"), 200e18, merkleProofBob);

        // Verify final state
        assertTrue(airdrop.isClaimed(1));
        assertEq(rewardToken.balanceOf(makeAddr("bob")), 200e18);
        assertEq(rewardToken.balanceOf(address(airdrop)), 800e18);

        vm.expectEmit(true, true, true, true);
        emit KarmaAirdrop.Claimed(2, makeAddr("charlie"), 300e18);
        airdrop.claim(2, makeAddr("charlie"), 300e18, merkleProofCharlie);
    }
}
