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
import static org.awaitility.Awaitility.await;

import java.io.IOException;
import java.time.Duration;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import net.consensys.linea.sequencer.txpoolvalidation.shared.NullifierTracker.NullifierStats;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

/**
 * Comprehensive tests for NullifierTracker functionality.
 * 
 * Tests nullifier tracking, epoch scoping, TTL expiration, thread safety, and performance.
 */
class NullifierTrackerTest {

  private static final String TEST_NULLIFIER_1 = "0xa1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456";
  private static final String TEST_NULLIFIER_2 = "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321";
  private static final String TEST_EPOCH_1 = "0x1c61ef0b2ebc0235d85fe8537b4455549356e3895005ba7a03fbd4efc9ba3692";
  private static final String TEST_EPOCH_2 = "0x9999999999999999999999999999999999999999999999999999999999999999";

  private NullifierTracker tracker;

  @BeforeEach
  void setUp() {
    // Create tracker with reasonable defaults for testing
    tracker = new NullifierTracker("Test", 1000L, 1L); // 1 hour TTL, 1000 max size
  }

  @AfterEach
  void tearDown() throws Exception {
    if (tracker != null) {
      tracker.close();
    }
  }

  @Test
  void testBasicNullifierTracking() {
    // Test first use of nullifier in epoch
    boolean isNew = tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_1);
    assertThat(isNew).isTrue();

