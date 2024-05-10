// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.24;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { L1MessageServiceV1 } from "./messageService/l1/v1/L1MessageServiceV1.sol";
import { IZkEvmV2 } from "./interfaces/l1/IZkEvmV2.sol";
import { IPlonkVerifier } from "./interfaces/l1/IPlonkVerifier.sol";

/**
 * @title Contract to manage cross-chain messaging on L1 and rollup proving.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract ZkEvmV2 is Initializable, AccessControlUpgradeable, L1MessageServiceV1, IZkEvmV2 {
  uint256 internal constant MODULO_R = 21888242871839275222246405745257275088548364400416034343698204186575808495617;
  bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

  uint256 public currentTimestamp;
  uint256 public currentL2BlockNumber;

  mapping(uint256 blockNumber => bytes32 stateRootHash) public stateRootHashes;
  mapping(uint256 proofType => address verifierAddress) public verifiers;

  /// @dev Total contract storage is 54 slots with the gap below.
  /// @dev Keep 50 free storage slots for future implementation updates to avoid storage collision.
  uint256[50] private __gap;

  /**
   * @notice Verifies the proof with locally computed public inputs.
   * @dev If the verifier based on proof type is not found, it reverts with InvalidProofType.
   * @param _publicInputHash The full BlockData collection - block, transaction and log data.
   * @param _proofType The proof type to determine which verifier contract to use.
   * @param _proof The proof to be verified with the proof type verifier contract.
   * @param _parentStateRootHash The beginning roothash to start with.
   * @param _finalizedL2BlockNumber The final L2 block number being finalized.
   */
  function _verifyProof(
    uint256 _publicInputHash,
    uint256 _proofType,
    bytes calldata _proof,
    bytes32 _parentStateRootHash,
    uint256 _finalizedL2BlockNumber
  ) internal {
    uint256[] memory input = new uint256[](1);
    input[0] = _publicInputHash;

    address verifierToUse = verifiers[_proofType];

    if (verifierToUse == address(0)) {
      revert InvalidProofType();
    }

    bool success = IPlonkVerifier(verifierToUse).Verify(_proof, input);
    if (!success) {
      revert InvalidProof();
    }

    emit BlocksVerificationDone(
      _finalizedL2BlockNumber,
      _parentStateRootHash,
      stateRootHashes[_finalizedL2BlockNumber]
    );
  }
}
