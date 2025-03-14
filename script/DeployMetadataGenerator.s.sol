// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { NFTMetadataGeneratorSVG } from "../src/nft-metadata-generators/NFTMetadataGeneratorSVG.sol";
import { INFTMetadataGenerator } from "../src/interfaces/INFTMetadataGenerator.sol";

contract DeployMetadataGenerator is BaseScript {
    function run() public returns (INFTMetadataGenerator, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (address deployer,,) = deploymentConfig.activeNetworkConfig();

        vm.startBroadcast(deployer);
        NFTMetadataGeneratorSVG metadataGenerator = new NFTMetadataGeneratorSVG("<svg>", "</svg>");
        vm.stopBroadcast();

        return (INFTMetadataGenerator(metadataGenerator), deploymentConfig);
    }
}
