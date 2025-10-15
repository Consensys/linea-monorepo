// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

import { KarmaTest } from "./Karma.t.sol";
import { KarmaDistributorMock } from "../mocks/KarmaDistributorMock.sol";

contract SetRewardTest is KarmaTest {
    address public distributor;

    function setUp() public virtual override {
        super.setUp();
        distributor = address(new KarmaDistributorMock(IERC20(address(karma))));
    }

    function test_RevertWhen_SenderIsNotDefaultAdmin() public {
        vm.prank(makeAddr("someone"));
        vm.expectRevert();
        karma.setReward(distributor, 0, 0);
    }

    function test_RevertWhen_SenderIsNotOperator() public {
        assert(karma.hasRole(karma.OPERATOR_ROLE(), operator) == false);

        vm.prank(operator);
        vm.expectRevert();
        karma.setReward(distributor, 0, 0);
    }

    function testSetRewardAsAdmin() public {
        vm.startPrank(owner);
        karma.addRewardDistributor(distributor);
        uint256 rewardAmount = 1000 ether;
        karma.setReward(distributor, rewardAmount, 0);
        vm.stopPrank();
        assertEq(karma.balanceOfRewardDistributor(distributor), rewardAmount);
    }

    function testSetRewardAsOtherAdmin() public {
        vm.startPrank(owner);
        karma.grantRole(karma.DEFAULT_ADMIN_ROLE(), operator);
        karma.addRewardDistributor(distributor);
        vm.stopPrank();

        uint256 rewardAmount = 1000 ether;
        vm.prank(operator);
        karma.setReward(distributor, rewardAmount, 0);
        assertEq(karma.balanceOfRewardDistributor(distributor), rewardAmount);
    }

    function testSetRewardAsOperator() public {
        // grant operator role
        assert(karma.hasRole(karma.DEFAULT_ADMIN_ROLE(), owner));

        // actually `vm.prank()` should be used here, but for some reason
        // foundry seems to mess up the context for what `owner` is
        vm.startPrank(owner);
        karma.grantRole(karma.OPERATOR_ROLE(), operator);
        vm.stopPrank();

        // set reward as operator
        vm.prank(operator);
        karma.setReward(address(distributor1), 0, 0);
    }
}
