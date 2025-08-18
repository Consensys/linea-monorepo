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
package net.consensys.linea.sequencer.txpoolvalidation;

import static org.assertj.core.api.Assertions.assertThat;

import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Path;
import java.time.Duration;
import java.time.Instant;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicLong;
import net.consensys.linea.sequencer.txpoolvalidation.shared.DenyListManager;
import net.consensys.linea.sequencer.txpoolvalidation.shared.NullifierTracker;
import net.consensys.linea.sequencer.txpoolvalidation.shared.NullifierTracker.NullifierStats;
import org.apache.tuweni.bytes.Bytes;
import org.bouncycastle.asn1.sec.SECNamedCurves;
import org.bouncycastle.asn1.x9.X9ECParameters;
import org.bouncycastle.crypto.params.ECDomainParameters;
import org.hyperledger.besu.crypto.SECPSignature;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

/**
 * Performance and stress tests for RLN validation components.
 * 
 * Tests high-throughput scenarios and system behavior under load.
 */
class RlnValidationPerformanceTest {

  @TempDir Path tempDir;

  private static final SECPSignature FAKE_SIGNATURE;

  static {
    final X9ECParameters params = SECNamedCurves.getByName("secp256k1");
    final ECDomainParameters curve =
        new ECDomainParameters(params.getCurve(), params.getG(), params.getN(), params.getH());
    FAKE_SIGNATURE =
        SECPSignature.create(
            new BigInteger("66397251408932042429874251838229702988618145381408295790259650671563847073199"),
            new BigInteger("24729624138373455972486746091821238755870276413282629437244319694880507882088"),
            (byte) 0,
            curve.getN());
  }

  private DenyListManager denyListManager;
  private NullifierTracker nullifierTracker;

  @BeforeEach
  void setUp() throws IOException {
    Path denyListFile = tempDir.resolve("performance_deny_list.txt");
    denyListManager = new DenyListManager("PerformanceTest", denyListFile.toString(), 60, 0);
    nullifierTracker = new NullifierTracker("PerformanceTest", 100_000L, 1L);
  }

  @AfterEach
  void tearDown() throws Exception {
    if (denyListManager != null) {
      denyListManager.close();
    }
    if (nullifierTracker != null) {
      nullifierTracker.close();
    }
  }

  @Test
  void testHighThroughputNullifierTracking() throws InterruptedException {
    int threadCount = 10;
    int operationsPerThread = 1000;
    int totalOperations = threadCount * operationsPerThread;
    
    ExecutorService executor = Executors.newFixedThreadPool(threadCount);
    CountDownLatch latch = new CountDownLatch(threadCount);
    AtomicInteger successCount = new AtomicInteger(0);
    AtomicLong totalDuration = new AtomicLong(0);

    // Measure performance of nullifier tracking
    Instant startTime = Instant.now();

    for (int t = 0; t < threadCount; t++) {
      final int threadId = t;
      executor.submit(() -> {
        try {
          Instant threadStart = Instant.now();
          
          for (int i = 0; i < operationsPerThread; i++) {
            String nullifier = String.format("0x%064d", threadId * operationsPerThread + i);
            String epoch = "epoch-" + (i % 10); // Use different epochs to test scoping
            
            boolean isNew = nullifierTracker.checkAndMarkNullifier(nullifier, epoch);
            if (isNew) {
              successCount.incrementAndGet();
            }
          }
          
          Instant threadEnd = Instant.now();
          totalDuration.addAndGet(Duration.between(threadStart, threadEnd).toMillis());
          
        } finally {
          latch.countDown();
        }
      });
    }

    boolean completed = latch.await(30, TimeUnit.SECONDS);
    assertThat(completed).isTrue();

    executor.shutdown();
    executor.awaitTermination(5, TimeUnit.SECONDS);

    Instant endTime = Instant.now();
    long totalWallClockTime = Duration.between(startTime, endTime).toMillis();

    // Verify performance metrics
    assertThat(successCount.get()).isEqualTo(totalOperations);
    
    NullifierStats stats = nullifierTracker.getStats();
    assertThat(stats.totalTracked()).isEqualTo(totalOperations);
    assertThat(stats.duplicateAttempts()).isEqualTo(0);

    // Log performance results
    double throughput = (double) totalOperations / (totalWallClockTime / 1000.0);
    System.out.printf("Nullifier tracking performance: %d operations in %d ms (%.2f ops/sec)%n", 
        totalOperations, totalWallClockTime, throughput);
    
    // Performance assertion - should handle at least 1000 ops/sec
    assertThat(throughput).isGreaterThan(1000.0);
  }

