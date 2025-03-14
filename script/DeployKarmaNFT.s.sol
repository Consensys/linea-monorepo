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
        NFTMetadataGeneratorSVG metadataGenerator = new NFTMetadataGeneratorSVG("<svg>", "</svg>");

        // Deploy KarmaNFT
        KarmaNFT karmaNFT = new KarmaNFT(karmaAddress, address(metadataGenerator));

        vm.stopBroadcast();

        return (karmaNFT, INFTMetadataGenerator(metadataGenerator), deploymentConfig);
    }
}
