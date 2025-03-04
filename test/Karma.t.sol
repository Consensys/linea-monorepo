// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";
import { DeployKarmaScript } from "../script/DeployKarma.s.sol";
import { DeploymentConfig } from "../script/DeploymentConfig.s.sol";
import { Karma } from "../src/Karma.sol";
import { KarmaDistributorMock } from "./mocks/KarmaDistributorMock.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";

contract KarmaTest is Test {
    Karma public karma;

    address public owner;
    address public alice = makeAddr("alice");
    address public bob = makeAddr("bob");

    KarmaDistributorMock public distributor1;
    KarmaDistributorMock public distributor2;

    function setUp() public virtual {
        DeployKarmaScript karmaDeployment = new DeployKarmaScript();
        (Karma _karma, DeploymentConfig deploymentConfig) = karmaDeployment.run();
        karma = _karma;
        (address deployer,,) = deploymentConfig.activeNetworkConfig();
        owner = deployer;

        distributor1 = new KarmaDistributorMock();
        distributor2 = new KarmaDistributorMock();

        vm.startBroadcast(owner);
        karma.addRewardDistributor(address(distributor1));
        karma.addRewardDistributor(address(distributor2));
        vm.stopBroadcast();
    }

    function testAddKarmaDistributorOnlyOwner() public {
        KarmaDistributorMock distributor3 = new KarmaDistributorMock();

        vm.prank(alice);
        vm.expectRevert("Ownable: caller is not the owner");
        karma.addRewardDistributor(address(distributor3));

        vm.prank(owner);
        karma.addRewardDistributor(address(distributor3));

        address[] memory distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 3);
        assertEq(distributors[0], address(distributor1));
        assertEq(distributors[1], address(distributor2));
        assertEq(distributors[2], address(distributor3));
    }

    function testRemoveKarmaDistributorOnlyOwner() public {
        vm.prank(alice);
        vm.expectRevert("Ownable: caller is not the owner");
        karma.removeRewardDistributor(address(distributor1));

        vm.prank(owner);
        karma.removeRewardDistributor(address(distributor1));

        address[] memory distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 1);
        assertEq(distributors[0], address(distributor2));
    }

    function testRemoveUnknownKarmaDistributor() public {
        vm.prank(owner);
        vm.expectRevert(Karma.Karma__UnknownDistributor.selector);
        karma.removeRewardDistributor(address(1));
    }

    function testTotalSupply() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), 1000 ether, 1000);
        karma.setReward(address(distributor2), 2000 ether, 2000);
        vm.stopBroadcast();

        distributor1.setTotalKarmaShares(1000 ether);
        distributor2.setTotalKarmaShares(2000 ether);

        vm.prank(owner);
        karma.mint(owner, 500 ether);

        uint256 totalSupply = karma.totalSupply();
        assertEq(totalSupply, 3500 ether);
    }

    function testBalanceOfWithNoSystemTotalKarma() public view {
        uint256 aliceBalance = karma.balanceOf(alice);
        assertEq(aliceBalance, 0);

        uint256 bobBalance = karma.balanceOf(bob);
        assertEq(bobBalance, 0);
    }

    function testBalanceOf() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), 1000 ether, 1000);
        karma.setReward(address(distributor2), 2000 ether, 2000);
        vm.stopBroadcast();

        distributor1.setTotalKarmaShares(1000 ether);
        distributor2.setTotalKarmaShares(2000 ether);

        distributor1.setUserKarmaShare(alice, 1000e18);
        distributor2.setUserKarmaShare(alice, 2000e18);

        vm.prank(owner);
        karma.mint(alice, 500e18);

        uint256 expectedBalance = 3500e18;

        uint256 balance = karma.balanceOf(alice);
        assertEq(balance, expectedBalance);
    }

    function testMintOnlyOwner() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), 1000 ether, 1000);
        karma.setReward(address(distributor2), 2000 ether, 2000);
        vm.stopBroadcast();

        distributor1.setTotalKarmaShares(1000 ether);
        distributor2.setTotalKarmaShares(2000 ether);
        assertEq(karma.totalSupply(), 3000 ether);

        vm.prank(alice);
        vm.expectRevert("Ownable: caller is not the owner");
        karma.mint(alice, 1000e18);

        vm.prank(owner);
        karma.mint(alice, 1000e18);
        assertEq(karma.totalSupply(), 4000e18);
    }

    function testTransfersNotAllowed() public {
        vm.expectRevert(Karma.Karma__TransfersNotAllowed.selector);
        karma.transfer(alice, 100e18);

        vm.expectRevert(Karma.Karma__TransfersNotAllowed.selector);
        karma.approve(alice, 100e18);

        vm.expectRevert(Karma.Karma__TransfersNotAllowed.selector);
        karma.transferFrom(alice, bob, 100e18);

        uint256 allowance = karma.allowance(alice, bob);
        assertEq(allowance, 0);
    }
}