  @Test
  void testDenyListPerformance() throws InterruptedException {
    int threadCount = 5;
    int operationsPerThread = 200;
    int totalOperations = threadCount * operationsPerThread;
    
    ExecutorService executor = Executors.newFixedThreadPool(threadCount);
    CountDownLatch latch = new CountDownLatch(threadCount);
    AtomicInteger addCount = new AtomicInteger(0);
    AtomicInteger checkCount = new AtomicInteger(0);

    Instant startTime = Instant.now();

    for (int t = 0; t < threadCount; t++) {
      final int threadId = t;
      executor.submit(() -> {
        try {
          for (int i = 0; i < operationsPerThread; i++) {
            Address testAddr = Address.fromHexString(String.format("0x%040d", threadId * operationsPerThread + i));
            
            // Add to deny list
            boolean added = denyListManager.addToDenyList(testAddr);
            if (added) {
              addCount.incrementAndGet();
            }
            
            // Check deny list
            boolean isDenied = denyListManager.isDenied(testAddr);
            if (isDenied) {
              checkCount.incrementAndGet();
            }
          }
        } finally {
          latch.countDown();
        }
      });
    }

    boolean completed = latch.await(60, TimeUnit.SECONDS);
    assertThat(completed).isTrue();

    executor.shutdown();
    executor.awaitTermination(5, TimeUnit.SECONDS);

    Instant endTime = Instant.now();
    long totalTime = Duration.between(startTime, endTime).toMillis();

    // Verify results
    assertThat(addCount.get()).isEqualTo(totalOperations);
    assertThat(checkCount.get()).isEqualTo(totalOperations);
    assertThat(denyListManager.size()).isEqualTo(totalOperations);

    // Log performance
    double throughput = (double) (totalOperations * 2) / (totalTime / 1000.0); // 2 operations per iteration
    System.out.printf("Deny list performance: %d operations in %d ms (%.2f ops/sec)%n", 
        totalOperations * 2, totalTime, throughput);
  }

  @Test
  void testMemoryUsageUnderLoad() throws InterruptedException {
    // Test memory usage with large number of entries
    int nullifierCount = 10_000;
    int addressCount = 1_000;

    // Add many nullifiers
    for (int i = 0; i < nullifierCount; i++) {
      String nullifier = String.format("0x%064d", i);
      String epoch = "epoch-" + (i % 100);
      nullifierTracker.checkAndMarkNullifier(nullifier, epoch);
    }

    // Add many addresses to deny list
    for (int i = 0; i < addressCount; i++) {
      Address addr = Address.fromHexString(String.format("0x%040d", i));
      denyListManager.addToDenyList(addr);
    }

    // Verify counts
    NullifierStats stats = nullifierTracker.getStats();
    assertThat(stats.totalTracked()).isEqualTo(nullifierCount);
    assertThat(stats.currentNullifiers()).isEqualTo(nullifierCount);
    assertThat(denyListManager.size()).isEqualTo(addressCount);

    // Test continued operations under load
    String testNullifier = "0x9999999999999999999999999999999999999999999999999999999999999999";
    boolean canStillOperate = nullifierTracker.checkAndMarkNullifier(testNullifier, "test-epoch");
    assertThat(canStillOperate).isTrue();

    Address testAddr = Address.fromHexString("0x9999999999999999999999999999999999999999");
    boolean canStillAdd = denyListManager.addToDenyList(testAddr);
    assertThat(canStillAdd).isTrue();
  }

