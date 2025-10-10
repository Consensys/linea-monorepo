// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { ERC1967Proxy } from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { SimpleKarmaDistributor } from "../src/SimpleKarmaDistributor.sol";

/**
 * @dev Script for deploying the SimpleKarmaDistributor contract as an upgradeable proxy.
 */
contract DeploySimpleKarmaDistributorScript is BaseScript {
    /**
     * @dev Deploys SimpleKarmaDistributor for production use.
     * Reads karma token address from the `KARMA_ADDRESS` env variable.
     */
    function run() public returns (SimpleKarmaDistributor distributor, address impl) {
        address karmaAddress = vm.envAddress("KARMA_ADDRESS");
        require(karmaAddress != address(0), "KARMA_ADDRESS is not set");
        return _run(karmaAddress);
    }

    /**
     * @dev Deploys SimpleKarmaDistributor for test environments and returns deployment config.
     */
    function runForTest(address karmaAddress)
        public
        returns (SimpleKarmaDistributor distributor, DeploymentConfig deploymentConfig)
    {
        deploymentConfig = new DeploymentConfig(broadcaster);
        (distributor,) = _run(karmaAddress);
    }

    function _run(address karmaAddress) internal broadcast returns (SimpleKarmaDistributor distributor, address impl) {
        return deploy(broadcaster, karmaAddress);
    }

    function deploy(address deployer, address karmaAddress) public returns (SimpleKarmaDistributor, address) {
        vm.startBroadcast(deployer);
        bytes memory initializeData = abi.encodeCall(SimpleKarmaDistributor.initialize, (deployer, karmaAddress));
        address impl = address(new SimpleKarmaDistributor());
        address proxy = address(new ERC1967Proxy(impl, initializeData));
        vm.stopBroadcast();

        return (SimpleKarmaDistributor(proxy), impl);
    }
}
