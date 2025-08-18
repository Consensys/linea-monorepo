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

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

/**
 * Simple integration tests to verify RLN service availability and basic functionality.
 */
class RlnServiceIntegrationTest {

  private JniRlnVerificationService service;

  @BeforeEach
  void setUp() {
    service = new JniRlnVerificationService();
  }

  @Test
  void testServiceInitialization() {
    assertThat(service).isNotNull();
    assertThat(service.getImplementationInfo()).isNotNull();
    assertThat(service.getImplementationInfo()).contains("JNI-based RLN verification service");
  }

  @Test
  void testServiceAvailabilityCheck() {
    boolean isAvailable = service.isAvailable();
    String info = service.getImplementationInfo();
    
    if (isAvailable) {
      assertThat(info).contains("native Rust implementation");
      assertThat(info).doesNotContain("UNAVAILABLE");
    } else {
      assertThat(info).contains("UNAVAILABLE");
    }
  }

  @Test
  void testErrorHandlingWhenServiceUnavailable() {
    if (!service.isAvailable()) {
      try {
        service.verifyRlnProof(new byte[0], new byte[0], new String[5]);
        assertThat(false).as("Expected exception when service unavailable").isTrue();
      } catch (RlnVerificationService.RlnVerificationException e) {
        assertThat(e.getMessage()).contains("JNI RLN verification service is not available");
      }
    }
  }
}