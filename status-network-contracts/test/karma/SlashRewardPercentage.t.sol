// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Math } from "@openzeppelin/contracts/utils/math/Math.sol";
import { Karma } from "../../src/Karma.sol";
import { KarmaTest } from "./Karma.t.sol";

contract SlashRewardPercentageTest is KarmaTest {
    function test_SetSlashRewardPercentageAsAdmin() public {
        vm.prank(owner);
        vm.expectEmit(true, true, true, true);
        emit Karma.SlashRewardPercentageUpdated(1000, 2000);
        karma.setSlashRewardPercentage(2000); // 20%

        assertEq(karma.slashRewardPercentage(), 2000);
    }

    function test_RevertWhen_SetSlashRewardPercentageAsOperator() public {
        bytes memory expectedError = _accessControlError(operator, karma.DEFAULT_ADMIN_ROLE());
        vm.prank(operator);
        vm.expectRevert(expectedError);
        karma.setSlashRewardPercentage(2000);
    }

    function test_RevertWhen_SetSlashRewardPercentageNotAuthorized() public {
        bytes memory expectedError = _accessControlError(alice, karma.DEFAULT_ADMIN_ROLE());
        vm.prank(alice);
        vm.expectRevert(expectedError);
        karma.setSlashRewardPercentage(2000);
    }

    function test_RevertWhen_SetSlashRewardPercentageExceedsMax() public {
        vm.prank(owner);
        vm.expectRevert(Karma.Karma__InvalidSlashRewardPercentage.selector);
        karma.setSlashRewardPercentage(10_001);
    }

    function test_SlashWithCustomRewardPercentage() public {
        // Set custom slash reward percentage to 20%
        vm.prank(owner);
        karma.setSlashRewardPercentage(2000);

        // Setup: mint karma and grant slasher role
        vm.startPrank(owner);
        karma.mint(alice, 100 ether);
        karma.grantRole(karma.SLASHER_ROLE(), owner);
        vm.stopPrank();

        uint256 slashAmount = karma.calculateSlashAmount(100 ether); // 50 ether (50%)

        address rewardRecipient = makeAddr("rewardRecipient");
        vm.prank(owner);
        karma.slash(alice, rewardRecipient);

        uint256 rewardAmount = Math.mulDiv(slashAmount, 2000, 10_000);

        assertEq(karma.balanceOf(rewardRecipient), rewardAmount);
        assertEq(karma.balanceOf(alice), 100 ether - slashAmount);
    }
}
