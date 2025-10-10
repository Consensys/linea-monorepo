// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";

import { Test } from "forge-std/Test.sol";

import { DeployKarmaScript } from "../script/DeployKarma.s.sol";
import { DeploySimpleKarmaDistributorScript } from "../script/DeploySimpleKarmaDistributor.s.sol";
import { DeploymentConfig } from "../script/DeploymentConfig.s.sol";
import { Karma } from "../src/Karma.sol";
import { SimpleKarmaDistributor } from "../src/SimpleKarmaDistributor.sol";

contract SimpleKarmaDistributorTest is Test {
    SimpleKarmaDistributor internal distributor;
    Karma internal karma;

    address internal owner;
    address internal operator = makeAddr("operator");
    address internal alice = makeAddr("alice");

    function setUp() public virtual {
        DeployKarmaScript karmaDeployment = new DeployKarmaScript();
        (Karma _karma, DeploymentConfig deploymentConfig) = karmaDeployment.runForTest();
        karma = _karma;
        (address deployer,) = deploymentConfig.activeNetworkConfig();
        owner = deployer;

        DeploySimpleKarmaDistributorScript distributorDeployment = new DeploySimpleKarmaDistributorScript();
        (distributor,) = distributorDeployment.deploy(owner, address(karma));

        vm.startPrank(owner);
        distributor.setRewardsSupplier(address(karma));
        distributor.grantRole(distributor.OPERATOR_ROLE(), operator);
        karma.addRewardDistributor(address(distributor));
        karma.setAllowedToTransfer(address(distributor), true);
        vm.stopPrank();
    }

    function _accessControlError(address account, bytes32 role) internal pure returns (bytes memory) {
        return bytes(
            string(
                abi.encodePacked(
                    "AccessControl: account ",
                    Strings.toHexString(uint160(account)),
                    " is missing role ",
                    Strings.toHexString(uint256(role), 32)
                )
            )
        );
    }

    function test_SetRewardsRevertsIfNotAdmin() public {
        bytes memory expectedError = _accessControlError(alice, distributor.DEFAULT_ADMIN_ROLE());
        vm.prank(alice);
        vm.expectRevert(expectedError);
        distributor.setRewardsSupplier(alice);
    }

    function test_SetRewardsUpdatesSupplierIfAdmin() public {
        assertNotEq(distributor.rewardsSupplier(), alice);
        vm.prank(owner);
        distributor.setRewardsSupplier(alice);
        assertEq(distributor.rewardsSupplier(), alice);
    }

    function test_GrantRoleRevertsIfNotAdmin() public {
        bytes memory expectedError = _accessControlError(alice, distributor.DEFAULT_ADMIN_ROLE());
        bytes32 operatorRole = distributor.OPERATOR_ROLE();
        vm.prank(alice);
        vm.expectRevert(expectedError);
        distributor.grantRole(operatorRole, alice);
    }

    function test_GrantRoleCanBeUsedByAdmin() public {
        bytes32 operatorRole = distributor.OPERATOR_ROLE();
        vm.prank(owner);
        distributor.grantRole(operatorRole, alice);
        assert(distributor.hasRole(distributor.OPERATOR_ROLE(), alice));
    }

    function test_SetRewardUpdatesAvailableSupply() public {
        uint256 amount = 100 ether;

        vm.prank(owner);
        karma.setReward(address(distributor), amount, 0);

        assertEq(distributor.availableSupply(), amount);
        assertEq(distributor.totalRewardsSupply(), 0);
    }

    function test_SetRewardOnlySupplier() public {
        uint256 amount = 10 ether;
        vm.prank(owner);
        distributor.setRewardsSupplier(address(karma));

        vm.prank(alice);
        vm.expectRevert(SimpleKarmaDistributor.SimpleKarmaDistributor__Unauthorized.selector);
        distributor.setReward(amount, 0);
    }

    function test_MintByAdminAdjustsSupply() public {
        uint256 rewards = 200 ether;
        uint256 mintAmount = 50 ether;

        vm.prank(owner);
        karma.setReward(address(distributor), rewards, 0);

        vm.prank(owner);
        distributor.mint(alice, mintAmount);

        assertEq(distributor.availableSupply(), rewards - mintAmount);
        assertEq(distributor.mintedSupply(), mintAmount);
        assertEq(distributor.totalRewardsSupply(), mintAmount);
        assertEq(distributor.rewardsBalanceOfAccount(alice), mintAmount);
    }

    function test_MintByOperator() public {
        uint256 rewards = 200 ether;
        uint256 mintAmount = 50 ether;

        vm.prank(owner);
        karma.setReward(address(distributor), rewards, 0);

        vm.prank(operator);
        distributor.mint(alice, mintAmount);

        assertEq(distributor.availableSupply(), rewards - mintAmount);
        assertEq(distributor.mintedSupply(), mintAmount);
        assertEq(distributor.totalRewardsSupply(), mintAmount);
        assertEq(distributor.rewardsBalanceOfAccount(alice), mintAmount);
    }

    function test_MintRevertsWhenInsufficientSupply() public {
        uint256 rewards = 200 ether;
        uint256 mintAmount = 300 ether;

        vm.prank(owner);
        karma.setReward(address(distributor), rewards, 0);

        vm.prank(operator);
        vm.expectRevert(SimpleKarmaDistributor.SimpleKarmaDistributor__InsufficientAvailableSupply.selector);
        distributor.mint(alice, mintAmount);
    }

    function test_MintRevertsWhenZeroAmount() public {
        vm.prank(operator);
        vm.expectRevert(SimpleKarmaDistributor.SimpleKarmaDistributor__AmountCannotBeZero.selector);
        distributor.mint(alice, 0);
    }

    function test_RedeemRewardsTransfersKarma() public {
        uint256 rewards = 200 ether;
        uint256 mintAmount = 50 ether;

        vm.prank(owner);
        karma.setReward(address(distributor), rewards, 0);

        vm.prank(operator);
        distributor.mint(alice, mintAmount);

        assertEq(karma.balanceOf(alice), mintAmount);
        assertEq(distributor.rewardsBalanceOfAccount(alice), mintAmount);

        vm.prank(alice);
        uint256 redeemed = distributor.redeemRewards(alice);

        assertEq(redeemed, mintAmount);
        assertEq(distributor.mintedSupply(), 0);
        assertEq(distributor.rewardsBalanceOfAccount(alice), 0);
        assertEq(karma.balanceOf(alice), mintAmount);
        assertEq(distributor.totalRewardsSupply(), 0);
    }

    function test_RedeemRewardsWhenNoBalance() public {
        uint256 redeemed = distributor.redeemRewards(alice);
        assertEq(redeemed, 0);
    }
}
