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

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Mock implementation of RlnVerificationService for testing purposes.
 *
 * <p>This implementation provides configurable behavior for testing different scenarios without
 * requiring the native RLN library. It's particularly useful for:
 *
 * <ul>
 *   <li>Unit testing in environments without native libraries
 *   <li>Integration testing with predictable verification results
 *   <li>Testing error handling and edge cases
 * </ul>
 */
public class MockRlnVerificationService implements RlnVerificationService {
  private static final Logger LOG = LoggerFactory.getLogger(MockRlnVerificationService.class);

  private boolean shouldVerifySuccessfully = true;
  private boolean shouldThrowException = false;
  private String exceptionMessage = "Mock verification error";
  private RlnProofData mockProofData =
      new RlnProofData(
          "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", // shareX
          "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321", // shareY
          "0x0000000000000000000000000000000000000000000000000000000000000001", // epoch
          "0x2b1c4b4e1f1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e", // root
          "0x3a2b3c4d5e6f7890123456789abcdef0123456789abcdef0123456789abcdef0", // nullifier
          true // isValid
          );

  public MockRlnVerificationService() {
    LOG.info("Mock RLN verification service initialized");
  }

  @Override
  public boolean verifyRlnProof(byte[] verifyingKeyBytes, byte[] proofBytes, String[] publicInputs)
      throws RlnVerificationException {

    LOG.debug(
        "Mock verification called with {} public inputs",
        publicInputs != null ? publicInputs.length : "null");

    if (shouldThrowException) {
      throw new RlnVerificationException(exceptionMessage);
    }

    if (publicInputs == null || publicInputs.length != 5) {
      throw new RlnVerificationException(
          "Expected exactly 5 public inputs, got: "
              + (publicInputs == null ? "null" : publicInputs.length));
    }

    return shouldVerifySuccessfully;
  }

  @Override
  public RlnProofData parseAndVerifyRlnProof(
      byte[] verifyingKeyBytes, byte[] combinedProofBytes, String currentEpochHex)
      throws RlnVerificationException {

    LOG.debug("Mock parseAndVerifyRlnProof called with epoch: {}", currentEpochHex);

    if (shouldThrowException) {
      throw new RlnVerificationException(exceptionMessage);
    }

    // Create a copy of the mock data with the correct epoch
    return new RlnProofData(
        mockProofData.shareX(),
        mockProofData.shareY(),
        currentEpochHex, // Use the provided epoch
        mockProofData.root(),
        mockProofData.nullifier(),
        shouldVerifySuccessfully);
  }

  @Override
  public boolean isAvailable() {
    return true; // Mock is always available
  }

  @Override
  public String getImplementationInfo() {
    return "Mock RLN verification service (for testing only)";
  }

  // Configuration methods for testing

  /**
   * Configures whether verification should succeed or fail.
   *
   * @param shouldSucceed true if verification should succeed, false otherwise
   */
  public void setVerificationResult(boolean shouldSucceed) {
    this.shouldVerifySuccessfully = shouldSucceed;
    LOG.debug("Mock verification result set to: {}", shouldSucceed);
  }

  /**
   * Configures whether to throw an exception during verification.
   *
   * @param shouldThrow true if an exception should be thrown, false otherwise
   * @param message the exception message to use
   */
  public void setThrowException(boolean shouldThrow, String message) {
    this.shouldThrowException = shouldThrow;
    this.exceptionMessage = message;
    LOG.debug("Mock exception throwing set to: {} with message: {}", shouldThrow, message);
  }

  /**
   * Sets the mock proof data to return from parseAndVerifyRlnProof.
   *
   * @param proofData the mock proof data to use
   */
  public void setMockProofData(RlnProofData proofData) {
    this.mockProofData = proofData;
    LOG.debug("Mock proof data updated: {}", proofData);
  }

  /**
   * Resets the mock service to default behavior.
   *
   * <p>This method restores the mock to its initial state: - Verification returns true - No
   * exceptions are thrown - Default proof data is used
   */
  public void reset() {
    this.shouldVerifySuccessfully = true;
    this.shouldThrowException = false;
    this.exceptionMessage = "Mock verification error";
    this.mockProofData =
        new RlnProofData(
            "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", // shareX
            "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321", // shareY
            "0x0000000000000000000000000000000000000000000000000000000000000001", // epoch
            "0x2b1c4b4e1f1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e", // root
            "0x3a2b3c4d5e6f7890123456789abcdef0123456789abcdef0123456789abcdef0", // nullifier
            true // isValid
            );
    LOG.debug("Mock RLN verification service reset to defaults");
  }
}
