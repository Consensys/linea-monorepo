// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.30;

import { IL1MessageService } from "../../messaging/l1/interfaces/IL1MessageService.sol";

/**
 * @title Interface for the L1 Linea Token Burner Contract.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IL1LineaTokenBurner {
  function claimMessageWithProof(IL1MessageService.ClaimMessageWithProofParams calldata _params) external;
}
