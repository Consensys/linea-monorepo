// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ITrustedCodehashAccess } from "./interfaces/ITrustedCodehashAccess.sol";

/**
 * @title TrustedCodehashAccess
 * @author Ricardo Guilherme Schmidt <ricardo3@status.im>
 * @notice Ensures that only specific contract bytecode hashes are trusted to
 *         interact with the functions using the `onlyTrustedCodehash` modifier.
 * @dev This contract is used to restrict access to functions based on the codehash of the caller.
 */
abstract contract TrustedCodehashAccess is ITrustedCodehashAccess, OwnableUpgradeable {
    /// @notice Whidelisted codehashes.
    mapping(bytes32 codehash => bool permission) private trustedCodehashes;
    /// @notice Gap for upgrade safety.
    uint256[10] private __gap;

    /**
     * @notice Restricts access based on the codehash of the caller.
     *         Only contracts with trusted codehashes can execute functions using this modifier.
     */
    modifier onlyTrustedCodehash() {
        bytes32 codehash = msg.sender.codehash;
        if (!trustedCodehashes[codehash]) {
            revert TrustedCodehashAccess__UnauthorizedCodehash();
        }
        _;
    }

    /**
     * @notice Initializes the contract with the provided owner address.
     * @dev This function is called only once during the contract deployment.
     * @param _initialOwner The address of the owner.
     */
    function __TrustedCodehashAccess_init(address _initialOwner) public onlyInitializing {
        _transferOwnership(_initialOwner);
    }

    /**
     * @notice Allows the owner to set or update the trust status for a contract's codehash.
     * @dev Emits the `TrustedCodehashUpdated` event whenever a codehash is updated.
     * @param _codehash The bytecode hash of the contract.
     * @param _trusted Boolean flag to designate the contract as trusted or not.
     */
    function setTrustedCodehash(bytes32 _codehash, bool _trusted) external onlyOwner {
        trustedCodehashes[_codehash] = _trusted;
        emit TrustedCodehashUpdated(_codehash, _trusted);
    }

    /**
     * @notice Checks if a contract's codehash is trusted to interact with protected functions.
     * @param _codehash The bytecode hash of the contract.
     * @return bool True if the codehash is trusted, false otherwise.
     */
    function isTrustedCodehash(bytes32 _codehash) external view returns (bool) {
        return trustedCodehashes[_codehash];
    }
}
