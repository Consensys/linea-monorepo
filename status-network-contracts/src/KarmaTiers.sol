// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title KarmaTiers
 * @dev Manages tier system based on Karma token balance thresholds
 * @notice This contract allows efficient tier lookup for L2 nodes and tier management
 */
contract KarmaTiers is Ownable {
    /// @notice Emitted when a new tier is added
    event TierAdded(uint8 indexed tierId, string name, uint256 minKarma, uint256 maxKarma, uint32 txPerEpoch);
    /// @notice Emitted when a tier is updated
    event TierUpdated(uint8 indexed tierId, string name, uint256 minKarma, uint256 maxKarma, uint32 txPerEpoch);
    /// @notice Emitted when a tier is deactivated
    event TierDeactivated(uint8 indexed tierId);
    /// @notice Emitted when a tier is activated
    event TierActivated(uint8 indexed tierId);

    /// @notice Emitted when a transaction amount is invalid
    error InvalidTxAmount();
    /// @notice Emitted when a tier name is empty
    error EmptyTierName();
    /// @notice Emitted when a tier is not found
    error TierNotFound();
    /// @notice Emitted when a tier name exceeds maximum length
    error TierNameTooLong(uint256 nameLength, uint256 maxLength);
    /// @notice Emitted when a new tier overlaps with an existing one
    error OverlappingTiers(uint8 existingTierId, uint256 newMinKarma, uint256 newMaxKarma);
    /// @notice Emitted when a tier's minKarma is greater than or equal to maxKarma
    error InvalidTierRange(uint256 minKarma, uint256 maxKarma);

    struct Tier {
        uint256 minKarma;
        uint256 maxKarma;
        string name;
        uint32 txPerEpoch;
        bool active;
    }

    modifier onlyValidTierId(uint8 tierId) {
        if (tierId == 0 || tierId > currentTierId) {
            revert TierNotFound();
        }
        _;
    }

    /*//////////////////////////////////////////////////////////////////////////
                                  CONSTANTS
    //////////////////////////////////////////////////////////////////////////*/

    /// @notice Maximum length for tier names
    uint256 public constant MAX_TIER_NAME_LENGTH = 32;

    /*//////////////////////////////////////////////////////////////////////////
                                  STATE VARIABLES
    //////////////////////////////////////////////////////////////////////////*/

    /// @notice Mapping of tier IDs to Tier structs
    mapping(uint8 id => Tier tier) public tiers;
    /// @notice Current tier ID, incremented with each new tier added
    uint8 public currentTierId;

    /*//////////////////////////////////////////////////////////////////////////
                                     CONSTRUCTOR
    //////////////////////////////////////////////////////////////////////////*/

    constructor() {
        transferOwnership(msg.sender);
    }

    /*//////////////////////////////////////////////////////////////////////////
                           USER-FACING FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    /**
     * @dev Add a new tier to the system
     * @param name The name of the tier
     * @param minKarma Minimum Karma required for this tier
     * @param maxKarma Maximum Karma for this tier (0 for unlimited)
     */
    function addTier(string calldata name, uint256 minKarma, uint256 maxKarma, uint32 txPerEpoch) external onlyOwner {
        if (bytes(name).length == 0) revert EmptyTierName();
        if (maxKarma != 0 && maxKarma <= minKarma) revert InvalidTierRange(minKarma, maxKarma);
        if (txPerEpoch == 0) revert InvalidTxAmount();

        // Check for overlaps with existing tiers
        _validateNoOverlap(minKarma, maxKarma, type(uint8).max);
        _validateTierName(name);

        currentTierId++;

        tiers[currentTierId] =
            Tier({ minKarma: minKarma, maxKarma: maxKarma, name: name, active: true, txPerEpoch: txPerEpoch });

        emit TierAdded(currentTierId, name, minKarma, maxKarma, txPerEpoch);
    }

    /**
     * @dev Update an existing tier
     * @param name The name of the tier to update
     * @param newMinKarma New minimum Karma requirement
     * @param newMaxKarma New maximum Karma (0 for unlimited)
     */
    function updateTier(
        uint8 tierId,
        string calldata name,
        uint256 newMinKarma,
        uint256 newMaxKarma,
        uint32 newTxPerEpoch
    )
        external
        onlyOwner
        onlyValidTierId(tierId)
    {
        if (newMaxKarma != 0 && newMaxKarma <= newMinKarma) revert InvalidTierRange(newMinKarma, newMaxKarma);
        if (newTxPerEpoch == 0) revert InvalidTxAmount();

        // Check for overlaps with other tiers (excluding the one being updated)
        _validateNoOverlap(newMinKarma, newMaxKarma, tierId);
        _validateTierName(name);

        tiers[tierId].name = name;
        tiers[tierId].minKarma = newMinKarma;
        tiers[tierId].maxKarma = newMaxKarma;
        tiers[tierId].txPerEpoch = newTxPerEpoch;

        emit TierUpdated(tierId, name, newMinKarma, newMaxKarma, newTxPerEpoch);
    }

    /**
     * @dev Deactivate a tier (keeps it in storage but marks as inactive)
     * @param tierId The ID of the tier to deactivate
     */
    function deactivateTier(uint8 tierId) external onlyOwner onlyValidTierId(tierId) {
        tiers[tierId].active = false;
        emit TierDeactivated(tierId);
    }

    /**
     * @dev Reactivate a tier
     * @param tierId The ID of the tier to reactivate
     */
    function activateTier(uint8 tierId) external onlyOwner onlyValidTierId(tierId) {
        tiers[tierId].active = true;
        emit TierActivated(tierId);
    }

    /*//////////////////////////////////////////////////////////////////////////
                           INTERNAL FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    /**
     * @dev Validate tier name length and content
     * @param name The tier name to validate
     */
    function _validateTierName(string calldata name) internal pure {
        bytes memory nameBytes = bytes(name);
        if (nameBytes.length == 0) revert EmptyTierName();
        if (nameBytes.length > MAX_TIER_NAME_LENGTH) {
            revert TierNameTooLong(nameBytes.length, MAX_TIER_NAME_LENGTH);
        }
    }

    /*//////////////////////////////////////////////////////////////////////////
                           VIEW FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    /**
     * @notice Get tier by karma balance.
     * @dev This function returns the highest tier ID that the user qualifies for based on their karma balance.
     * @param karmaBalance The karma balance to check
     * @return tierId The tier id that matches the karma balance
     */
    function getTierIdByKarmaBalance(uint256 karmaBalance) external view returns (uint8) {
        uint8 bestTierId = 0;
        uint256 bestMinKarma = 0;

        for (uint8 i = 1; i <= currentTierId; i++) {
            Tier memory currentTier = tiers[i];
            if (!currentTier.active) continue;

            // Check if user meets the minimum requirement for this tier
            if (karmaBalance >= currentTier.minKarma) {
                // Only update if this tier has a higher minKarma requirement
                if (currentTier.minKarma > bestMinKarma) {
                    bestTierId = i;
                    bestMinKarma = currentTier.minKarma;
                }
            }
        }
        return bestTierId;
    }

    /**
     * @dev Get tier count
     * @return count Total number of tiers (including inactive)
     */
    function getTierCount() external view returns (uint256 count) {
        return currentTierId;
    }

    /**
     * @dev Get tier by id
     * @param tierId The ID of the tier to retrieve
     * @return tier The tier information
     */
    function getTierById(uint8 tierId) external view onlyValidTierId(tierId) returns (Tier memory tier) {
        return tiers[tierId];
    }

    /**
     * @dev Internal function to validate no overlap exists
     * @param minKarma Minimum Karma for the tier
     * @param maxKarma Maximum Karma for the tier (0 = unlimited)
     */
    function _validateNoOverlap(uint256 minKarma, uint256 maxKarma, uint8 excludeTierId) internal view {
        for (uint8 i = 1; i <= currentTierId; i++) {
            if (i == excludeTierId || !tiers[i].active) continue;

            Tier memory existingTier = tiers[i];
            uint256 existingMax = existingTier.maxKarma == 0 ? type(uint256).max : existingTier.maxKarma;
            uint256 newMax = maxKarma == 0 ? type(uint256).max : maxKarma;

            // Check for overlap using: NOT (no overlap) = overlap
            // No overlap means: newMax < existingMin OR newMin > existingMax
            if (!(newMax < existingTier.minKarma || minKarma > existingMax)) {
                revert OverlappingTiers(i, minKarma, maxKarma);
            }
        }
    }
}
