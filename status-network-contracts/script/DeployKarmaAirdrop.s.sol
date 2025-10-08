// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { KarmaAirdrop } from "../src/KarmaAirdrop.sol";

/**
 * @dev This script deploys the KarmaAirdrop contract.
 * It requires the address of the Karma token contract and the owner of the airdrop to be provided via environment
 * variables.
 * The deploy function handles the deployment of the KarmaAirdrop contract.
 * The address of the Karma contract must be provided via the "KARMA_ADDRESS" environment variable.
 * The owner of the KarmaAirdrop contract must be provided via the "KARMA_AIRDROP_OWNER" environment variable.
 * The default delegatee address must be provided via the "DEFAULT_DELEGATEE" environment variable.
 * The allowMerkleRootUpdate flag can be optionally provided via the "ALLOW_MERKLE_ROOT_UPDATE" environment variable
 * (defaults to false).
 */
contract DeployKarmaAirdropScript is BaseScript {
    /**
     * @dev Deploys KarmaAirdrop contract and returns the instance.
     * The address of the Karma contract must be provided via the "KARMA_ADDRESS" environment variable.
     * The owner of the KarmaAirdrop contract must be provided via the "KARMA_AIRDROP_OWNER" environment variable.
     * The default delegatee address must be provided via the "DEFAULT_DELEGATEE" environment variable.
     * The allowMerkleRootUpdate flag can be optionally provided via the "ALLOW_MERKLE_ROOT_UPDATE" environment
     * variable
     * (defaults to false).
     * The deployer/owner of the KarmaAirdrop contract will be set to the address specified in "KARMA_AIRDROP_OWNER".
     * @return karmaAirdrop The deployed KarmaAirdrop contract instance.
     */
    function run() public returns (KarmaAirdrop karmaAirdrop) {
        address karmaAddress = vm.envAddress("KARMA_ADDRESS");
        require(karmaAddress != address(0), "KARMA_ADDRESS is not set");

        address ownerAddress = vm.envAddress("KARMA_AIRDROP_OWNER");
        require(ownerAddress != address(0), "KARMA_AIRDROP_OWNER is not set");

        address defaultDelegatee = vm.envAddress("DEFAULT_DELEGATEE");
        require(defaultDelegatee != address(0), "DEFAULT_DELEGATEE is not set");

        bool allowMerkleRootUpdate = vm.envOr("ALLOW_MERKLE_ROOT_UPDATE", false);
        karmaAirdrop = _run(karmaAddress, ownerAddress, allowMerkleRootUpdate, defaultDelegatee);
    }

    /**
     * @dev Deploys KarmaAirdrop contract for testing purposes with custom defaultDelegatee.
     * @param karmaAddress The address of the Karma token contract.
     * @param owner The address that will be set as the owner of the KarmaAirdrop contract.
     * @param defaultDelegatee The default delegatee address for new claimers.
     * @return karmaAirdrop The deployed KarmaAirdrop contract instance.
     * @return deploymentConfig The DeploymentConfig instance for the current network.
     */
    function runForTest(
        address karmaAddress,
        address owner,
        address defaultDelegatee
    )
        public
        returns (KarmaAirdrop karmaAirdrop, DeploymentConfig deploymentConfig)
    {
        deploymentConfig = new DeploymentConfig(broadcaster);
        karmaAirdrop = _run(karmaAddress, owner, false, defaultDelegatee);
    }

    /**
     * @dev Deploys KarmaAirdrop contract for testing purposes with custom allowMerkleRootUpdate flag.
     * @param karmaAddress The address of the Karma token contract.
     * @param owner The address that will be set as the owner of the KarmaAirdrop contract.
     * @param allowMerkleRootUpdate Whether to allow merkle root updates.
     * @return karmaAirdrop The deployed KarmaAirdrop contract instance.
     * @return deploymentConfig The DeploymentConfig instance for the current network.
     */
    function runForTest(
        address karmaAddress,
        address owner,
        bool allowMerkleRootUpdate
    )
        public
        returns (KarmaAirdrop karmaAirdrop, DeploymentConfig deploymentConfig)
    {
        deploymentConfig = new DeploymentConfig(broadcaster);
        karmaAirdrop = _run(karmaAddress, owner, allowMerkleRootUpdate, address(0));
    }

    /**
     * @dev Deploys KarmaAirdrop contract within a broadcast context and returns the instance.
     * @param karmaAddress The address of the Karma token contract.
     * @param owner The address that will be set as the owner of the KarmaAirdrop contract.
     * @param allowMerkleRootUpdate Whether to allow merkle root updates.
     * @param defaultDelegatee The default delegatee address for new claimers.
     * @return karmaAirdrop The deployed KarmaAirdrop contract instance.
     */
    function _run(
        address karmaAddress,
        address owner,
        bool allowMerkleRootUpdate,
        address defaultDelegatee
    )
        internal
        broadcast
        returns (KarmaAirdrop karmaAirdrop)
    {
        karmaAirdrop = new KarmaAirdrop(karmaAddress, owner, allowMerkleRootUpdate, defaultDelegatee);
    }
}
