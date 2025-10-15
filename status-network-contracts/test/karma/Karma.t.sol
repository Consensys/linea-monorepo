// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

import { Test, console } from "forge-std/Test.sol";
import { DeployKarmaScript } from "../../script/DeployKarma.s.sol";
import { DeploymentConfig } from "../../script/DeploymentConfig.s.sol";
import { Karma } from "../../src/Karma.sol";
import { KarmaDistributorMock } from "../mocks/KarmaDistributorMock.sol";

contract KarmaTest is Test {
    Karma public karma;

    address public owner;
    address public alice = makeAddr("alice");
    address public bob = makeAddr("bob");

    KarmaDistributorMock public distributor1;
    KarmaDistributorMock public distributor2;

    address public operator = makeAddr("operator");

    function setUp() public virtual {
        DeployKarmaScript karmaDeployment = new DeployKarmaScript();
        (Karma _karma, DeploymentConfig deploymentConfig) = karmaDeployment.runForTest();
        karma = _karma;
        (address deployer,) = deploymentConfig.activeNetworkConfig();
        owner = deployer;

        distributor1 = new KarmaDistributorMock(IERC20(address(karma)));
        distributor2 = new KarmaDistributorMock(IERC20(address(karma)));

        vm.startBroadcast(owner);
        karma.addRewardDistributor(address(distributor1));
        karma.addRewardDistributor(address(distributor2));
        karma.setAllowedToTransfer(address(distributor1), true);
        karma.setAllowedToTransfer(address(distributor2), true);
        vm.stopBroadcast();
    }

    function _accessControlError(address account, bytes32 role) internal pure returns (bytes memory) {
        string memory expectedError = string(
            abi.encodePacked(
                "AccessControl: account ",
                Strings.toHexString(uint160(account)),
                " is missing role ",
                Strings.toHexString(uint256(role), 32)
            )
        );
        return bytes(expectedError);
    }

    function testAddKarmaDistributorOnlyAdmin() public {
        KarmaDistributorMock distributor3 = new KarmaDistributorMock(IERC20(address(karma)));

        bytes memory expectedError = _accessControlError(alice, karma.DEFAULT_ADMIN_ROLE());
        vm.prank(alice);
        vm.expectRevert(expectedError);
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
        bytes memory expectedError = _accessControlError(alice, karma.DEFAULT_ADMIN_ROLE());
        vm.prank(alice);
        vm.expectRevert(expectedError);
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

    function testActualTokenBalanceOf() public {
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

        uint256 balance = karma.balanceOf(alice);
        uint256 actualBalance = karma.actualTokenBalanceOf(alice);

        assertEq(balance, 3500e18);
        assertEq(actualBalance, 500e18);
    }

    function testMintOnlyAdmin() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), 1000 ether, 1000);
        karma.setReward(address(distributor2), 2000 ether, 2000);
        vm.stopBroadcast();

        distributor1.setTotalKarmaShares(1000 ether);
        distributor2.setTotalKarmaShares(2000 ether);
        assertEq(karma.totalSupply(), 3000 ether);

        vm.prank(alice);
        vm.expectRevert(Karma.Karma__Unauthorized.selector);
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
