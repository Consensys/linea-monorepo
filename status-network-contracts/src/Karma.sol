// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { ERC20Upgradeable } from "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import { ERC20VotesUpgradeable } from "./utils/ERC20VotesUpgradeable.sol";
import { IRewardDistributor } from "./interfaces/IRewardDistributor.sol";
import { EnumerableSet } from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import { Math } from "@openzeppelin/contracts/utils/math/Math.sol";

/**
 * @title Karma
 * @notice This contract allows for setting rewards for reward distributors.
 * @dev Implementation of the Karma token
 */
contract Karma is Initializable, ERC20VotesUpgradeable, UUPSUpgradeable, AccessControlUpgradeable {
    using EnumerableSet for EnumerableSet.AddressSet;

    /// @notice Emitted when the address is invalid
    error Karma__InvalidAddress();
    /// @notice Emitted because transfers are not allowed
    error Karma__TransfersNotAllowed();
    /// @notice Emitted when distributor is already added
    error Karma__DistributorAlreadyAdded();
    /// @notice Emitted when distributor is not found
    error Karma__UnknownDistributor();
    /// @notice Emitted sender does not have the required role
    error Karma__Unauthorized();
    /// @notice Emitted when slash percentage to set is invalid
    error Karma__InvalidSlashPercentage();
    /// @notice Emitted when slash reward percentage is invalid
    error Karma__InvalidSlashRewardPercentage();
    /// @notice Emitted when balance to slash is invalid
    error Karma__CannotSlashZeroBalance();
    /// @notice Emitted when there are insufficient funds to transfer
    error Karma__InsufficientTransferBalance();

    /// @notice Emitted when a reward distributor is added
    event RewardDistributorAdded(address distributor);
    /// @notice Emitted when a reward distributor is removed
    event RewardDistributorRemoved(address distributor);
    /// @notice Emitted when an account is slashed
    event AccountSlashed(
        address indexed account, uint256 amount, address indexed rewardRecipient, uint256 rewardAmount
    );
    /// @notice Emitted when the slash percentage is updated
    event SlashPercentageUpdated(uint256 oldPercentage, uint256 newPercentage);
    /// @notice Emitted when the slash reward percentage is updated
    event SlashRewardPercentageUpdated(uint256 oldPercentage, uint256 newPercentage);

    /*//////////////////////////////////////////////////////////////////////////
                                  CONSTANTS
    //////////////////////////////////////////////////////////////////////////*/

    /// @notice Maximum slash percentage (in basis points: 100% = 10000)
    uint256 public constant MAX_SLASH_PERCENTAGE = 10_000;
    /// @notice Minimum slash amount
    uint256 public constant MIN_SLASH_AMOUNT = 1 ether;

    /*//////////////////////////////////////////////////////////////////////////
                                  STATE VARIABLES
    //////////////////////////////////////////////////////////////////////////*/

    /// @notice The name of the token
    string public constant NAME = "Karma";
    /// @notice The symbol of the token
    string public constant SYMBOL = "KARMA";
    /// @notice Set of reward distributors
    EnumerableSet.AddressSet private rewardDistributors;
    /// @notice Maps addresses to their transfer permission
    mapping(address account => bool whitelisted) public allowedToTransfer;
    /// @notice Percentage of Karma to slash (in basis points: 1% = 100, 10% = 1000, 100% = 10000)
    uint256 public slashPercentage;
    /// @notice Percentage of slashed amount to allocate for rewards (in basis points: 1% = 100, 10% = 1000, 100% =
    /// 10000)
    uint256 public slashRewardPercentage;

    /// @notice Operator role keccak256("OPERATOR_ROLE")
    bytes32 public constant OPERATOR_ROLE = 0x97667070c54ef182b0f5858b034beac1b6f3089aa2d3188bb1e8929f4fa9b929;
    /// @notice Slasher role keccak256("SLASHER_ROLE")
    bytes32 public constant SLASHER_ROLE = 0x12b42e8a160f6064dc959c6f251e3af0750ad213dbecf573b4710d67d6c28e39;

    /// @notice Gap for upgrade safety.
    // solhint-disable-next-line
    uint256[30] private __gap_Karma;

    /// @notice Modifier to check if sender is admin or operator
    modifier onlyAdminOrOperator() {
        if (!hasRole(DEFAULT_ADMIN_ROLE, msg.sender) && !hasRole(OPERATOR_ROLE, msg.sender)) {
            revert Karma__Unauthorized();
        }
        _;
    }

    /// @notice Modifier to check if sender has slasher role
    modifier onlySlasher() {
        if (!hasRole(DEFAULT_ADMIN_ROLE, msg.sender) && !hasRole(SLASHER_ROLE, msg.sender)) {
            revert Karma__Unauthorized();
        }
        _;
    }

    /*//////////////////////////////////////////////////////////////////////////
                                     CONSTRUCTOR
    //////////////////////////////////////////////////////////////////////////*/

    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Initializes the contract with the provided owner.
     * @param _owner Address of the owner of the contract.
     */
    function initialize(address _owner) public initializer {
        if (_owner == address(0)) {
            revert Karma__InvalidAddress();
        }
        __ERC20_init(NAME, SYMBOL);
        __ERC20Votes_init();
        __UUPSUpgradeable_init();
        __AccessControl_init();

        _setupRole(DEFAULT_ADMIN_ROLE, _owner);
        slashPercentage = 5000; // 50%
        slashRewardPercentage = 1000; // 10%
    }

    /*//////////////////////////////////////////////////////////////////////////
                           USER-FACING FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    /**
     * @notice Adds a reward distributor to the set of reward distributors.
     * @dev Only the owner can add a reward distributor.
     * @dev Emits a `RewardDistributorAdded` event when a distributor is added.
     * @param distributor The address of the reward distributor.
     */
    function addRewardDistributor(address distributor) public virtual onlyRole(DEFAULT_ADMIN_ROLE) {
        _addRewardDistributor(distributor);
    }

    /**
     * @notice Removes a reward distributor from the set of reward distributors.
     * @dev Only the owner can remove a reward distributor.
     * @dev Burns all karma from the distributor.
     * @param distributor The address of the reward distributor.
     */
    function removeRewardDistributor(address distributor) public virtual onlyRole(DEFAULT_ADMIN_ROLE) {
        _removeRewardDistributor(distributor);
    }

    /**
     * @notice Sets the slash percentage for the contract.
     * @dev Only the admin can configure the slash percentage
     * @param percentage The percentage to set (in basis points: 1% = 100, 10% = 1000, 100% = 10000)
     */
    function setSlashPercentage(uint256 percentage) public onlyRole(DEFAULT_ADMIN_ROLE) {
        if (percentage > 10_000) {
            revert Karma__InvalidSlashPercentage();
        }
        uint256 oldPercentage = slashPercentage;
        slashPercentage = percentage;
        emit SlashPercentageUpdated(oldPercentage, percentage);
    }

    /**
     * @notice Sets the slash reward percentage for the contract.
     * @dev Only the admin or operator can configure the slash reward percentage
     * @param percentage The percentage to set (in basis points: 1% = 100, 10% = 1000, 100% = 10000)
     */
    function setSlashRewardPercentage(uint256 percentage) public onlyRole(DEFAULT_ADMIN_ROLE) {
        if (percentage > 10_000) {
            revert Karma__InvalidSlashRewardPercentage();
        }
        uint256 oldPercentage = slashRewardPercentage;
        slashRewardPercentage = percentage;
        emit SlashRewardPercentageUpdated(oldPercentage, percentage);
    }

    /**
     * @notice Sets whether an account is allowed to transfer tokens.
     * @dev Only the admin can set transfer permissions.
     * @param account The address of the account.
     * @param allowed Boolean indicating whether the account is allowed to transfer tokens.
     */
    function setAllowedToTransfer(address account, bool allowed) public onlyRole(DEFAULT_ADMIN_ROLE) {
        if (account == address(0)) {
            revert Karma__InvalidAddress();
        }
        allowedToTransfer[account] = allowed;
    }

    /**
     * @notice Sets the reward for a reward distributor.
     * @dev Only the owner can set the reward for a reward distributor.
     * @dev The total allocation for all reward distributors is updated.
     * @param rewardsDistributor The address of the reward distributor.
     * @param amount The amount of rewards to set.
     * @param duration The duration of the rewards.
     */
    function setReward(
        address rewardsDistributor,
        uint256 amount,
        uint256 duration
    )
        public
        virtual
        onlyAdminOrOperator
    {
        _setReward(rewardsDistributor, amount, duration);
    }

    /**
     * @notice Mints tokens to an account.
     * @dev Only the owner can mint tokens.
     * @dev The amount minted must not exceed the mint allowance.
     * @param account The account to mint tokens to.
     * @param amount The amount of tokens to mint.
     */
    function mint(address account, uint256 amount) public virtual onlyAdminOrOperator {
        _mint(account, amount);
    }

    /**
     * @notice Slashes karma from an account based on the current slashing percentage
     * @dev Only accounts with the SLASHER_ROLE can call this function
     * @param account Account to slash
     * @param rewardRecipient Address that will receive the slash reward
     * @return slashedAmount The amount of karma that was slashed
     */
    function slash(address account, address rewardRecipient) public virtual onlySlasher returns (uint256) {
        return _slash(account, rewardRecipient);
    }

    function calculateSlashAmount(uint256 value) public view returns (uint256) {
        return _calculateSlashAmount(value);
    }

    /**
     * @notice Transfers tokens from the caller to a specified address.
     * @dev Transfers are only allowed if the caller is whitelisted in `allowedToTransfer`.
     * @param to The address to transfer tokens to.
     * @param amount The amount of tokens to transfer.
     * @return A boolean value indicating whether the operation succeeded.
     */
    function transfer(address to, uint256 amount) public override returns (bool) {
        address owner = _msgSender();
        _transfer(owner, to, amount);
        return true;
    }

    function approve(address, uint256) public pure override returns (bool) {
        revert Karma__TransfersNotAllowed();
    }

    function transferFrom(address, address, uint256) public pure override returns (bool) {
        revert Karma__TransfersNotAllowed();
    }

    /*//////////////////////////////////////////////////////////////////////////
                           INTERNAL FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    function _beforeTokenTransfer(address from, address to, uint256) internal view override {
        if (from != address(0) && to != address(0)) {
            if (!allowedToTransfer[msg.sender]) {
                revert Karma__TransfersNotAllowed();
            }

            if (rewardDistributors.contains(to)) {
                revert Karma__TransfersNotAllowed();
            }
        }
    }

    /**
     * @notice Authorizes contract upgrades via UUPS.
     * @dev This function is only callable by the owner.
     */
    function _authorizeUpgrade(address) internal view override {
        if (!hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert Karma__Unauthorized();
        }
    }

    /**
     * @notice Adds a reward distributor to the set of reward distributors.
     * @param distributor The address of the reward distributor.
     */
    function _addRewardDistributor(address distributor) internal virtual {
        if (rewardDistributors.contains(distributor)) {
            revert Karma__DistributorAlreadyAdded();
        }
        rewardDistributors.add(distributor);
        emit RewardDistributorAdded(distributor);
    }

    /**
     * @notice Removes a reward distributor from the set of reward distributors.
     * @param distributor The address of the reward distributor.
     */
    function _removeRewardDistributor(address distributor) internal virtual {
        if (!rewardDistributors.contains(distributor)) {
            revert Karma__UnknownDistributor();
        }
        _burn(distributor, super.balanceOf(distributor));
        rewardDistributors.remove(distributor);
        emit RewardDistributorRemoved(distributor);
    }

    /**
     * @notice Sets the reward for a reward distributor.
     * @dev Mints actual tokens to the reward distributor.
     * @param rewardsDistributor The address of the reward distributor.
     * @param amount The amount of rewards to set.
     * @param duration The duration of the rewards.
     */
    function _setReward(address rewardsDistributor, uint256 amount, uint256 duration) internal virtual {
        if (!rewardDistributors.contains(rewardsDistributor)) {
            revert Karma__UnknownDistributor();
        }
        _mint(rewardsDistributor, amount);
        IRewardDistributor(rewardsDistributor).setReward(amount, duration);
    }

    /**
     * @notice Slashes karma from an account based on the current slashing percentage
     * @param account Account to slash
     * @param rewardRecipient Address that will receive the slash reward
     * @return slashedAmount The amount of karma that was slashed
     */
    function _slash(address account, address rewardRecipient) internal virtual returns (uint256) {
        uint256 currentBalance = _balanceOf(account);
        if (currentBalance == 0) {
            revert Karma__CannotSlashZeroBalance();
        }

        // first, calculate the total amount to slash from the actual reward tokens
        uint256 totalAmountToSlash = _calculateSlashAmount(super.balanceOf(account));

        for (uint256 i = 0; i < rewardDistributors.length(); i++) {
            address distributor = rewardDistributors.at(i);
            uint256 currentDistributorAccountBalance = IRewardDistributor(distributor).rewardsBalanceOfAccount(account);

            // then, calculate the amount to slash from each reward distributor
            totalAmountToSlash += _calculateSlashAmount(currentDistributorAccountBalance);

            // turn virtual Karma into real Karma for slashing
            IRewardDistributor(distributor).redeemRewards(account);
        }

        // Calculate reward from the slashed amount
        // slashRewardPercentage of the slashed amount goes to the reward recipient
        uint256 rewardAmount = Math.mulDiv(totalAmountToSlash, slashRewardPercentage, MAX_SLASH_PERCENTAGE);

        // Burn the entire slashed amount from the account
        _burn(account, totalAmountToSlash);

        // Mint the reward to the recipient (if recipient is specified)
        if (rewardAmount > 0 && rewardRecipient != address(0)) {
            _mint(rewardRecipient, rewardAmount);
        }

        emit AccountSlashed(account, totalAmountToSlash, rewardRecipient, rewardAmount);
        return totalAmountToSlash;
    }

    /**
     * @notice Calculates the amount to slash from a given balance, falling back to the minimum slash amount if
     * necessary.
     */
    function _calculateSlashAmount(uint256 balance) internal view returns (uint256) {
        uint256 amountToSlash = Math.mulDiv(balance, slashPercentage, MAX_SLASH_PERCENTAGE);
        if (amountToSlash < MIN_SLASH_AMOUNT) {
            if (balance < MIN_SLASH_AMOUNT) {
                // Not enough balance for minimum slash, slash entire balance
                amountToSlash = balance;
            } else {
                amountToSlash = MIN_SLASH_AMOUNT;
            }
        }
        return amountToSlash;
    }

    function _balanceOf(address account) internal view returns (uint256) {
        uint256 externalBalance = 0;

        // first, aggregate all slashed amounts from reward distributors
        for (uint256 i = 0; i < rewardDistributors.length(); i++) {
            address distributor = rewardDistributors.at(i);
            externalBalance += IRewardDistributor(distributor).rewardsBalanceOfAccount(account);
        }

        return (super.balanceOf(account) + externalBalance);
    }

    /*//////////////////////////////////////////////////////////////////////////
                           VIEW FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    /**
     * @notice Returns the total supply of the token.
     * @dev The total supply is the sum of the token supply and the external supply.
     * @return The total supply of the token.
     */
    function totalSupply() public view override returns (uint256) {
        uint256 externalSupply = 0;
        uint256 totalDistributorBalance = 0;

        for (uint256 i = 0; i < rewardDistributors.length(); i++) {
            IRewardDistributor distributor = IRewardDistributor(rewardDistributors.at(i));
            externalSupply += distributor.totalRewardsSupply();
            totalDistributorBalance += super.balanceOf(address(distributor));
        }

        if (externalSupply > totalDistributorBalance) {
            externalSupply = totalDistributorBalance;
        }

        // subtract the distributor balances to avoid double counting
        return super.totalSupply() - totalDistributorBalance + externalSupply;
    }

    /**
     * @notice Returns the reward distributors.
     * @return The reward distributors.
     */
    function getRewardDistributors() external view returns (address[] memory) {
        return rewardDistributors.values();
    }

    /**
     * @notice Returns the balance of an account.
     * @dev The balance of an account is the sum of the balance of the account and the external rewards
     * @param account The account to get the balance of.
     * @return The balance of the account.
     */
    function balanceOf(address account) public view override returns (uint256) {
        if (rewardDistributors.contains(account)) {
            return 0;
        }
        return _balanceOf(account);
    }

    /**
     * @notice Returns the actual token balance of an account.
     * @dev This value excludes virtual Karma distributed by reward distributors.
     * @param account The account to get the token balance of.
     * @return The actual token balance of the account.
     */
    function actualTokenBalanceOf(address account) public view returns (uint256) {
        return super.balanceOf(account);
    }

    /**
     * @notice Returns the balance of a reward distributor.
     * @dev The balance of a reward distributor is the balance of the actual tokens held by the distributor.
     * @param distributor The address of the reward distributor.
     * @return The balance of the reward distributor.
     */
    function balanceOfRewardDistributor(address distributor) external view returns (uint256) {
        if (!rewardDistributors.contains(distributor)) {
            revert Karma__UnknownDistributor();
        }
        return super.balanceOf(distributor);
    }

    /**
     * @notice Returns the allowance of an account.
     * @dev Allowances are not used in this contract, so this function always returns 0.
     * @return Always returns 0.
     */
    function allowance(address, address) public pure override returns (uint256) {
        return 0;
    }
}
