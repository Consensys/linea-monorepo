// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.26;

import { Karma } from "../Karma.sol";

import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";

/// @title Rate-Limiting Nullifier registry contract
/// @dev This contract allows you to register RLN commitment and withdraw/slash.
contract RLN is Initializable, UUPSUpgradeable, AccessControlUpgradeable {
    bytes32 public constant SLASHER_ROLE = keccak256("SLASHER_ROLE");
    bytes32 public constant REGISTER_ROLE = keccak256("REGISTER_ROLE");

    error RLN__MemberNotFound();
    error RLN__IdCommitmentAlreadyRegistered();
    error RLN__SetIsFull();
    error RLN__Unauthorized();

    /// @dev User metadata struct.
    /// @param userAddress: address of depositor;
    struct User {
        address userAddress;
        uint256 index;
    }

    /// @dev Registry set size (1 << DEPTH).
    uint256 public SET_SIZE;

    /// @dev Current index where identityCommitment will be stored.
    uint256 public identityCommitmentIndex;

    /// @dev Registry set. The keys are `identityCommitment`s.
    /// The values are addresses of accounts that call `register` transaction.
    mapping(uint256 commitment => User user) public members;

    /// @dev Karma Token used for registering.
    Karma public karma;

    /// @dev Emmited when a new member registered.
    /// @param identityCommitment: `identityCommitment`;
    /// @param index: idCommitmentIndex value.
    event MemberRegistered(uint256 identityCommitment, uint256 index);

    /// @dev Emmited when a member was slashed.
    /// @param index: index of `identityCommitment`;
    /// @param slasher: address of slasher (msg.sender).
    event MemberSlashed(uint256 index, address slasher);

    constructor() {
        _disableInitializers();
    }

    /// @dev Constructor.
    /// @param _owner: address of the owner of the contract;
    /// @param _slasher: address of the slasher;
    /// @param _register: address of the register;
    /// @param depth: depth of the merkle tree;
    /// @param _token: address of the ERC20 contract;
    function initialize(
        address _owner,
        address _slasher,
        address _register,
        uint256 depth,
        address _token
    )
        public
        initializer
    {
        __UUPSUpgradeable_init();
        __AccessControl_init();
        _setupRole(DEFAULT_ADMIN_ROLE, _owner);
        _setupRole(SLASHER_ROLE, _slasher);
        _setupRole(REGISTER_ROLE, _register);
        SET_SIZE = 1 << depth;

        karma = Karma(_token);
    }

    /**
     * @notice Authorizes contract upgrades via UUPS.
     * @dev This function is only callable by the owner.
     */
    function _authorizeUpgrade(address) internal view override {
        if (!hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert RLN__Unauthorized();
        }
    }

    /// @dev Adds `identityCommitment` to the registry set and takes the necessary stake amount.
    ///
    /// NOTE: The set must not be full.
    ///
    /// @param identityCommitment: `identityCommitment`;
    function register(uint256 identityCommitment, address user) external onlyRole(REGISTER_ROLE) {
        uint256 index = identityCommitmentIndex;
        if (index >= SET_SIZE) {
            revert RLN__SetIsFull();
        }
        if (members[identityCommitment].userAddress != address(0)) {
            revert RLN__IdCommitmentAlreadyRegistered();
        }

        members[identityCommitment] = User(user, index);
        emit MemberRegistered(identityCommitment, index);

        unchecked {
            identityCommitmentIndex = index + 1;
        }
    }

    /// @dev Slashes identity with identityCommitment.
    /// @param identityCommitment: `identityCommitment`;
    function slash(uint256 identityCommitment) external onlyRole(SLASHER_ROLE) {
        User memory member = members[identityCommitment];
        if (member.userAddress == address(0)) {
            revert RLN__MemberNotFound();
        }
        karma.slash(member.userAddress);
        delete members[identityCommitment];

        emit MemberSlashed(member.index, msg.sender);
    }
}
