// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

/**
 * @title TrustedCodehashAccess
 * @author Ricardo Guilherme Schmidt <ricardo3@status.im>
 * @notice Ensures that only specific contract bytecode hashes are trusted to
 *         interact with the functions using the `onlyTrustedCodehash` modifier.
 */
interface ITrustedCodehashAccess {
    error TrustedCodehashAccess__UnauthorizedCodehash();

    event TrustedCodehashUpdated(bytes32 indexed codehash, bool trusted);

    /**
     * @notice Allows the owner to set or update the trust status for a contract's codehash.
     * @dev Emits the `TrustedCodehashUpdated` event whenever a codehash is updated.
     * @param _codehash The bytecode hash of the contract.
     * @param _trusted Boolean flag to designate the contract as trusted or not.
     */
    function setTrustedCodehash(bytes32 _codehash, bool _trusted) external;

    /**
     * @notice Checks if a contract's codehash is trusted to interact with protected functions.
     * @param _codehash The bytecode hash of the contract.
     * @return bool True if the codehash is trusted, false otherwise.
     */
    function isTrustedCodehash(bytes32 _codehash) external view returns (bool);
}
