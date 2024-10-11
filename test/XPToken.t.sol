// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";
import { XPToken } from "../src/XPToken.sol";
import { XPProviderMock } from "./mocks/XPProviderMock.sol";
import { IXPProvider } from "../src/interfaces/IXPProvider.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";

contract XPTokenTest is Test {
    XPToken xpToken;

    address owner = makeAddr("owner");
    address alice = makeAddr("alice");
    address bob = makeAddr("bob");

    XPProviderMock provider1;
    XPProviderMock provider2;

    function setUp() public {
        vm.prank(owner);
        xpToken = new XPToken(1000e18);

        provider1 = new XPProviderMock();
        provider2 = new XPProviderMock();

        vm.prank(owner);
        xpToken.addXPProvider(provider1);

        vm.prank(owner);
        xpToken.addXPProvider(provider2);
    }

    function testAddXPProviderOnlyOwner() public {
        XPProviderMock provider3 = new XPProviderMock();

        vm.prank(alice);
        vm.expectPartialRevert(Ownable.OwnableUnauthorizedAccount.selector);
        xpToken.addXPProvider(provider3);

        vm.prank(owner);
        xpToken.addXPProvider(provider3);

        IXPProvider[] memory providers = xpToken.getXPProviders();
        assertEq(providers.length, 3);
        assertEq(address(providers[0]), address(provider1));
        assertEq(address(providers[1]), address(provider2));
        assertEq(address(providers[2]), address(provider3));
    }

    function testRemoveXPProviderOnlyOwner() public {
        vm.prank(alice);
        vm.expectPartialRevert(Ownable.OwnableUnauthorizedAccount.selector);
        xpToken.removeXPProvider(0);

        vm.prank(owner);
        xpToken.removeXPProvider(0);

        IXPProvider[] memory providers = xpToken.getXPProviders();
        assertEq(providers.length, 1);
        assertEq(address(providers[0]), address(provider2));
    }

    function testRemoveXPProviderIndexOutOfBounds() public {
        vm.prank(owner);
        vm.expectRevert(XPToken.XPProvider__IndexOutOfBounds.selector);
        xpToken.removeXPProvider(10);
    }

    function testTotalSupply() public view {
        uint256 totalSupply = xpToken.totalSupply();
        assertEq(totalSupply, 1000 ether);
    }

    function testBalanceOfWithNoSystemTotalXP() public view {
        uint256 aliceBalance = xpToken.balanceOf(alice);
        assertEq(aliceBalance, 0);

        uint256 bobBalance = xpToken.balanceOf(bob);
        assertEq(bobBalance, 0);
    }

    function testBalanceOf() public {
        provider1.setUserXPShare(alice, 100e18);
        provider1.setTotalXPShares(1000e18);

        provider2.setUserXPShare(alice, 200e18);
        provider2.setTotalXPShares(2000e18);

        // Expected balance calculation
        uint256 userTotalXP = 100e18 + 200e18;
        uint256 systemTotalXP = 1000e18 + 2000e18;

        uint256 expectedBalance = (xpToken.totalSupply() * userTotalXP) / systemTotalXP;

        uint256 balance = xpToken.balanceOf(alice);
        assertEq(balance, expectedBalance);
    }

    function testSetTotalSupplyOnlyOwner() public {
        uint256 totalSupply = xpToken.totalSupply();
        assertEq(totalSupply, 1000e18);

        vm.prank(alice);
        vm.expectPartialRevert(Ownable.OwnableUnauthorizedAccount.selector);
        xpToken.setTotalSupply(2000e18);

        vm.prank(owner);
        xpToken.setTotalSupply(2000e18);
        totalSupply = xpToken.totalSupply();
        assertEq(totalSupply, 2000e18);
    }

    function testTransfersNotAllowed() public {
        vm.expectRevert(XPToken.XPToken__TransfersNotAllowed.selector);
        xpToken.transfer(alice, 100e18);

        vm.expectRevert(XPToken.XPToken__TransfersNotAllowed.selector);
        xpToken.approve(alice, 100e18);

        vm.expectRevert(XPToken.XPToken__TransfersNotAllowed.selector);
        xpToken.transferFrom(alice, bob, 100e18);

        uint256 allowance = xpToken.allowance(alice, bob);
        assertEq(allowance, 0);
    }
}

contract XPTokenOwnershipTest is Test {
    XPToken xpToken;

    address owner = makeAddr("owner");
    address alice = makeAddr("alice");

    function setUp() public {
        vm.prank(owner);
        xpToken = new XPToken(1000e18);
    }

    function testInitialOwner() public view {
        assertEq(xpToken.owner(), owner);
    }

    function testOwnershipTransfer() public {
        vm.prank(owner);
        xpToken.transferOwnership(alice);
        assertEq(xpToken.owner(), owner);

        vm.prank(alice);
        xpToken.acceptOwnership();
        assertEq(xpToken.owner(), alice);
    }
}
