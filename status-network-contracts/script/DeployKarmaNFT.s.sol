// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { KarmaNFT } from "../src/KarmaNFT.sol";
import { NFTMetadataGeneratorSVG } from "../src/nft-metadata-generators/NFTMetadataGeneratorSVG.sol";
import { INFTMetadataGenerator } from "../src/interfaces/INFTMetadataGenerator.sol";

/**
 * @dev Script for deploying KarmaNFT contract and its dependencies.
 */
contract DeployKarmaNFTScript is BaseScript {
    /**
     * @dev Deploys NFT metadata generator and KarmaNFT contract for production use and returns the instances.
     * The address of the Karma contract must be provided via the "KARMA_ADDRESS" environment variable.
     * @return karmaNFT The deployed KarmaNFT contract instance.
     */
    function run() public returns (KarmaNFT) {
        address metadataGenerator = vm.envAddress("NFT_METADATA_GENERATOR_ADDRESS");
        require(metadataGenerator != address(0), "NFT_METADATA_GENERATOR_ADDRESS is not set");

        address karmaAddress = vm.envAddress("KARMA_ADDRESS");
        require(karmaAddress != address(0), "KARMA_ADDRESS is not set");

        return _run(metadataGenerator, karmaAddress);
    }

    /**
     * @dev Deploys KarmaNFT contract for testing purposes and returns the instances along with deployment config.
     * @param metadataGenerator The address of the NFT metadata generator contract.
     * @param karmaAddress The address of the Karma contract.
     * @return karmaNFT The deployed KarmaNFT contract instance.
     * @return deploymentConfig The DeploymentConfig instance for the current network.
     */
    function runForTest(address metadataGenerator, address karmaAddress) public returns (KarmaNFT, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        KarmaNFT karmaNFT = _run(metadataGenerator, karmaAddress);
        return (karmaNFT, deploymentConfig);
    }

    /**
     * @dev Deploys NFT metadata generator and KarmaNFT contract within a broadcast context and returns the instances.
     * @param metadataGenerator The address of the NFT metadata generator contract.
     * @param karmaAddress The address of the Karma contract.
     * @return karmaNFT The deployed KarmaNFT contract instance.
     */
    function _run(address metadataGenerator, address karmaAddress) internal broadcast returns (KarmaNFT) {
        return deploy(metadataGenerator, karmaAddress);
    }

    /**
     * @dev Deploys NFT metadata generator and KarmaNFT contract and returns the instances.
     * Note: This function does not handle broadcasting; it should be called within a broadcast context.
     * @param metadataGenerator The address of the NFT metadata generator contract.
     * @param karmaAddress The address of the Karma contract.
     * @return karmaNFT The deployed KarmaNFT contract instance.
     */
    function deploy(address metadataGenerator, address karmaAddress) public returns (KarmaNFT karmaNFT) {
        karmaNFT = new KarmaNFT(karmaAddress, metadataGenerator);
    }
}
