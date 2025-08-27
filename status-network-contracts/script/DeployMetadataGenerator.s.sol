// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { NFTMetadataGeneratorSVG } from "../src/nft-metadata-generators/NFTMetadataGeneratorSVG.sol";
import { INFTMetadataGenerator } from "../src/interfaces/INFTMetadataGenerator.sol";

contract DeployMetadataGenerator is BaseScript {
    function run() public returns (INFTMetadataGenerator, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (address deployer,) = deploymentConfig.activeNetworkConfig();

        vm.startBroadcast(deployer);
        string memory svgPrefix =
        // solhint-disable-next-line
            "<svg width=\"200\" height=\"200\" viewBox=\"0 0 200 200\"><rect x=\"0\" y=\"0\" width=\"100%\" height=\"100%\" stroke=\"black\" stroke-width=\"3px\" fill=\"white\"/><text x=\"50%\" y=\"50%\" dominant-baseline=\"middle\" text-anchor=\"middle\">";
        string memory svgSuffix = "</text></svg>";
        NFTMetadataGeneratorSVG metadataGenerator = new NFTMetadataGeneratorSVG(svgPrefix, svgSuffix);
        vm.stopBroadcast();

        return (INFTMetadataGenerator(metadataGenerator), deploymentConfig);
    }
}
