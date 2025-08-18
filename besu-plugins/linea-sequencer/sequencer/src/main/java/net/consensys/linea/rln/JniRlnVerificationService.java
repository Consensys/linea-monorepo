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

import net.consensys.linea.rln.jni.RlnBridge;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * JNI-based implementation of RlnVerificationService.
 *
 * <p>This implementation wraps the existing RlnBridge JNI calls and provides better error handling
 * and abstraction. It isolates the JNI complexity from the rest of the application.
 */
public class JniRlnVerificationService implements RlnVerificationService {
  private static final Logger LOG = LoggerFactory.getLogger(JniRlnVerificationService.class);

  private final boolean isAvailable;

  /**
   * Creates a new JNI-based RLN verification service.
   *
   * <p>The constructor tests if the native library is available and logs the result for debugging
   * purposes.
   */
  public JniRlnVerificationService() {
    boolean available = false;
    try {
      // Test if the native library is loaded and working
      // We'll try a simple call to see if JNI is working
      available = testNativeLibraryAvailability();
      LOG.info("JNI RLN verification service initialized successfully");
    } catch (UnsatisfiedLinkError e) {
      LOG.error(
          "JNI RLN verification service unavailable - native library not loaded: {}",
          e.getMessage());
    } catch (Exception e) {
      LOG.error(
          "JNI RLN verification service unavailable - initialization error: {}", e.getMessage(), e);
    }
    this.isAvailable = available;
  }

  /** Tests if the native library is available by attempting a simple JNI call. */
  private boolean testNativeLibraryAvailability() {
    try {
      // Try to load the RlnBridge class which triggers native library loading
      Class.forName("net.consensys.linea.rln.jni.RlnBridge");
      return true;
    } catch (ClassNotFoundException | UnsatisfiedLinkError e) {
      return false;
    }
  }

  @Override
  public boolean verifyRlnProof(byte[] verifyingKeyBytes, byte[] proofBytes, String[] publicInputs)
      throws RlnVerificationException {

    if (!isAvailable) {
      throw new RlnVerificationException("JNI RLN verification service is not available");
    }

    if (publicInputs == null || publicInputs.length != 5) {
      throw new RlnVerificationException(
          "Expected exactly 5 public inputs, got: "
              + (publicInputs == null ? "null" : publicInputs.length));
    }

    try {
      return RlnBridge.verifyRlnProof(verifyingKeyBytes, proofBytes, publicInputs);
    } catch (RuntimeException e) {
      throw new RlnVerificationException("Native RLN proof verification failed", e);
    }
  }

  @Override
  public RlnProofData parseAndVerifyRlnProof(
      byte[] verifyingKeyBytes, byte[] combinedProofBytes, String currentEpochHex)
      throws RlnVerificationException {

    if (!isAvailable) {
      throw new RlnVerificationException("JNI RLN verification service is not available");
    }

    try {
      String[] result =
          RlnBridge.parseAndVerifyRlnProof(verifyingKeyBytes, combinedProofBytes, currentEpochHex);

      if (result == null) {
        throw new RlnVerificationException("Native proof parsing returned null");
      }

      if (result.length != 6) {
        throw new RlnVerificationException("Expected 6 result values, got: " + result.length);
      }

      boolean isValid = "true".equals(result[5]);

      return new RlnProofData(
          result[0], // shareX
          result[1], // shareY
          result[2], // epoch
          result[3], // root
          result[4], // nullifier
          isValid);
    } catch (RuntimeException e) {
      throw new RlnVerificationException("Native RLN proof parsing and verification failed", e);
    }
  }

  @Override
  public boolean isAvailable() {
    return isAvailable;
  }

  @Override
  public String getImplementationInfo() {
    return isAvailable
        ? "JNI-based RLN verification service (native Rust implementation)"
        : "JNI-based RLN verification service (UNAVAILABLE - native library not loaded)";
  }
}
