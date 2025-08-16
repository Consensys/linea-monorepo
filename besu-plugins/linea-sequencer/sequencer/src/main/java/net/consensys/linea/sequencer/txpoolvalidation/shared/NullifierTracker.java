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

import com.github.benmanes.caffeine.cache.Cache;
import com.github.benmanes.caffeine.cache.Caffeine;
import com.github.benmanes.caffeine.cache.RemovalCause;
import com.github.benmanes.caffeine.cache.RemovalListener;
import com.github.benmanes.caffeine.cache.Scheduler;
import java.io.Closeable;
import java.io.IOException;
import java.time.Duration;
import java.time.Instant;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * High-performance nullifier tracking service using Caffeine cache.
 *
 * <p><strong>Security Critical:</strong> This component is essential for RLN security. Nullifier
 * tracking prevents replay attacks and enforces transaction rate limiting by detecting when users
 * reuse nullifiers within the same epoch.
 *
 * <p><strong>Performance Optimized:</strong> Uses Caffeine cache for high-throughput, low-latency
 * operations. Eliminates file I/O bottlenecks present in naive implementations.
 *
 * <p><strong>Epoch Scoping:</strong> Nullifiers are tracked per epoch. The same nullifier can be
 * reused across different epochs but not within the same epoch, enabling proper rate limiting.
 *
 * <p><strong>Automatic Cleanup:</strong> Expired nullifiers are automatically evicted based on
 * configured TTL to prevent unbounded memory growth.
 *
 * <p><strong>Thread Safety:</strong> All operations are thread-safe and lock-free, suitable for
 * high-concurrency transaction validation.
 *
 * @author Status Network Development Team
 * @since 1.0
 */
public class NullifierTracker implements Closeable {
  private static final Logger LOG = LoggerFactory.getLogger(NullifierTracker.class);

  private final String serviceName;
  private final Cache<String, NullifierData> nullifierCache;

  // Metrics for monitoring and debugging
  private final AtomicLong totalNullifiersTracked = new AtomicLong(0);
  private final AtomicLong nullifierHits = new AtomicLong(0);
  private final AtomicLong expiredNullifiers = new AtomicLong(0);

  /** Represents a tracked nullifier with its metadata. */
  private record NullifierData(String nullifier, String epochId, Instant timestamp) {}

  /**
   * Creates a new high-performance nullifier tracker using Caffeine cache.
   *
   * @param serviceName Service name for logging identification
   * @param maxSize Maximum number of nullifiers to track simultaneously (cache size)
   * @param nullifierExpiryHours Hours after which nullifiers expire and are evicted
   */
  public NullifierTracker(String serviceName, long maxSize, long nullifierExpiryHours) {
    this.serviceName = serviceName;

    // Configure Caffeine cache for optimal performance
    this.nullifierCache =
        Caffeine.newBuilder()
            .maximumSize(maxSize)
            .expireAfterWrite(Duration.ofHours(nullifierExpiryHours))
            .scheduler(Scheduler.systemScheduler()) // Use system scheduler for automatic cleanup
            .removalListener(new NullifierRemovalListener())
            .build();

    LOG.info(
        "{}: High-performance nullifier tracker initialized. MaxSize: {}, TTL: {} hours",
        serviceName,
        maxSize,
        nullifierExpiryHours);
  }

  /**
   * Legacy constructor for backward compatibility with file-based configuration.
   *
   * <p><strong>Note:</strong> The storageFilePath is ignored in this implementation. Nullifiers are
   * stored in memory only for maximum performance.
   *
   * @param serviceName Service name for logging identification
   * @param storageFilePath Ignored - kept for backward compatibility
   * @param nullifierExpiryHours Hours after which nullifiers expire and are evicted
   */
  public NullifierTracker(String serviceName, String storageFilePath, long nullifierExpiryHours) {
    this(serviceName, 1_000_000L, nullifierExpiryHours); // Default to 1M capacity
    LOG.info(
        "{}: Using in-memory nullifier tracking (file path ignored for performance)", serviceName);
  }

