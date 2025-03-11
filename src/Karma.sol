// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Ownable2StepUpgradeable } from "@openzeppelin/contracts-upgradeable/access/Ownable2StepUpgradeable.sol";
import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { ERC20Upgradeable } from "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import { IRewardDistributor } from "./interfaces/IRewardDistributor.sol";
import { EnumerableSet } from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";

/**
 * @title Karma
 * @notice This contract allows for setting rewards for reward distributors.
 * @dev Implementation of the Karma token
 */
contract Karma is Initializable, ERC20Upgradeable, Ownable2StepUpgradeable, UUPSUpgradeable {
    using EnumerableSet for EnumerableSet.AddressSet;

    error Karma__TransfersNotAllowed();
    error Karma__MintAllowanceExceeded();
    error Karma__DistributorAlreadyAdded();
    error Karma__UnknownDistributor();

    event RewardDistributorAdded(address distributor);

    /*//////////////////////////////////////////////////////////////////////////
                                  STATE VARIABLES
    //////////////////////////////////////////////////////////////////////////*/

    /// @notice The name of the token
    string public constant NAME = "Karma";
    /// @notice The symbol of the token
    string public constant SYMBOL = "KARMA";
    /// @notice The total allocation for all reward distributors
    uint256 public totalDistributorAllocation;

    /// @notice Set of reward distributors
    EnumerableSet.AddressSet private rewardDistributors;
    /// @notice Mapping of reward distributor to allocation
    mapping(address distributor => uint256 allocation) public rewardDistributorAllocations;

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
        __ERC20_init(NAME, SYMBOL);
        _transferOwnership(_owner);
        __UUPSUpgradeable_init();
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
    function addRewardDistributor(address distributor) external onlyOwner {
        if (rewardDistributors.contains(distributor)) {
            revert Karma__DistributorAlreadyAdded();
        }

        rewardDistributors.add(address(distributor));
        emit RewardDistributorAdded(distributor);
    }

    /**
     * @notice Removes a reward distributor from the set of reward distributors.
     * @dev Only the owner can remove a reward distributor.
     * @param distributor The address of the reward distributor.
     */
    function removeRewardDistributor(address distributor) external onlyOwner {
        if (!rewardDistributors.contains(distributor)) {
            revert Karma__UnknownDistributor();
        }

        rewardDistributors.remove(distributor);
    }

    /**
     * @notice Sets the reward for a reward distributor.
     * @dev Only the owner can set the reward for a reward distributor.
     * @dev The total allocation for all reward distributors is updated.
     * @param rewardsDistributor The address of the reward distributor.
     * @param amount The amount of rewards to set.
     * @param duration The duration of the rewards.
     */
    function setReward(address rewardsDistributor, uint256 amount, uint256 duration) external onlyOwner {
        if (!rewardDistributors.contains(rewardsDistributor)) {
            revert Karma__UnknownDistributor();
        }

        rewardDistributorAllocations[rewardsDistributor] += amount;
        totalDistributorAllocation += amount;
        IRewardDistributor(rewardsDistributor).setReward(amount, duration);
    }

    /**
     * @notice Mints tokens to an account.
     * @dev Only the owner can mint tokens.
     * @dev The amount minted must not exceed the mint allowance.
     * @param account The account to mint tokens to.
     * @param amount The amount of tokens to mint.
     */
    function mint(address account, uint256 amount) external onlyOwner {
        if (amount > _mintAllowance()) {
            revert Karma__MintAllowanceExceeded();
        }

        _mint(account, amount);
    }

    function transfer(address, uint256) public pure override returns (bool) {
        revert Karma__TransfersNotAllowed();
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

    function _totalSupply() internal view returns (uint256) {
        return super.totalSupply() + _externalSupply();
    }

    function _mintAllowance() internal view returns (uint256) {
        uint256 maxSupply = _externalSupply() * 3;
        uint256 fullTotalSupply = _totalSupply();
        if (maxSupply <= fullTotalSupply) {
            return 0;
        }

        return maxSupply - fullTotalSupply;
    }

    /**
     * @notice Returns the external supply of the token.
     * @dev The external supply is the sum of the rewards from all reward distributors.
     * @return The external supply of the token.
     */
    function _externalSupply() internal view returns (uint256) {
        uint256 externalSupply;

        for (uint256 i = 0; i < rewardDistributors.length(); i++) {
            IRewardDistributor distributor = IRewardDistributor(rewardDistributors.at(i));
            uint256 supply = distributor.totalRewardsSupply();
            if (supply > rewardDistributorAllocations[address(distributor)]) {
                supply = rewardDistributorAllocations[address(distributor)];
            }

            externalSupply += supply;
        }
        return externalSupply;
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
     * @notice Returns the total supply of the token.
     * @dev The total supply is the sum of the token supply and the external supply.
     * @return The total supply of the token.
     */
    function totalSupply() public view override returns (uint256) {
        return _totalSupply();
    }

    /**
     * @notice Returns the reward distributors.
     * @return The reward distributors.
     */
    function getRewardDistributors() external view returns (address[] memory) {
        return rewardDistributors.values();
    }

    /**
     * @notice Returns the mint allowance.
     * @dev The mint allowance is the difference between the external supply and the total supply.
     * @return The mint allowance.
     */
    function mintAllowance() public view returns (uint256) {
        return _mintAllowance();
    }

    /**
     * @notice Returns the balance of an account.
     * @dev The balance of an account is the sum of the balance of the account and the external rewards
     * @param account The account to get the balance of.
     * @return The balance of the account.
     */
    function balanceOf(address account) public view override returns (uint256) {
        uint256 externalBalance;

        for (uint256 i = 0; i < rewardDistributors.length(); i++) {
            address distributor = rewardDistributors.at(i);
            externalBalance += IRewardDistributor(distributor).rewardsBalanceOfAccount(account);
        }

        return super.balanceOf(account) + externalBalance;
    }

    function allowance(address, address) public pure override returns (uint256) {
        return 0;
    }
}
