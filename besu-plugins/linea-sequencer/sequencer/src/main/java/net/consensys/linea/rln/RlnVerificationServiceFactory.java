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
 * Factory for creating RLN verification service implementations.
 *
 * <p>This factory provides a centralized way to create the appropriate RLN verification service
 * based on the environment and configuration:
 *
 * <ul>
 *   <li>Production: Uses JNI-based native implementation
 *   <li>Testing: Uses mock implementation when native library is unavailable
 *   <li>Fallback: Provides graceful degradation options
 * </ul>
 */
public class RlnVerificationServiceFactory {
  private static final Logger LOG = LoggerFactory.getLogger(RlnVerificationServiceFactory.class);

  public enum ServiceType {
    /** Native JNI implementation using Rust cryptography */
    NATIVE,
    /** Mock implementation for testing */
    MOCK,
    /** Automatically choose best available implementation */
    AUTO
  }

  /**
   * Creates an RLN verification service based on the requested type.
   *
   * @param serviceType The type of service to create
   * @return The created service instance
   * @throws IllegalStateException if the requested service type cannot be created
   */
  public static RlnVerificationService create(ServiceType serviceType) {
    switch (serviceType) {
      case NATIVE:
        return createNativeService();
      case MOCK:
        return createMockService();
      case AUTO:
        return createAutoService();
      default:
        throw new IllegalArgumentException("Unknown service type: " + serviceType);
    }
  }

  /**
   * Creates an RLN verification service with automatic implementation selection.
   *
   * <p>This method tries to create a native service first, falling back to mock if the native
   * library is not available.
   *
   * @return The best available service implementation
   */
  public static RlnVerificationService createAutoService() {
    try {
      RlnVerificationService nativeService = createNativeService();
      if (nativeService.isAvailable()) {
        LOG.info("Using native RLN verification service");
        return nativeService;
      }
    } catch (Exception e) {
      LOG.warn("Native RLN verification service not available: {}", e.getMessage());
    }

    LOG.info("Falling back to mock RLN verification service");
    return createMockService();
  }

  /**
   * Creates a native JNI-based RLN verification service.
   *
   * @return Native service implementation
   * @throws IllegalStateException if native service cannot be created
   */
  private static RlnVerificationService createNativeService() {
    return new JniRlnVerificationService();
  }

  /**
   * Creates a mock RLN verification service for testing.
   *
   * @return Mock service implementation
   */
  private static RlnVerificationService createMockService() {
    return new MockRlnVerificationService();
  }
}
