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
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Duration;
import java.time.Instant;
import java.time.temporal.ChronoUnit;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

/**
 * Comprehensive tests for DenyListManager functionality.
 * 
 * Tests file I/O, TTL expiration, thread safety, and all core operations.
 */
class DenyListManagerTest {

  @TempDir Path tempDir;

  private static final Address TEST_ADDRESS_1 = Address.fromHexString("0x1234567890123456789012345678901234567890");
  private static final Address TEST_ADDRESS_2 = Address.fromHexString("0x9876543210987654321098765432109876543210");
  private static final Address TEST_ADDRESS_3 = Address.fromHexString("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd");

  private DenyListManager denyListManager;
  private Path denyListFile;

  @BeforeEach
  void setUp() {
    denyListFile = tempDir.resolve("test_deny_list.txt");
  }

  @AfterEach
  void tearDown() throws Exception {
    if (denyListManager != null) {
      denyListManager.close();
    }
  }

  @Test
  void testBasicDenyListOperations() {
    denyListManager = new DenyListManager(
        "Test",
        denyListFile.toString(),
        60, // 60 minutes TTL
        0   // No auto-refresh
    );

    // Initially empty
    assertThat(denyListManager.size()).isEqualTo(0);
    assertThat(denyListManager.isDenied(TEST_ADDRESS_1)).isFalse();

    // Add address
    boolean added = denyListManager.addToDenyList(TEST_ADDRESS_1);
    assertThat(added).isTrue();
    assertThat(denyListManager.size()).isEqualTo(1);
    assertThat(denyListManager.isDenied(TEST_ADDRESS_1)).isTrue();

    // Add same address again
    boolean addedAgain = denyListManager.addToDenyList(TEST_ADDRESS_1);
    assertThat(addedAgain).isFalse(); // Already present
    assertThat(denyListManager.size()).isEqualTo(1);

    // Remove address
    boolean removed = denyListManager.removeFromDenyList(TEST_ADDRESS_1);
    assertThat(removed).isTrue();
    assertThat(denyListManager.size()).isEqualTo(0);
    assertThat(denyListManager.isDenied(TEST_ADDRESS_1)).isFalse();

    // Remove non-existent address
    boolean removedAgain = denyListManager.removeFromDenyList(TEST_ADDRESS_2);
    assertThat(removedAgain).isFalse();
  }

  @Test
  void testFilePersistence() throws IOException {
    denyListManager = new DenyListManager(
        "Test",
        denyListFile.toString(),
        60,
        0
    );

    // Add multiple addresses
    denyListManager.addToDenyList(TEST_ADDRESS_1);
    denyListManager.addToDenyList(TEST_ADDRESS_2);

    // Verify file was created and contains entries
    assertThat(Files.exists(denyListFile)).isTrue();
    String fileContent = Files.readString(denyListFile);
    assertThat(fileContent).contains(TEST_ADDRESS_1.toHexString().toLowerCase());
    assertThat(fileContent).contains(TEST_ADDRESS_2.toHexString().toLowerCase());

    // Close and recreate manager to test loading from file
    denyListManager.close();
    denyListManager = new DenyListManager(
        "Test",
        denyListFile.toString(),
        60,
        0
    );

    // Should load from file
    assertThat(denyListManager.size()).isEqualTo(2);
    assertThat(denyListManager.isDenied(TEST_ADDRESS_1)).isTrue();
    assertThat(denyListManager.isDenied(TEST_ADDRESS_2)).isTrue();
  }

  @Test
  void testTtlExpiration() throws IOException {
    // Create manager with very short TTL for testing
    denyListManager = new DenyListManager(
        "Test",
        denyListFile.toString(),
        0, // 0 minutes TTL - everything expires immediately
        0
    );

    // Add address - it should be immediately expired
    denyListManager.addToDenyList(TEST_ADDRESS_1);
    
    // Check that it's marked as expired when checked
    assertThat(denyListManager.isDenied(TEST_ADDRESS_1)).isFalse();
    assertThat(denyListManager.size()).isEqualTo(0); // Should be cleaned up
  }

  @Test
  void testFileRefresh() throws Exception {
    // Create manager with auto-refresh
    denyListManager = new DenyListManager(
        "Test",
        denyListFile.toString(),
        60,
        1 // Refresh every 1 second
    );

    // Manually add entry to file
    Instant now = Instant.now();
    String fileEntry = TEST_ADDRESS_3.toHexString().toLowerCase() + "," + now.toString();
    Files.writeString(denyListFile, fileEntry);

    // Wait for refresh to pick up the change
    await().atMost(Duration.ofSeconds(3))
        .until(() -> denyListManager.isDenied(TEST_ADDRESS_3));

    assertThat(denyListManager.size()).isEqualTo(1);
    assertThat(denyListManager.isDenied(TEST_ADDRESS_3)).isTrue();
  }