  /**
   * Checks if a nullifier has been used before within the given epoch and marks it as used if new.
   *
   * <p><strong>Thread-safe and atomic:</strong> This operation is atomic to prevent race conditions
   * where multiple transactions with the same nullifier could pass validation simultaneously.
   *
   * <p><strong>Epoch Scoping:</strong> Nullifiers are scoped by epoch. The same nullifier can be
   * reused across different epochs but not within the same epoch.
   *
   * @param nullifierHex Hex-encoded nullifier to check/register
   * @param epochId Current epoch identifier for scoping
   * @return true if nullifier is new within this epoch (transaction should be allowed), false if
   *     already used in this epoch
   */
  public boolean checkAndMarkNullifier(String nullifierHex, String epochId) {
    if (nullifierHex == null || nullifierHex.trim().isEmpty()) {
      LOG.warn("{}: Invalid nullifier provided: {}", serviceName, nullifierHex);
      return false;
    }

    if (epochId == null || epochId.trim().isEmpty()) {
      LOG.warn("{}: Invalid epoch ID provided: {}", serviceName, epochId);
      return false;
    }

    String normalizedNullifier = nullifierHex.toLowerCase().trim();
    String normalizedEpochId = epochId.trim();
    String epochScopedKey = normalizedNullifier + ":" + normalizedEpochId;

    Instant now = Instant.now();
    NullifierData nullifierData = new NullifierData(normalizedNullifier, normalizedEpochId, now);

    // Atomic check-and-set using Caffeine's get() with loader pattern
    NullifierData existingData = nullifierCache.get(epochScopedKey, key -> nullifierData);

    if (existingData != nullifierData) {
      // Nullifier was already present (existingData is the previous value)
      nullifierHits.incrementAndGet();
      LOG.warn(
          "{}: Nullifier reuse detected within epoch! Nullifier: {}, Epoch: {}, Previous use: {}",
          serviceName,
          normalizedNullifier,
          normalizedEpochId,
          existingData.timestamp());
      return false;
    }

    // New nullifier for this epoch
    totalNullifiersTracked.incrementAndGet();
    LOG.debug(
        "{}: New nullifier registered: {}, Epoch: {}, Cache size: {}",
        serviceName,
        normalizedNullifier,
        normalizedEpochId,
        nullifierCache.estimatedSize());

    return true;
  }

  /**
   * Checks if a nullifier has been used within the given epoch without marking it as used.
   *
   * @param nullifierHex Hex-encoded nullifier to check
   * @param epochId Epoch identifier for scoping
   * @return true if nullifier has been used within this epoch, false if new
   */
  public boolean isNullifierUsed(String nullifierHex, String epochId) {
    if (nullifierHex == null
        || nullifierHex.trim().isEmpty()
        || epochId == null
        || epochId.trim().isEmpty()) {
      return false;
    }
    String epochScopedKey = nullifierHex.toLowerCase().trim() + ":" + epochId.trim();
    return nullifierCache.getIfPresent(epochScopedKey) != null;
  }

  /**
   * Batch validation of multiple nullifiers for improved performance.
   * Optimized for scenarios where multiple transactions need validation simultaneously.
   *
   * @param nullifierEpochPairs List of nullifier-epoch pairs to validate
   * @return Map of results where key is "nullifier:epoch" and value is validation result
   */
  public Map<String, Boolean> checkAndMarkNullifiersBatch(
      List<Map.Entry<String, String>> nullifierEpochPairs) {
    
    Map<String, Boolean> results = new ConcurrentHashMap<>();
    Instant now = Instant.now();
    
    // Process all pairs in a single pass for better cache efficiency
    for (Map.Entry<String, String> pair : nullifierEpochPairs) {
      String nullifierHex = pair.getKey();
      String epochId = pair.getValue();
      
      if (nullifierHex == null || nullifierHex.trim().isEmpty() || 
          epochId == null || epochId.trim().isEmpty()) {
        results.put(nullifierHex + ":" + epochId, false);
        continue;
      }
      
      String normalizedNullifier = nullifierHex.toLowerCase().trim();
      String normalizedEpochId = epochId.trim();
      String epochScopedKey = normalizedNullifier + ":" + normalizedEpochId;
      
      NullifierData nullifierData = new NullifierData(normalizedNullifier, normalizedEpochId, now);
      NullifierData existingData = nullifierCache.get(epochScopedKey, key -> nullifierData);
      
      boolean isNew = (existingData == nullifierData);
      results.put(epochScopedKey, isNew);
      
      if (isNew) {
        totalNullifiersTracked.incrementAndGet();
      } else {
        nullifierHits.incrementAndGet();
      }
    }
    
    return results;
  }

  /**
   * Gets current statistics for monitoring and debugging.
   *
   * @return Statistics including cache size, total tracked, hits, and expiration count
   */
  public NullifierStats getStats() {
    return new NullifierStats(
        (int) nullifierCache.estimatedSize(),
        totalNullifiersTracked.get(),
        nullifierHits.get(),
        expiredNullifiers.get());
  }

  /** Statistics record for nullifier tracking metrics. */
  public record NullifierStats(
      int currentNullifiers, long totalTracked, long duplicateAttempts, long expiredCount) {}

  /** Removal listener for tracking cache evictions and expiration events. */
  private class NullifierRemovalListener implements RemovalListener<String, NullifierData> {
    @Override
    public void onRemoval(String key, NullifierData value, RemovalCause cause) {
      if (cause == RemovalCause.EXPIRED) {
        expiredNullifiers.incrementAndGet();
        if (LOG.isTraceEnabled()) {
          LOG.trace("{}: Nullifier expired and evicted: {}", serviceName, key);
        }
      }
    }
  }

  @Override
  public void close() throws IOException {
    if (nullifierCache != null) {
      nullifierCache.invalidateAll();
      nullifierCache.cleanUp();
    }
    LOG.info("{}: Nullifier tracker closed. Final stats: {}", serviceName, getStats());
  }
}
