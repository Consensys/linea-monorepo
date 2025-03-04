// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { ERC1967Proxy } from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

import { KarmaNFT } from "../src/KarmaNFT.sol";
import { NFTMetadataGeneratorSVG } from "../src/nft-metadata-generators/NFTMetadataGeneratorSVG.sol";

contract DeployKarmaNFTScript is BaseScript {
    function run() public returns (KarmaNFT, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (address deployer,,) = deploymentConfig.activeNetworkConfig();

        address karmaAddress = vm.envAddress("KARMA_ADDRESS");
        require(karmaAddress != address(0), "KARMA_ADDRESS is not set");

        vm.startBroadcast(deployer);

        // Deploy NFT metadata generator
        NFTMetadataGeneratorSVG metadataGenerator = new NFTMetadataGeneratorSVG("<svg>", "</svg>");

        // Deploy KarmaNFT
        KarmaNFT karmaNFT = new KarmaNFT(karmaAddress, address(metadataGenerator));

        vm.stopBroadcast();

        return (karmaNFT, deploymentConfig);
    }
}
