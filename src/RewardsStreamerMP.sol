// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Math } from "@openzeppelin/contracts/utils/math/Math.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { IStakeManager } from "./interfaces/IStakeManager.sol";
import { IStakeVault } from "./interfaces/IStakeVault.sol";
import { IRewardDistributor } from "./interfaces/IRewardDistributor.sol";
import { TrustedCodehashAccess } from "./TrustedCodehashAccess.sol";
import { StakeMath } from "./math/StakeMath.sol";

// solhint-disable max-states-count
/**
 * @title RewardsStreamerMP
 * @notice A contract that manages staking and rewards for multiple vaults.
 * @dev This contract is upgradeable using the UUPS pattern.
 * @dev Uses `TrustedCodeHashAccess` to whitelist trusted vaults.
 */
contract RewardsStreamerMP is
    Initializable,
    UUPSUpgradeable,
    IStakeManager,
    TrustedCodehashAccess,
    IRewardDistributor,
    StakeMath
{
    struct VaultData {
        uint256 stakedBalance;
        uint256 rewardIndex;
        uint256 mpAccrued;
        uint256 maxMP;
        uint256 lastMPUpdateTime;
        uint256 lockUntil;
        uint256 mpStaked;
        uint256 rewardsAccrued;
    }

    /*//////////////////////////////////////////////////////////////////////////
                                  STATE VARIABLES
    //////////////////////////////////////////////////////////////////////////*/

    // solhint-disable var-name-mixedcase
    /// @notice Token that is staked in the vaults (SNT).
    IERC20 public STAKING_TOKEN;
    /// @notice Scale factor used for rewards calculation.
    uint256 public constant SCALE_FACTOR = 1e27;
    /// @notice Total staked balance in the system.
    uint256 public totalStaked;
    /// @notice Total amount of staked multiplier points
    uint256 public totalMPStaked;
    /// @notice Total multiplier points accrued.
    uint256 public totalMPAccrued;
    /// @notice Total rewards accrued in the system.
    uint256 public totalRewardsAccrued;
    /// @notice Total maximum multiplier points that can be accrued.
    uint256 public totalMaxMP;
    /// @notice Time of the last multiplier points update.
    uint256 public lastMPUpdatedTime;
    /// @notice Index of the current reward period.
    uint256 public lastRewardIndex;
    /// @notice The address that can set rewards
    address public rewardsSupplier;
    /// @notice Amount of rewards available for distribution.
    uint256 public rewardAmount;
    /// @notice Time of the last reward update.
    uint256 public lastRewardTime;
    /// @notice Time when rewards start.
    uint256 public rewardStartTime;
    /// @notice Time when rewards end.
    uint256 public rewardEndTime;
    /// @notice Maps vault addresses to vault data
    mapping(address vault => VaultData data) public vaultData;
    /// @notice Maps Account address to a list of vaults
    mapping(address owner => address[] vault) public vaults;
    /// @notice Maps vault addresses to their owners
    mapping(address vault => address owner) public vaultOwners;
    /// @notice Flag to enable emergency mode.
    bool public emergencyModeEnabled;

    modifier onlyRegisteredVault() {
        if (vaultOwners[msg.sender] == address(0)) {
            revert StakingManager__VaultNotRegistered();
        }
        _;
    }

    modifier onlyNotEmergencyMode() {
        if (emergencyModeEnabled) {
            revert StakingManager__EmergencyModeEnabled();
        }
        _;
    }

    modifier onlyRewardsSupplier() {
        if (msg.sender != rewardsSupplier) {
            revert StakingManager__Unauthorized();
        }
        _;
    }

    /*//////////////////////////////////////////////////////////////////////////
                                     CONSTRUCTOR
    //////////////////////////////////////////////////////////////////////////*/

    /**
     * @notice Initializes the contract.
     * @dev Disables initializers to prevent reinitialization.
     */
    constructor() {
        _disableInitializers();
    }

    /*//////////////////////////////////////////////////////////////////////////
                           USER-FACING FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    /**
     * @notice Initializes the contract with the provided owner and staking token.
     * @dev Also sets the initial `lastMPUpdatedTime`
     * @param _owner Address of the owner of the contract.
     * @param _stakingToken Address of the staking token.
     */
    function initialize(address _owner, address _stakingToken) external initializer {
        __TrustedCodehashAccess_init(_owner);
        __UUPSUpgradeable_init();

        STAKING_TOKEN = IERC20(_stakingToken);
        lastMPUpdatedTime = block.timestamp;
    }

    /**
     * @notice Allows the owner to set the rewards supplier.
     * @dev The supplier is going to be the `Karma` token.
     * @param _rewardsSupplier The address of the rewards supplier.
     */
    function setRewardsSupplier(address _rewardsSupplier) external onlyOwner onlyNotEmergencyMode {
        rewardsSupplier = _rewardsSupplier;
    }

    /**
     * @notice Registers a vault with its owner. Called by the vault itself during initialization.
     * @dev Only callable by contracts with trusted codehash
     */
    function registerVault() external onlyTrustedCodehash onlyNotEmergencyMode {
        address vault = msg.sender;
        address owner = IStakeVault(vault).owner();

        if (vaultOwners[vault] != address(0)) {
            revert StakingManager__VaultAlreadyRegistered();
        }

        // Verify this is a legitimate vault by checking it points to stakeManager
        if (address(IStakeVault(vault).stakeManager()) != address(this)) {
            revert StakingManager__InvalidVault();
        }

        vaultOwners[vault] = owner;
        vaults[owner].push(vault);
        emit VaultRegistered(vault, owner);
    }

    /**
     * @notice Allows users to stake and start accruing MPs.
     * @dev Called by trusted `StakeVault`.
     * @dev Can only be called when emergency mode is disabled.
     * @dev Only registered vaults are allowed to stake.
     * @param amount The amount of tokens to stake
     * @param lockPeriod The duration to lock the stake
     */
    function stake(
        uint256 amount,
        uint256 lockPeriod
    )
        external
        onlyTrustedCodehash
        onlyNotEmergencyMode
        onlyRegisteredVault
    {
        if (amount == 0) {
            revert StakingManager__AmountCannotBeZero();
        }

        _updateGlobalState();
        _updateVault(msg.sender, true);

        VaultData storage vault = vaultData[msg.sender];

        (uint256 _deltaMpTotal, uint256 _deltaMPMax, uint256 _newLockEnd) =
            _calculateStake(vault.stakedBalance, vault.maxMP, vault.lockUntil, block.timestamp, amount, lockPeriod);

        vault.stakedBalance += amount;
        totalStaked += amount;
        totalMPStaked += _deltaMpTotal;

        if (lockPeriod != 0) {
            vault.lockUntil = _newLockEnd;
        }

        vault.mpAccrued += _deltaMpTotal;
        vault.mpStaked += _deltaMpTotal;
        totalMPAccrued += _deltaMpTotal;

        vault.maxMP += _deltaMPMax;
        totalMaxMP += _deltaMPMax;

        vault.rewardIndex = lastRewardIndex;

        emit Staked(msg.sender, amount, lockPeriod);
    }

    /**
     * @notice Allows users to lock their staked balance for a specified duration.
     * @dev Called by trusted `StakeVault`.
     * @dev Can only be called when emergency mode is disabled.
     * @dev Only registered vaults are allowed to lock.
     * @param lockPeriod The duration to lock the stake
     */
    function lock(uint256 lockPeriod) external onlyTrustedCodehash onlyNotEmergencyMode onlyRegisteredVault {
        VaultData storage vault = vaultData[msg.sender];

        if (lockPeriod == 0) {
            revert StakingManager__DurationCannotBeZero();
        }

        _updateGlobalState();
        _updateVault(msg.sender, false);
        (uint256 deltaMp, uint256 newLockEnd) =
            _calculateLock(vault.stakedBalance, vault.maxMP, vault.lockUntil, block.timestamp, lockPeriod);

        // Update account state
        vault.lockUntil = newLockEnd;
        vault.mpAccrued += deltaMp;
        vault.mpStaked += deltaMp;
        vault.maxMP += deltaMp;

        // Update global state
        totalMPAccrued += deltaMp;
        totalMPStaked += deltaMp;
        totalMaxMP += deltaMp;

        vault.rewardIndex = lastRewardIndex;

        emit Locked(msg.sender, lockPeriod, newLockEnd);
    }

    /**
     * @notice Allows users to unstake their staked balance.
     * @dev Called by trusted `StakeVault`.
     * @dev Can only be called when emergency mode is disabled.
     * @dev Only registered vaults are allowed to unstake.
     * @dev Unstaking reduces accrued MPs proportionally.
     * @param amount The amount of tokens to unstake
     */
    function unstake(uint256 amount) external onlyTrustedCodehash onlyNotEmergencyMode onlyRegisteredVault {
        VaultData storage vault = vaultData[msg.sender];
        _unstake(amount, vault, msg.sender);
        emit Unstaked(msg.sender, amount);
    }

    // @notice Allows an account to leave the system. This can happen when a
    //         user doesn't agree with an upgrade of the stake manager.
    // @dev This function is protected by whitelisting the codehash of the caller.
    //      This ensures `StakeVault`s will call this function only if they don't
    //      trust the `StakeManager` (e.g. in case of an upgrade).
    function leave() external onlyTrustedCodehash {
        VaultData storage vault = vaultData[msg.sender];

        if (vault.stakedBalance > 0) {
            //updates lockuntil to allow unstake early
            vault.lockUntil = block.timestamp;
            // calling `_unstake` to update accounting accordingly
            _unstake(vault.stakedBalance, vault, msg.sender);

            // further cleanup that isn't done in `_unstake`
            vault.rewardIndex = 0;
            vault.lockUntil = 0;
            vault.lastMPUpdateTime = 0;
        }

        emit AccountLeft(msg.sender);
    }

    /**
     * @notice Allows the owner to update the global state.
     * @dev This function is only callable when emergency mode is disabled.
     * @dev Takes care of updating the global MP and reward index.
     */
    function updateGlobalState() external onlyNotEmergencyMode {
        _updateGlobalState();
    }

    /**
     * @notice Allows an admin to set the reward amount and duration.
     * @dev This function is only callable by the owner.
     * @param amount The amount of rewards to distribute.
     * @param duration The duration of the reward period.
     */
    function setReward(uint256 amount, uint256 duration) external onlyRewardsSupplier {
        if (rewardEndTime > block.timestamp) {
            revert StakingManager__RewardPeriodNotEnded();
        }

        if (duration == 0) {
            revert StakingManager__DurationCannotBeZero();
        }

        if (amount == 0) {
            revert StakingManager__AmountCannotBeZero();
        }

        // this will call updateRewardIndex and update the totalRewardsAccrued
        _updateGlobalState();

        // in case updateRewardIndex returns earlier,
        // we still update the lastRewardTime
        lastRewardTime = block.timestamp;
        rewardAmount = amount;
        rewardStartTime = block.timestamp;
        rewardEndTime = block.timestamp + duration;
    }

    /**
     * @notice Allows any user to compound accrued MP for any user.
     * @dev This function is only callable when emergency mode is disabled.
     * @dev Anyone can compound MPs for account.
     */
    function updateAccount(address account) external onlyNotEmergencyMode {
        _updateGlobalState();
        address[] memory accountVaults = vaults[account];
        for (uint256 i = 0; i < accountVaults.length; i++) {
            _updateVault(accountVaults[i], false);
        }
    }

    /**
     * @notice Allows users to claim their accrued rewards.
     * @dev This function is only callable when emergency mode is disabled.
     * @dev Anyone can claim rewards on behalf of any vault
     */
    function updateVault(address vaultAddress) external onlyNotEmergencyMode {
        _updateGlobalState();
        _updateVault(vaultAddress, false);
    }

    /**
     * @notice Enables emergency mode.
     * @dev This function is only callable when emergency mode is disabled.
     * @dev Only the owner of the contract can call this function.
     */
    function enableEmergencyMode() external onlyOwner onlyNotEmergencyMode {
        emergencyModeEnabled = true;
        emit EmergencyModeEnabled();
    }

    /**
     * @notice Migrate the staked balance of a vault to another vault.
     * @param migrateTo The address of the vault to migrate to.
     * @dev This function is only callable by trusted stake vaults.
     * @dev Reverts if the vault to migrate to is not owned by the same user.
     * @dev Revets if the vault to migrate to has a non-zero staked balance.
     */
    function migrateToVault(address migrateTo) external onlyNotEmergencyMode onlyTrustedCodehash onlyRegisteredVault {
        // first ensure the vault to migrate to is actually owned by the same user
        if (IStakeVault(msg.sender).owner() != IStakeVault(migrateTo).owner()) {
            revert StakingManager__Unauthorized();
        }

        if (vaultData[migrateTo].stakedBalance > 0) {
            revert StakingManager__MigrationTargetHasFunds();
        }

        _updateGlobalState();
        _updateVault(msg.sender, false);

        VaultData storage oldVault = vaultData[msg.sender];
        VaultData storage newVault = vaultData[migrateTo];

        // migrate vault data to new vault
        newVault.stakedBalance = oldVault.stakedBalance;
        newVault.rewardIndex = oldVault.rewardIndex;
        newVault.mpStaked = oldVault.mpStaked;
        newVault.mpAccrued = oldVault.mpAccrued;
        newVault.maxMP = oldVault.maxMP;
        newVault.lastMPUpdateTime = oldVault.lastMPUpdateTime;
        newVault.lockUntil = oldVault.lockUntil;
        newVault.rewardsAccrued = oldVault.rewardsAccrued;

        delete vaultData[msg.sender];

        emit VaultMigrated(msg.sender, migrateTo);
    }

    /*//////////////////////////////////////////////////////////////////////////
                           INTERNAL FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    function _updateGlobalState() internal {
        _updateGlobalMP();
        _updateRewardIndex();
    }

    function _updateGlobalMP() internal {
        uint256 newTotalMPAccrued = _totalMP();
        if (newTotalMPAccrued > totalMPAccrued) {
            totalMPAccrued = newTotalMPAccrued;
            lastMPUpdatedTime = block.timestamp;
        }
    }

    function _updateVault(address vaultAddress, bool forceMPUpdate) internal {
        VaultData storage vault = vaultData[vaultAddress];
        uint256 accruedMP = _vaultPendingMP(vault);
        if (accruedMP > 0 || forceMPUpdate) {
            vault.mpAccrued += accruedMP;
            vault.lastMPUpdateTime = block.timestamp;
        }

        uint256 rewardsAccrued = _vaultPendingRewards(vault);
        vault.rewardsAccrued += rewardsAccrued;
        vault.rewardIndex = lastRewardIndex;

        uint256 mpToStake = vault.mpAccrued - vault.mpStaked;
        vault.mpStaked += mpToStake;
        totalMPStaked += mpToStake;
        emit Compound(vaultAddress, mpToStake);
    }

    function _updateRewardIndex() internal {
        uint256 accruedRewards;
        uint256 newRewardIndex;

        (accruedRewards, newRewardIndex) = _rewardIndex();
        totalRewardsAccrued += accruedRewards;

        if (newRewardIndex > lastRewardIndex) {
            lastRewardIndex = newRewardIndex;
            lastRewardTime = block.timestamp < rewardEndTime ? block.timestamp : rewardEndTime;
        }
    }

    function _unstake(uint256 amount, VaultData storage vault, address vaultAddress) internal {
        _updateGlobalState();
        _updateVault(vaultAddress, false);

        (uint256 _deltaMpTotal, uint256 _deltaMpMax) = _calculateUnstake(
            vault.stakedBalance, vault.lockUntil, block.timestamp, vault.mpAccrued, vault.maxMP, amount
        );
        vault.stakedBalance -= amount;
        vault.maxMP -= _deltaMpMax;
        vault.rewardIndex = lastRewardIndex;
        vault.mpAccrued -= _deltaMpTotal;

        // A vault's "staked" MP will be reduced if the accrued MP is less than the staked MP.
        // This is because the accrued MP represents the vault's total MP
        if (vault.mpAccrued < vault.mpStaked) {
            totalMPStaked -= vault.mpStaked - vault.mpAccrued;
            vault.mpStaked = vault.mpAccrued;
        }

        totalMPAccrued -= _deltaMpTotal;
        totalMaxMP -= _deltaMpMax;
        totalStaked -= amount;

        // if the user can unstake it means the lock period has ended
        // and we can reset lockUntil
        vault.lockUntil = 0;
    }

    function _totalShares() internal view returns (uint256) {
        return totalStaked + totalMPStaked;
    }

    function _vaultShares(address vaultAddress) internal view returns (uint256) {
        VaultData storage vault = vaultData[vaultAddress];
        return vault.stakedBalance + vault.mpStaked;
    }

    function _totalMP() internal view returns (uint256) {
        if (totalMaxMP == 0) {
            return totalMPAccrued;
        }

        uint256 currentTime = block.timestamp;
        uint256 timeDiff = currentTime - lastMPUpdatedTime;
        if (timeDiff == 0) {
            return totalMPAccrued;
        }

        uint256 accruedMP = _accrueMP(totalStaked, timeDiff);
        if (totalMPAccrued + accruedMP > totalMaxMP) {
            accruedMP = totalMaxMP - totalMPAccrued;
        }

        uint256 newTotalMPAccrued = totalMPAccrued + accruedMP;

        return newTotalMPAccrued;
    }

    function _rewardIndex() internal view returns (uint256, uint256) {
        uint256 shares = _totalShares();

        if (shares == 0) {
            return (0, lastRewardIndex);
        }

        uint256 currentTime = block.timestamp;
        uint256 applicableTime = rewardEndTime > currentTime ? currentTime : rewardEndTime;
        uint256 elapsedTime = applicableTime - lastRewardTime;

        if (elapsedTime == 0) {
            return (0, lastRewardIndex);
        }

        uint256 accruedRewards = _calculatePendingRewards();
        if (accruedRewards == 0) {
            return (0, lastRewardIndex);
        }

        uint256 newRewardIndex = lastRewardIndex + Math.mulDiv(accruedRewards, SCALE_FACTOR, shares);

        return (accruedRewards, newRewardIndex);
    }

    function _calculatePendingRewards() internal view returns (uint256) {
        if (rewardEndTime <= rewardStartTime) {
            // No active reward period
            return 0;
        }

        uint256 currentTime = block.timestamp < rewardEndTime ? block.timestamp : rewardEndTime;

        if (currentTime <= lastRewardTime) {
            // No new rewards have accrued since lastRewardTime
            return 0;
        }

        uint256 timeElapsed = currentTime - lastRewardTime;
        uint256 duration = rewardEndTime - rewardStartTime;

        if (duration == 0) {
            // Prevent division by zero
            return 0;
        }

        uint256 accruedRewards = Math.mulDiv(timeElapsed, rewardAmount, duration);
        return accruedRewards;
    }

    function _vaultPendingMP(VaultData storage vault) internal view returns (uint256) {
        if (block.timestamp == vault.lastMPUpdateTime) {
            return 0;
        }
        if (vault.maxMP == 0 || vault.stakedBalance == 0) {
            return 0;
        }

        uint256 deltaMP = _calculateAccrual(
            vault.stakedBalance, vault.mpAccrued, vault.maxMP, vault.lastMPUpdateTime, block.timestamp
        );

        return deltaMP;
    }

    function _mpBalanceOf(address vaultAddress) internal view returns (uint256) {
        VaultData storage vault = vaultData[vaultAddress];
        return vault.mpAccrued + _vaultPendingMP(vault);
    }

    function _vaultPendingRewards(VaultData storage vault) internal view returns (uint256) {
        uint256 newRewardIndex;
        (, newRewardIndex) = _rewardIndex();

        uint256 accountShares = vault.stakedBalance + vault.mpStaked;
        uint256 deltaRewardIndex = newRewardIndex - vault.rewardIndex;

        return (accountShares * deltaRewardIndex) / SCALE_FACTOR;
    }

    /**
     * @notice Authorizes contract upgrades via UUPS.
     * @dev This function is only callable by the owner.
     */
    function _authorizeUpgrade(address) internal view override {
        _checkOwner();
    }

    /*//////////////////////////////////////////////////////////////////////////
                           VIEW FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    /**
     * @notice Get the vaults owned by a user
     * @param account The address of the user
     * @return The vaults owned by the user
     */
    function getAccountVaults(address account) external view returns (address[] memory) {
        return vaults[account];
    }

    /**
     * @notice Returns the total multiplier points accrued in the system.
     * @return The total multiplier points accrued in the system.
     */
    function totalMP() external view returns (uint256) {
        return _totalMP();
    }

    /**
     * @notice Returns the total shares in the system.
     * @dev Total shares are total staked tokens and total multiplier points staked.
     * @return The total shares in the system.
     */
    function totalShares() external view returns (uint256) {
        return _totalShares();
    }

    /**
     * @notice Returns total rewards supply in the system.
     * @return Total rewards supply in the system.
     */
    function totalRewardsSupply() external view returns (uint256) {
        return totalRewardsAccrued + _calculatePendingRewards();
    }

    /**
     * @notice Returns the total shares of a given vault.
     * @return The total vault shares
     */
    function vaultShares(address vaultAddress) external view returns (uint256) {
        return _vaultShares(vaultAddress);
    }

    /**
     * @notice Returns vault data for a given vault address.
     * @return Vault data for the given vault address
     */
    function getVault(address vaultAddress) external view returns (VaultData memory) {
        return vaultData[vaultAddress];
    }

    /**
     * @notice Returns the staked balance of a vault.
     * @param vaultAddress The address of the vault.
     */
    function getStakedBalance(address vaultAddress) external view returns (uint256) {
        return vaultData[vaultAddress].stakedBalance;
    }

    /**
     * @notice Returns the rewards balance of a vault.
     * @return Rewards balance of the vault.
     */
    function rewardsBalanceOf(address vaultAddress) public view returns (uint256) {
        VaultData storage vault = vaultData[vaultAddress];
        return vault.rewardsAccrued + _vaultPendingRewards(vault);
    }

    /**
     * @notice Returns the multiplier points balance of a vault.
     * @return Multiplier points balance of the vault.
     */
    function mpBalanceOf(address vaultAddress) external view returns (uint256) {
        return _mpBalanceOf(vaultAddress);
    }

    /**
     * @notice Returns staked multiplier points of a vault.
     * @return Staked multiplier points of the vault.
     */
    function mpStakedOf(address vaultAddress) external view returns (uint256) {
        VaultData storage vault = vaultData[vaultAddress];
        return vault.mpStaked;
    }

    /**
     * @notice Get the total multiplier points for a user
     * @dev Iterates over all vaults owned by the user and sums the multiplier points
     * @param account The address of the user
     * @return The total multiplier points for the user
     */
    function mpBalanceOfAccount(address account) external view returns (uint256) {
        address[] memory accountVaults = vaults[account];
        uint256 accountTotalMP = 0;

        for (uint256 i = 0; i < accountVaults.length; i++) {
            VaultData storage vault = vaultData[accountVaults[i]];
            accountTotalMP += vault.mpAccrued + _vaultPendingMP(vault);
        }
        return accountTotalMP;
    }

    /**
     * @notice Get the total maximum multiplier points for a user
     * @dev Iterates over all vaults owned by the user and sums the maximum multiplier points
     * @param account The address of the user
     * @return The total maximum multiplier points for the user
     */
    function getAccountTotalMaxMP(address account) external view returns (uint256) {
        address[] memory accountVaults = vaults[account];
        uint256 accountTotalMaxMP = 0;

        for (uint256 i = 0; i < accountVaults.length; i++) {
            accountTotalMaxMP += vaultData[accountVaults[i]].maxMP;
        }
        return accountTotalMaxMP;
    }

    /**
     * @notice Get the total staked balance for a user
     * @dev Iterates over all vaults owned by the user and sums the staked balances
     * @param account The address of the user
     * @return The total staked balance for the user
     */
    function getAccountTotalStakedBalance(address account) external view returns (uint256) {
        address[] memory accountVaults = vaults[account];
        uint256 accountTotalStake = 0;

        for (uint256 i = 0; i < accountVaults.length; i++) {
            accountTotalStake += vaultData[accountVaults[i]].stakedBalance;
        }
        return accountTotalStake;
    }

    /**
     * @notice Returns the rewards balance of an account.
     * @dev Iterates over all vaults owned by the account and sums the rewards.
     * @return Rewards balance of the account.
     */
    function rewardsBalanceOfAccount(address account) external view returns (uint256) {
        address[] memory accountVaults = vaults[account];
        uint256 accountTotalRewards = 0;

        for (uint256 i = 0; i < accountVaults.length; i++) {
            accountTotalRewards += rewardsBalanceOf(accountVaults[i]);
        }
        return accountTotalRewards;
    }
}
