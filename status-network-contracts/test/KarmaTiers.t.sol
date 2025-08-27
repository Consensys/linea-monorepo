// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

import { Test } from "forge-std/Test.sol";
import { DeploymentConfig } from "../script/DeploymentConfig.s.sol";
import { DeployKarmaTiersScript } from "../script/DeployKarmaTiers.s.sol";
import { KarmaTiers } from "../src/KarmaTiers.sol";

contract KarmaTiersTest is Test {
    KarmaTiers public karmaTiers;
    address public owner;
    address public nonOwner = address(0xBEEF);

    function setUp() public virtual {
        DeployKarmaTiersScript deployment = new DeployKarmaTiersScript();
        (KarmaTiers _karmaTiers, DeploymentConfig deploymentConfig) = deployment.run();
        (address _deployer,) = deploymentConfig.activeNetworkConfig();
        owner = _deployer;
        karmaTiers = _karmaTiers;
    }

    function test_Revert_When_UpdateTiersCalledByNonOwner() public {
        KarmaTiers.Tier[] memory newTiers = new KarmaTiers.Tier[](1);
        newTiers[0] = KarmaTiers.Tier({ minKarma: 0, maxKarma: 100, name: "Bronze", txPerEpoch: 5 });

        vm.prank(nonOwner);
        vm.expectRevert("Ownable: caller is not the owner");
        karmaTiers.updateTiers(newTiers);
    }

    function test_Revert_When_TiersAreEmpty() public {
        KarmaTiers.Tier[] memory newTiers = new KarmaTiers.Tier[](0);

        vm.prank(owner);
        vm.expectRevert(); // generic revert
        karmaTiers.updateTiers(newTiers);
    }

    function test_Revert_When_TiersNotStartingAtZero() public {
        KarmaTiers.Tier[] memory newTiers = new KarmaTiers.Tier[](1);
        newTiers[0] = KarmaTiers.Tier({ minKarma: 1, maxKarma: 100, name: "Bronze", txPerEpoch: 5 });

        vm.prank(owner);
        vm.expectRevert(abi.encodeWithSelector(KarmaTiers.NonContiguousTiers.selector, 0, 0, 1));
        karmaTiers.updateTiers(newTiers);
    }

    function test_Revert_When_TierNameEmpty() public {
        KarmaTiers.Tier[] memory newTiers = new KarmaTiers.Tier[](1);
        newTiers[0] = KarmaTiers.Tier({ minKarma: 0, maxKarma: 100, name: "", txPerEpoch: 5 });

        vm.prank(owner);
        vm.expectRevert(KarmaTiers.EmptyTierName.selector);
        karmaTiers.updateTiers(newTiers);
    }

    function test_Revert_When_TierNameTooLong() public {
        string memory longName = new string(33);
        KarmaTiers.Tier[] memory newTiers = new KarmaTiers.Tier[](1);
        newTiers[0] = KarmaTiers.Tier({ minKarma: 0, maxKarma: 100, name: longName, txPerEpoch: 5 });

        vm.prank(owner);
        vm.expectRevert(abi.encodeWithSelector(KarmaTiers.TierNameTooLong.selector, 33, 32));
        karmaTiers.updateTiers(newTiers);
    }

    function test_Revert_When_TiersNotContiguous() public {
        KarmaTiers.Tier[] memory newTiers = new KarmaTiers.Tier[](2);
        newTiers[0] = KarmaTiers.Tier({ minKarma: 0, maxKarma: 100, name: "Bronze", txPerEpoch: 5 });
        newTiers[1] = KarmaTiers.Tier({ minKarma: 102, maxKarma: 200, name: "Silver", txPerEpoch: 5 });

        vm.prank(owner);
        vm.expectRevert(abi.encodeWithSelector(KarmaTiers.NonContiguousTiers.selector, 1, 101, 102));
        karmaTiers.updateTiers(newTiers);
    }

    function test_Success_When_TiersAreContiguous() public {
        KarmaTiers.Tier[] memory newTiers = new KarmaTiers.Tier[](3);
        newTiers[0] = KarmaTiers.Tier({ minKarma: 0, maxKarma: 100, name: "Bronze", txPerEpoch: 5 });
        newTiers[1] = KarmaTiers.Tier({ minKarma: 101, maxKarma: 200, name: "Silver", txPerEpoch: 10 });
        newTiers[2] = KarmaTiers.Tier({ minKarma: 201, maxKarma: 999, name: "Gold", txPerEpoch: 15 });

        vm.expectEmit(false, false, false, true);
        emit KarmaTiers.TiersUpdated();

        vm.prank(owner);
        karmaTiers.updateTiers(newTiers);

        assertEq(karmaTiers.getTierCount(), 3);
        (uint8 tierId) = karmaTiers.getTierIdByKarmaBalance(150);
        assertEq(tierId, 1);
    }

    function test_Success_When_LastTierIsUnlimited() public {
        KarmaTiers.Tier[] memory newTiers = new KarmaTiers.Tier[](2);
        newTiers[0] = KarmaTiers.Tier({ minKarma: 0, maxKarma: 100, name: "Bronze", txPerEpoch: 5 });
        newTiers[1] =
            KarmaTiers.Tier({ minKarma: 101, maxKarma: type(uint256).max, name: "Unlimited", txPerEpoch: 100 });

        vm.expectEmit(false, false, false, true);
        emit KarmaTiers.TiersUpdated();

        vm.prank(owner);
        karmaTiers.updateTiers(newTiers);

        assertEq(karmaTiers.getTierCount(), 2);
        assertEq(karmaTiers.getTierIdByKarmaBalance(500_000), 1);
    }

    function test_GetTierIdByKarmaBalance_EdgeCases() public {
        KarmaTiers.Tier[] memory newTiers = new KarmaTiers.Tier[](2);
        newTiers[0] = KarmaTiers.Tier({ minKarma: 0, maxKarma: 100, name: "Bronze", txPerEpoch: 5 });
        newTiers[1] = KarmaTiers.Tier({ minKarma: 101, maxKarma: 200, name: "Silver", txPerEpoch: 5 });

        vm.prank(owner);
        karmaTiers.updateTiers(newTiers);

        assertEq(karmaTiers.getTierIdByKarmaBalance(0), 0);
        assertEq(karmaTiers.getTierIdByKarmaBalance(99), 0);
        assertEq(karmaTiers.getTierIdByKarmaBalance(100), 0);
        assertEq(karmaTiers.getTierIdByKarmaBalance(101), 1);
        assertEq(karmaTiers.getTierIdByKarmaBalance(200), 1);
        assertEq(karmaTiers.getTierIdByKarmaBalance(250), 1);
    }
}
