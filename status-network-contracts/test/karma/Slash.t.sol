// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Karma } from "../../src/Karma.sol";
import { KarmaTest } from "./Karma.t.sol";

contract SlashTest is KarmaTest {
    address public slasher = makeAddr("slasher");

    function _mintKarmaToAccount(address account, uint256 amount) internal {
        vm.startPrank(owner);
        karma.mint(account, amount);
        vm.stopPrank();
    }

    function setUp() public override {
        super.setUp();

        vm.startPrank(owner);
        karma.grantRole(karma.SLASHER_ROLE(), slasher);
        vm.stopPrank();
    }

    function test_RevertWhen_SenderIsNotDefaultAdminOrSlasher() public {
        vm.prank(makeAddr("someone"));
        vm.expectRevert(Karma.Karma__Unauthorized.selector);
        karma.slash(alice, address(0));
    }

    function test_RevertWhen_KarmaBalanceIsInvalid() public {
        vm.prank(slasher);
        vm.expectRevert(Karma.Karma__CannotSlashZeroBalance.selector);
        karma.slash(alice, address(0));
    }

    function test_SlashRemainingBalanceIfBalanceIsLow() public {
        uint256 initialBalance = karma.MIN_SLASH_AMOUNT() - 1;
        _mintKarmaToAccount(alice, initialBalance);

        address rewardRecipient = makeAddr("rewardRecipient");

        vm.prank(slasher);
        uint256 slashed = karma.slash(alice, rewardRecipient);

        // The entire balance should be slashed
        // slashRewardPercentage (default 10%) goes to recipient
        uint256 rewardAmount = (slashed * karma.slashRewardPercentage()) / 10_000;

        // Verify recipient received the reward
        assertEq(karma.balanceOf(rewardRecipient), rewardAmount);
        // Alice should have 0 balance (everything slashed, reward went to recipient)
        assertEq(karma.balanceOf(alice), 0);
    }

    function test_Slash() public {
        // ensure rewards
        uint256 currentBalance = 100 ether;
        _mintKarmaToAccount(alice, currentBalance);
        uint256 slashedAmount = karma.calculateSlashAmount(currentBalance);

        // slash the account with no reward recipient
        vm.prank(slasher);
        karma.slash(alice, address(0));

        // With address(0) recipient, entire amount is burned (no reward minted back)
        assertEq(karma.balanceOf(alice), currentBalance - slashedAmount);

        currentBalance = karma.balanceOf(alice);
        slashedAmount = karma.calculateSlashAmount(currentBalance);

        // slash again
        vm.prank(slasher);
        karma.slash(alice, address(0));

        // Same - entire amount burned
        assertEq(karma.balanceOf(alice), currentBalance - slashedAmount);
    }

    function testFuzz_Slash(uint256 rewardsAmount) public {
        vm.assume(rewardsAmount > 0);
        vm.assume(rewardsAmount <= type(uint128).max);
        _mintKarmaToAccount(alice, rewardsAmount);
        uint256 slashAmount = karma.calculateSlashAmount(rewardsAmount);

        vm.prank(slasher);
        karma.slash(alice, address(0));

        // With address(0) recipient, entire amount is burned (no reward minted back)
        assertEq(karma.balanceOf(alice), rewardsAmount - slashAmount);
    }
}
