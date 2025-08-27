// SPDX-License-Identifier: MIT

pragma solidity 0.8.26;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { IStakeManagerProxy } from "./interfaces/IStakeManagerProxy.sol";
import { IStakeVault } from "./interfaces/IStakeVault.sol";

/**
 * @title StakeVault
 * @author Ricardo Guilherme Schmidt <ricardo3@status.im>
 * @notice A contract to secure user stakes and manage staking with IStakeManager.
 * @dev This contract is owned by the user and allows staking, unstaking, and withdrawing tokens.
 * @dev The only reason this is `OwnableUpgradeable` is because we use proxy clones
 * to create stake vault instances. Hence, we need to use `Initializeable` to set the owner.
 */
contract StakeVault is IStakeVault, Initializable, OwnableUpgradeable {
    /// @notice Emitted when not enough balance to withdraw
    error StakeVault__NotEnoughAvailableBalance();
    /// @notice Emitted when destination address is invalid
    error StakeVault__InvalidDestinationAddress();
    /// @notice Emitted when staking was unsuccessful
    error StakeVault__StakingFailed();
    /// @notice Emitted when the funds are locked
    error StakeVault__FundsLocked();
    /// @notice Emitted when unstaking was unsuccessful
    error StakeVault__UnstakingFailed();
    /// @notice Emitted when not allowed to exit the system
    error StakeVault__NotAllowedToExit();
    /// @notice Emitted when not allowed to leave the system
    error StakeVault__NotAllowedToLeave();
    /// @notice Emitted when migration failed
    error StakeVault__MigrationFailed();
    /// @notice Emitted when the caller is not the owner of the vault
    error StakeVault__NotAuthorized();
    /// @notice Emitted when withdrawing funds from vault fails
    error StakeVault__WithdrawFromVaultFailed();

    /*//////////////////////////////////////////////////////////////////////////
                                  STATE VARIABLES
    //////////////////////////////////////////////////////////////////////////*/

    /// @notice Staking token - must be set immutable due to codehash check in StakeManager
    IERC20 public immutable STAKING_TOKEN;
    /// @notice Stake manager proxy contract
    IStakeManagerProxy public stakeManager;
    /// @notice Timestamp until the funds are locked
    uint256 public lockUntil;

    modifier validDestination(address _destination) {
        if (_destination == address(0)) {
            revert StakeVault__InvalidDestinationAddress();
        }
        _;
    }

    /*//////////////////////////////////////////////////////////////////////////
                                     CONSTRUCTOR
    //////////////////////////////////////////////////////////////////////////*/

    /**
     * @notice Initializes the contract with the staking token address.
     * @dev The staking token address is immutable and cannot be changed after deployment.
     * @dev Contract will be initialized via `initialize` function.
     * @param token The address of the staking token.
     */
    constructor(IERC20 token) {
        STAKING_TOKEN = token;
        _disableInitializers();
    }

    /*//////////////////////////////////////////////////////////////////////////
                           USER-FACING FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    /**
     * @notice Initializes the contract with the owner and the stake manager.
     * @dev Ensures that the stake manager implementation is trusted.
     * @dev Initializion is done on proxy clones.
     * @param _owner The address of the owner.
     * @param _stakeManager The address of the StakeManager contract.
     */
    function initialize(address _owner, address _stakeManager) public initializer {
        _transferOwnership(_owner);
        stakeManager = IStakeManagerProxy(_stakeManager);
    }

    /**
     * @notice Registers the vault with the stake manager.
     * @dev This is necessary to allow the stake manager to interact with the vault.
     */
    function register() public {
        stakeManager.registerVault();
    }

    /**
     * @notice Stake tokens for a specified time.
     * @dev This function is only callable by the owner.
     * @dev Can only be called if the stake manager is trusted.
     * @dev Reverts if the staking token transfer fails.
     * @param _amount The amount of tokens to stake.
     * @param _seconds The time period to stake for.
     */
    function stake(uint256 _amount, uint256 _seconds) external onlyOwner {
        _stake(_amount, _seconds, msg.sender);
    }

    /**
     * @notice Stake tokens from a specified address for a specified time.
     * @dev Overloads the `stake` function to allow staking from a specified address.
     * @dev This function is only callable by the owner.
     * @dev Can only be called if the stake manager is trusted.
     * @dev Reverts if the staking token transfer fails.
     * @param _amount The amount of tokens to stake.
     * @param _seconds The time period to stake for.
     * @param _from The address from which tokens will be transferred.
     */
    function stake(uint256 _amount, uint256 _seconds, address _from) external onlyOwner {
        _stake(_amount, _seconds, _from);
    }

    /**
     * @notice Lock the staked amount for a specified time.
     * @dev This function is only callable by the owner.
     * @dev Can only be called if the stake manager is trusted.
     * @param _seconds The time period to lock the staked amount for.
     */
    function lock(uint256 _seconds) external onlyOwner {
        lockUntil = stakeManager.lock(_seconds, lockUntil);
    }

    /**
     * @notice Unstake a specified amount of tokens and send to the owner.
     * @dev This function is only callable by the owner.
     * @dev Can only be called if the stake manager is trusted.
     * @dev Reverts if the staking token transfer fails.
     * @param _amount The amount of tokens to unstake.
     */
    function unstake(uint256 _amount) external onlyOwner {
        _unstake(_amount, msg.sender);
    }

    /**
     * @notice Unstake a specified amount of tokens and send to a destination address.
     * @dev Overloads the `unstake` function to allow unstaking to a specified address.
     * @dev This function is only callable by the owner.
     * @dev Can only be called if the stake manager is trusted.
     * @dev Reverts if the staking token transfer fails.
     * @param _amount The amount of tokens to unstake.
     * @param _destination The address to receive the unstaked tokens.
     */
    function unstake(uint256 _amount, address _destination) external onlyOwner validDestination(_destination) {
        _unstake(_amount, _destination);
    }

    /**
     * @notice Allows the vault to leave the system and withdraw all funds.
     * @dev This function is only callable by the owner.
     * @dev Vaults can only leave the system if the stake manager is not trusted.
     * @param _destination The address to receive the funds.
     */
    function leave(address _destination) external onlyOwner validDestination(_destination) {
        // We have to `try/catch` here in case the upgrade was bad and `leave()`
        // either doesn't exist on the new stake manager or reverts for some reason.
        // If it was a good upgrade, it will cause the stake manager to properly update
        // its internal accounting before we move the funds out.
        try stakeManager.leave() {
            if (lockUntil <= block.timestamp) {
                STAKING_TOKEN.transfer(_destination, STAKING_TOKEN.balanceOf(address(this)));
            }
        } catch {
            if (lockUntil <= block.timestamp) {
                STAKING_TOKEN.transfer(_destination, STAKING_TOKEN.balanceOf(address(this)));
            }
        }
    }

    /**
     * @notice Migrate all funds to a new vault.
     * @dev This function is only callable by the owner.
     * @dev This function is only callable if the current stake manager is trusted.
     * @dev Reverts when the stake manager reverts or the funds can't be transferred.
     * @param migrateTo The address of the new vault.
     */
    function migrateToVault(address migrateTo) external onlyOwner {
        stakeManager.migrateToVault(migrateTo);
        bool success = STAKING_TOKEN.transfer(migrateTo, STAKING_TOKEN.balanceOf(address(this)));
        if (!success) {
            revert StakeVault__MigrationFailed();
        }
    }

    /**
     * @notice Updates the lock until timestamp.
     * @dev This function is only callable by the stake manager.
     * @param _lockUntil The new lock until timestamp.
     */
    function migrateFromVault(uint256 _lockUntil) external {
        if (msg.sender != address(stakeManager)) {
            revert StakeVault__NotAuthorized();
        }
        lockUntil = _lockUntil;
    }

    /**
     * @notice Withdraw tokens from the contract.
     * @dev This function is only callable by the owner.
     * @dev Only withdraws excess staking token amounts.
     * @param _token The IERC20 token to withdraw.
     * @param _amount The amount of tokens to withdraw.
     */
    function withdraw(IERC20 _token, uint256 _amount) external onlyOwner {
        _withdraw(_token, _amount, msg.sender);
    }

    /**
     * @notice Withdraw tokens from the contract to a destination address.
     * @dev Overloads the `withdraw` function to allow withdrawing to a specified address.
     * @dev This function is only callable by the owner.
     * @dev Only withdraws excess staking token amounts.
     * @param _token The IERC20 token to withdraw.
     * @param _amount The amount of tokens to withdraw.
     * @param _destination The address to receive the tokens.
     */
    function withdraw(
        IERC20 _token,
        uint256 _amount,
        address _destination
    )
        external
        onlyOwner
        validDestination(_destination)
    {
        _withdraw(_token, _amount, _destination);
    }

    function withdrawFromVault(
        uint256 _amount,
        address _destination
    )
        external
        onlyOwner
        validDestination(_destination)
    {
        if (lockUntil > block.timestamp) {
            revert StakeVault__FundsLocked();
        }
        bool success = STAKING_TOKEN.transfer(_destination, _amount);
        if (!success) {
            revert StakeVault__WithdrawFromVaultFailed();
        }
    }

    /**
     * @notice Returns the available amount of a token that can be withdrawn.
     * @dev Returns only excess amount if token is staking token.
     * @param _token The IERC20 token to check.
     * @return The amount of token available for withdrawal.
     */
    function availableWithdraw(IERC20 _token) external view returns (uint256) {
        if (_token == STAKING_TOKEN) {
            return STAKING_TOKEN.balanceOf(address(this)) - amountStaked();
        }
        return _token.balanceOf(address(this));
    }

    /*//////////////////////////////////////////////////////////////////////////
                           INTERNAL FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    /**
     * @notice Stakes tokens for a specified time.
     * @dev Reverts if the staking token transfer fails.
     * @param _amount The amount of tokens to stake.
     * @param _seconds The time period to stake for.
     * @param _source The address from which tokens will be transferred.
     */
    function _stake(uint256 _amount, uint256 _seconds, address _source) internal {
        lockUntil = stakeManager.stake(_amount, _seconds, lockUntil);
        bool success = STAKING_TOKEN.transferFrom(_source, address(this), _amount);
        if (!success) {
            revert StakeVault__StakingFailed();
        }
    }

    /**
     * @notice Unstakes tokens to a specified address.
     * @dev Reverts if the staking token transfer fails.
     * @param _amount The amount of tokens to unstake.
     * @param _destination The address to receive the unstaked tokens.
     */
    function _unstake(uint256 _amount, address _destination) internal {
        if (lockUntil > block.timestamp) {
            revert StakeVault__FundsLocked();
        }
        stakeManager.unstake(_amount);
        bool success = STAKING_TOKEN.transfer(_destination, _amount);
        if (!success) {
            revert StakeVault__UnstakingFailed();
        }
    }

    /**
     * @notice Withdraws tokens to a specified address.
     * @dev Reverts if the staking token transfer fails.
     * @dev Only withdraws excess staking token amounts.
     * @param _token The IERC20 token to withdraw.
     * @param _amount The amount of tokens to withdraw.
     * @param _destination The address to receive the tokens.
     */
    function _withdraw(IERC20 _token, uint256 _amount, address _destination) internal {
        if (_token == STAKING_TOKEN && STAKING_TOKEN.balanceOf(address(this)) - amountStaked() < _amount) {
            revert StakeVault__NotEnoughAvailableBalance();
        }
        _token.transfer(_destination, _amount);
    }

    /**
     * @notice Allows vaults to exit the system in case of emergency or the system is rigged.
     * @param _destination The address to receive the funds.
     * @dev This function tries to read `IStakeManager.emergencyModeEnabeled()` to check if an
     *      emergency mode is enabled. If the call fails, it will still transfer the funds to the
     *      destination address.
     * @dev This function is only callable by the owner.
     * @dev Reverts when `emergencyModeEnabled()` returns false.
     */
    function emergencyExit(address _destination) external onlyOwner validDestination(_destination) {
        try stakeManager.emergencyModeEnabled() returns (bool enabled) {
            if (!enabled) {
                revert StakeVault__NotAllowedToExit();
            }
            STAKING_TOKEN.transfer(_destination, STAKING_TOKEN.balanceOf(address(this)));
        } catch {
            STAKING_TOKEN.transfer(_destination, STAKING_TOKEN.balanceOf(address(this)));
        }
    }

    /*//////////////////////////////////////////////////////////////////////////
                           VIEW FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    /**
     * @notice Returns the address of the current owner.
     * @return The address of the owner.
     */
    function owner() public view override(OwnableUpgradeable, IStakeVault) returns (address) {
        return super.owner();
    }

    /**
     * @notice Returns the amount of tokens staked by the vault.
     * @return The amount of tokens staked.
     */
    function amountStaked() public view returns (uint256) {
        return stakeManager.stakedBalanceOf(address(this));
    }
}
