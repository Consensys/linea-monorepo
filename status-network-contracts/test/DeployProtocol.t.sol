// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";
import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";

import { DeployProtocolScript } from "../script/DeployProtocol.s.sol";
import { DeploymentConfig } from "../script/DeploymentConfig.s.sol";
import { MockToken } from "./mocks/MockToken.sol";

import { Karma } from "../src/Karma.sol";
import { KarmaNFT } from "../src/KarmaNFT.sol";
import { StakeManager } from "../src/StakeManager.sol";
import { VaultFactory } from "../src/VaultFactory.sol";
import { INFTMetadataGenerator } from "../src/interfaces/INFTMetadataGenerator.sol";

contract DeployProtocolTest is Test {
    DeployProtocolScript deployProtocol;
    MockToken stakingToken;
    address deployer;

    Karma karma;
    address karmaImpl;
    INFTMetadataGenerator metadataGenerator;
    KarmaNFT karmaNFT;
    StakeManager stakeManager;
    VaultFactory vaultFactory;
    address vaultImpl;
    DeploymentConfig deploymentConfig;

    function setUp() public {
        stakingToken = new MockToken("Staking Token", "STK");

        deployProtocol = new DeployProtocolScript();

        (karma, metadataGenerator, karmaNFT, stakeManager, vaultFactory, vaultImpl, deploymentConfig) =
            deployProtocol.runForTest(address(stakingToken));

        (deployer,) = deploymentConfig.activeNetworkConfig();
    }

    function testKarmaOwnership() public view {
        bytes32 defaultAdminRole = karma.DEFAULT_ADMIN_ROLE();
        assertTrue(karma.hasRole(defaultAdminRole, deployer), "Deployer should have admin role on Karma");
    }

    function testStakeManagerOwnership() public {
        vm.prank(deployer);
        assertTrue(
            AccessControlUpgradeable(address(stakeManager)).hasRole(stakeManager.DEFAULT_ADMIN_ROLE(), deployer),
            "Deployer should be default admin"
        );
    }

    function testVaultFactoryOwnership() public view {
        assertEq(vaultFactory.owner(), deployer, "Deployer should be owner of VaultFactory");
    }

    function testKarmaNFTConfiguration() public view {
        assertEq(
            address(karmaNFT.metadataGenerator()),
            address(metadataGenerator),
            "KarmaNFT should use correct metadata generator"
        );
    }

    function testStakeManagerConfiguration() public view {
        assertEq(
            address(stakeManager.STAKING_TOKEN()),
            address(stakingToken),
            "StakeManager should use correct staking token"
        );
        assertEq(
            address(stakeManager.rewardsSupplier()),
            address(karma),
            "StakeManager should use Karma as rewards supplier"
        );
    }

    function testVaultFactoryConfiguration() public view {
        assertEq(
            address(vaultFactory.stakeManager()),
            address(stakeManager),
            "VaultFactory should reference correct StakeManager"
        );
        assertEq(
            address(vaultFactory.vaultImplementation()),
            vaultImpl,
            "VaultFactory should reference correct Vault implementation"
        );
    }

    function testKarmaRewardDistributorSetup() public view {
        assertTrue(karma.allowedToTransfer(address(stakeManager)), "StakeManager should be allowed to transfer");
    }

    function testContractInitialization() public {
        // Test Karma initialization
        assertEq(karma.name(), "Karma", "Karma should have correct name");
        assertEq(karma.symbol(), "KARMA", "Karma should have correct symbol");

        // Test that contracts are properly initialized (not zero addresses for critical references)
        assertTrue(address(stakeManager.STAKING_TOKEN()) != address(0), "StakeManager staking token should be set");
        assertTrue(
            address(stakeManager.rewardsSupplier()) != address(0), "StakeManager rewards supplier should be set"
        );
        assertTrue(address(vaultFactory.stakeManager()) != address(0), "VaultFactory stake manager should be set");
    }
}
