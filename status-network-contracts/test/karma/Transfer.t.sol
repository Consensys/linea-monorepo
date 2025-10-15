// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Karma } from "../../src/Karma.sol";
import { KarmaTest } from "./Karma.t.sol";

contract TransferTest is KarmaTest {
    function setUp() public override {
        super.setUp();
    }

    function test_RevertWhen_TransferIsNotAllowed() public {
        vm.expectRevert(Karma.Karma__TransfersNotAllowed.selector);
        karma.transfer(alice, 100e18);
    }

    function test_RevertWhen_AmountExceedsAccountSlashAmount() public {
        uint256 amount = 1000e18;
        vm.startPrank(owner);

        karma.mint(alice, amount);
        assertEq(karma.balanceOf(alice), amount);

        uint256 slashedAmount = karma.calculateSlashAmount(karma.balanceOf(alice));

        karma.grantRole(karma.SLASHER_ROLE(), owner);
        karma.slash(alice, address(0));

        // With address(0) recipient, entire amount is burned (no reward minted back)
        uint256 transferableAmount = amount - slashedAmount;
        assertEq(karma.balanceOf(alice), transferableAmount);

        karma.setAllowedToTransfer(alice, true);
        vm.stopPrank();

        vm.prank(alice);
        vm.expectRevert("ERC20: transfer amount exceeds balance");
        karma.transfer(bob, transferableAmount + 1);
    }

    function test_Transfer() public {
        uint256 amount = 1000e18;
        vm.prank(owner);
        karma.mint(alice, amount);
        assertEq(karma.balanceOf(alice), amount);

        vm.prank(owner);
        karma.setAllowedToTransfer(alice, true);

        vm.prank(alice);
        karma.transfer(bob, amount);

        assertEq(karma.balanceOf(alice), 0);
        assertEq(karma.balanceOf(bob), amount);
    }
}
