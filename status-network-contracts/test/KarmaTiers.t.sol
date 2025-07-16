// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

import { Test } from "forge-std/Test.sol";
import { DeploymentConfig } from "../script/DeploymentConfig.s.sol";
import { DeployKarmaTiersScript } from "../script/DeployKarmaTiers.s.sol";
import { KarmaTiers } from "../src/KarmaTiers.sol";

contract KarmaTiersTest is Test {
    KarmaTiers public karmaTiers;
    address public deployer;
    address public nonOwner = makeAddr("nonOwner");

    event TierAdded(uint8 indexed tierId, string name, uint256 minKarma, uint256 maxKarma, uint32 txPerEpoch);
    event TierUpdated(uint8 indexed tierId, string name, uint256 minKarma, uint256 maxKarma, uint32 txPerEpoch);
    event TierDeactivated(uint8 indexed tierId);
    event TierActivated(uint8 indexed tierId);

    function setUp() public virtual {
        DeployKarmaTiersScript deployment = new DeployKarmaTiersScript();
        (KarmaTiers _karmaTiers, DeploymentConfig deploymentConfig) = deployment.run();
        (address _deployer,) = deploymentConfig.activeNetworkConfig();
        deployer = _deployer;
        karmaTiers = _karmaTiers;
    }
}

contract AddTierTests is KarmaTiersTest {
    function setUp() public override {
        super.setUp();
    }

    function test_AddTier_RevertWhen_EmptyName() public {
        vm.prank(deployer);
        vm.expectRevert(KarmaTiers.EmptyTierName.selector);
        karmaTiers.addTier("", 100, 500, 5);
    }

    function test_AddTier_RevertWhen_InvalidRange() public {
        vm.prank(deployer);
        vm.expectRevert(abi.encodeWithSelector(KarmaTiers.InvalidTierRange.selector, 500, 100));
        karmaTiers.addTier("Invalid", 500, 100, 5);
    }

    function test_AddTier_RevertWhen_InvalidRangeEqual() public {
        vm.prank(deployer);
        vm.expectRevert(abi.encodeWithSelector(KarmaTiers.InvalidTierRange.selector, 500, 500));
        karmaTiers.addTier("Invalid", 500, 500, 5);
    }

    function test_AddTier_RevertWhen_OverlappingTiers() public {
        vm.prank(deployer);
        karmaTiers.addTier("Bronze", 100, 500, 5);

        vm.prank(deployer);
        vm.expectRevert(abi.encodeWithSelector(KarmaTiers.OverlappingTiers.selector, 1, 400, 600));
        karmaTiers.addTier("Silver", 400, 600, 5);
    }

    function test_AddTier_RevertWhen_NotOwner() public {
        vm.prank(nonOwner);
        vm.expectRevert("Ownable: caller is not the owner");
        karmaTiers.addTier("Bronze", 100, 500, 5);
    }

    function test_AddTier_RevertWhen_TierNameTooLong() public {
        string memory longName = "ThisIsAVeryLongTierNameThatExceedsTheMaximumAllowedLength";

        vm.expectRevert(
            abi.encodeWithSelector(
                KarmaTiers.TierNameTooLong.selector, bytes(longName).length, karmaTiers.MAX_TIER_NAME_LENGTH()
            )
        );
        vm.prank(deployer);
        karmaTiers.addTier(longName, 100, 500, 5);
    }

    function test_AddTier_Success() public {
        string memory tierName = "Bronze";
        uint256 minKarma = 100;
        uint256 maxKarma = 500;
        uint32 txPerEpoch = 5;

        vm.expectEmit(true, false, false, true);
        emit TierAdded(1, tierName, minKarma, maxKarma, txPerEpoch);

        vm.prank(deployer);
        karmaTiers.addTier(tierName, minKarma, maxKarma, txPerEpoch);

        assertEq(karmaTiers.currentTierId(), 1);

        KarmaTiers.Tier memory tier = karmaTiers.getTierById(1);
        assertEq(tier.name, tierName);
        assertEq(tier.minKarma, minKarma);
        assertEq(tier.maxKarma, maxKarma);
        assertTrue(tier.active);
    }

    function test_AddTier_UnlimitedMaxKarma() public {
        string memory tierName = "Unlimited";
        uint256 minKarma = 1000;
        uint256 maxKarma = 0; // 0 means unlimited
        uint32 txPerEpoch = 5;

        vm.prank(deployer);
        karmaTiers.addTier(tierName, minKarma, maxKarma, txPerEpoch);

        KarmaTiers.Tier memory tier = karmaTiers.getTierById(1);
        assertEq(tier.maxKarma, 0);
    }

    function test_AddTier_MultipleSuccessiveTiers() public {
        vm.startPrank(deployer);
        karmaTiers.addTier("Bronze", 0, 100, 5);
        karmaTiers.addTier("Silver", 101, 500, 5);
        karmaTiers.addTier("Gold", 501, 1000, 5);
        vm.stopPrank();

        assertEq(karmaTiers.currentTierId(), 3);
        assertEq(karmaTiers.getTierCount(), 3);
    }
}

