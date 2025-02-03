// SPDX-License-Identifier: MIT

pragma solidity ^0.8.26;

import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
// import { TransparentClones } from "./TransparentClones.sol";
import { Clones } from "@openzeppelin/contracts/proxy/Clones.sol";
import { StakeVault } from "./StakeVault.sol";

/**
 * @title VaultFactory
 * @author 0x-r4bbit
 *
 * This contract is reponsible for creating staking vaults for users.
 * A user of the staking protocol is able to create multiple vaults to facilitate
 * different strategies. For example, a user may want to create a vault for
 * a long-term lock period, while also creating a vault that has no lock period
 * at all.
 *
 * @notice This contract is used by users to create staking vaults.
 * @dev This contract will be deployed by Status, making Status the owner of the contract.
 * @dev A contract address for a `StakeManager` has to be provided to create this contract.
 * @dev Reverts with {VaultFactory__InvalidStakeManagerAddress} if the provided
 * `StakeManager` address is zero.
 * @dev The `StakeManager` contract address can be changed by the owner.
 */
contract VaultFactory is Ownable {
    error VaultFactory__InvalidStakeManagerAddress();

    event VaultCreated(address indexed vault, address indexed owner);
    event StakeManagerAddressChanged(address indexed newStakeManagerAddress);
    event VaultImplementationChanged(address indexed newVaultImplementation);

    /// @dev Address of the `StakeManager` contract instance.
    address public stakeManager;
    /// @dev Address of the `StakeVault` implementation contract.
    address public vaultImplementation;

    /// @param _stakeManager Address of the `StakeManager` contract instance.
    constructor(address _owner, address _stakeManager, address _vaultImplementation) Ownable(_owner) {
        if (_stakeManager == address(0)) {
            revert VaultFactory__InvalidStakeManagerAddress();
        }
        stakeManager = _stakeManager;
        vaultImplementation = _vaultImplementation;
    }

    /// @notice Sets the `StakeManager` contract address.
    /// @dev Only the owner can call this function.
    /// @dev Emits a {StakeManagerAddressChanged} event.
    /// @param _stakeManager Address of the `StakeManager` contract instance.
    function setStakeManager(address _stakeManager) external onlyOwner {
        stakeManager = _stakeManager;
        emit StakeManagerAddressChanged(_stakeManager);
    }

    /// @notice Sets the `StakeVault` implementation contract address.
    /// @dev Only the owner can call this function.
    /// @dev Emits a {VaultImplementationChanged} event.
    /// @param _vaultImplementation Address of the `StakeVault` implementation contract.
    /// @dev This function is used to change the implementation of the `StakeVault` contract.
    function setVaultImplementation(address _vaultImplementation) external onlyOwner {
        vaultImplementation = _vaultImplementation;
        emit VaultImplementationChanged(_vaultImplementation);
    }

    /// @notice Creates an instance of a `StakeVault` contract.
    /// @dev Anyone can call this function.
    /// @dev Emits a {VaultCreated} event.
    function createVault() external returns (StakeVault clone) {
        clone = StakeVault(Clones.clone(vaultImplementation));
        clone.initialize(msg.sender, stakeManager);
        clone.register();
        emit VaultCreated(address(clone), msg.sender);
    }
}
