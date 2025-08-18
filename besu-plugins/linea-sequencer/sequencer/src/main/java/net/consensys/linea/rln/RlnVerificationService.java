/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package net.consensys.linea.rln;

/**
 * Service interface for RLN (Rate Limiting Nullifier) proof verification.
 *
 * <p>This abstraction allows for different implementations:
 *
 * <ul>
 *   <li>Production: JNI-based implementation using Rust cryptography
 *   <li>Testing: Mock implementation for unit tests
 *   <li>Fallback: Alternative verification backends
 * </ul>
 */
public interface RlnVerificationService {

  /** Represents the result of parsing and verifying an RLN proof. */
  record RlnProofData(
      String shareX, String shareY, String epoch, String root, String nullifier, boolean isValid) {}

  /** Exception thrown when RLN verification operations fail. */
  class RlnVerificationException extends Exception {
    public RlnVerificationException(String message) {
      super(message);
    }

    public RlnVerificationException(String message, Throwable cause) {
      super(message, cause);
    }
  }

  /**
   * Verifies an RLN proof with explicit public inputs.
   *
   * @param verifyingKeyBytes The serialized verifying key
   * @param proofBytes The serialized proof
   * @param publicInputs Array of public input hex strings: [share_x, share_y, epoch, root,
   *     nullifier]
   * @return true if the proof is valid, false otherwise
   * @throws RlnVerificationException if verification cannot be performed
   */
  boolean verifyRlnProof(byte[] verifyingKeyBytes, byte[] proofBytes, String[] publicInputs)
      throws RlnVerificationException;

  /**
   * Parses and verifies an RLN proof from combined format.
   *
   * @param verifyingKeyBytes The serialized verifying key
   * @param combinedProofBytes The combined proof data (proof + proof values)
   * @param currentEpochHex The current epoch identifier as hex string
   * @return RlnProofData containing extracted values and verification result
   * @throws RlnVerificationException if parsing or verification fails
   */
  RlnProofData parseAndVerifyRlnProof(
      byte[] verifyingKeyBytes, byte[] combinedProofBytes, String currentEpochHex)
      throws RlnVerificationException;

  /**
   * Checks if the verification service is available and ready to use.
   *
   * @return true if the service is available, false otherwise
   */
  boolean isAvailable();

  /**
   * Gets a human-readable description of this verification service implementation.
   *
   * @return description string
   */
  String getImplementationInfo();
}
