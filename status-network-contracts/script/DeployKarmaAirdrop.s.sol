// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { Karma } from "../src/Karma.sol";
import { KarmaAirdrop } from "../src/KarmaAirdrop.sol";

/**
 * @dev This script deploys the KarmaAirdrop contract.
 * It requires the address of the Karma token contract and the owner of the airdrop to be provided via environment
 * variables.
 * The deploy function handles the deployment of the KarmaAirdrop contract.
 * The address of the Karma contract must be provided via the "KARMA_ADDRESS" environment variable.
 * The owner of the KarmaAirdrop contract must be provided via the "KARMA_AIRDROP_OWNER" environment variable.
 */
contract DeployKarmaAirdropScript is BaseScript {
    /**
     * @dev Deploys KarmaAirdrop contract and returns the instance.
     * The address of the Karma contract must be provided via the "KARMA_ADDRESS" environment variable.
     * The owner of the KarmaAirdrop contract must be provided via the "KARMA_AIRDROP_OWNER" environment variable.
     * The deployer/owner of the KarmaAirdrop contract will be set to the address specified in "KARMA_AIRDROP_OWNER".
     * @return karmaAirdrop The deployed KarmaAirdrop contract instance.
     */
    function run() public returns (KarmaAirdrop karmaAirdrop) {
        address karmaAddress = vm.envAddress("KARMA_ADDRESS");
        require(karmaAddress != address(0), "KARMA_ADDRESS is not set");

        address ownerAddress = vm.envAddress("KARMA_AIRDROP_OWNER");
        require(ownerAddress != address(0), "KARMA_AIRDROP_OWNER is not set");
        karmaAirdrop = _run(karmaAddress, ownerAddress);
    }

    /**
     * @dev Deploys KarmaAirdrop contract for testing purposes and returns the instance along with deployment config.
     * @param karmaAddress The address of the Karma token contract.
     * @param owner The address that will be set as the owner of the KarmaAirdrop contract.
     * @return karmaAirdrop The deployed KarmaAirdrop contract instance.
     * @return deploymentConfig The DeploymentConfig instance for the current network.
     */
    function runForTest(
        address karmaAddress,
        address owner
    )
        public
        returns (KarmaAirdrop karmaAirdrop, DeploymentConfig deploymentConfig)
    {
        deploymentConfig = new DeploymentConfig(broadcaster);
        karmaAirdrop = _run(karmaAddress, owner);
    }

    /**
     * @dev Deploys KarmaAirdrop contract within a broadcast context and returns the instance.
     * @param karmaAddress The address of the Karma token contract.
     * @param owner The address that will be set as the owner of the KarmaAirdrop contract.
     * @return karmaAirdrop The deployed KarmaAirdrop contract instance.
     */
    function _run(address karmaAddress, address owner) internal broadcast returns (KarmaAirdrop karmaAirdrop) {
        karmaAirdrop = new KarmaAirdrop(karmaAddress, owner);
    }
}
