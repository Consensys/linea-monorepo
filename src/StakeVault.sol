// SPDX-License-Identifier: MIT

pragma solidity ^0.8.26;

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
    error StakeVault__NotEnoughAvailableBalance();
    error StakeVault__InvalidDestinationAddress();
    error StakeVault__UpdateNotAvailable();
    error StakeVault__StakingFailed();
    error StakeVault__UnstakingFailed();
    error StakeVault__NotAllowedToExit();
    error StakeVault__NotAllowedToLeave();
    error StakeVault__StakeManagerImplementationNotTrusted();
    error StakeVault__MigrationFailed();

    //STAKING_TOKEN must be kept as an immutable, otherwise, StakeManager would accept StakeVaults with any token
    //if is needed that STAKING_TOKEN to be a variable, StakeManager should be changed to check codehash and
    //StakeVault(msg.sender).STAKING_TOKEN()
    IERC20 public immutable STAKING_TOKEN;
    IStakeManagerProxy public stakeManager;
    address public stakeManagerImplementationAddress;

    /**
     * @dev Emitted when tokens are staked.
     * @param from The address from which tokens are transferred.
     * @param to The address receiving the staked tokens (this contract).
     * @param amount The amount of tokens staked.
     * @param time The time period for which tokens are staked.
     */
    event Staked(address indexed from, address indexed to, uint256 amount, uint256 time);

    modifier validDestination(address _destination) {
        if (_destination == address(0)) {
            revert StakeVault__InvalidDestinationAddress();
        }
        _;
    }

    modifier onlyTrustedStakeManager() {
        if (!_stakeManagerImplementationTrusted()) {
            revert StakeVault__StakeManagerImplementationNotTrusted();
        }
        _;
    }

    /**
     * @notice Initializes the contract with the owner, staked token, and stake manager.
     */
    constructor(IERC20 token) {
        STAKING_TOKEN = token;
        _disableInitializers();
    }

    /**
     * @param _owner The address of the owner.
     * @param _stakeManager The address of the StakeManager contract.
     */
    function initialize(address _owner, address _stakeManager) public initializer {
        __Ownable_init(_owner);
        stakeManager = IStakeManagerProxy(_stakeManager);
        stakeManagerImplementationAddress = stakeManager.implementation();
    }

    /**
     * @notice Allows the owner to trust a new stake manager implementation.
     * @param stakeManagerAddress The address of the new stake manager implementation.
     */
    function trustStakeManager(address stakeManagerAddress) external onlyOwner {
        stakeManagerImplementationAddress = stakeManagerAddress;
    }

    /**
     * @notice Registers the vault with the stake manager.
     */
    function register() public {
        stakeManager.registerVault();
    }

    /**
     * @notice Returns the address of the current owner.
     */
    function owner() public view override(OwnableUpgradeable, IStakeVault) returns (address) {
        return super.owner();
    }

    /**
     * @notice Stake tokens for a specified time.
     * @param _amount The amount of tokens to stake.
     * @param _seconds The time period to stake for.
     */
    function stake(uint256 _amount, uint256 _seconds) external onlyOwner onlyTrustedStakeManager {
        _stake(_amount, _seconds, msg.sender);
    }

    /**
     * @notice Stake tokens from a specified address for a specified time.
     * @param _amount The amount of tokens to stake.
     * @param _seconds The time period to stake for.
     * @param _from The address from which tokens will be transferred.
     */
    function stake(uint256 _amount, uint256 _seconds, address _from) external onlyOwner onlyTrustedStakeManager {
        _stake(_amount, _seconds, _from);
    }

    /**
     * @notice Lock the staked amount for a specified time.
     * @param _seconds The time period to lock the staked amount for.
     */
    function lock(uint256 _seconds) external onlyOwner onlyTrustedStakeManager {
        stakeManager.lock(_seconds);
    }

    /**
     * @notice Unstake a specified amount of tokens and send to the owner.
     * @param _amount The amount of tokens to unstake.
     */
    function unstake(uint256 _amount) external onlyOwner onlyTrustedStakeManager {
        _unstake(_amount, msg.sender);
    }

    /**
     * @notice Unstake a specified amount of tokens and send to a destination address.
     * @param _amount The amount of tokens to unstake.
     * @param _destination The address to receive the unstaked tokens.
     */
    function unstake(
        uint256 _amount,
        address _destination
    )
        external
        onlyOwner
        validDestination(_destination)
        onlyTrustedStakeManager
    {
        _unstake(_amount, _destination);
    }

    /**
     * @notice Withdraw all tokens from the contract to the owner.
     */
    function leave(address _destination) external onlyOwner validDestination(_destination) {
        if (_stakeManagerImplementationTrusted()) {
            // If the stakeManager is trusted, the vault cannot leave the system
            // and has to properly unstake instead (which might not be possible if
            // funds are locked).
            revert StakeVault__NotAllowedToLeave();
        }

        // If the stakeManager is not trusted, we know there was an upgrade.
        // In this case, vaults are free to leave the system and move their funds back
        // to the owner.
        //
        // We have to `try/catch` here in case the upgrade was malicious and `leave()`
        // either doesn't exist on the new stake manager or reverts for some reason.
        // If it was a benign upgrade, it will cause the stake manager to properly update
        // its internal accounting before we move the funds out.
        try stakeManager.leave() {
            STAKING_TOKEN.transfer(_destination, STAKING_TOKEN.balanceOf(address(this)));
        } catch {
            STAKING_TOKEN.transfer(_destination, STAKING_TOKEN.balanceOf(address(this)));
        }
    }

    /**
     * @notice Withdraw tokens from the contract.
     * @param _token The IERC20 token to withdraw.
     * @param _amount The amount of tokens to withdraw.
     */
    function withdraw(IERC20 _token, uint256 _amount) external onlyOwner {
        _withdraw(_token, _amount, msg.sender);
    }

    /**
     * @notice Withdraw tokens from the contract to a destination address.
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

    /**
     * @notice Returns the available amount of a token that can be withdrawn.
     * @param _token The IERC20 token to check.
     * @return The amount of token available for withdrawal.
     */
    function availableWithdraw(IERC20 _token) external view returns (uint256) {
        if (_token == STAKING_TOKEN) {
            return STAKING_TOKEN.balanceOf(address(this)) - amountStaked();
        }
        return _token.balanceOf(address(this));
    }

    function _stake(uint256 _amount, uint256 _seconds, address _source) internal {
        stakeManager.stake(_amount, _seconds);
        bool success = STAKING_TOKEN.transferFrom(_source, address(this), _amount);
        if (!success) {
            revert StakeVault__StakingFailed();
        }
        emit Staked(_source, address(this), _amount, _seconds);
    }

    function _unstake(uint256 _amount, address _destination) internal {
        stakeManager.unstake(_amount);
        bool success = STAKING_TOKEN.transfer(_destination, _amount);
        if (!success) {
            revert StakeVault__UnstakingFailed();
        }
    }

    function _withdraw(IERC20 _token, uint256 _amount, address _destination) internal {
        if (_token == STAKING_TOKEN && STAKING_TOKEN.balanceOf(address(this)) - amountStaked() < _amount) {
            revert StakeVault__NotEnoughAvailableBalance();
        }
        _token.transfer(_destination, _amount);
    }

    function amountStaked() public view returns (uint256) {
        return stakeManager.getStakedBalance(address(this));
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

    function _stakeManagerImplementationTrusted() internal view virtual returns (bool) {
        return stakeManagerImplementationAddress == stakeManager.implementation();
    }

    /**
     * @notice Migrate all funds to a new vault.
     * @param migrateTo The address of the new vault.
     * @dev This function is only callable by the owner.
     * @dev This function is only callable if the current stake manager is trusted.
     * @dev Reverts when the stake manager reverts or the funds can't be transferred.
     */
    function migrateToVault(address migrateTo) external onlyOwner onlyTrustedStakeManager {
        stakeManager.migrateToVault(migrateTo);
        bool success = STAKING_TOKEN.transfer(migrateTo, STAKING_TOKEN.balanceOf(address(this)));
        if (!success) {
            revert StakeVault__MigrationFailed();
        }
    }
}