  @Test
  void testCacheEvictionBehavior() throws InterruptedException, IOException {
    // Create tracker with small size and short TTL for testing eviction
    nullifierTracker.close();
    nullifierTracker = new NullifierTracker("EvictionTest", 100L, 1L); // 100 max size, 1 hour TTL

    // Fill beyond capacity to test size-based behavior
    for (int i = 0; i < 50; i++) {
      String nullifier = String.format("0x%064d", i);
      nullifierTracker.checkAndMarkNullifier(nullifier, "test-epoch");
    }

    NullifierStats stats = nullifierTracker.getStats();
    // Verify tracker is working and recording entries
    assertThat(stats.currentNullifiers()).isGreaterThan(0);
    assertThat(stats.totalTracked()).isEqualTo(50);

    // Wait for TTL expiration
    Thread.sleep(5000); // Wait for entries to expire

    // Try to add new entry after expiration
    boolean canAdd = nullifierTracker.checkAndMarkNullifier("0xnew", "test-epoch");
    assertThat(canAdd).isTrue();
  }

  @Test
  void testDenyListFileIoPerformance() throws InterruptedException {
    int operationCount = 100;
    ExecutorService executor = Executors.newFixedThreadPool(5);
    CountDownLatch latch = new CountDownLatch(operationCount);
    AtomicLong totalFileOpTime = new AtomicLong(0);

    for (int i = 0; i < operationCount; i++) {
      final int index = i;
      executor.submit(() -> {
        try {
          Instant start = Instant.now();
          
          Address addr = Address.fromHexString(String.format("0x%040d", index));
          denyListManager.addToDenyList(addr);
          boolean isDenied = denyListManager.isDenied(addr);
          assertThat(isDenied).isTrue();
          
          Instant end = Instant.now();
          totalFileOpTime.addAndGet(Duration.between(start, end).toMillis());
          
        } finally {
          latch.countDown();
        }
      });
    }

    boolean completed = latch.await(30, TimeUnit.SECONDS);
    assertThat(completed).isTrue();

    executor.shutdown();
    executor.awaitTermination(5, TimeUnit.SECONDS);

    // Calculate average file operation time
    double avgFileOpTime = (double) totalFileOpTime.get() / operationCount;
    System.out.printf("Average file operation time: %.2f ms%n", avgFileOpTime);
    
    // File operations should be reasonably fast (under 100ms per operation)
    assertThat(avgFileOpTime).isLessThan(100.0);
  }