contract UpdateTierTests is KarmaTiersTest {
    function setUp() public override {
        super.setUp();
        vm.prank(deployer);
        karmaTiers.addTier("Bronze", 100, 500, 5); // Add a tier to update
    }

    function test_UpdateTier_RevertWhen_InvalidTierId() public {
        vm.expectRevert(KarmaTiers.TierNotFound.selector);
        vm.prank(deployer);
        karmaTiers.updateTier(2, "Bronze", 100, 500, 5);
    }

    function test_UpdateTier_RevertWhen_InvalidRange() public {
        vm.expectRevert(abi.encodeWithSelector(KarmaTiers.InvalidTierRange.selector, 600, 400));
        vm.prank(deployer);
        karmaTiers.updateTier(1, "Bronze", 600, 400, 5);
    }

    function test_UpdateTier_RevertWhen_OverlapWithOtherTier() public {
        vm.startPrank(deployer);
        karmaTiers.addTier("Silver", 600, 1000, 5);
        vm.stopPrank();

        vm.expectRevert(abi.encodeWithSelector(KarmaTiers.OverlappingTiers.selector, 2, 550, 800));
        vm.prank(deployer);
        karmaTiers.updateTier(1, "Bronze", 550, 800, 5);
    }

    function test_UpdateTier_RevertWhen_NotOwner() public {
        vm.prank(nonOwner);
        vm.expectRevert("Ownable: caller is not the owner");
        karmaTiers.updateTier(1, "Updated Bronze", 150, 600, 5);
    }

    function test_UpdateTier_Success() public {
        string memory newName = "Updated Bronze";
        uint256 newMinKarma = 150;
        uint256 newMaxKarma = 600;
        uint32 newTxPerEpoch = 10;

        vm.expectEmit(true, false, false, true);
        emit TierUpdated(1, newName, newMinKarma, newMaxKarma, newTxPerEpoch);

        vm.prank(deployer);
        karmaTiers.updateTier(1, newName, newMinKarma, newMaxKarma, newTxPerEpoch);

        KarmaTiers.Tier memory tier = karmaTiers.getTierById(1);
        assertEq(tier.name, newName);
        assertEq(tier.minKarma, newMinKarma);
        assertEq(tier.maxKarma, newMaxKarma);
    }
}

contract DeactivateActivateTierTests is KarmaTiersTest {
    function setUp() public override {
        super.setUp();
    }

    function test_DeactivateTier_RevertWhen_InvalidTierId() public {
        vm.expectRevert(KarmaTiers.TierNotFound.selector);
        vm.prank(deployer);
        karmaTiers.deactivateTier(1);
    }

    function test_DeactivateTier_RevertWhen_NotOwner() public {
        vm.prank(deployer);
        karmaTiers.addTier("Bronze", 100, 500, 5);

        vm.prank(nonOwner);
        vm.expectRevert("Ownable: caller is not the owner");
        karmaTiers.deactivateTier(1);
    }

    function test_DeactivateTier_Success() public {
        vm.prank(deployer);
        karmaTiers.addTier("Bronze", 100, 500, 5);

        vm.expectEmit(true, false, false, false);
        emit TierDeactivated(1);

        vm.prank(deployer);
        karmaTiers.deactivateTier(1);

        KarmaTiers.Tier memory tier = karmaTiers.getTierById(1);
        assertFalse(tier.active);
    }
}

