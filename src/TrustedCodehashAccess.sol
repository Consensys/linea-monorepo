// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ITrustedCodehashAccess } from "./interfaces/ITrustedCodehashAccess.sol";
/**
 * @title TrustedCodehashAccess
 * @author Ricardo Guilherme Schmidt <ricardo3@status.im>
 * @notice Ensures that only specific contract bytecode hashes are trusted to
 *         interact with the functions using the `onlyTrustedCodehash` modifier.
 */

abstract contract TrustedCodehashAccess is ITrustedCodehashAccess, OwnableUpgradeable {
    mapping(bytes32 codehash => bool permission) private trustedCodehashes;

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

    function __TrustedCodehashAccess_init(address _initialOwner) public onlyInitializing {
        __Ownable_init(_initialOwner);
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
