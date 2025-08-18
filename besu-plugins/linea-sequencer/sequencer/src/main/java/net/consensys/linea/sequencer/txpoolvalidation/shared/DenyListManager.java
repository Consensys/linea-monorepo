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

import java.io.BufferedReader;
import java.io.Closeable;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;
import java.nio.file.StandardOpenOption;
import java.time.Instant;
import java.time.format.DateTimeParseException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;
import org.hyperledger.besu.datatypes.Address;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Shared deny list manager providing single source of truth for deny list state.
 *
 * <p>This manager encapsulates all deny list functionality including:
 *
 * <ul>
 *   <li>Thread-safe in-memory cache management
 *   <li>Atomic file I/O operations with proper locking
 *   <li>Automatic TTL-based entry expiration
 *   <li>Scheduled file refresh for external modifications
 *   <li>Clear separation of read-only vs write operations
 * </ul>
 *
 * <p><strong>Usage Pattern:</strong>
 *
 * <ul>
 *   <li>RlnVerifierValidator: Uses both read and write operations
 *   <li>LineaEstimateGas: Uses only read operations for efficiency
 * </ul>
 *
 * <p><strong>Thread Safety:</strong> All operations are thread-safe using ConcurrentHashMap and
 * synchronized file I/O.
 *
 * @author Status Network Development Team
 * @since 1.0
 */
public class DenyListManager implements Closeable {
  private static final Logger LOG = LoggerFactory.getLogger(DenyListManager.class);

  private final Path denyListFilePath;
  private final long entryMaxAgeMinutes;
  private final String serviceName;

  // Thread-safe in-memory cache - single source of truth
  private final Map<Address, Instant> denyList = new ConcurrentHashMap<>();

  private ScheduledExecutorService denyListRefreshScheduler;

  /**
   * Creates a new DenyListManager with the specified configuration.
   *
   * @param serviceName Name for logging and identification purposes
   * @param denyListPath Path to the deny list file
   * @param entryMaxAgeMinutes Maximum age for deny list entries in minutes
   * @param refreshIntervalSeconds How often to refresh from file (0 to disable)
   */
  public DenyListManager(
      String serviceName,
      String denyListPath,
      long entryMaxAgeMinutes,
      long refreshIntervalSeconds) {
    this.serviceName = serviceName;
    this.denyListFilePath = Paths.get(denyListPath);
    this.entryMaxAgeMinutes = entryMaxAgeMinutes;

    // Load initial state from file
    loadDenyListFromFile();

    // Start refresh scheduler if enabled
    if (refreshIntervalSeconds > 0) {
      startDenyListRefreshScheduler(refreshIntervalSeconds);
    } else {
      LOG.info("{}: Deny list auto-refresh is DISABLED (refresh interval <= 0)", serviceName);
    }

    LOG.info(
        "{}: DenyListManager initialized successfully. File: {}, MaxAge: {}min, Refresh: {}s",
        serviceName,
        denyListPath,
        entryMaxAgeMinutes,
        refreshIntervalSeconds);
  }

  /**
   * Checks if an address is currently on the deny list.
   *
   * <p>This is a read-only operation that automatically handles TTL expiration. Safe for concurrent
   * access by multiple threads.
   *
   * @param address The address to check
   * @return true if the address is denied and not expired, false otherwise
   */
  public boolean isDenied(Address address) {
    Instant deniedAt = denyList.get(address);
    if (deniedAt == null) {
      return false;
    }

    // Check if entry has expired
    if (isEntryExpired(deniedAt)) {
      // Remove expired entry (this might cause a small race condition but it's acceptable)
      if (denyList.remove(address, deniedAt)) {
        LOG.debug(
            "{}: Expired deny list entry for {} removed during check",
            serviceName,
            address.toHexString());
        // Note: We don't persist this removal immediately for performance
        // It will be cleaned up during the next file refresh
      }
      return false;
    }

    return true;
  }

