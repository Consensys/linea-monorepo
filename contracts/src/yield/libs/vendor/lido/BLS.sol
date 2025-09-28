// SPDX-License-Identifier: MIT

// See contracts/COMPILERS.md
// solhint-disable-next-line lido/fixed-compiler-version
pragma solidity ^0.8.25;

/**
 * @notice Modified & stripped BLS Lib to support ETH beacon spec for validator deposit message verification.
 * @author Lido
 * @author Solady (https://github.com/Vectorized/solady/blob/dcdfab80f4e6cb9ac35c91610b2a2ec42689ec79/src/utils/ext/ithaca/BLS.sol)
 * @author Ithaca (https://github.com/ithacaxyz/odyssey-examples/blob/main/chapter1/contracts/src/libraries/BLS.sol)
 */
// solhint-disable contract-name-capwords
library BLS12_381 {
    /*´:°•.°+.*•´.*:˚.°*.˚•´.°:°•.°•.*•´.*:˚.°*.˚•´.°:°•.°+.*•´.*:*/
    /*                    PRECOMPILE ADDRESSES                    */
    /*.•°:°.´+˚.*°.˚:*.´•*.+°.•°:´*.´•*.•°.•°:°.´:•˚°.*°.˚:*.´+°.•*/
    /// @dev SHA256 precompile address.
    address internal constant SHA256 = 0x0000000000000000000000000000000000000002;

    /*´:°•.°+.*•´.*:˚.°*.˚•´.°:°•.°•.*•´.*:˚.°*.˚•´.°:°•.°+.*•´.*:*/
    /*                        CUSTOM ERRORS                       */
    /*.•°:°.´+˚.*°.˚:*.´•*.+°.•°:´*.´•*.•°.•°:°.´:•˚°.*°.˚:*.´+°.•*/

    // A custom error for each precompile helps us in debugging which precompile has failed.

    /// @dev provided pubkey length is not 48
    error InvalidPubkeyLength();

    /*´:°•.°+.*•´.*:˚.°*.˚•´.°:°•.°•.*•´.*:˚.°*.˚•´.°:°•.°+.*•´.*:*/
    /*                         UTILITY                            */
    /*.•°:°.´+˚.*°.˚:*.´•*.+°.•°:´*.´•*.•°.•°:°.´:•˚°.*°.˚:*.´+°.•*/

    /// @notice Extracted part from `SSZ.verifyProof` for hashing two leaves
    /// @dev Combines 2 bytes32 in 64 bytes input for sha256 precompile
    function sha256Pair(bytes32 left, bytes32 right) internal view returns (bytes32 result) {
        /// @solidity memory-safe-assembly
        assembly {
            // Store `left` at memory position 0x00
            mstore(0x00, left)
            // Store `right` at memory position 0x20
            mstore(0x20, right)

            // Call SHA-256 precompile (0x02) with 64-byte input at memory 0x00
            let success := staticcall(gas(), 0x02, 0x00, 0x40, 0x00, 0x20)
            if iszero(success) {
                revert(0, 0)
            }

            // Load the resulting hash from memory
            result := mload(0x00)
        }
    }

    /// @notice Extracted and modified part from `SSZ.hashTreeRoot` for hashing validator pubkey from calldata
    /// @dev Reverts if `pubkey` length is not 48
    function pubkeyRoot(bytes memory pubkey) internal view returns (bytes32 _pubkeyRoot) {
        if (pubkey.length != 48) revert InvalidPubkeyLength();

        /// @solidity memory-safe-assembly
        assembly {
            // Dynamic data types such as bytes are stored at the specified offset.
            let offset := mload(pubkey)
            // Copy the pubkey to the scratch space.
            mcopy(0x00, add(offset, 32), 48)
            // Clear the last 16 bytes.
            mcopy(48, 0x60, 16)

            // Call the SHA-256 precompile (0x02) with the 64-byte input
            if iszero(staticcall(gas(), 0x02, 0x00, 0x40, 0x00, 0x20)) {
                revert(0, 0)
            }

            // Load the resulting SHA-256 hash
            _pubkeyRoot := mload(0x00)
        }
    }
}
