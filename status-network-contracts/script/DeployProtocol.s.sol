// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { console } from "forge-std/Test.sol";

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { DeployKarmaScript } from "./DeployKarma.s.sol";
import { DeployMetadataGeneratorScript } from "./DeployMetadataGenerator.s.sol";
import { DeployKarmaNFTScript } from "./DeployKarmaNFT.s.sol";
import { DeployStakeManagerScript } from "./DeployStakeManager.s.sol";
import { DeployVaultFactoryScript } from "./DeployVaultFactory.s.sol";

import { INFTMetadataGenerator } from "../src/interfaces/INFTMetadataGenerator.sol";
import { Karma } from "../src/Karma.sol";
import { KarmaNFT } from "../src/KarmaNFT.sol";
import { StakeManager } from "../src/StakeManager.sol";
import { VaultFactory } from "../src/VaultFactory.sol";

/**
 * @dev This script deploys the entire protocol including Karma, KarmaNFT, StakeManager, and VaultFactory.
 * It uses the DeploymentConfig to get network-specific parameters.
 * The script assumes that the staking token address is provided in the active network configuration.
 */
contract DeployProtocolScript is BaseScript {
    DeployKarmaScript deployKarma;

    DeployMetadataGeneratorScript deployMetadataGenerator;

    DeployKarmaNFTScript deployKarmaNFT;

    DeployStakeManagerScript deployStakeManager;

    DeployVaultFactoryScript deployVaultFactory;

    constructor() BaseScript() {
        deployKarma = new DeployKarmaScript();
        deployMetadataGenerator = new DeployMetadataGeneratorScript();
        deployKarmaNFT = new DeployKarmaNFTScript();
        deployStakeManager = new DeployStakeManagerScript();
        deployVaultFactory = new DeployVaultFactoryScript();
    }

    /**
     * @dev Deploys protocol for production use and returns the instances.
     * The address of the staking token must be provided via the active network configuration.
     * @return karma The deployed Karma contract instance.
     * @return karmaImpl The address of the Karma logic contract.
     * @return metadataGenerator The deployed NFT metadata generator contract instance.
     * @return karmaNFT The deployed KarmaNFT contract instance.
     * @return stakeManager The deployed StakeManager contract instance.
     * @return stakeManagerImpl The address of the StakeManager logic contract.
     * @return vaultFactory The deployed VaultFactory contract instance.
     * @return vaultImpl The address of the StakeVault logic contract.
     * @return vaultProxyClone The address of the StakeVault proxy clone used by the VaultFactory.
     */
    function run()
        public
        returns (
            Karma,
            address,
            INFTMetadataGenerator,
            KarmaNFT,
            StakeManager,
            address,
            VaultFactory,
            address,
            address
        )
    {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (, address stakingToken) = deploymentConfig.activeNetworkConfig();
        return _run(stakingToken);
    }

    /**
     * @dev Deploys protocol within a broadcast context and returns the instances.
     * @param stakingToken The address of the staking token to be used in the StakeManager and VaultFactory.
     * @return karma The deployed Karma contract instance.
     * @return karmaImpl The address of the Karma logic contract.
     * @return metadataGenerator The deployed NFT metadata generator contract instance.
     * @return karmaNFT The deployed KarmaNFT contract instance.
     * @return stakeManager The deployed StakeManager contract instance.
     * @return stakeManagerImpl The address of the StakeManager logic contract.
     * @return vaultFactory The deployed VaultFactory contract instance.
     * @return vaultImpl The address of the StakeVault logic contract.
     * @return vaultProxyClone The address of the StakeVault proxy clone used by the VaultFactory.
     */
    function _run(address stakingToken)
        internal
        broadcast
        returns (
            Karma karma,
            address karmaImpl,
            INFTMetadataGenerator metadataGenerator,
            KarmaNFT karmaNFT,
            StakeManager stakeManager,
            address stakeManagerImpl,
            VaultFactory vaultFactory,
            address vaultImpl,
            address vaultProxyClone
        )
    {
        console.log("Deploying Karma...");
        (karma, karmaImpl) = deployKarma.deploy(broadcaster);

        console.log("Deploying NFTMetadataGeneratorSVG...");
        metadataGenerator = deployMetadataGenerator.deploy();

        console.log("Deploying KarmaNFT...");
        karmaNFT = deployKarmaNFT.deploy(address(metadataGenerator), address(karma));

        console.log("Deploying StakeManager...");
        (stakeManager, stakeManagerImpl) = deployStakeManager.deploy(broadcaster, stakingToken, address(karma));

        console.log("Deploying VaultFactory...");
        (vaultFactory, vaultImpl, vaultProxyClone) =
            deployVaultFactory.deploy(broadcaster, address(stakeManager), stakingToken);

        console.log("\nContract addresses:");
        console.log(address(karma), ": Karma (proxy)");
        console.log(karmaImpl, ": Karma (implementation)");
        console.log(address(metadataGenerator), ": NFTMetadataGeneratorSVG");
        console.log(address(karmaNFT), ": KarmaNFT");
        console.log(address(stakeManager), ": StakeManager (proxy)");
        console.log(stakeManagerImpl, ": StakeManager (implementation)");
        console.log(address(vaultFactory), ": VaultFactory");
        console.log(vaultImpl, ": StakeVault (implementation)");
        console.log(vaultProxyClone, ": StakeVault (proxy clone)");

        /// INITIALIZATION
        console.log("\nInitializing contracts...");

        karma.addRewardDistributor(address(stakeManager));
        console.log("Added reward distributor (StakeManager)", address(stakeManager));

        karma.setAllowedToTransfer(address(stakeManager), true);
        console.log("Whitelisted reward distributor", address(stakeManager), "for transfer");

        stakeManager.setRewardsSupplier(address(karma));
        console.log("Set rewards supplier (Karma) for StakeManager:", address(karma));

        stakeManager.setTrustedCodehash(vaultProxyClone.codehash, true);
        console.log("Set trusted codehash for StakeVault proxy clone:", vaultProxyClone);
    }
}
