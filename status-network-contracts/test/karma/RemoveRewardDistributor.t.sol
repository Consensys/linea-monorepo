// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

import { KarmaTest } from "./Karma.t.sol";
import { KarmaDistributorMock } from "../mocks/KarmaDistributorMock.sol";

contract RemoveRewardDistributorTest is KarmaTest {
    address public distributor;

    function setUp() public virtual override {
        super.setUp();
        distributor = address(new KarmaDistributorMock(IERC20(address(karma))));
    }

    function test_RevertWhen_SenderIsNotDefaultAdmin() public {
        vm.expectRevert();
        karma.removeRewardDistributor(distributor);
    }

    function testRemoveRewardDistributor() public {
        // add a distributor
        vm.prank(owner);
        karma.addRewardDistributor(distributor);
        address[] memory distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 3);
        assertEq(distributors[2], distributor);

        // set some rewards
        uint256 rewardAmount = 1000 ether;
        vm.prank(owner);
        karma.setReward(distributor, rewardAmount, 0);
        // mock distributor distributes all its karma immediately
        uint256 totalSupply = karma.totalSupply();
        assertEq(totalSupply, rewardAmount);

        // remove the distributor
        vm.prank(owner);
        karma.removeRewardDistributor(distributor);
        distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 2);

        assertEq(karma.totalSupply(), totalSupply - rewardAmount);
    }

    function testRemoveRewardDistributorAsOtherAdmin() public {
        // add a distributor
        vm.prank(owner);
        karma.addRewardDistributor(distributor);
        address[] memory distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 3);
        assertEq(distributors[2], distributor);

        // grant admin role
        address otherAdmin = makeAddr("otherAdmin");
        vm.startPrank(owner);
        karma.grantRole(karma.DEFAULT_ADMIN_ROLE(), otherAdmin);
        vm.stopPrank();

        // remove the distributor
        vm.prank(otherAdmin);
        karma.removeRewardDistributor(address(distributor1));
        distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 2);
    }
}