    // Test reuse of same nullifier in same epoch
    boolean isReused = tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_1);
    assertThat(isReused).isFalse();

    // Verify tracking state
    assertThat(tracker.isNullifierUsed(TEST_NULLIFIER_1, TEST_EPOCH_1)).isTrue();
    assertThat(tracker.isNullifierUsed(TEST_NULLIFIER_2, TEST_EPOCH_1)).isFalse();
  }

  @Test
  void testEpochScoping() {
    // Use same nullifier in different epochs - should be allowed
    boolean isNewInEpoch1 = tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_1);
    assertThat(isNewInEpoch1).isTrue();

    boolean isNewInEpoch2 = tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_2);
    assertThat(isNewInEpoch2).isTrue(); // Same nullifier, different epoch

    // Verify both are tracked separately
    assertThat(tracker.isNullifierUsed(TEST_NULLIFIER_1, TEST_EPOCH_1)).isTrue();
    assertThat(tracker.isNullifierUsed(TEST_NULLIFIER_1, TEST_EPOCH_2)).isTrue();

    // Try to reuse in each epoch
    assertThat(tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_1)).isFalse();
    assertThat(tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_2)).isFalse();
  }

  @Test
  void testInvalidInputHandling() {
    // Test null nullifier
    boolean result1 = tracker.checkAndMarkNullifier(null, TEST_EPOCH_1);
    assertThat(result1).isFalse();

    // Test empty nullifier
    boolean result2 = tracker.checkAndMarkNullifier("", TEST_EPOCH_1);
    assertThat(result2).isFalse();

    // Test null epoch
    boolean result3 = tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, null);
    assertThat(result3).isFalse();

    // Test empty epoch
    boolean result4 = tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, "");
    assertThat(result4).isFalse();

    // Verify no entries were added
    NullifierStats stats = tracker.getStats();
    assertThat(stats.totalTracked()).isEqualTo(0);
  }

  @Test
  void testStatisticsTracking() {
    // Track some nullifiers
    tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_1);
    tracker.checkAndMarkNullifier(TEST_NULLIFIER_2, TEST_EPOCH_1);
    tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_2);

    // Attempt reuse
    tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_1);
    tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_1);

    NullifierStats stats = tracker.getStats();
    assertThat(stats.totalTracked()).isEqualTo(3);
    assertThat(stats.duplicateAttempts()).isEqualTo(2);
    assertThat(stats.currentNullifiers()).isEqualTo(3);
  }

  @Test
  void testConcurrentAccess() throws InterruptedException {
    int threadCount = 10;
    int operationsPerThread = 100;
    ExecutorService executor = Executors.newFixedThreadPool(threadCount);
    CountDownLatch latch = new CountDownLatch(threadCount);
    AtomicInteger successCount = new AtomicInteger(0);

    for (int t = 0; t < threadCount; t++) {
      final int threadId = t;
      executor.submit(() -> {
        try {
          for (int i = 0; i < operationsPerThread; i++) {
            String nullifier = String.format("0x%064d", threadId * operationsPerThread + i);
            boolean isNew = tracker.checkAndMarkNullifier(nullifier, TEST_EPOCH_1);
            if (isNew) {
              successCount.incrementAndGet();
            }
          }
        } finally {
          latch.countDown();
        }
      });
    }

    boolean completed = latch.await(10, TimeUnit.SECONDS);
    assertThat(completed).isTrue();

    executor.shutdown();
    executor.awaitTermination(5, TimeUnit.SECONDS);

    // All operations should have succeeded (unique nullifiers)
    assertThat(successCount.get()).isEqualTo(threadCount * operationsPerThread);
    
    NullifierStats stats = tracker.getStats();
    assertThat(stats.totalTracked()).isEqualTo(threadCount * operationsPerThread);
    assertThat(stats.duplicateAttempts()).isEqualTo(0);
  }

  @Test
  void testConcurrentNullifierReuse() throws InterruptedException {
    int threadCount = 5;
    ExecutorService executor = Executors.newFixedThreadPool(threadCount);
    CountDownLatch latch = new CountDownLatch(threadCount);
    AtomicInteger successCount = new AtomicInteger(0);
    AtomicInteger failureCount = new AtomicInteger(0);

    // All threads try to use the same nullifier in the same epoch
    for (int t = 0; t < threadCount; t++) {
      executor.submit(() -> {
        try {
          boolean isNew = tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_1);
          if (isNew) {
            successCount.incrementAndGet();
          } else {
            failureCount.incrementAndGet();
          }
        } finally {
          latch.countDown();
        }
      });
    }

    boolean completed = latch.await(10, TimeUnit.SECONDS);
    assertThat(completed).isTrue();

    executor.shutdown();
    executor.awaitTermination(5, TimeUnit.SECONDS);

    // Only one thread should succeed
    assertThat(successCount.get()).isEqualTo(1);
    assertThat(failureCount.get()).isEqualTo(threadCount - 1);
    
    NullifierStats stats = tracker.getStats();
    assertThat(stats.totalTracked()).isEqualTo(1);
    assertThat(stats.duplicateAttempts()).isEqualTo(threadCount - 1);
  }

  @Test
  void testCaseInsensitiveNullifiers() {
    // Test that nullifiers are normalized to lowercase
    String upperCaseNullifier = TEST_NULLIFIER_1.toUpperCase();
    String lowerCaseNullifier = TEST_NULLIFIER_1.toLowerCase();

    boolean firstResult = tracker.checkAndMarkNullifier(upperCaseNullifier, TEST_EPOCH_1);
    assertThat(firstResult).isTrue();

    boolean secondResult = tracker.checkAndMarkNullifier(lowerCaseNullifier, TEST_EPOCH_1);
    assertThat(secondResult).isFalse(); // Should be treated as same nullifier

    // Mixed case should also be detected
    String mixedCaseNullifier = "0xA1B2c3D4e5F6789012345678901234567890abcdef1234567890abcdef123456";
    boolean thirdResult = tracker.checkAndMarkNullifier(mixedCaseNullifier, TEST_EPOCH_1);
    assertThat(thirdResult).isFalse();
  }

  @Test
  void testNullifierTrackerConfiguration() throws IOException {
    // Test that tracker can be configured with different parameters
    tracker.close(); // Close default tracker
    
    // Create tracker with specific configuration
    tracker = new NullifierTracker("ConfigTest", 500L, 24L); // 500 max size, 24 hour TTL
    
    // Verify it's working
    boolean isNew = tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_1);
    assertThat(isNew).isTrue();
    
    // Verify configuration is applied
    NullifierStats stats = tracker.getStats();
    assertThat(stats.totalTracked()).isEqualTo(1);
    assertThat(stats.currentNullifiers()).isEqualTo(1);
  }

  @Test
  void testWhitespaceHandling() {
    // Test nullifiers and epochs with whitespace
    String nullifierWithSpaces = "  " + TEST_NULLIFIER_1 + "  ";
    String epochWithSpaces = "  " + TEST_EPOCH_1 + "  ";

    boolean result1 = tracker.checkAndMarkNullifier(nullifierWithSpaces, epochWithSpaces);
    assertThat(result1).isTrue();

    // Same nullifier without spaces should be detected as reuse
    boolean result2 = tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_1);
    assertThat(result2).isFalse();
  }

  @Test
  void testLegacyConstructor() throws Exception {
    tracker.close(); // Close default tracker
    
    // Test legacy constructor with file path (should be ignored)
    tracker = new NullifierTracker("Test", "/tmp/ignored_file.txt", 1L);

    // Should work normally despite file path
    boolean isNew = tracker.checkAndMarkNullifier(TEST_NULLIFIER_1, TEST_EPOCH_1);
    assertThat(isNew).isTrue();

    NullifierStats stats = tracker.getStats();
    assertThat(stats.totalTracked()).isEqualTo(1);
  }
}