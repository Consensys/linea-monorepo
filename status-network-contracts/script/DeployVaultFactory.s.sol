// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { Clones } from "@openzeppelin/contracts/proxy/Clones.sol";

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { VaultFactory } from "../src/VaultFactory.sol";
import { StakeVault } from "../src/StakeVault.sol";

/**
 * @title DeployVaultFactoryScript
 * @dev Script to deploy the VaultFactory contract.
 */
contract DeployVaultFactoryScript is BaseScript {
    /**
     * @dev Deploys VaultFactory contract for production use and returns the instance
     * @return vaultFactory The deployed VaultFactory contract instance.
     * @return vaultImplementation The address of the StakeVault logic contract.
     * @return vaultProxyClone The address of the StakeVault proxy clone used by the VaultFactory.
     */
    function run() public returns (VaultFactory vaultFactory, address vaultImplementation, address vaultProxyClone) {
        address stakeManager = vm.envAddress("STAKE_MANAGER_PROXY_ADDRESS");
        require(stakeManager != address(0), "STAKE_MANAGER_PROXY_ADDRESS is not set");

        address stakingToken = vm.envAddress("STAKING_TOKEN_ADDRESS");
        require(stakingToken != address(0), "STAKING_TOKEN_ADDRESS is not set");

        return _run(stakeManager, stakingToken);
    }

    /**
     * @dev Deploys VaultFactory contract for testing purposes and returns the instance along with deployment config.
     * @param stakeManager The address of the StakeManager contract.
     * @param stakingToken The address of the staking token.
     * @return vaultFactory The deployed VaultFactory contract instance.
     * @return vaultImplementation The address of the StakeVault logic contract.
     * @return vaultProxyClone The address of the StakeVault proxy clone used by the VaultFactory.
     * @return deploymentConfig The DeploymentConfig instance for the current network.
     */
    function runForTest(
        address stakeManager,
        address stakingToken
    )
        public
        returns (
            VaultFactory vaultFactory,
            address vaultImplementation,
            address vaultProxyClone,
            DeploymentConfig deploymentConfig
        )
    {
        deploymentConfig = new DeploymentConfig(broadcaster);
        (vaultFactory, vaultImplementation, vaultProxyClone) = _run(stakeManager, stakingToken);
        return (vaultFactory, vaultImplementation, vaultProxyClone, deploymentConfig);
    }

    /**
     * @dev Deploys VaultFactory contract within a broadcast context and returns the instance.
     * @param stakeManager The address of the StakeManager contract.
     * @param stakingToken The address of the staking token.
     * @return vaultFactory The deployed VaultFactory contract instance.
     * @return vaultImplementation The address of the StakeVault logic contract.
     * @return vaultProxyClone The address of the StakeVault proxy clone used by the VaultFactory.
     */
    function _run(
        address stakeManager,
        address stakingToken
    )
        internal
        broadcast
        returns (VaultFactory vaultFactory, address vaultImplementation, address vaultProxyClone)
    {
        return deploy(broadcaster, stakeManager, stakingToken);
    }

    /**
     * @dev Deploys VaultFactory contract and returns the instance.
     * Note: This function does not handle broadcasting; it should be called within a broadcast context.
     * @param deployer The address that will be set as the deployer/owner of the VaultFactory contract.
     * @param stakeManager The address of the StakeManager contract.
     * @param stakingToken The address of the staking token.
     * @return vaultFactory The deployed VaultFactory contract instance.
     * @return vaultImplementation The address of the StakeVault logic contract.
     * @return vaultProxyClone The address of the StakeVault proxy clone used by the VaultFactory.
     */
    function deploy(
        address deployer,
        address stakeManager,
        address stakingToken
    )
        public
        returns (VaultFactory vaultFactory, address vaultImplementation, address vaultProxyClone)
    {
        // Create vault implementation for proxy clones
        vaultImplementation = address(new StakeVault(IERC20(stakingToken)));
        vaultProxyClone = Clones.clone(vaultImplementation);
        // Create vault factory
        vaultFactory = new VaultFactory(deployer, stakeManager, vaultImplementation);
    }
}
