// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity ^0.8.26;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import { IVerifier } from "./IVerifier.sol";

/// @title Rate-Limiting Nullifier registry contract
/// @dev This contract allows you to register RLN commitment and withdraw/slash.
contract RLN {
    using SafeERC20 for IERC20;

    /// @dev User metadata struct.
    /// @param userAddress: address of depositor;
    /// @param messageLimit: user's message limit (stakeAmount / MINIMAL_DEPOSIT).
    struct User {
        address userAddress;
        uint256 messageLimit;
        uint256 index;
    }

    /// @dev Withdrawal time-lock struct
    /// @param blockNumber: number of block when a withdraw was initialized;
    /// @param messageLimit: amount of tokens to freeze/release;
    /// @param receiver: address of receiver.
    struct Withdrawal {
        uint256 blockNumber;
        uint256 amount;
        address receiver;
    }

    /// @dev Minimal membership deposit (stake amount) value - cost of 1 message.
    uint256 public immutable MINIMAL_DEPOSIT;

    /// @dev Maximal rate.
    uint256 public immutable MAXIMAL_RATE;

    /// @dev Registry set size (1 << DEPTH).
    uint256 public immutable SET_SIZE;

    /// @dev Address of the fee receiver.
    address public immutable FEE_RECEIVER;

    /// @dev Fee percentage.
    uint8 public immutable FEE_PERCENTAGE;

    /// @dev Freeze period - number of blocks for which the withdrawal of money is frozen.
    uint256 public immutable FREEZE_PERIOD;

    /// @dev Current index where identityCommitment will be stored.
    uint256 public identityCommitmentIndex;

    /// @dev Registry set. The keys are `identityCommitment`s.
    /// The values are addresses of accounts that call `register` transaction.
    mapping(uint256 => User) public members;

    /// @dev Withdrawals logic.
    mapping(uint256 => Withdrawal) public withdrawals;

    /// @dev ERC20 Token used for staking.
    IERC20 public immutable token;

    /// @dev Groth16 verifier.
    IVerifier public immutable verifier;

    /// @dev Emmited when a new member registered.
    /// @param identityCommitment: `identityCommitment`;
    /// @param messageLimit: user's message limit;
    /// @param index: idCommitmentIndex value.
    event MemberRegistered(uint256 identityCommitment, uint256 messageLimit, uint256 index);

    /// @dev Emmited when a member was withdrawn.
    /// @param index: index of `identityCommitment`;
    event MemberWithdrawn(uint256 index);

    /// @dev Emmited when a member was slashed.
    /// @param index: index of `identityCommitment`;
    /// @param slasher: address of slasher (msg.sender).
    event MemberSlashed(uint256 index, address slasher);

    /// @param minimalDeposit: minimal membership deposit;
    /// @param maximalRate: maximal rate;
    /// @param depth: depth of the merkle tree;
    /// @param feePercentage: fee percentage;
    /// @param feeReceiver: address of the fee receiver;
    /// @param freezePeriod: amount of blocks for withdrawal time-lock;
    /// @param _token: address of the ERC20 contract;
    /// @param _verifier: address of the Groth16 Verifier.
    constructor(
        uint256 minimalDeposit,
        uint256 maximalRate,
        uint256 depth,
        uint8 feePercentage,
        address feeReceiver,
        uint256 freezePeriod,
        address _token,
        address _verifier
    ) {
        require(feeReceiver != address(0), "RLN, constructor: fee receiver cannot be 0x0 address");

        MINIMAL_DEPOSIT = minimalDeposit;
        MAXIMAL_RATE = maximalRate;
        SET_SIZE = 1 << depth;

        FEE_PERCENTAGE = feePercentage;
        FEE_RECEIVER = feeReceiver;
        FREEZE_PERIOD = freezePeriod;

        token = IERC20(_token);
        verifier = IVerifier(_verifier);
    }

    /// @dev Adds `identityCommitment` to the registry set and takes the necessary stake amount.
    ///
    /// NOTE: The set must not be full.
    ///
    /// @param identityCommitment: `identityCommitment`;
    /// @param amount: stake amount.
    function register(uint256 identityCommitment, uint256 amount) external {
        uint256 index = identityCommitmentIndex;

        require(index < SET_SIZE, "RLN, register: set is full");
        require(amount >= MINIMAL_DEPOSIT, "RLN, register: amount is lower than minimal deposit");
        require(amount % MINIMAL_DEPOSIT == 0, "RLN, register: amount should be a multiple of minimal deposit");
        require(members[identityCommitment].userAddress == address(0), "RLN, register: idCommitment already registered");

        uint256 messageLimit = amount / MINIMAL_DEPOSIT;
        require(messageLimit <= MAXIMAL_RATE, "RLN, register: message limit cannot be more than MAXIMAL_RATE");

        token.safeTransferFrom(msg.sender, address(this), amount);

        members[identityCommitment] = User(msg.sender, messageLimit, index);
        emit MemberRegistered(identityCommitment, messageLimit, index);

        unchecked {
            identityCommitmentIndex = index + 1;
        }
    }

    /// @dev Request for withdraw and freeze the stake to prevent self-slashing. Stake can be
    /// released after FREEZE_PERIOD blocks.
    /// @param identityCommitment: `identityCommitment`;
    /// @param proof: snarkjs's format generated proof (without public inputs) packed consequently.
    function withdraw(uint256 identityCommitment, uint256[8] calldata proof) external {
        User memory member = members[identityCommitment];
        require(member.userAddress != address(0), "RLN, withdraw: member doesn't exist");
        require(withdrawals[identityCommitment].blockNumber == 0, "RLN, release: such withdrawal exists");
        require(_verifyProof(identityCommitment, member.userAddress, proof), "RLN, withdraw: invalid proof");

        uint256 withdrawAmount = member.messageLimit * MINIMAL_DEPOSIT;
        withdrawals[identityCommitment] = Withdrawal(block.number, withdrawAmount, member.userAddress);
        emit MemberWithdrawn(member.index);
    }

    /// @dev Releases stake amount.
    /// @param identityCommitment: `identityCommitment` of withdrawn user.
    function release(uint256 identityCommitment) external {
        Withdrawal memory withdrawal = withdrawals[identityCommitment];
        require(withdrawal.blockNumber != 0, "RLN, release: no such withdrawals");
        require(block.number - withdrawal.blockNumber > FREEZE_PERIOD, "RLN, release: cannot release yet");

        delete withdrawals[identityCommitment];
        delete members[identityCommitment];

        token.safeTransfer(withdrawal.receiver, withdrawal.amount);
    }

    /// @dev Slashes identity with identityCommitment.
    /// @param identityCommitment: `identityCommitment`;
    /// @param receiver: stake receiver;
    /// @param proof: snarkjs's format generated proof (without public inputs) packed consequently.
    function slash(uint256 identityCommitment, address receiver, uint256[8] calldata proof) external {
        require(receiver != address(0), "RLN, slash: empty receiver address");

        User memory member = members[identityCommitment];
        require(member.userAddress != address(0), "RLN, slash: member doesn't exist");
        require(member.userAddress != receiver, "RLN, slash: self-slashing is prohibited");

        require(_verifyProof(identityCommitment, receiver, proof), "RLN, slash: invalid proof");

        delete members[identityCommitment];
        delete withdrawals[identityCommitment];

        uint256 withdrawAmount = member.messageLimit * MINIMAL_DEPOSIT;
        uint256 feeAmount = (FEE_PERCENTAGE * withdrawAmount) / 100;

        token.safeTransfer(receiver, withdrawAmount - feeAmount);
        token.safeTransfer(FEE_RECEIVER, feeAmount);
        emit MemberSlashed(member.index, receiver);
    }

    /// @dev Groth16 proof verification
    function _verifyProof(
        uint256 identityCommitment,
        address receiver,
        uint256[8] calldata proof
    )
        internal
        view
        returns (bool)
    {
        return verifier.verifyProof(
            [proof[0], proof[1]],
            [[proof[2], proof[3]], [proof[4], proof[5]]],
            [proof[6], proof[7]],
            [identityCommitment, uint256(uint160(receiver))]
        );
    }
}
