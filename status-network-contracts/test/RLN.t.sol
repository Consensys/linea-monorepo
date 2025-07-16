// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";
import { ERC1967Proxy } from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import { RLN } from "../src/rln/RLN.sol";
import { Karma } from "../src/Karma.sol";
import { KarmaDistributorMock } from "./mocks/KarmaDistributorMock.sol";
import { DeployKarmaScript } from "../script/DeployKarma.s.sol";

import { DeploymentConfig } from "../script/DeploymentConfig.s.sol";

contract RLNTest is Test {
    RLN public rln;

    uint256 private constant DEPTH = 2; // for most tests
    uint256 private constant SMALL_DEPTH = 1; // for “full” test

    // Sample identity commitments
    uint256 private identityCommitment0 = 1234;
    uint256 private identityCommitment1 = 5678;
    uint256 private identityCommitment2 = 9999;

    // Role‐holders
    address private owner;
    Karma private karma;
    KarmaDistributorMock public distributor1;
    KarmaDistributorMock public distributor2;

    address private adminAddr;
    address private registerAddr;
    address private slasherAddr;

    address private user1Addr = makeAddr("user1");
    address private user2Addr = makeAddr("user2");
    address private user3Addr = makeAddr("user3");

    function setUp() public {
        DeployKarmaScript karmaDeployment = new DeployKarmaScript();
        (Karma _karma, DeploymentConfig deploymentConfig) = karmaDeployment.run();
        karma = _karma;
        (address deployer,) = deploymentConfig.activeNetworkConfig();
        owner = deployer;
        distributor1 = new KarmaDistributorMock();
        distributor2 = new KarmaDistributorMock();

        // Assign deterministic addresses
        adminAddr = makeAddr("admin");
        registerAddr = makeAddr("register");
        slasherAddr = makeAddr("slasher");

        // Deploy RLN via UUPS proxy with DEPTH = 2
        rln = _deployRLN(DEPTH, karma);

        // Sanity‐check that roles were assigned correctly
        assertTrue(rln.hasRole(rln.DEFAULT_ADMIN_ROLE(), adminAddr));
        assertTrue(rln.hasRole(rln.REGISTER_ROLE(), registerAddr));
        assertTrue(rln.hasRole(rln.SLASHER_ROLE(), slasherAddr));

        vm.startBroadcast(owner);
        karma.addRewardDistributor(address(distributor1));
        karma.addRewardDistributor(address(distributor2));
        karma.grantRole(karma.SLASHER_ROLE(), address(rln));
        vm.stopBroadcast();
    }

    /// @dev Deploys a new RLN instance (behind ERC1967Proxy).
    function _deployRLN(uint256 depth, Karma karmaToken) internal returns (RLN) {
        bytes memory initData = abi.encodeCall(
            RLN.initialize,
            (
                adminAddr,
                slasherAddr,
                registerAddr,
                depth,
                address(karmaToken) // token address unused in these tests
            )
        );
        address impl = address(new RLN());
        address proxy = address(new ERC1967Proxy(impl, initData));
        return RLN(proxy);
    }

    /* ---------- INITIAL STATE ---------- */

    function test_initial_state() public {
        // SET_SIZE should be 2^DEPTH = 4
        assertEq(rln.SET_SIZE(), uint256(1) << DEPTH);

        // No identities registered yet
        assertEq(rln.identityCommitmentIndex(), 0);

        // members(...) should return (address(0), 0) for any commitment
        (address user0, uint256 idx0) = _memberData(identityCommitment0);
        assertEq(user0, address(0));
        assertEq(idx0, 0);
    }

    /* ---------- REGISTER ---------- */

    function test_register_succeeds() public {
        // Register first identity
        uint256 indexBefore = rln.identityCommitmentIndex();
        vm.startPrank(registerAddr);
        vm.expectEmit(true, false, false, true);
        emit RLN.MemberRegistered(identityCommitment0, indexBefore);
        rln.register(identityCommitment0, user1Addr);
        vm.stopPrank();

        assertEq(rln.identityCommitmentIndex(), indexBefore + 1);
        (address u0, uint256 i0) = _memberData(identityCommitment0);
        assertEq(u0, user1Addr);
        assertEq(i0, indexBefore);

        // Register second identity
        indexBefore = rln.identityCommitmentIndex();
        vm.startPrank(registerAddr);
        vm.expectEmit(true, false, false, true);
        emit RLN.MemberRegistered(identityCommitment1, indexBefore);
        rln.register(identityCommitment1, user2Addr);
        vm.stopPrank();

        assertEq(rln.identityCommitmentIndex(), indexBefore + 1);
        (address u1, uint256 i1) = _memberData(identityCommitment1);
        assertEq(u1, user2Addr);
        assertEq(i1, indexBefore);
    }

    function test_register_fails_when_index_exceeds_set_size() public {
        // Deploy a small RLN with depth = 1 => SET_SIZE = 2
        RLN smallRLN = _deployRLN(SMALL_DEPTH, karma);
        address smallRegister = registerAddr;

        // Fill up both slots
        vm.startPrank(smallRegister);
        smallRLN.register(identityCommitment0, user1Addr);
        smallRLN.register(identityCommitment1, user2Addr);
        vm.stopPrank();

        // Now the set is full (2 members). Attempt a third registration.
        vm.startPrank(smallRegister);
        vm.expectRevert(RLN.RLN__SetIsFull.selector);
        smallRLN.register(identityCommitment2, user3Addr);
        vm.stopPrank();
    }

    function test_register_fails_when_duplicate_identity_commitment() public {
        // Register once
        vm.startPrank(registerAddr);
        rln.register(identityCommitment0, user1Addr);
        vm.stopPrank();

        // Attempt to register the same commitment again
        vm.startPrank(registerAddr);
        vm.expectRevert(RLN.RLN__IdCommitmentAlreadyRegistered.selector);
        rln.register(identityCommitment0, user1Addr);
        vm.stopPrank();
    }

    /* ---------- SLASH ---------- */

    function test_slash_succeeds() public {
        uint256 distributorBalance = 50 ether;
        vm.startPrank(owner);
        karma.mint(user2Addr, 10 ether); // Mint Karma tokens to user2
        distributor1.setUserKarmaShare(user2Addr, distributorBalance);
        vm.stopPrank();

        // Register the identity first
        vm.startPrank(registerAddr);
        rln.register(identityCommitment1, user2Addr);
        vm.stopPrank();

        // Retrieve the assigned index
        (, uint256 index1) = _memberData(identityCommitment1);

        // Slash with a valid proof
        vm.startPrank(slasherAddr);
        vm.expectEmit(false, true, false, true);
        emit RLN.MemberSlashed(index1, slasherAddr);
        rln.slash(identityCommitment1);
        vm.stopPrank();

        // After slash, the member record should be cleared
        (address u1, uint256 i1) = _memberData(identityCommitment1);
        assertEq(u1, address(0));
        assertEq(i1, 0);
    }

    function test_slash_fails_when_not_registered() public {
        // Attempt to slash a non‐existent identity
        vm.startPrank(slasherAddr);
        vm.expectRevert(RLN.RLN__MemberNotFound.selector);
        rln.slash(identityCommitment0);
        vm.stopPrank();
    }

    /* ========== HELPERS ========== */

    /// @dev Returns (userAddress, index) for a given identityCommitment.
    function _memberData(uint256 commitment) internal view returns (address userAddress, uint256 index) {
        (userAddress, index) = rln.members(commitment);
        return (userAddress, index);
    }
}