contract ActivateTierTests is KarmaTiersTest {
    function setUp() public override {
        super.setUp();
    }

    function test_ActivateTier_Success() public {
        vm.startPrank(deployer);
        karmaTiers.addTier("Bronze", 100, 500, 5);
        karmaTiers.deactivateTier(1);
        vm.stopPrank();

        vm.expectEmit(true, false, false, false);
        emit TierActivated(1);

        vm.prank(deployer);
        karmaTiers.activateTier(1);

        KarmaTiers.Tier memory tier = karmaTiers.getTierById(1);
        assertTrue(tier.active);
    }

    function test_ActivateTier_RevertWhen_InvalidTierId() public {
        vm.expectRevert(KarmaTiers.TierNotFound.selector);
        vm.prank(deployer);
        karmaTiers.activateTier(1);
    }

    function test_ActivateTier_RevertWhen_NotOwner() public {
        vm.prank(deployer);
        karmaTiers.addTier("Bronze", 100, 500, 5);
        vm.prank(deployer);
        karmaTiers.deactivateTier(1);

        vm.prank(nonOwner);
        vm.expectRevert("Ownable: caller is not the owner");
        karmaTiers.activateTier(1);
    }
}

contract ViewFunctionsTest is KarmaTiersTest {
    function setUp() public override {
        super.setUp();
    }

    function test_GetTierCount_InitiallyZero() public {
        assertEq(karmaTiers.getTierCount(), 0);
    }

    function test_GetTierCount_IncreasesWithTiers() public {
        vm.prank(deployer);
        karmaTiers.addTier("Bronze", 100, 500, 5);
        assertEq(karmaTiers.getTierCount(), 1);

        vm.prank(deployer);
        karmaTiers.addTier("Silver", 600, 1000, 5);
        assertEq(karmaTiers.getTierCount(), 2);
    }

    function test_GetTierById_Success() public {
        string memory tierName = "Bronze";
        uint256 minKarma = 100;
        uint256 maxKarma = 500;
        uint32 txPerEpoch = 5;

        vm.prank(deployer);
        karmaTiers.addTier(tierName, minKarma, maxKarma, txPerEpoch);

        KarmaTiers.Tier memory tier = karmaTiers.getTierById(1);
        assertEq(tier.name, tierName);
        assertEq(tier.minKarma, minKarma);
        assertEq(tier.maxKarma, maxKarma);
        assertTrue(tier.active);
    }

    function test_GetTierById_RevertWhen_InvalidTierId() public {
        vm.expectRevert(KarmaTiers.TierNotFound.selector);
        karmaTiers.getTierById(1);
    }

    function test_GetTierById_RevertWhen_TierIdZero() public {
        vm.expectRevert(KarmaTiers.TierNotFound.selector);
        karmaTiers.getTierById(0);
    }
}

contract EdgeCasesTest is KarmaTiersTest {
    function setUp() public override {
        super.setUp();
    }

    function test_OverlapValidation_EdgeCases() public {
        // Test adjacent ranges (should not overlap)
        vm.startPrank(deployer);
        karmaTiers.addTier("Tier1", 0, 100, 5);
        karmaTiers.addTier("Tier2", 101, 200, 5);
        vm.stopPrank();

        // Test touching ranges (100 and 101 are adjacent, should not overlap)
        assertEq(karmaTiers.getTierCount(), 2);

        // Test exact boundary overlap (should fail)
        vm.expectRevert(abi.encodeWithSelector(KarmaTiers.OverlappingTiers.selector, 1, 100, 150));
        vm.prank(deployer);
        karmaTiers.addTier("Tier3", 100, 150, 5);
    }

    function test_UnlimitedTierOverlap() public {
        // Add unlimited tier
        vm.prank(deployer);
        karmaTiers.addTier("Unlimited", 1000, 0, 5);

        // Try to add tier that overlaps with unlimited tier
        vm.expectRevert(abi.encodeWithSelector(KarmaTiers.OverlappingTiers.selector, 1, 1500, 2000));
        vm.prank(deployer);
        karmaTiers.addTier("Overlap", 1500, 2000, 5);

        // Try to add tier that starts before unlimited tier
        vm.expectRevert(abi.encodeWithSelector(KarmaTiers.OverlappingTiers.selector, 1, 500, 1500));
        vm.prank(deployer);
        karmaTiers.addTier("Before", 500, 1500, 5);
    }
}

