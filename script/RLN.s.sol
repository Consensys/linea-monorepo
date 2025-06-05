// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { ERC1967Proxy } from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

import { Groth16Verifier } from "../src/rln/Verifier.sol";
import { RLN } from "../src/rln/RLN.sol";

contract DeployRLNScript is BaseScript {
    function run() public returns (RLN, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (address deployer,) = deploymentConfig.activeNetworkConfig();

        uint256 depth = vm.envUint("DEPTH");
        address karmaAddress = vm.envAddress("KARMA_ADDRESS");

        vm.startBroadcast(deployer);
        address verifier = (address)(new Groth16Verifier());
        // Deploy Karma logic contract
        bytes memory initializeData =
            abi.encodeCall(RLN.initialize, (deployer, deployer, deployer, depth, verifier, karmaAddress));
        address impl = address(new RLN());
        // Create upgradeable proxy
        address proxy = address(new ERC1967Proxy(impl, initializeData));

        vm.stopBroadcast();

        return (RLN(proxy), deploymentConfig);
    }
}
