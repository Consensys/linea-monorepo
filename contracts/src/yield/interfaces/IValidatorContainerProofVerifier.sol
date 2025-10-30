// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

/**
 * @title Interface for ValidatorContainerProofVerifier.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IValidatorContainerProofVerifier {
  /**
   * @notice Input for validator container proof verification
   * @custom:proof array of merkle proofs from Validator container root to Beacon block root
   * @custom:validatorIndex of validator in CL state tree
   * @custom:effectiveBalance of validator in CL state tree
   * @custom:childBlockTimestamp of EL block that has parent block beacon root in BEACON_ROOTS contract
   * @custom:slot of the beacon block for which the proof is generated
   * @custom:proposerIndex of the beacon block for which the proof is generated
   * @custom:activationEpoch of validator in CL state tree
   * @custom:activationEligibilityEpoch of validator in CL state tree
   */
  struct ValidatorContainerWitness {
    bytes32[] proof;
    uint256 validatorIndex;
    uint64 effectiveBalance;
    uint64 childBlockTimestamp;
    uint64 slot;
    uint64 proposerIndex;
    uint64 activationEpoch;
    uint64 activationEligibilityEpoch;
  }

  /// @notice Thrown when the slot and proposer index branch does not align with the supplied Merkle proof.
  error InvalidSlot();

  /// @notice Thrown when no beacon block root is found for the supplied child block timestamp in the EIP-4788 contract.
  error RootNotFound();

  /// @notice Thrown when the validator has not been active for the full `SHARD_COMMITTEE_PERIOD`.
  error ValidatorNotActiveForLongEnough();

  /**
   * @notice validates proof of active validator container in CL against Beacon block root
   * @param _witness object containing user input passed as calldata
   * @param _pubkey of validator to verify proof for.
   * @param _withdrawalCredentials to verify proof with
   * @dev reverts with `InvalidProof` when provided input cannot be proven to Beacon block root
   */
  function verifyActiveValidatorContainer(
    ValidatorContainerWitness calldata _witness,
    bytes memory _pubkey,
    bytes32 _withdrawalCredentials
  ) external view;
}
