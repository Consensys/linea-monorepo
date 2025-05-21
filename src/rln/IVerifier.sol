// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity ^0.8.17;

interface IVerifier {
    function verifyProof(
        uint256[2] calldata a,
        uint256[2][2] calldata b,
        uint256[2] calldata c,
        uint256[2] calldata input
    )
        external
        view
        returns (bool);
}
