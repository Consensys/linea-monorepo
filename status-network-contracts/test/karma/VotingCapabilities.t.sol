// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Vm } from "forge-std/Vm.sol";
import { IVotesUpgradeable } from "@openzeppelin/contracts-upgradeable/governance/utils/IVotesUpgradeable.sol";
import { Karma } from "../../src/Karma.sol";
import { KarmaTest } from "./Karma.t.sol";

contract VotingCapabilityTest is KarmaTest {
    function setUp() public override {
        super.setUp();
    }

    function test_InitialNonceIsZero() public view {
        uint256 nonce = karma.nonces(alice);
        assertEq(nonce, 0);
    }

    function test_RevertWhen_MintingRisksOverflowingVotes() public {
        uint256 amount = 2 ** 224;
        vm.prank(owner);
        vm.expectRevert("ERC20Votes: total supply risks overflowing votes");
        karma.mint(alice, amount);
    }

    function test_RecentCheckpoints() public {
        vm.prank(alice);
        karma.delegate(alice);

        for (uint256 i = 0; i < 6; i++) {
            vm.roll(block.number + 1);
            vm.prank(owner);
            karma.mint(alice, 1);
        }

        uint256 currentBlock = block.number;
        assertEq(karma.numCheckpoints(alice), 6);
        // recent
        assertEq(karma.getPastVotes(alice, currentBlock - 1), 5);
        // non-recent
        assertEq(karma.getPastVotes(alice, currentBlock - 6), 0);
    }

    function test_DelegationWithBalance() public {
        uint256 supply = 1000 ether;

        // Mint tokens to holder
        vm.prank(owner);
        karma.mint(alice, supply);

        // Initially, delegates should be zero address
        assertEq(karma.delegates(alice), address(0));

        // Delegate to self and capture events
        vm.prank(alice);
        vm.expectEmit(true, true, true, true);
        emit IVotesUpgradeable.DelegateChanged(alice, address(0), alice);
        vm.expectEmit(true, true, true, true);
        emit IVotesUpgradeable.DelegateVotesChanged(alice, 0, supply);
        karma.delegate(alice);

        // After delegation, delegates should return alice
        assertEq(karma.delegates(alice), alice);

        // Current votes should equal supply
        assertEq(karma.getVotes(alice), supply);

        // Advance block
        vm.roll(block.number + 1);

        uint256 delegationBlock = block.number - 1;

        // Past votes before delegation should be 0
        assertEq(karma.getPastVotes(alice, delegationBlock - 1), 0);

        // Past votes at delegation block should equal supply
        assertEq(karma.getPastVotes(alice, delegationBlock), supply);
    }

    function test_DelegationWithoutBalance() public {
        // Initially, delegates should be zero address
        assertEq(karma.delegates(alice), address(0));

        // Delegate to self without having any balance
        vm.prank(alice);
        vm.expectEmit(true, true, true, true);
        emit IVotesUpgradeable.DelegateChanged(alice, address(0), alice);
        vm.recordLogs();
        karma.delegate(alice);

        // Verify DelegateVotesChanged was NOT emitted
        Vm.Log[] memory entries = vm.getRecordedLogs();
        for (uint256 i = 0; i < entries.length; i++) {
            // DelegateVotesChanged event signature
            assertNotEq(
                entries[i].topics[0],
                keccak256("DelegateVotesChanged(address,uint256,uint256)"),
                "DelegateVotesChanged should not be emitted"
            );
        }

        // After delegation, delegates should return alice
        assertEq(karma.delegates(alice), alice);
    }

    function test_AcceptSignedDelegation() public {
        uint256 supply = 1000 ether;
        uint256 delegatorPrivateKey = 0xd1e9a70;
        address delegatorAddress = vm.addr(delegatorPrivateKey);

        // Mint tokens to delegator
        vm.prank(owner);
        karma.mint(delegatorAddress, supply);

        // Advance block to separate mint from delegation
        vm.roll(block.number + 1);

        // Initially, delegates should be zero address
        assertEq(karma.delegates(delegatorAddress), address(0));

        // Generate delegation signature
        uint256 nonce = 0;
        uint256 expiry = type(uint256).max;
        bytes32 domainSeparator = karma.DOMAIN_SEPARATOR();
        bytes32 structHash = keccak256(
            abi.encode(
                keccak256("Delegation(address delegatee,uint256 nonce,uint256 expiry)"),
                delegatorAddress,
                nonce,
                expiry
            )
        );
        bytes32 digest = keccak256(abi.encodePacked("\x19\x01", domainSeparator, structHash));
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(delegatorPrivateKey, digest);

        // Execute delegateBySig and capture events
        vm.expectEmit(true, true, true, true);
        emit IVotesUpgradeable.DelegateChanged(delegatorAddress, address(0), delegatorAddress);
        vm.expectEmit(true, true, true, true);
        emit IVotesUpgradeable.DelegateVotesChanged(delegatorAddress, 0, supply);
        karma.delegateBySig(delegatorAddress, nonce, expiry, v, r, s);

        // After delegation, delegates should return delegatorAddress
        assertEq(karma.delegates(delegatorAddress), delegatorAddress);

        // Current votes should equal supply
        assertEq(karma.getVotes(delegatorAddress), supply);

        // Advance block to be able to query past votes
        vm.roll(block.number + 1);

        uint256 delegationBlock = block.number - 1;

        // Past votes before delegation should be 0
        assertEq(karma.getPastVotes(delegatorAddress, delegationBlock - 1), 0);

        // Past votes at delegation block should equal supply
        assertEq(karma.getPastVotes(delegatorAddress, delegationBlock), supply);
    }

    function test_RevertWhen_ReusingSignature() public {
        uint256 supply = 1000 ether;
        uint256 delegatorPrivateKey = 0xd1e9a70;
        address delegatorAddress = vm.addr(delegatorPrivateKey);

        // Mint tokens to delegator
        vm.prank(owner);
        karma.mint(delegatorAddress, supply);

        // Generate delegation signature
        uint256 nonce = 0;
        uint256 expiry = type(uint256).max;
        bytes32 domainSeparator = karma.DOMAIN_SEPARATOR();
        bytes32 structHash = keccak256(
            abi.encode(
                keccak256("Delegation(address delegatee,uint256 nonce,uint256 expiry)"),
                delegatorAddress,
                nonce,
                expiry
            )
        );
        bytes32 digest = keccak256(abi.encodePacked("\x19\x01", domainSeparator, structHash));
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(delegatorPrivateKey, digest);

        // First delegation should succeed
        karma.delegateBySig(delegatorAddress, nonce, expiry, v, r, s);

        // Attempt to reuse the same signature should fail
        vm.expectRevert("ERC20Votes: invalid nonce");
        karma.delegateBySig(delegatorAddress, nonce, expiry, v, r, s);
    }

    function test_RejectBadDelegatee() public {
        uint256 delegatorPrivateKey = 0xd1e9a70;
        address delegatorAddress = vm.addr(delegatorPrivateKey);
        address holderDelegatee = makeAddr("holderDelegatee");

        // Mint tokens to delegator
        vm.prank(owner);
        karma.mint(delegatorAddress, 1000 ether);

        // Generate delegation signature for delegatorAddress
        bytes32 digest = keccak256(
            abi.encodePacked(
                "\x19\x01",
                karma.DOMAIN_SEPARATOR(),
                keccak256(
                    abi.encode(
                        keccak256("Delegation(address delegatee,uint256 nonce,uint256 expiry)"),
                        delegatorAddress,
                        0,
                        type(uint256).max
                    )
                )
            )
        );
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(delegatorPrivateKey, digest);

        // Call delegateBySig with a different delegatee than what was signed
        vm.recordLogs();
        karma.delegateBySig(holderDelegatee, 0, type(uint256).max, v, r, s);

        // Find and verify the DelegateChanged event
        Vm.Log[] memory entries = vm.getRecordedLogs();
        for (uint256 i = 0; i < entries.length; i++) {
            if (entries[i].topics[0] == keccak256("DelegateChanged(address,address,address)")) {
                // The delegator should NOT be the delegatorAddress
                assertNotEq(address(uint160(uint256(entries[i].topics[1]))), delegatorAddress);
                // From delegate should be zero address
                assertEq(address(uint160(uint256(entries[i].topics[2]))), address(0));
                // To delegate should be holderDelegatee
                assertEq(address(uint160(uint256(entries[i].topics[3]))), holderDelegatee);
                return;
            }
        }
    }

    function test_RevertWhen_BadNonce() public {
        uint256 delegatorPrivateKey = 0xd1e9a70;
        address delegatorAddress = vm.addr(delegatorPrivateKey);

        // Mint tokens to delegator
        vm.prank(owner);
        karma.mint(delegatorAddress, 1000 ether);

        // Generate delegation signature with nonce 0
        bytes32 digest = keccak256(
            abi.encodePacked(
                "\x19\x01",
                karma.DOMAIN_SEPARATOR(),
                keccak256(
                    abi.encode(
                        keccak256("Delegation(address delegatee,uint256 nonce,uint256 expiry)"),
                        delegatorAddress,
                        0,
                        type(uint256).max
                    )
                )
            )
        );
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(delegatorPrivateKey, digest);

        // Attempt to use the signature with a different nonce (1 instead of 0)
        vm.expectRevert("ERC20Votes: invalid nonce");
        karma.delegateBySig(delegatorAddress, 1, type(uint256).max, v, r, s);
    }

    function test_RevertWhen_ExpiredSignature() public {
        uint256 delegatorPrivateKey = 0xd1e9a70;
        address delegatorAddress = vm.addr(delegatorPrivateKey);

        // Mint tokens to delegator
        vm.prank(owner);
        karma.mint(delegatorAddress, 1000 ether);

        // Set expiry to 1 week in the past
        uint256 expiry = block.timestamp - 1 weeks;

        // Generate delegation signature with expired timestamp
        bytes32 digest = keccak256(
            abi.encodePacked(
                "\x19\x01",
                karma.DOMAIN_SEPARATOR(),
                keccak256(
                    abi.encode(
                        keccak256("Delegation(address delegatee,uint256 nonce,uint256 expiry)"),
                        delegatorAddress,
                        0,
                        expiry
                    )
                )
            )
        );
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(delegatorPrivateKey, digest);

        // Attempt to use the expired signature
        vm.expectRevert("ERC20Votes: signature expired");
        karma.delegateBySig(delegatorAddress, 0, expiry, v, r, s);
    }

    function test_ChangeDelegation() public {
        uint256 supply = 1000 ether;
        address holderDelegatee = makeAddr("holderDelegatee");

        // Setup: mint tokens to alice and delegate to self
        vm.prank(owner);
        karma.mint(alice, supply);

        vm.prank(alice);
        karma.delegate(alice);

        // Verify initial delegation
        assertEq(karma.delegates(alice), alice);

        // Move to next block to be able to check past votes
        vm.roll(block.number + 1);

        uint256 initialDelegationBlock = block.number - 1;

        // Change delegation to holderDelegatee
        vm.prank(alice);
        vm.expectEmit(true, true, true, true);
        emit IVotesUpgradeable.DelegateChanged(alice, alice, holderDelegatee);
        vm.expectEmit(true, true, true, true);
        emit IVotesUpgradeable.DelegateVotesChanged(alice, supply, 0);
        vm.expectEmit(true, true, true, true);
        emit IVotesUpgradeable.DelegateVotesChanged(holderDelegatee, 0, supply);
        karma.delegate(holderDelegatee);

        // After delegation change, delegates should return holderDelegatee
        assertEq(karma.delegates(alice), holderDelegatee);

        // Current votes should be 0 for alice and supply for holderDelegatee
        assertEq(karma.getVotes(alice), 0);
        assertEq(karma.getVotes(holderDelegatee), supply);

        // Advance block
        vm.roll(block.number + 1);

        uint256 delegationChangeBlock = block.number - 1;

        // Past votes before delegation change (at the initial delegation block)
        assertEq(karma.getPastVotes(alice, initialDelegationBlock), supply);
        assertEq(karma.getPastVotes(holderDelegatee, initialDelegationBlock), 0);

        // Past votes at delegation change block
        assertEq(karma.getPastVotes(alice, delegationChangeBlock), 0);
        assertEq(karma.getPastVotes(holderDelegatee, delegationChangeBlock), supply);
    }

    function test_TransferWithoutDelegation() public {
        uint256 supply = 1000 ether;

        // Mint tokens to alice
        vm.prank(owner);
        karma.mint(alice, supply);

        // Enable transfers for alice
        vm.prank(owner);
        karma.setAllowedToTransfer(alice, true);

        // Transfer 1 token from alice to bob (neither have delegated)
        vm.prank(alice);
        vm.recordLogs();
        karma.transfer(bob, 1);

        // Verify DelegateVotesChanged was NOT emitted
        Vm.Log[] memory entries = vm.getRecordedLogs();
        for (uint256 i = 0; i < entries.length; i++) {
            assertNotEq(
                entries[i].topics[0],
                keccak256("DelegateVotesChanged(address,uint256,uint256)"),
                "DelegateVotesChanged should not be emitted"
            );
        }

        // Both alice and bob should have 0 votes (no delegation)
        assertEq(karma.getVotes(alice), 0);
        assertEq(karma.getVotes(bob), 0);
    }

    function test_TransferWithSenderDelegation() public {
        uint256 supply = 1000 ether;

        // Mint tokens to alice
        vm.prank(owner);
        karma.mint(alice, supply);

        // Alice delegates to herself
        vm.prank(alice);
        karma.delegate(alice);

        // Enable transfers for alice
        vm.prank(owner);
        karma.setAllowedToTransfer(alice, true);

        // Transfer 1 token from alice to bob
        vm.prank(alice);
        vm.recordLogs();
        karma.transfer(bob, 1);

        // Find the Transfer and DelegateVotesChanged events
        Vm.Log[] memory entries = vm.getRecordedLogs();
        uint256 transferLogIndex = type(uint256).max;
        uint256 delegateVotesChangedLogIndex = type(uint256).max;
        bool foundDelegateVotesChanged = false;

        for (uint256 i = 0; i < entries.length; i++) {
            if (entries[i].topics[0] == keccak256("Transfer(address,address,uint256)")) {
                transferLogIndex = i;
            } else if (entries[i].topics[0] == keccak256("DelegateVotesChanged(address,uint256,uint256)")) {
                // Verify the DelegateVotesChanged event details
                address delegate = address(uint160(uint256(entries[i].topics[1])));
                (uint256 previousBalance, uint256 newBalance) = abi.decode(entries[i].data, (uint256, uint256));

                assertEq(delegate, alice);
                assertEq(previousBalance, supply);
                assertEq(newBalance, supply - 1);

                delegateVotesChangedLogIndex = i;
                foundDelegateVotesChanged = true;
            }
        }

        // Verify Transfer event comes before DelegateVotesChanged
        assertTrue(foundDelegateVotesChanged, "DelegateVotesChanged event should be emitted");
        assertTrue(transferLogIndex < delegateVotesChangedLogIndex, "Transfer should come before DelegateVotesChanged");

        // Alice should have supply - 1 votes, bob should have 0 votes
        assertEq(karma.getVotes(alice), supply - 1);
        assertEq(karma.getVotes(bob), 0);
    }

    function test_TransferWithReceiverDelegation() public {
        uint256 supply = 1000 ether;

        // Mint tokens to alice
        vm.prank(owner);
        karma.mint(alice, supply);

        // Bob delegates to himself (before receiving any tokens)
        vm.prank(bob);
        karma.delegate(bob);

        // Enable transfers for alice
        vm.prank(owner);
        karma.setAllowedToTransfer(alice, true);

        // Transfer 1 token from alice to bob
        vm.prank(alice);
        vm.recordLogs();
        karma.transfer(bob, 1);

        // Find the Transfer and DelegateVotesChanged events
        Vm.Log[] memory entries = vm.getRecordedLogs();
        uint256 transferLogIndex = type(uint256).max;
        uint256 delegateVotesChangedLogIndex = type(uint256).max;
        bool foundDelegateVotesChanged = false;

        for (uint256 i = 0; i < entries.length; i++) {
            if (entries[i].topics[0] == keccak256("Transfer(address,address,uint256)")) {
                transferLogIndex = i;
            } else if (entries[i].topics[0] == keccak256("DelegateVotesChanged(address,uint256,uint256)")) {
                // Verify the DelegateVotesChanged event details
                address delegate = address(uint160(uint256(entries[i].topics[1])));
                (uint256 previousBalance, uint256 newBalance) = abi.decode(entries[i].data, (uint256, uint256));

                assertEq(delegate, bob);
                assertEq(previousBalance, 0);
                assertEq(newBalance, 1);

                delegateVotesChangedLogIndex = i;
                foundDelegateVotesChanged = true;
            }
        }

        // Verify Transfer event comes before DelegateVotesChanged
        assertTrue(foundDelegateVotesChanged, "DelegateVotesChanged event should be emitted");
        assertTrue(transferLogIndex < delegateVotesChangedLogIndex, "Transfer should come before DelegateVotesChanged");

        // Alice should have 0 votes (no delegation), bob should have 1 vote
        assertEq(karma.getVotes(alice), 0);
        assertEq(karma.getVotes(bob), 1);
    }

    function test_TransferWithFullDelegation() public {
        uint256 supply = 1000 ether;

        // Mint tokens to alice
        vm.prank(owner);
        karma.mint(alice, supply);

        // Both alice and bob delegate to themselves
        vm.prank(alice);
        karma.delegate(alice);

        vm.prank(bob);
        karma.delegate(bob);

        // Enable transfers for alice
        vm.prank(owner);
        karma.setAllowedToTransfer(alice, true);

        // Transfer 1 token from alice to bob
        vm.prank(alice);
        vm.recordLogs();
        karma.transfer(bob, 1);

        // Find the Transfer and DelegateVotesChanged events
        Vm.Log[] memory entries = vm.getRecordedLogs();
        uint256 transferLogIndex = type(uint256).max;
        uint256 aliceDelegateVotesChangedLogIndex = type(uint256).max;
        uint256 bobDelegateVotesChangedLogIndex = type(uint256).max;
        bool foundAliceDelegateVotesChanged = false;
        bool foundBobDelegateVotesChanged = false;

        for (uint256 i = 0; i < entries.length; i++) {
            if (entries[i].topics[0] == keccak256("Transfer(address,address,uint256)")) {
                transferLogIndex = i;
            } else if (entries[i].topics[0] == keccak256("DelegateVotesChanged(address,uint256,uint256)")) {
                address delegate = address(uint160(uint256(entries[i].topics[1])));
                (uint256 previousBalance, uint256 newBalance) = abi.decode(entries[i].data, (uint256, uint256));

                if (delegate == alice) {
                    assertEq(previousBalance, supply);
                    assertEq(newBalance, supply - 1);
                    aliceDelegateVotesChangedLogIndex = i;
                    foundAliceDelegateVotesChanged = true;
                } else if (delegate == bob) {
                    assertEq(previousBalance, 0);
                    assertEq(newBalance, 1);
                    bobDelegateVotesChangedLogIndex = i;
                    foundBobDelegateVotesChanged = true;
                }
            }
        }

        // Verify both DelegateVotesChanged events were emitted
        assertTrue(foundAliceDelegateVotesChanged, "Alice's DelegateVotesChanged event should be emitted");
        assertTrue(foundBobDelegateVotesChanged, "Bob's DelegateVotesChanged event should be emitted");

        // Verify Transfer event comes before both DelegateVotesChanged events
        assertTrue(
            transferLogIndex < aliceDelegateVotesChangedLogIndex,
            "Transfer should come before Alice's DelegateVotesChanged"
        );
        assertTrue(
            transferLogIndex < bobDelegateVotesChangedLogIndex,
            "Transfer should come before Bob's DelegateVotesChanged"
        );

        // Alice should have supply - 1 votes, bob should have 1 vote
        assertEq(karma.getVotes(alice), supply - 1);
        assertEq(karma.getVotes(bob), 1);
    }

    function test_NumCheckpointsForDelegate() public {
        address other1 = makeAddr("other1");
        address other2 = makeAddr("other2");

        // Mint tokens to alice
        vm.prank(owner);
        karma.mint(alice, 1000 ether);

        // Enable transfers for alice and bob
        vm.startPrank(owner);
        karma.setAllowedToTransfer(alice, true);
        karma.setAllowedToTransfer(bob, true);
        vm.stopPrank();

        // Transfer 100 tokens to bob
        vm.prank(alice);
        karma.transfer(bob, 100);

        // Initially other1 has no checkpoints
        assertEq(karma.numCheckpoints(other1), 0);

        // t1: Bob delegates to other1
        vm.prank(bob);
        karma.delegate(other1);

        assertEq(karma.numCheckpoints(other1), 1);

        // t2: Bob transfers 10 tokens to other2
        vm.roll(block.number + 1);
        vm.prank(bob);
        karma.transfer(other2, 10);

        assertEq(karma.numCheckpoints(other1), 2);

        // t3: Bob transfers another 10 tokens to other2
        vm.roll(block.number + 1);
        vm.prank(bob);
        karma.transfer(other2, 10);

        assertEq(karma.numCheckpoints(other1), 3);

        // t4: Alice transfers 20 tokens back to bob
        vm.roll(block.number + 1);
        vm.prank(alice);
        karma.transfer(bob, 20);

        assertEq(karma.numCheckpoints(other1), 4);

        // Advance block to query past votes
        vm.roll(block.number + 1);

        uint256 t4Block = block.number - 1;
        uint256 t3Block = t4Block - 1;
        uint256 t2Block = t3Block - 1;
        uint256 t1Block = t2Block - 1;

        // Verify checkpoint data
        Karma.Checkpoint memory checkpoint0 = karma.checkpoints(other1, 0);
        assertEq(checkpoint0.fromBlock, t1Block);
        assertEq(checkpoint0.votes, 100);

        Karma.Checkpoint memory checkpoint1 = karma.checkpoints(other1, 1);
        assertEq(checkpoint1.fromBlock, t2Block);
        assertEq(checkpoint1.votes, 90);

        Karma.Checkpoint memory checkpoint2 = karma.checkpoints(other1, 2);
        assertEq(checkpoint2.fromBlock, t3Block);
        assertEq(checkpoint2.votes, 80);

        Karma.Checkpoint memory checkpoint3 = karma.checkpoints(other1, 3);
        assertEq(checkpoint3.fromBlock, t4Block);
        assertEq(checkpoint3.votes, 100);

        // Verify past votes at each checkpoint block
        assertEq(karma.getPastVotes(other1, t1Block), 100);
        assertEq(karma.getPastVotes(other1, t2Block), 90);
        assertEq(karma.getPastVotes(other1, t3Block), 80);
        assertEq(karma.getPastVotes(other1, t4Block), 100);
    }

    function test_DoesNotAddMoreThanOneCheckpointInBlock() public {
        address other1 = makeAddr("other1");
        address other2 = makeAddr("other2");

        // Mint tokens to alice
        vm.prank(owner);
        karma.mint(alice, 1000 ether);

        // Enable transfers for alice and bob
        vm.startPrank(owner);
        karma.setAllowedToTransfer(alice, true);
        karma.setAllowedToTransfer(bob, true);
        vm.stopPrank();

        // Transfer 100 tokens to bob
        vm.prank(alice);
        karma.transfer(bob, 100);

        // Initially other1 has no checkpoints
        assertEq(karma.numCheckpoints(other1), 0);

        // Perform multiple operations in the same block:
        // 1. Bob delegates to other1
        // 2. Bob transfers 10 tokens to other2
        // 3. Bob transfers another 10 tokens to other2
        // All in the same block should result in only ONE checkpoint
        uint256 blockNumber = block.number;

        vm.prank(bob);
        karma.delegate(other1);

        vm.prank(bob);
        karma.transfer(other2, 10);

        vm.prank(bob);
        karma.transfer(other2, 10);

        // Should only have 1 checkpoint despite 3 operations
        assertEq(karma.numCheckpoints(other1), 1);

        // The checkpoint should reflect the final state (80 tokens)
        Karma.Checkpoint memory checkpoint0 = karma.checkpoints(other1, 0);
        assertEq(checkpoint0.fromBlock, blockNumber);
        assertEq(checkpoint0.votes, 80);

        // Move to next block and perform another operation
        vm.roll(block.number + 1);
        vm.prank(alice);
        karma.transfer(bob, 20);
        uint256 t4Block = block.number;

        // Now should have 2 checkpoints
        assertEq(karma.numCheckpoints(other1), 2);

        // Verify the second checkpoint
        Karma.Checkpoint memory checkpoint1 = karma.checkpoints(other1, 1);
        assertEq(checkpoint1.fromBlock, t4Block);
        assertEq(checkpoint1.votes, 100);
    }

    function test_RevertWhen_GetPastVotesForFutureBlock() public {
        address other1 = makeAddr("other1");

        // Attempt to get past votes for a block number far in the future
        vm.expectRevert("ERC20Votes: block not yet mined");
        karma.getPastVotes(other1, 5e10);
    }

    function test_GetPastVotesReturnsZeroWhenNoCheckpoints() public {
        address other1 = makeAddr("other1");

        // Query past votes at block 0 for an account with no checkpoints
        assertEq(karma.getPastVotes(other1, 0), 0);
    }

    function test_GetPastVotesReturnsLatestIfAfterLastCheckpoint() public {
        address other1 = makeAddr("other1");
        uint256 supply = 1000 ether;

        // Mint tokens to alice
        vm.prank(owner);
        karma.mint(alice, supply);

        // Alice delegates to other1
        vm.prank(alice);
        karma.delegate(other1);

        // Advance 2 blocks
        vm.roll(block.number + 1);
        vm.roll(block.number + 1);

        uint256 t1Block = block.number - 2;

        // Queries at or after the checkpoint block should return the same value
        assertEq(karma.getPastVotes(other1, t1Block), supply);
        assertEq(karma.getPastVotes(other1, t1Block + 1), supply);
    }

    function test_GetPastVotesReturnsZeroBeforeFirstCheckpoint() public {
        address other1 = makeAddr("other1");
        uint256 supply = 1000 ether;

        // Mint tokens to alice
        vm.prank(owner);
        karma.mint(alice, supply);

        // Advance a block before delegating
        vm.roll(block.number + 1);

        // Alice delegates to other1
        vm.prank(alice);
        karma.delegate(other1);

        // Advance 2 more blocks
        vm.roll(block.number + 1);
        vm.roll(block.number + 1);

        uint256 t1Block = block.number - 2;

        // Query before the first checkpoint should return 0
        assertEq(karma.getPastVotes(other1, t1Block - 1), 0);

        // Query after the first checkpoint should return supply
        assertEq(karma.getPastVotes(other1, t1Block + 1), supply);
    }

    function test_GetPastVotesReturnsCorrectBalanceAtCheckpoints() public {
        address other1 = makeAddr("other1");
        address other2 = makeAddr("other2");
        uint256 supply = 1000 ether;

        // Mint tokens to alice
        vm.prank(owner);
        karma.mint(alice, supply);

        // Enable transfers for alice and other2
        vm.startPrank(owner);
        karma.setAllowedToTransfer(alice, true);
        karma.setAllowedToTransfer(other2, true);
        vm.stopPrank();

        // t1: Alice delegates to other1
        vm.prank(alice);
        karma.delegate(other1);

        // Advance 2 blocks
        vm.roll(block.number + 1);
        vm.roll(block.number + 1);

        // t2: Alice transfers 10 tokens to other2
        vm.prank(alice);
        karma.transfer(other2, 10);

        // Advance 2 blocks
        vm.roll(block.number + 1);
        vm.roll(block.number + 1);

        // t3: Alice transfers another 10 tokens to other2
        vm.prank(alice);
        karma.transfer(other2, 10);

        // Advance 2 blocks
        vm.roll(block.number + 1);
        vm.roll(block.number + 1);

        // t4: other2 transfers 20 tokens back to alice
        vm.prank(other2);
        karma.transfer(alice, 20);

        // Advance 2 blocks
        vm.roll(block.number + 1);
        vm.roll(block.number + 1);

        uint256 t4Block = block.number - 2;
        uint256 t3Block = t4Block - 2;
        uint256 t2Block = t3Block - 2;
        uint256 t1Block = t2Block - 2;

        // Verify past votes at various checkpoints
        assertEq(karma.getPastVotes(other1, t1Block - 1), 0);
        assertEq(karma.getPastVotes(other1, t1Block), supply);
        assertEq(karma.getPastVotes(other1, t1Block + 1), supply);
        assertEq(karma.getPastVotes(other1, t2Block), supply - 10);
        assertEq(karma.getPastVotes(other1, t2Block + 1), supply - 10);
        assertEq(karma.getPastVotes(other1, t3Block), supply - 20);
        assertEq(karma.getPastVotes(other1, t3Block + 1), supply - 20);
        assertEq(karma.getPastVotes(other1, t4Block), supply);
        assertEq(karma.getPastVotes(other1, t4Block + 1), supply);
    }
}
