// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.19;

/**
 * @title Contract to manage cross-chain messaging on L1 and rollup proving
 * @author ConsenSys Software Inc.
 */
interface IPlonkVerifier {
  /**
   * @notice Interface for verifier contracts.
   * @param _proof The proof used to verify.
   * @param _public_inputs The computed public inputs for the proof verification.
   */
  function Verify(bytes calldata _proof, uint256[] calldata _public_inputs) external returns (bool);
}