contract GetTierIdByKarmaBalanceTest is KarmaTiersTest {
    function setUp() public override {
        super.setUp();
        vm.startPrank(deployer);
        karmaTiers.addTier("Bronze", 100, 500, 5);
        karmaTiers.addTier("Silver", 501, 1000, 5);
        karmaTiers.addTier("Gold", 1001, 1500, 5);
        karmaTiers.addTier("Platinum", 5001, 10_000, 5); // creating a gap
        // let's also take into account that tiers aren't sorted
        karmaTiers.addTier("Wood", 10, 99, 5);

        karmaTiers.deactivateTier(3); // Deactivate Gold tier for testing
        vm.stopPrank();
    }

    function test_GetTierIdByKarmaBalance_BelowMinKarma() public {
        uint256 karmaBalance = 5;
        uint8 tierId = karmaTiers.getTierIdByKarmaBalance(karmaBalance);
        assertEq(tierId, 0); // Should return 0 for no tier
    }

    function test_GetTierIdByKarmaBalance_InBronzeTier() public {
        uint256 karmaBalance = 300;
        uint8 tierId = karmaTiers.getTierIdByKarmaBalance(karmaBalance);
        assertEq(tierId, 1);
    }

    function test_GetTierIdByKarmaBalance_InSilverTier() public {
        uint256 karmaBalance = 800;
        uint8 tierId = karmaTiers.getTierIdByKarmaBalance(karmaBalance);
        assertEq(tierId, 2);
    }

    function test_GetTierIdByKarmaBalance_InGoldTier() public {
        uint256 karmaBalance = 1200;
        uint8 tierId = karmaTiers.getTierIdByKarmaBalance(karmaBalance);
        assertEq(tierId, 2); // Since Gold is deactivated, should return 2 for Silver
    }
}

contract FuzzTests is KarmaTiersTest {
    function setUp() public override {
        super.setUp();
    }

    function testFuzz_AddTier_ValidInputs(
        string calldata name,
        uint256 minKarma,
        uint256 maxKarma,
        uint32 txPerEpoch
    )
        public
    {
        vm.assume(bytes(name).length > 0 && bytes(name).length <= 32);
        vm.assume(maxKarma == 0 || maxKarma > minKarma);
        vm.assume(minKarma < type(uint256).max);
        vm.assume(txPerEpoch > 0);

        vm.prank(deployer);
        karmaTiers.addTier(name, minKarma, maxKarma, txPerEpoch);

        KarmaTiers.Tier memory tier = karmaTiers.getTierById(1);
        assertEq(tier.name, name);
        assertEq(tier.minKarma, minKarma);
        assertEq(tier.maxKarma, maxKarma);
        assertTrue(tier.active);
    }

    function testFuzz_UpdateTier_ValidInputs(
        string calldata initialName,
        uint256 initialMinKarma,
        uint256 initialMaxKarma,
        string calldata newName,
        uint256 newMinKarma,
        uint256 newMaxKarma,
        uint32 initialTxPerEpoch,
        uint32 newTxPerEpoch
    )
        public
    {
        // Setup constraints for initial tier
        vm.assume(bytes(initialName).length > 0 && bytes(initialName).length <= 32);
        vm.assume(initialMaxKarma == 0 || initialMaxKarma > initialMinKarma);

        // Setup constraints for new tier
        vm.assume(bytes(newName).length > 0 && bytes(newName).length <= 32);
        vm.assume(newMaxKarma == 0 || newMaxKarma > newMinKarma);
        vm.assume(initialTxPerEpoch > 0);
        vm.assume(newTxPerEpoch > 0);

        vm.startPrank(deployer);
        karmaTiers.addTier(initialName, initialMinKarma, initialMaxKarma, initialTxPerEpoch);
        karmaTiers.updateTier(1, newName, newMinKarma, newMaxKarma, newTxPerEpoch);
        vm.stopPrank();

        KarmaTiers.Tier memory tier = karmaTiers.getTierById(1);
        assertEq(tier.name, newName);
        assertEq(tier.minKarma, newMinKarma);
        assertEq(tier.maxKarma, newMaxKarma);
    }
}
