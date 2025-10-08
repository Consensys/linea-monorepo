// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { NFTMetadataGeneratorSVG } from "../src/nft-metadata-generators/NFTMetadataGeneratorSVG.sol";
import { INFTMetadataGenerator } from "../src/interfaces/INFTMetadataGenerator.sol";

/**
 * @dev Script for deploying NFT metadata generator contract.
 */
contract DeployMetadataGeneratorScript is BaseScript {
    string public svgPrefix =
    // solhint-disable-next-line
        "<svg width=\"200\" height=\"200\" viewBox=\"0 0 200 200\"><rect x=\"0\" y=\"0\" width=\"100%\" height=\"100%\" stroke=\"black\" stroke-width=\"3px\" fill=\"white\"/><text x=\"50%\" y=\"50%\" dominant-baseline=\"middle\" text-anchor=\"middle\">";
    string public svgSuffix = "</text></svg>";

    /**
     * @dev Deploys NFT metadata generator contract for production use and returns the instances.
     * The address of the Karma contract must be provided via the "KARMA_ADDRESS" environment variable.
     * @return metadataGenerator The deployed NFT metadata generator contract instance.
     */
    function run() public returns (INFTMetadataGenerator metadataGenerator) {
        metadataGenerator = deploy(broadcaster);
    }

    /**
     * @dev Deploys NFT metadata generator contract for testing purposes and returns the instances along with
     * deployment config.
     * @param _svgPrefix The SVG prefix string to be used in the metadata generator.
     * @param _svgSuffix The SVG suffix string to be used in the metadata generator.
     * @return metadataGenerator The deployed NFT metadata generator contract instance.
     * @return deploymentConfig The DeploymentConfig instance for the current network.
     */
    function runForTest(
        string memory _svgPrefix,
        string memory _svgSuffix
    )
        public
        returns (INFTMetadataGenerator metadataGenerator, DeploymentConfig deploymentConfig)
    {
        deploymentConfig = new DeploymentConfig(broadcaster);
        svgPrefix = _svgPrefix;
        svgSuffix = _svgSuffix;
        metadataGenerator = deploy(broadcaster);
    }

    /**
     * @dev Deploys NFT metadata generator contract and returns the instances.
     * @param deployer The address that will be set as the deployer/owner of the NFT metadata generator contract.
     * @return metadataGenerator The deployed NFT metadata generator contract instance.
     */
    function deploy(address deployer) public returns (INFTMetadataGenerator) {
        vm.startBroadcast(deployer);
        NFTMetadataGeneratorSVG metadataGenerator = new NFTMetadataGeneratorSVG(svgPrefix, svgSuffix);
        vm.stopBroadcast();
        return INFTMetadataGenerator(metadataGenerator);
    }
}
