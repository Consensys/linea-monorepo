// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity ^0.8.26;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

import "forge-std/Test.sol";

import "../src/rln/RLN.sol";
import { IVerifier } from "../src/rln/IVerifier.sol";

// A ERC20 token contract which allows arbitrary minting for testing
contract TestERC20 is ERC20 {
    constructor() ERC20("TestERC20", "TST") { }

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

// A mock verifier which makes us skip the proof verification.
contract MockVerifier is IVerifier {
    bool public result;

    constructor() {
        result = true;
    }

    function verifyProof(
        uint256[2] memory,
        uint256[2][2] memory,
        uint256[2] memory,
        uint256[2] memory
    )
        external
        view
        returns (bool)
    {
        return result;
    }

    function changeResult(bool _result) external {
        result = _result;
    }
}

contract RLNTest is Test {
    event MemberRegistered(uint256 identityCommitment, uint256 messageLimit, uint256 index);
    event MemberWithdrawn(uint256 index);
    event MemberSlashed(uint256 index, address slasher);

    RLN rln;
    TestERC20 token;
    MockVerifier verifier;

    uint256 rlnInitialTokenBalance = 1_000_000;
    uint256 minimalDeposit = 100;
    uint256 maximalRate = 1 << 16 - 1;
    uint256 depth = 20;
    uint8 feePercentage = 10;
    address feeReceiver = makeAddr("feeReceiver");
    uint256 freezePeriod = 1;

    uint256 identityCommitment0 = 1234;
    uint256 identityCommitment1 = 5678;

    address user0 = makeAddr("user0");
    address user1 = makeAddr("user1");
    address slashedReceiver = makeAddr("slashedReceiver");

    uint256 messageLimit0 = 2;
    uint256 messageLimit1 = 3;

    uint256[8] mockProof =
        [uint256(0), uint256(1), uint256(2), uint256(3), uint256(4), uint256(5), uint256(6), uint256(7)];

    function setUp() public {
        token = new TestERC20();
        verifier = new MockVerifier();
        rln = new RLN(
            minimalDeposit,
            maximalRate,
            depth,
            feePercentage,
            feeReceiver,
            freezePeriod,
            address(token),
            address(verifier)
        );
    }

    function test_initial_state() public {
        assertEq(rln.MINIMAL_DEPOSIT(), minimalDeposit);
        assertEq(rln.SET_SIZE(), 1 << depth);
        assertEq(rln.FEE_PERCENTAGE(), feePercentage);
        assertEq(rln.FEE_RECEIVER(), feeReceiver);
        assertEq(rln.FREEZE_PERIOD(), freezePeriod);
        assertEq(address(rln.token()), address(token));
        assertEq(address(rln.verifier()), address(verifier));
        assertEq(rln.identityCommitmentIndex(), 0);
    }

    /* register */

    function test_register_succeeds() public {
        // Test: register one user
        register(user0, identityCommitment0, messageLimit0);
        // Test: register second user
        register(user1, identityCommitment1, messageLimit1);
    }

    function test_register_fails_when_index_exceeds_set_size() public {
        // Set size is (1 << smallDepth) = 2, and thus there can
        // only be 2 members, otherwise reverts.
        uint256 smallDepth = 1;
        TestERC20 _token = new TestERC20();
        RLN smallRLN = new RLN(
            minimalDeposit, maximalRate, smallDepth, feePercentage, feeReceiver, 0, address(_token), address(verifier)
        );

        // Register the first user
        _token.mint(user0, minimalDeposit);
        vm.startPrank(user0);
        _token.approve(address(smallRLN), minimalDeposit);
        smallRLN.register(identityCommitment0, minimalDeposit);
        vm.stopPrank();
        // Register the second user
        _token.mint(user1, minimalDeposit);
        vm.startPrank(user1);
        _token.approve(address(smallRLN), minimalDeposit);
        smallRLN.register(identityCommitment1, minimalDeposit);
        vm.stopPrank();
        // Now tree (set) is full. Try register the third. It should revert.
        address user2 = makeAddr("user2");
        uint256 identityCommitment2 = 9999;
        token.mint(user2, minimalDeposit);
        vm.startPrank(user2);
        token.approve(address(smallRLN), minimalDeposit);
        // `register` should revert
        vm.expectRevert("RLN, register: set is full");
        smallRLN.register(identityCommitment2, minimalDeposit);
        vm.stopPrank();
    }

    function test_register_fails_when_amount_lt_minimal_deposit() public {
        uint256 insufficientAmount = minimalDeposit - 1;
        token.mint(user0, rlnInitialTokenBalance);
        vm.startPrank(user0);
        token.approve(address(rln), rlnInitialTokenBalance);
        vm.expectRevert("RLN, register: amount is lower than minimal deposit");
        rln.register(identityCommitment0, insufficientAmount);
        vm.stopPrank();
    }

    function test_register_fails_when_duplicate_identity_commitments() public {
        // Register first with user0 with identityCommitment0
        register(user0, identityCommitment0, messageLimit0);
        // Register again with user1 with identityCommitment0
        token.mint(user1, rlnInitialTokenBalance);
        vm.startPrank(user1);
        token.approve(address(rln), rlnInitialTokenBalance);
        // `register` should revert
        vm.expectRevert("RLN, register: idCommitment already registered");
        rln.register(identityCommitment0, rlnInitialTokenBalance);
        vm.stopPrank();
    }

    /* withdraw */

    function test_withdraw_succeeds() public {
        // Register first
        register(user0, identityCommitment0, messageLimit0);
        // Make sure proof verification is skipped
        assertEq(verifier.result(), true);

        // Withdraw user0
        // Ensure event is emitted
        (,, uint256 index) = rln.members(identityCommitment0);
        vm.expectEmit(true, true, false, true);
        emit MemberWithdrawn(index);
        rln.withdraw(identityCommitment0, mockProof);
        // Check withdrawal entry is set correctly
        (uint256 blockNumber, uint256 amount, address receiver) = rln.withdrawals(identityCommitment0);
        assertEq(blockNumber, block.number);
        assertEq(amount, getRegisterAmount(messageLimit0));
        assertEq(receiver, user0);
    }

    function test_withdraw_fails_when_not_registered() public {
        // Withdraw fails if the user has not registered before
        vm.expectRevert("RLN, withdraw: member doesn't exist");
        rln.withdraw(identityCommitment0, mockProof);
    }

    function test_withdraw_fails_when_already_underways() public {
        // Register first
        register(user0, identityCommitment0, messageLimit0);
        // Withdraw user0
        rln.withdraw(identityCommitment0, mockProof);
        // Withdraw again and it should fail
        vm.expectRevert("RLN, release: such withdrawal exists");
        rln.withdraw(identityCommitment0, mockProof);
    }

    function test_withdraw_fails_when_invalid_proof() public {
        // Register first
        register(user0, identityCommitment0, messageLimit0);
        // Make sure mock verifier always return false
        // And thus the proof is always considered invalid
        verifier.changeResult(false);
        assertEq(verifier.result(), false);
        vm.expectRevert("RLN, withdraw: invalid proof");
        rln.withdraw(identityCommitment0, mockProof);
    }

    /* release */

    function test_release_succeeds() public {
        // Register first
        register(user0, identityCommitment0, messageLimit0);
        // Withdraw user0
        // Make sure proof verification is skipped
        assertEq(verifier.result(), true);
        rln.withdraw(identityCommitment0, mockProof);

        // Test: release succeeds after freeze period
        // Set block.number to `blockNumbersToRelease`
        uint256 blockNumbersToRelease = getUnfrozenBlockHeight();
        vm.roll(blockNumbersToRelease);

        uint256 user0BalanceBefore = token.balanceOf(user0);
        uint256 rlnBalanceBefore = token.balanceOf(address(rln));
        // Calls release and check balances
        rln.release(identityCommitment0);
        uint256 user0BalanceDiff = token.balanceOf(user0) - user0BalanceBefore;
        uint256 rlnBalanceDiff = rlnBalanceBefore - token.balanceOf(address(rln));
        uint256 expectedUser0BalanceDiff = getRegisterAmount(messageLimit0);
        assertEq(user0BalanceDiff, expectedUser0BalanceDiff);
        assertEq(rlnBalanceDiff, expectedUser0BalanceDiff);
        checkUserIsDeleted(identityCommitment0);
    }

    function test_release_fails_when_no_withdrawal() public {
        // Release fails if there is no withdrawal for the user
        vm.expectRevert("RLN, release: no such withdrawals");
        rln.release(identityCommitment0);
    }

    function test_release_fails_when_freeze_period() public {
        // Register first
        register(user0, identityCommitment0, messageLimit0);
        // Make sure mock verifier always return true to skip proof verification
        assertEq(verifier.result(), true);
        // Withdraw user0
        rln.withdraw(identityCommitment0, mockProof);
        // Ensure withdrawal is set
        (uint256 blockNumber, uint256 amount, address receiver) = rln.withdrawals(identityCommitment0);
        assertEq(blockNumber, block.number);
        assertEq(amount, getRegisterAmount(messageLimit0));
        assertEq(receiver, user0);

        // Test: release fails in freeze period
        vm.expectRevert("RLN, release: cannot release yet");
        rln.release(identityCommitment0);
        // Set block.number to blockNumbersToRelease - 1, which is still in freeze period
        uint256 blockNumbersToRelease = getUnfrozenBlockHeight();
        vm.roll(blockNumbersToRelease - 1);
        vm.expectRevert("RLN, release: cannot release yet");
        rln.release(identityCommitment0);
    }

    /* slash */

    function test_slash_succeeds() public {
        // Test: register and get slashed
        register(user0, identityCommitment0, messageLimit0);
        uint256 registerAmount = getRegisterAmount(messageLimit0);
        uint256 slashFee = getSlashFee(registerAmount);
        uint256 slashReward = registerAmount - slashFee;
        uint256 slashedReceiverBalanceBefore = token.balanceOf(slashedReceiver);
        uint256 rlnBalanceBefore = token.balanceOf(address(rln));
        uint256 feeReceiverBalanceBefore = token.balanceOf(feeReceiver);
        // ensure event is emitted
        (,, uint256 index) = rln.members(identityCommitment0);
        vm.expectEmit(true, true, false, true);
        emit MemberSlashed(index, slashedReceiver);
        // Slash and check balances
        rln.slash(identityCommitment0, slashedReceiver, mockProof);
        uint256 slashedReceiverBalanceDiff = token.balanceOf(slashedReceiver) - slashedReceiverBalanceBefore;
        uint256 rlnBalanceDiff = rlnBalanceBefore - token.balanceOf(address(rln));
        uint256 feeReceiverBalanceDiff = token.balanceOf(feeReceiver) - feeReceiverBalanceBefore;
        assertEq(slashedReceiverBalanceDiff, slashReward);
        assertEq(rlnBalanceDiff, registerAmount);
        assertEq(feeReceiverBalanceDiff, slashFee);
        // Check the record of user0 has been deleted
        checkUserIsDeleted(identityCommitment0);

        // Test: register, withdraw, ang get slashed before release
        register(user1, identityCommitment1, messageLimit1);
        rln.withdraw(identityCommitment1, mockProof);
        rln.slash(identityCommitment1, slashedReceiver, mockProof);
        // Check the record of user1 has been deleted
        checkUserIsDeleted(identityCommitment1);
    }

    function test_slash_fails_when_receiver_is_zero() public {
        // Register first
        register(user0, identityCommitment0, messageLimit0);
        // Try slash user0 and it fails because of the zero address
        vm.expectRevert("RLN, slash: empty receiver address");
        rln.slash(identityCommitment0, address(0), mockProof);
    }

    function test_slash_fails_when_not_registered() public {
        // It fails if the user is not registered yet
        vm.expectRevert("RLN, slash: member doesn't exist");
        rln.slash(identityCommitment0, slashedReceiver, mockProof);
    }

    function test_slash_fails_when_self_slashing() public {
        // `slash` fails when receiver is the same as the registered msg.sender
        register(user0, identityCommitment0, messageLimit0);
        vm.expectRevert("RLN, slash: self-slashing is prohibited");
        rln.slash(identityCommitment0, user0, mockProof);
    }

    function test_slash_fails_when_invalid_proof() public {
        // It fails if the proof is invalid
        // Register first
        register(user0, identityCommitment0, messageLimit0);
        // Make sure mock verifier always return false
        // And thus the proof is always considered invalid
        verifier.changeResult(false);
        assertEq(verifier.result(), false);
        vm.expectRevert("RLN, slash: invalid proof");
        // Slash fails because of the invalid proof
        rln.slash(identityCommitment0, slashedReceiver, mockProof);
    }

    /* Helpers */
    function getRegisterAmount(uint256 messageLimit) public view returns (uint256) {
        return messageLimit * minimalDeposit;
    }

    function register(address user, uint256 identityCommitment, uint256 messageLimit) public {
        // Mint to user first
        uint256 registerTokenAmount = getRegisterAmount(messageLimit);
        token.mint(user, registerTokenAmount);
        // Remember the balance for later check
        uint256 tokenRLNBefore = token.balanceOf(address(rln));
        uint256 tokenUserBefore = token.balanceOf(user);
        uint256 identityCommitmentIndexBefore = rln.identityCommitmentIndex();
        // User approves to rln and calls register
        vm.startPrank(user);
        token.approve(address(rln), registerTokenAmount);
        // Ensure event is emitted
        vm.expectEmit(true, true, false, true);
        emit MemberRegistered(identityCommitment, messageLimit, identityCommitmentIndexBefore);
        rln.register(identityCommitment, registerTokenAmount);
        vm.stopPrank();

        // Check states
        uint256 tokenRLNDiff = token.balanceOf(address(rln)) - tokenRLNBefore;
        uint256 tokenUserDiff = tokenUserBefore - token.balanceOf(user);
        // RLN state
        assertEq(rln.identityCommitmentIndex(), identityCommitmentIndexBefore + 1);
        assertEq(tokenRLNDiff, registerTokenAmount);
        // User state
        (address userAddress, uint256 actualMessageLimit, uint256 index) = rln.members(identityCommitment);
        assertEq(userAddress, user);
        assertEq(actualMessageLimit, messageLimit);
        assertEq(index, identityCommitmentIndexBefore);
        assertEq(tokenUserDiff, registerTokenAmount);
    }

    function getUnfrozenBlockHeight() public view returns (uint256) {
        return block.number + freezePeriod + 1;
    }

    function checkUserIsDeleted(uint256 identityCommitment) public {
        // User state
        (address userAddress, uint256 actualMessageLimit, uint256 index) = rln.members(identityCommitment);
        assertEq(userAddress, address(0));
        assertEq(actualMessageLimit, 0);
        assertEq(index, 0);
        // Withdrawal state
        (uint256 blockNumber, uint256 amount, address receiver) = rln.withdrawals(identityCommitment);
        assertEq(blockNumber, 0);
        assertEq(amount, 0);
        assertEq(receiver, address(0));
    }

    function getSlashFee(uint256 registerAmount) public view returns (uint256) {
        return registerAmount * feePercentage / 100;
    }
}