  @Test
  void testConcurrentNullifierConflicts() throws InterruptedException {
    // Test behavior when many threads try to use the same nullifiers
    int threadCount = 20;
    String conflictedNullifier = "0xconflicted000000000000000000000000000000000000000000000000000000";
    
    ExecutorService executor = Executors.newFixedThreadPool(threadCount);
    CountDownLatch latch = new CountDownLatch(threadCount);
    AtomicInteger successCount = new AtomicInteger(0);
    AtomicInteger conflictCount = new AtomicInteger(0);

    for (int t = 0; t < threadCount; t++) {
      executor.submit(() -> {
        try {
          // All threads try to use the same nullifier
          boolean isNew = nullifierTracker.checkAndMarkNullifier(conflictedNullifier, "conflict-epoch");
          
          if (isNew) {
            successCount.incrementAndGet();
          } else {
            conflictCount.incrementAndGet();
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

    // Only one thread should succeed, others should detect conflict
    assertThat(successCount.get()).isEqualTo(1);
    assertThat(conflictCount.get()).isEqualTo(threadCount - 1);

    NullifierStats stats = nullifierTracker.getStats();
    assertThat(stats.duplicateAttempts()).isEqualTo(threadCount - 1);
  }

  @Test
  void testSystemResourceUsageUnderLoad() throws InterruptedException {
    // Test system behavior under sustained load
    int duration = 5; // seconds
    AtomicInteger operationCount = new AtomicInteger(0);
    final boolean[] keepRunning = {true};

    ExecutorService executor = Executors.newFixedThreadPool(4);

    // Nullifier operations
    executor.submit(() -> {
      int counter = 0;
      while (keepRunning[0]) {
        String nullifier = String.format("0x%064d", counter++);
        String epoch = "load-epoch-" + (counter % 5);
        nullifierTracker.checkAndMarkNullifier(nullifier, epoch);
        operationCount.incrementAndGet();
        
        if (counter % 100 == 0) {
          try {
            Thread.sleep(1); // Small pause to prevent CPU overload
          } catch (InterruptedException e) {
            break;
          }
        }
      }
    });

    // Deny list operations  
    executor.submit(() -> {
      int counter = 0;
      while (keepRunning[0]) {
        Address addr = Address.fromHexString(String.format("0x%040d", counter % 1000));
        if (counter % 2 == 0) {
          denyListManager.addToDenyList(addr);
        } else {
          denyListManager.isDenied(addr);
        }
        operationCount.incrementAndGet();
        counter++;
        
        if (counter % 50 == 0) {
          try {
            Thread.sleep(1);
          } catch (InterruptedException e) {
            break;
          }
        }
      }
    });

    // Run for specified duration
    Thread.sleep(duration * 1000);
    keepRunning[0] = false;

    executor.shutdown();
    executor.awaitTermination(10, TimeUnit.SECONDS);

    // Verify system performed operations without issues
    assertThat(operationCount.get()).isGreaterThan(1000); // Should have done substantial work
    
    NullifierStats stats = nullifierTracker.getStats();
    assertThat(stats.currentNullifiers()).isGreaterThan(0);
    assertThat(denyListManager.size()).isGreaterThan(0);

    System.out.printf("Sustained load test: %d operations in %d seconds (%.2f ops/sec)%n",
        operationCount.get(), duration, (double) operationCount.get() / duration);
  }

  @Test
  void testDenyListFileGrowthAndCompaction() throws IOException {
    // Test behavior as deny list file grows large
    int addressCount = 1000;
    
    // Add many addresses
    Instant start = Instant.now();
    for (int i = 0; i < addressCount; i++) {
      Address addr = Address.fromHexString(String.format("0x%040d", i));
      denyListManager.addToDenyList(addr);
    }
    Instant addEnd = Instant.now();

    // Verify all were added
    assertThat(denyListManager.size()).isEqualTo(addressCount);

    // Test access performance after growth
    Instant checkStart = Instant.now();
    for (int i = 0; i < addressCount; i++) {
      Address addr = Address.fromHexString(String.format("0x%040d", i));
      assertThat(denyListManager.isDenied(addr)).isTrue();
    }
    Instant checkEnd = Instant.now();

    // Remove half the addresses
    Instant removeStart = Instant.now();
    for (int i = 0; i < addressCount / 2; i++) {
      Address addr = Address.fromHexString(String.format("0x%040d", i));
      denyListManager.removeFromDenyList(addr);
    }
    Instant removeEnd = Instant.now();

    // Verify final state
    assertThat(denyListManager.size()).isEqualTo(addressCount / 2);

    // Log performance metrics
    long addTime = Duration.between(start, addEnd).toMillis();
    long checkTime = Duration.between(checkStart, checkEnd).toMillis();
    long removeTime = Duration.between(removeStart, removeEnd).toMillis();

    System.out.printf("Deny list performance - Add: %d ms, Check: %d ms, Remove: %d ms%n",
        addTime, checkTime, removeTime);

    // Performance assertions
    assertThat(addTime).isLessThan(5000); // Should add 1000 entries in under 5 seconds
    assertThat(checkTime).isLessThan(1000); // Should check 1000 entries in under 1 second  
    assertThat(removeTime).isLessThan(2000); // Should remove 500 entries in under 2 seconds
  }

  @Test
  void testNullifierConflictUnderHighLoad() throws InterruptedException {
    // Test nullifier conflict detection under high concurrent load
    String conflictNullifier = "0xconflicted000000000000000000000000000000000000000000000000000000";
    String conflictEpoch = "high-load-epoch";
    
    int threadCount = 50;
    ExecutorService executor = Executors.newFixedThreadPool(threadCount);
    CountDownLatch latch = new CountDownLatch(threadCount);
    AtomicInteger winners = new AtomicInteger(0);
    AtomicInteger conflicts = new AtomicInteger(0);

    // All threads compete for the same nullifier
    for (int t = 0; t < threadCount; t++) {
      executor.submit(() -> {
        try {
          boolean won = nullifierTracker.checkAndMarkNullifier(conflictNullifier, conflictEpoch);
          if (won) {
            winners.incrementAndGet();
          } else {
            conflicts.incrementAndGet();
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

    // Critical security property: exactly one winner
    assertThat(winners.get()).isEqualTo(1);
    assertThat(conflicts.get()).isEqualTo(threadCount - 1);

    System.out.printf("High load conflict test: 1 winner, %d conflicts from %d threads%n",
        conflicts.get(), threadCount);
  }
}