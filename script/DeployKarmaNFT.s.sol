// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { KarmaNFT } from "../src/KarmaNFT.sol";
import { NFTMetadataGeneratorSVG } from "../src/nft-metadata-generators/NFTMetadataGeneratorSVG.sol";
import { INFTMetadataGenerator } from "../src/interfaces/INFTMetadataGenerator.sol";

contract DeployKarmaNFTScript is BaseScript {
    function run() public returns (KarmaNFT, INFTMetadataGenerator, DeploymentConfig) {
        address karmaAddress = vm.envAddress("KARMA_ADDRESS");
        require(karmaAddress != address(0), "KARMA_ADDRESS is not set");

        return _run(karmaAddress);
    }

    function runForTest(address karmaAddress) public returns (KarmaNFT, INFTMetadataGenerator, DeploymentConfig) {
        return _run(karmaAddress);
    }

    function _run(address karmaAddress) public returns (KarmaNFT, INFTMetadataGenerator, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (address deployer,,) = deploymentConfig.activeNetworkConfig();

        vm.startBroadcast(deployer);

        // Deploy NFT metadata generator
        string memory svgPrefix =
        // solhint-disable-next-line
            "<svg width=\"200\" height=\"200\" viewBox=\"0 0 200 200\"><rect x=\"0\" y=\"0\" width=\"100%\" height=\"100%\" stroke=\"black\" stroke-width=\"3px\" fill=\"white\"/><text x=\"50%\" y=\"50%\" dominant-baseline=\"middle\" text-anchor=\"middle\">";
        string memory svgSuffix = "</text></svg>";
        NFTMetadataGeneratorSVG metadataGenerator = new NFTMetadataGeneratorSVG(svgPrefix, svgSuffix);

        // Deploy KarmaNFT
        KarmaNFT karmaNFT = new KarmaNFT(karmaAddress, address(metadataGenerator));

        vm.stopBroadcast();

        return (karmaNFT, INFTMetadataGenerator(metadataGenerator), deploymentConfig);
    }
}
