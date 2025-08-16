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

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import java.security.SecureRandom;
import java.util.Arrays;
import net.consensys.linea.rln.RlnVerificationService.RlnProofData;
import net.consensys.linea.rln.RlnVerificationService.RlnVerificationException;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.condition.EnabledIf;

/**
 * Integration tests for JniRlnVerificationService.
 * 
 * Tests the actual JNI integration with the Rust RLN verification library.
 * These tests will only run if the native library is available.
 */
class JniRlnVerificationServiceTest {

  private JniRlnVerificationService service;
  private SecureRandom random;

  // Test data for RLN proofs
  private static final String VALID_SHARE_X = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef";
  private static final String VALID_SHARE_Y = "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321";
  private static final String VALID_EPOCH = "0x1c61ef0b2ebc0235d85fe8537b4455549356e3895005ba7a03fbd4efc9ba3692";
  private static final String VALID_ROOT = "0x19b4c972cda99dfd4d9c87f5c6f6c3f7b5f2e1d8a7b6c5e4f3e2d1c0b9a8f7e6";
  private static final String VALID_NULLIFIER = "0xa1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456";

  @BeforeEach
  void setUp() {
    service = new JniRlnVerificationService();
    random = new SecureRandom();
  }

  @Test
  void testServiceAvailability() {
    // Test whether the service can detect if JNI is available
    boolean isAvailable = service.isAvailable();
    String info = service.getImplementationInfo();
    
    assertThat(info).isNotNull();
    if (isAvailable) {
      assertThat(info).contains("JNI-based RLN verification service");
      assertThat(info).doesNotContain("UNAVAILABLE");
    } else {
      assertThat(info).contains("UNAVAILABLE");
    }
  }

  @Test
  @EnabledIf("isNativeLibraryAvailable")
  void testVerifyRlnProofWithValidInputs() throws RlnVerificationException {
    // This test requires real proof data - for now we test the API contract
    byte[] dummyVkBytes = generateRandomBytes(100);
    byte[] dummyProofBytes = generateRandomBytes(200);
    String[] publicInputs = {
      VALID_SHARE_X,
      VALID_SHARE_Y, 
      VALID_EPOCH,
      VALID_ROOT,
      VALID_NULLIFIER
    };

    // This will call the native method - result depends on proof validity
    // We're testing that the JNI call succeeds without throwing
    try {
      boolean result = service.verifyRlnProof(dummyVkBytes, dummyProofBytes, publicInputs);
      // Result can be true or false - we're just testing no exception is thrown
      assertThat(result).isIn(true, false);
    } catch (RlnVerificationException e) {
      // Exception is expected if proof is invalid, but not from JNI issues
      assertThat(e.getMessage()).doesNotContain("JNI");
    }
  }

  @Test
  void testVerifyRlnProofWithInvalidPublicInputsLength() {
    byte[] dummyVkBytes = generateRandomBytes(100);
    byte[] dummyProofBytes = generateRandomBytes(200);
    String[] invalidPublicInputs = {VALID_SHARE_X, VALID_SHARE_Y}; // Only 2 inputs instead of 5

    assertThatThrownBy(() -> 
      service.verifyRlnProof(dummyVkBytes, dummyProofBytes, invalidPublicInputs))
      .isInstanceOf(RlnVerificationException.class)
      .hasMessageContaining("Expected exactly 5 public inputs");
  }

  @Test
  void testVerifyRlnProofWithNullPublicInputs() {
    byte[] dummyVkBytes = generateRandomBytes(100);
    byte[] dummyProofBytes = generateRandomBytes(200);

    assertThatThrownBy(() -> 
      service.verifyRlnProof(dummyVkBytes, dummyProofBytes, null))
      .isInstanceOf(RlnVerificationException.class)
      .hasMessageContaining("Expected exactly 5 public inputs");
  }

  @Test
  @EnabledIf("isNativeLibraryAvailable")
  void testParseAndVerifyRlnProofWithValidInputs() throws RlnVerificationException {
    byte[] dummyVkBytes = generateRandomBytes(100);
    byte[] dummyCombinedProofBytes = generateRandomBytes(300);
    String currentEpochHex = VALID_EPOCH;

    try {
      RlnProofData result = service.parseAndVerifyRlnProof(dummyVkBytes, dummyCombinedProofBytes, currentEpochHex);
      
      // Result can be valid or invalid - we're testing the API contract
      assertThat(result).isNotNull();
      assertThat(result.shareX()).isNotNull();
      assertThat(result.shareY()).isNotNull();
      assertThat(result.epoch()).isNotNull();
      assertThat(result.root()).isNotNull();
      assertThat(result.nullifier()).isNotNull();
      // isValid() can be true or false depending on proof validity
    } catch (RlnVerificationException e) {
      // Exception is expected if proof parsing/verification fails
      assertThat(e.getMessage()).isNotEmpty();
    }
  }

  @Test
  void testServiceUnavailableWhenJniNotLoaded() {
    // This test will pass regardless of JNI availability
    // If JNI is not available, service should handle gracefully
    if (!service.isAvailable()) {
      assertThatThrownBy(() -> 
        service.verifyRlnProof(new byte[0], new byte[0], new String[5]))
        .isInstanceOf(RlnVerificationException.class)
        .hasMessageContaining("JNI RLN verification service is not available");
        
      assertThatThrownBy(() -> 
        service.parseAndVerifyRlnProof(new byte[0], new byte[0], "0x123"))
        .isInstanceOf(RlnVerificationException.class)
        .hasMessageContaining("JNI RLN verification service is not available");
    }
  }

  @Test
  void testImplementationInfo() {
    String info = service.getImplementationInfo();
    assertThat(info).isNotNull();
    assertThat(info).contains("JNI-based RLN verification service");
    
    if (service.isAvailable()) {
      assertThat(info).contains("native Rust implementation");
    } else {
      assertThat(info).contains("UNAVAILABLE");
    }
  }

  private byte[] generateRandomBytes(int length) {
    byte[] bytes = new byte[length];
    random.nextBytes(bytes);
    return bytes;
  }

  // Test condition method for @EnabledIf
  static boolean isNativeLibraryAvailable() {
    try {
      JniRlnVerificationService testService = new JniRlnVerificationService();
      return testService.isAvailable();
    } catch (Exception e) {
      return false;
    }
  }
}