contract KarmaOwnershipTest is KarmaTest {
    function setUp() public override {
        super.setUp();
    }

    function testInitialOwner() public view {
        assertEq(karma.owner(), owner);
    }

    function testOwnershipTransfer() public {
        vm.prank(owner);
        karma.transferOwnership(alice);
        assertEq(karma.owner(), owner);

        vm.prank(alice);
        karma.acceptOwnership();
        assertEq(karma.owner(), alice);
    }
}

contract KarmaMintAllowanceTest is KarmaTest {
    function setUp() public override {
        super.setUp();
    }

    function testMintAllowance_Available() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), 1000 ether, 1000);
        karma.setReward(address(distributor2), 2000 ether, 2000);
        vm.stopBroadcast();
        // 3000 external => maxSupply = 9000
        distributor1.setTotalKarmaShares(1000 ether);
        distributor2.setTotalKarmaShares(2000 ether);

        vm.prank(owner);
        karma.mint(owner, 500 ether);
        // totalSupply = 3500

        uint256 mintAllowance = karma.mintAllowance();
        assertEq(mintAllowance, 5500 ether);
    }

    function testMintAllowance_NotAvailable() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), 1000 ether, 1000);
        karma.setReward(address(distributor2), 2000 ether, 2000);
        vm.stopBroadcast();
        // 3000 external => maxSupply = 9000
        distributor1.setTotalKarmaShares(1000 ether);
        distributor2.setTotalKarmaShares(2000 ether);

        vm.prank(owner);
        karma.mint(owner, 6000 ether);
        // totalSupply = 9_000

        uint256 mintAllowance = karma.mintAllowance();
        assertEq(mintAllowance, 0);
    }

    function testMint_RevertWithAllowanceExceeded() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), 1000 ether, 1000);
        karma.setReward(address(distributor2), 2000 ether, 2000);
        vm.stopBroadcast();
        // 3000 external => maxSupply = 9000
        distributor1.setTotalKarmaShares(1000 ether);
        distributor2.setTotalKarmaShares(2000 ether);

        vm.prank(owner);
        karma.mint(owner, 500 ether);
        // totalSupply = 3500
        // allowed to mint 5500

        vm.prank(owner);
        vm.expectRevert(Karma.Karma__MintAllowanceExceeded.selector);
        karma.mint(owner, 6000 ether);
    }

    function testMint_Ok() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), 1000 ether, 1000);
        karma.setReward(address(distributor2), 2000 ether, 2000);
        vm.stopBroadcast();
        // 3000 external => maxSupply = 9000
        distributor1.setTotalKarmaShares(1000 ether);
        distributor2.setTotalKarmaShares(2000 ether);

        vm.prank(owner);
        karma.mint(owner, 500 ether);
        assertEq(karma.totalSupply(), 3500 ether);
        // totalSupply = 3500
        // allowed to mint 5500

        vm.prank(owner);
        karma.mint(owner, 5500 ether);
        assertEq(karma.totalSupply(), 9000 ether);
    }
}
