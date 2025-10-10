// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

import { IRewardDistributor } from "./interfaces/IRewardDistributor.sol";

/**
 * @title SimpleKarmaDistributor
 * @notice Upgradeable reward distributor that tracks virtual balances before redeeming Karma.
 */
contract SimpleKarmaDistributor is Initializable, UUPSUpgradeable, AccessControlUpgradeable, IRewardDistributor {
    using SafeERC20 for IERC20;

    /// @notice Emitted when an invalid zero address is provided.
    error SimpleKarmaDistributor__InvalidAddress();
    /// @notice Emitted when the caller does not have the required role.
    error SimpleKarmaDistributor__Unauthorized();
    /// @notice Emitted when trying to mint more Karma than the available supply.
    error SimpleKarmaDistributor__InsufficientAvailableSupply();

    /// @notice Emitted when the rewards supplier address is updated.
    event RewardsSupplierSet(address rewardsSupplier);
    /// @notice Emitted when rewards are supplied to this distributor.
    event RewardSet(uint256 amount);
    /// @notice Emitted when virtual rewards are minted for an account.
    event RewardMinted(address indexed account, uint256 amount);
    /// @notice Emitted when an account redeems its virtual rewards for Karma.
    event RewardsRedeemed(address indexed account, uint256 amount);

    /// @notice Operator role identifier (keccak256("OPERATOR_ROLE")).
    bytes32 public constant OPERATOR_ROLE = 0x97667070c54ef182b0f5858b034beac1b6f3089aa2d3188bb1e8929f4fa9b929;

    /// @notice The Karma token handled by this distributor.
    IERC20 public karmaToken;
    /// @notice Address that is allowed to provide additional rewards.
    address public rewardsSupplier;
    /// @notice Total amount of Karma available to be minted.
    uint256 public availableSupply;
    /// @notice Total amount of Karma virtually minted but not yet redeemed.
    uint256 public mintedSupply;
    /// @notice Virtual reward balances per account.
    mapping(address account => uint256 amount) public balances;

    /// @notice Storage gap for upgrade safety.
    // solhint-disable-next-line
    uint256[44] private __gap_SimpleKarmaDistributor;

    modifier onlyAdminOrOperator() {
        if (!hasRole(DEFAULT_ADMIN_ROLE, msg.sender) && !hasRole(OPERATOR_ROLE, msg.sender)) {
            revert SimpleKarmaDistributor__Unauthorized();
        }
        _;
    }

    modifier onlyRewardsSupplier() {
        if (msg.sender != rewardsSupplier) {
            revert SimpleKarmaDistributor__Unauthorized();
        }
        _;
    }

    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Initializes the reward distributor.
     * @param _owner Address with admin privileges for this contract.
     * @param _karmaToken Address of the Karma token contract.
     */
    function initialize(address _owner, address _karmaToken) external initializer {
        if (_owner == address(0) || _karmaToken == address(0)) {
            revert SimpleKarmaDistributor__InvalidAddress();
        }

        __UUPSUpgradeable_init();
        __AccessControl_init();

        karmaToken = IERC20(_karmaToken);

        _grantRole(DEFAULT_ADMIN_ROLE, _owner);
    }

    /**
     * @notice Allows the admin to set the rewards supplier address.
     * @param _rewardsSupplier The new rewards supplier address.
     */
    function setRewardsSupplier(address _rewardsSupplier) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (_rewardsSupplier == address(0)) {
            revert SimpleKarmaDistributor__InvalidAddress();
        }
        rewardsSupplier = _rewardsSupplier;
        emit RewardsSupplierSet(_rewardsSupplier);
    }

    /**
     * @notice Supplies rewards to this distributor.
     * @dev The Karma token should mint the `amount` before calling this.
     * @param amount Amount of Karma supplied.
     */
    function setReward(uint256 amount, uint256) external onlyRewardsSupplier {
        if (amount == 0) {
            return;
        }

        availableSupply += amount;
        emit RewardSet(amount);
    }

    /**
     * @notice Mints virtual rewards for the provided account.
     * @dev Callable by admins and operators. Requires sufficient available supply.
     * @param account Address that will receive the virtual rewards.
     * @param amount Amount of Karma to mint virtually.
     */
    function mint(address account, uint256 amount) external onlyAdminOrOperator {
        if (account == address(0)) {
            revert SimpleKarmaDistributor__InvalidAddress();
        }

        if (amount == 0) {
            return;
        }

        if (availableSupply < amount) {
            revert SimpleKarmaDistributor__InsufficientAvailableSupply();
        }

        balances[account] += amount;
        mintedSupply += amount;
        availableSupply -= amount;

        emit RewardMinted(account, amount);
    }

    /**
     * @notice Allows anyone to redeem the virtual rewards for an account.
     * @param account Address whose rewards should be redeemed.
     * @return The amount of Karma redeemed.
     */
    function redeemRewards(address account) external returns (uint256) {
        uint256 amount = balances[account];
        if (amount == 0) {
            return 0;
        }

        balances[account] = 0;
        mintedSupply -= amount;

        karmaToken.safeTransfer(account, amount);

        emit RewardsRedeemed(account, amount);
        return amount;
    }

    /**
     * @notice Returns the total supply of virtual rewards (minted).
     */
    function totalRewardsSupply() external view returns (uint256) {
        return mintedSupply;
    }

    /**
     * @notice Returns the virtual rewards balance for an account.
     * @param account Account to query.
     */
    function rewardsBalanceOf(address account) external view returns (uint256) {
        return balances[account];
    }

    /**
     * @notice Returns the virtual rewards balance for an account.
     * @param user Address to query.
     */
    function rewardsBalanceOfAccount(address user) external view returns (uint256) {
        return balances[user];
    }

    /**
     * @inheritdoc UUPSUpgradeable
     */
    function _authorizeUpgrade(address) internal view override {
        if (!hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert SimpleKarmaDistributor__Unauthorized();
        }
    }
}
