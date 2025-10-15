// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

import { KarmaTest } from "./Karma.t.sol";
import { KarmaDistributorMock } from "../mocks/KarmaDistributorMock.sol";

contract AddRewardDistributorTest is KarmaTest {
    address public distributor;

    function setUp() public virtual override {
        super.setUp();
        distributor = address(new KarmaDistributorMock(IERC20(address(karma))));
    }

    function test_RevertWhen_SenderIsNotDefaultAdmin() public {
        vm.prank(makeAddr("someone"));
        vm.expectRevert();
        karma.addRewardDistributor(distributor);
    }

    function testAddRewardDistributorAsOtherAdmin() public {
        address otherAdmin = makeAddr("otherAdmin");
        vm.startPrank(owner);
        karma.grantRole(karma.DEFAULT_ADMIN_ROLE(), otherAdmin);
        vm.stopPrank();

        vm.startPrank(otherAdmin);
        karma.addRewardDistributor(distributor);
        vm.stopPrank();
        address[] memory distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 3);
        assertEq(distributors[2], distributor);
    }
}