  @Test
  void testMalformedFileHandling() throws IOException {
    // Create file with malformed entries
    String malformedContent = "invalid-address,2023-01-01T00:00:00Z\n" +
                              "0x1234567890123456789012345678901234567890,invalid-timestamp\n" +
                              "incomplete-line\n" +
                              TEST_ADDRESS_1.toHexString().toLowerCase() + "," + Instant.now().toString();
    
    Files.writeString(denyListFile, malformedContent);

    denyListManager = new DenyListManager(
        "Test",
        denyListFile.toString(),
        60,
        0
    );

    // Should load only the valid entry
    assertThat(denyListManager.size()).isEqualTo(1);
    assertThat(denyListManager.isDenied(TEST_ADDRESS_1)).isTrue();
  }

  @Test
  void testExpiredEntriesCleanupOnLoad() throws IOException {
    // Create file with expired and valid entries
    Instant expired = Instant.now().minus(2, ChronoUnit.HOURS);
    Instant valid = Instant.now();
    
    String fileContent = TEST_ADDRESS_1.toHexString().toLowerCase() + "," + expired.toString() + "\n" +
                        TEST_ADDRESS_2.toHexString().toLowerCase() + "," + valid.toString();
    
    Files.writeString(denyListFile, fileContent);

    denyListManager = new DenyListManager(
        "Test",
        denyListFile.toString(),
        60, // 60 minutes TTL
        0
    );

    // Should load only the non-expired entry
    assertThat(denyListManager.size()).isEqualTo(1);
    assertThat(denyListManager.isDenied(TEST_ADDRESS_1)).isFalse(); // Expired
    assertThat(denyListManager.isDenied(TEST_ADDRESS_2)).isTrue();   // Valid

    // File should be cleaned up automatically
    String cleanedContent = Files.readString(denyListFile);
    assertThat(cleanedContent).doesNotContain(TEST_ADDRESS_1.toHexString());
    assertThat(cleanedContent).contains(TEST_ADDRESS_2.toHexString());
  }

  @Test
  void testConcurrentOperations() throws InterruptedException {
    denyListManager = new DenyListManager(
        "Test",
        denyListFile.toString(),
        60,
        0
    );

    // Test concurrent operations
    Thread[] threads = new Thread[10];
    
    for (int i = 0; i < threads.length; i++) {
      final int threadId = i;
      threads[i] = new Thread(() -> {
        Address testAddr = Address.fromHexString(String.format("0x%040d", threadId));
        denyListManager.addToDenyList(testAddr);
        assertThat(denyListManager.isDenied(testAddr)).isTrue();
      });
    }

    // Start all threads
    for (Thread thread : threads) {
      thread.start();
    }

    // Wait for all threads to complete
    for (Thread thread : threads) {
      thread.join();
    }

    // Verify all entries were added
    assertThat(denyListManager.size()).isEqualTo(10);
  }

  @Test
  void testReloadFromFile() throws IOException {
    denyListManager = new DenyListManager(
        "Test",
        denyListFile.toString(),
        60,
        0
    );

    // Add entry via manager
    denyListManager.addToDenyList(TEST_ADDRESS_1);
    assertThat(denyListManager.size()).isEqualTo(1);

    // Manually modify file to add another entry
    String existingContent = Files.readString(denyListFile);
    String newEntry = TEST_ADDRESS_2.toHexString().toLowerCase() + "," + Instant.now().toString();
    Files.writeString(denyListFile, existingContent + "\n" + newEntry);

    // Reload from file
    denyListManager.reloadFromFile();

    // Should now have both entries
    assertThat(denyListManager.size()).isEqualTo(2);
    assertThat(denyListManager.isDenied(TEST_ADDRESS_1)).isTrue();
    assertThat(denyListManager.isDenied(TEST_ADDRESS_2)).isTrue();
  }

  @Test
  void testNonExistentFile() {
    // Create manager with non-existent file
    Path nonExistentFile = tempDir.resolve("non_existent.txt");
    
    denyListManager = new DenyListManager(
        "Test",
        nonExistentFile.toString(),
        60,
        0
    );

    // Should initialize with empty list
    assertThat(denyListManager.size()).isEqualTo(0);

    // Adding entry should create the file
    denyListManager.addToDenyList(TEST_ADDRESS_1);
    assertThat(Files.exists(nonExistentFile)).isTrue();
    assertThat(denyListManager.size()).isEqualTo(1);
  }

  @Test 
  void testAtomicFileOperations() throws IOException {
    denyListManager = new DenyListManager(
        "Test",
        denyListFile.toString(),
        60,
        0
    );

    // Add entry and verify atomic operation
    denyListManager.addToDenyList(TEST_ADDRESS_1);
    
    // File should exist and be readable
    assertThat(Files.exists(denyListFile)).isTrue();
    assertThat(Files.isReadable(denyListFile)).isTrue();
    
    // Content should be valid
    String content = Files.readString(denyListFile);
    assertThat(content).contains(TEST_ADDRESS_1.toHexString().toLowerCase());
    // Verify it contains a timestamp (year 2025)
    assertThat(content).contains("2025-");
    assertThat(content).contains("T");
    assertThat(content).contains("Z");
  }
}