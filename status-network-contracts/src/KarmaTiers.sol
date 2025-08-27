// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title KarmaTiers
 * @dev Manages tier system based on Karma token balance thresholds
 * @notice This contract allows efficient tier lookup for L2 nodes and tier management
 */
contract KarmaTiers is Ownable {
    /// @notice Emitted when a tier list is updated
    event TiersUpdated();
    /// @notice Emitted when a transaction amount is invalid

    error InvalidTxAmount();
    /// @notice Emitted when a tier name is empty
    error EmptyTierName();
    /// @notice Emitted when a tier array is empty
    error EmptyTiersArray();
    /// @notice Emitted when a tier is not found
    error TierNotFound();
    /// @notice Emitted when a tier name exceeds maximum length
    error TierNameTooLong(uint256 nameLength, uint256 maxLength);
    /// @notice Emitted when tiers are not contiguous
    error NonContiguousTiers(uint8 index, uint256 expectedMinKarma, uint256 actualMinKarma);
    /// @notice Emitted when a tier's minKarma is greater than or equal to maxKarma
    error InvalidTierRange(uint256 minKarma, uint256 maxKarma);

    struct Tier {
        uint256 minKarma;
        uint256 maxKarma;
        string name;
        uint32 txPerEpoch;
    }

    modifier onlyValidTierId(uint8 tierId) {
        if (tierId >= tiers.length) {
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

    /// @notice Array of tiers
    Tier[] public tiers;

    /*//////////////////////////////////////////////////////////////////////////
                                     CONSTRUCTOR
    //////////////////////////////////////////////////////////////////////////*/

    constructor() {
        transferOwnership(msg.sender);
    }

    /*//////////////////////////////////////////////////////////////////////////
                           USER-FACING FUNCTIONS
    //////////////////////////////////////////////////////////////////////////*/

    function updateTiers(Tier[] calldata newTiers) external onlyOwner {
        if (newTiers.length == 0) {
            revert EmptyTiersArray();
        }
        // Ensure the first tier starts at minKarma = 0
        if (newTiers[0].minKarma != 0) {
            revert NonContiguousTiers(0, 0, newTiers[0].minKarma);
        }

        delete tiers; // Clear existing tiers

        uint256 lastMaxKarma = 0;
        for (uint8 i = 0; i < newTiers.length; i++) {
            Tier calldata input = newTiers[i];

            _validateTierName(input.name);
            if (input.maxKarma <= input.minKarma) {
                revert InvalidTierRange(input.minKarma, input.maxKarma);
            }

            if (i > 0) {
                uint256 expectedMinKarma = lastMaxKarma + 1;
                if (input.minKarma != expectedMinKarma) {
                    revert NonContiguousTiers(i, expectedMinKarma, input.minKarma);
                }
            }
            lastMaxKarma = input.maxKarma;
            tiers.push(input);
        }

        emit TiersUpdated();
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
        for (uint8 i = 0; i < tiers.length; i++) {
            // Check if user meets the minimum requirement for this tier
            if (karmaBalance < tiers[i].minKarma) {
                // Can't run into underflow here because `karmaBalance == 0 => karmaBalance < tiers[0].minKarma`
                return i - 1; // Return the previous tier if this one is not met
            }
        }
        return uint8(tiers.length - 1); // If all tiers are met, return the highest tier
    }

    /**
     * @dev Get tier count
     * @return count Total number of tiers (including inactive)
     */
    function getTierCount() external view returns (uint256 count) {
        return tiers.length;
    }

    /**
     * @dev Get tier by id
     * @param tierId The ID of the tier to retrieve
     * @return tier The tier information
     */
    function getTierById(uint8 tierId) external view onlyValidTierId(tierId) returns (Tier memory tier) {
        return tiers[tierId];
    }
}
