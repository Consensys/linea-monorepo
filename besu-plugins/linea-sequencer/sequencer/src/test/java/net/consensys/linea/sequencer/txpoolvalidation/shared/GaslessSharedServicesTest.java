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
package net.consensys.linea.sequencer.txpoolvalidation.shared;

import static org.assertj.core.api.Assertions.assertThat;

import java.io.IOException;
import java.nio.file.Path;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

/**
 * Simple tests to verify gasless shared services work correctly together.
 */
class GaslessSharedServicesTest {

  @TempDir Path tempDir;

  private static final Address TEST_ADDRESS = Address.fromHexString("0x1234567890123456789012345678901234567890");
  private static final String TEST_NULLIFIER = "0xa1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456";
  private static final String TEST_EPOCH = "0x1c61ef0b2ebc0235d85fe8537b4455549356e3895005ba7a03fbd4efc9ba3692";

  private DenyListManager denyListManager;
  private NullifierTracker nullifierTracker;
  private KarmaServiceClient karmaServiceClient;

  @BeforeEach
  void setUp() throws IOException {
    Path denyListFile = tempDir.resolve("test_deny_list.txt");
    denyListManager = new DenyListManager("Test", denyListFile.toString(), 60, 0);
    nullifierTracker = new NullifierTracker("Test", 1000L, 1L);
    karmaServiceClient = new KarmaServiceClient("Test", "localhost", 8545, false, 5000);
  }

  @AfterEach
  void tearDown() throws Exception {
    if (denyListManager != null) {
      denyListManager.close();
    }
    if (nullifierTracker != null) {
      nullifierTracker.close();
    }
    if (karmaServiceClient != null) {
      karmaServiceClient.close();
    }
  }

  @Test
  void testServicesInitialization() {
    assertThat(denyListManager).isNotNull();
    assertThat(nullifierTracker).isNotNull();
    assertThat(karmaServiceClient).isNotNull();
    
    assertThat(denyListManager.size()).isEqualTo(0);
    assertThat(nullifierTracker.getStats().currentNullifiers()).isEqualTo(0);
    assertThat(karmaServiceClient.isAvailable()).isTrue();
  }

  @Test
  void testDenyListBasicOperations() {
    // Initially not denied
    assertThat(denyListManager.isDenied(TEST_ADDRESS)).isFalse();
    
    // Add to deny list
    boolean added = denyListManager.addToDenyList(TEST_ADDRESS);
    assertThat(added).isTrue();
    assertThat(denyListManager.isDenied(TEST_ADDRESS)).isTrue();
    assertThat(denyListManager.size()).isEqualTo(1);
    
    // Remove from deny list
    boolean removed = denyListManager.removeFromDenyList(TEST_ADDRESS);
    assertThat(removed).isTrue();
    assertThat(denyListManager.isDenied(TEST_ADDRESS)).isFalse();
    assertThat(denyListManager.size()).isEqualTo(0);
  }

  @Test
  void testNullifierTrackingBasics() {
    // First use should be allowed
    boolean isNew = nullifierTracker.checkAndMarkNullifier(TEST_NULLIFIER, TEST_EPOCH);
    assertThat(isNew).isTrue();
    
    // Reuse should be blocked
    boolean isReused = nullifierTracker.checkAndMarkNullifier(TEST_NULLIFIER, TEST_EPOCH);
    assertThat(isReused).isFalse();
    
    // Verify tracking
    assertThat(nullifierTracker.isNullifierUsed(TEST_NULLIFIER, TEST_EPOCH)).isTrue();
    
    NullifierTracker.NullifierStats stats = nullifierTracker.getStats();
    assertThat(stats.totalTracked()).isEqualTo(1);
    assertThat(stats.duplicateAttempts()).isEqualTo(1);
  }

  @Test
  void testKarmaServiceClientConfiguration() {
    assertThat(karmaServiceClient.isAvailable()).isTrue();
    
    // Test that service handles unavailable scenarios gracefully
    java.util.Optional<KarmaServiceClient.KarmaInfo> result = 
        karmaServiceClient.fetchKarmaInfo(TEST_ADDRESS);
    
    // Since no actual service is running, should return empty
    assertThat(result).isEmpty();
  }

  @Test
  void testServicesResourceCleanup() throws Exception {
    // Verify services can be closed without errors
    denyListManager.close();
    nullifierTracker.close();
    karmaServiceClient.close();
    
    // After closing, karma service should not be available
    assertThat(karmaServiceClient.isAvailable()).isFalse();
  }
}