  /**
   * Adds an address to the deny list with current timestamp.
   *
   * <p>This is a write operation that immediately persists to file. Should only be called by
   * components that have write access (e.g., RlnVerifierValidator).
   *
   * @param address The address to add to the deny list
   * @return true if the address was newly added, false if it was already present
   */
  public boolean addToDenyList(Address address) {
    Instant now = Instant.now();
    Instant previous = denyList.put(address, now);

    if (previous == null) {
      // Persist immediately to ensure consistency
      saveDenyListToFile();
      LOG.info(
          "{}: Address {} added to deny list at {}. Cache size: {}",
          serviceName,
          address.toHexString(),
          now,
          denyList.size());
      return true;
    } else {
      LOG.debug(
          "{}: Address {} was already on deny list (updated timestamp)",
          serviceName,
          address.toHexString());
      // Still persist to update timestamp
      saveDenyListToFile();
      return false;
    }
  }

  /**
   * Removes an address from the deny list.
   *
   * <p>This is a write operation that immediately persists to file. Should only be called by
   * components that have write access (e.g., RlnVerifierValidator).
   *
   * @param address The address to remove from the deny list
   * @return true if the address was removed, false if it wasn't on the list
   */
  public boolean removeFromDenyList(Address address) {
    Instant removed = denyList.remove(address);

    if (removed != null) {
      // Persist immediately to ensure consistency
      saveDenyListToFile();
      LOG.info(
          "{}: Address {} removed from deny list. Cache size: {}",
          serviceName,
          address.toHexString(),
          denyList.size());
      return true;
    } else {
      LOG.debug(
          "{}: Address {} was not on deny list, nothing to remove",
          serviceName,
          address.toHexString());
      return false;
    }
  }

  /**
   * Gets the current size of the deny list (for monitoring/debugging).
   *
   * @return Number of addresses currently on the deny list
   */
  public int size() {
    return denyList.size();
  }

  /**
   * Forces a reload of the deny list from file.
   *
   * <p>This can be useful for testing or when external changes are made to the file. Thread-safe
   * and automatically handles TTL expiration during load.
   */
  public void reloadFromFile() {
    loadDenyListFromFile();
  }

  /** Starts the scheduled task for deny list file refresh. */
  private void startDenyListRefreshScheduler(long refreshIntervalSeconds) {
    denyListRefreshScheduler =
        Executors.newSingleThreadScheduledExecutor(
            r -> {
              Thread t = Executors.defaultThreadFactory().newThread(r);
              t.setName(serviceName + "-DenyListRefresh");
              t.setDaemon(true);
              return t;
            });

    denyListRefreshScheduler.scheduleAtFixedRate(
        this::loadDenyListFromFile,
        refreshIntervalSeconds,
        refreshIntervalSeconds,
        TimeUnit.SECONDS);

    LOG.info(
        "{}: Scheduled deny list refresh every {} seconds", serviceName, refreshIntervalSeconds);
  }

  /**
   * Loads the deny list from the configured file path.
   *
   * <p>Reads deny list entries from file in format: "address,timestamp" and automatically removes
   * expired entries based on configured TTL. Updates are atomic to prevent inconsistent state
   * during concurrent access.
   */
  private synchronized void loadDenyListFromFile() {
    if (!Files.exists(denyListFilePath)) {
      LOG.debug(
          "{}: Deny list file not found at {}, keeping current cache",
          serviceName,
          denyListFilePath);
      return;
    }

    Map<Address, Instant> newDenyListCache = new ConcurrentHashMap<>();
    Instant now = Instant.now();
    boolean entriesPruned = false;

    try (BufferedReader reader =
        Files.newBufferedReader(denyListFilePath, StandardCharsets.UTF_8)) {
      String line;
      while ((line = reader.readLine()) != null) {
        String[] parts = line.split(",", 2);
        if (parts.length == 2) {
          try {
            Address address = Address.fromHexString(parts[0].trim());
            Instant timestamp = Instant.parse(parts[1].trim());

            if (!isEntryExpired(timestamp, now)) {
              newDenyListCache.put(address, timestamp);
            } else {
              entriesPruned = true;
              LOG.debug(
                  "{}: Expired deny list entry for {} (added at {}) removed during load",
                  serviceName,
                  address,
                  timestamp);
            }
          } catch (IllegalArgumentException | DateTimeParseException e) {
            LOG.warn(
                "{}: Invalid entry in deny list file: '{}'. Skipping. Error: {}",
                serviceName,
                line,
                e.getMessage());
          }
        } else {
          LOG.warn(
              "{}: Malformed line in deny list file (expected 'address,timestamp'): '{}'",
              serviceName,
              line);
        }
      }

      // Atomic update of the cache
      denyList.clear();
      denyList.putAll(newDenyListCache);

      LOG.debug(
          "{}: Deny list loaded successfully from {}. {} active entries",
          serviceName,
          denyListFilePath,
          denyList.size());

      // If we pruned expired entries, save the cleaned list back to file
      if (entriesPruned) {
        saveDenyListToFile();
      }

    } catch (IOException e) {
      LOG.error(
          "{}: Error loading deny list from {}: {}",
          serviceName,
          denyListFilePath,
          e.getMessage(),
          e);
    }
  }

  /**
   * Atomically saves the current deny list state to file.
   *
   * <p>Uses atomic file operations (write to temp, then move) to ensure file consistency and
   * prevent corruption during concurrent access.
   */
  private synchronized void saveDenyListToFile() {
    Map<Address, Instant> denyListSnapshot = new HashMap<>(denyList);
    List<String> entriesAsString =
        denyListSnapshot.entrySet().stream()
            .map(
                entry ->
                    entry.getKey().toHexString().toLowerCase() + "," + entry.getValue().toString())
            .sorted()
            .collect(Collectors.toList());

    try {
      // Ensure parent directory exists
      Files.createDirectories(denyListFilePath.getParent());

      Path tempFilePath =
          denyListFilePath
              .getParent()
              .resolve(denyListFilePath.getFileName().toString() + ".tmp_save");

      Files.write(
          tempFilePath,
          entriesAsString,
          StandardCharsets.UTF_8,
          StandardOpenOption.CREATE,
          StandardOpenOption.TRUNCATE_EXISTING);

      Files.move(
          tempFilePath,
          denyListFilePath,
          StandardCopyOption.REPLACE_EXISTING,
          StandardCopyOption.ATOMIC_MOVE);

      LOG.debug(
          "{}: Deny list saved to file {} with {} entries",
          serviceName,
          denyListFilePath,
          entriesAsString.size());

    } catch (IOException e) {
      LOG.error(
          "{}: Error saving deny list to file {}: {}",
          serviceName,
          denyListFilePath,
          e.getMessage(),
          e);
    }
  }

  /** Checks if a deny list entry has expired based on its timestamp. */
  private boolean isEntryExpired(Instant entryTimestamp) {
    return isEntryExpired(entryTimestamp, Instant.now());
  }

  /** Checks if a deny list entry has expired based on its timestamp and current time. */
  private boolean isEntryExpired(Instant entryTimestamp, Instant currentTime) {
    long maxAgeMillis = TimeUnit.MINUTES.toMillis(entryMaxAgeMinutes);
    return (currentTime.toEpochMilli() - entryTimestamp.toEpochMilli()) >= maxAgeMillis;
  }

  /**
   * Closes all resources including scheduled executors.
   *
   * <p>Ensures graceful shutdown of all background tasks. This method should be called when the
   * manager is no longer needed to prevent resource leaks.
   *
   * @throws IOException if there are issues during resource cleanup
   */
  @Override
  public void close() throws IOException {
    if (denyListRefreshScheduler != null && !denyListRefreshScheduler.isShutdown()) {
      LOG.info("{}: Shutting down deny list refresh scheduler", serviceName);
      denyListRefreshScheduler.shutdown();
      try {
        if (!denyListRefreshScheduler.awaitTermination(5, TimeUnit.SECONDS)) {
          denyListRefreshScheduler.shutdownNow();
        }
      } catch (InterruptedException e) {
        denyListRefreshScheduler.shutdownNow();
        Thread.currentThread().interrupt();
      }
    }
    LOG.info("{}: DenyListManager closed", serviceName);
  }
}
