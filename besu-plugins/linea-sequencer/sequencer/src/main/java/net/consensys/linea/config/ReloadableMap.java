/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.config;

import java.util.Collection;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.atomic.AtomicReference;
import java.util.function.Supplier;
import org.jetbrains.annotations.NotNull;

/**
 * A thread-safe Map wrapper that supports atomic swapping of the underlying map for configuration
 * reload purposes. Implements Map&lt;K,V&gt; for drop-in compatibility.
 *
 * <p>This class is immutable from the perspective of Map operations - all mutation methods throw
 * {@link UnsupportedOperationException}. The underlying map can only be changed via {@link
 * #reload()}.
 *
 * <p>Example usage:
 *
 * <pre>{@code
 * // Create with initial map and reloader that reads from a file
 * ReloadableMap<Address, Set<EventFilter>> deniedEvents = new ReloadableMap<>(
 *     initialMap,
 *     () -> parseEventsDenyList(config.eventsDenyListPath())
 * );
 *
 * // Use as a Map (implements Map<K,V>)
 * Set<EventFilter> filters = deniedEvents.get(address);
 *
 * // Reload from source when configuration changes
 * deniedEvents.reload();
 *
 * // Pass to selectors that expect AtomicReference
 * new TransactionEventSelector(deniedEvents.getReference());
 * }</pre>
 *
 * @param <K> the type of keys in this map
 * @param <V> the type of values in this map
 */
public class ReloadableMap<K, V> implements Map<K, V> {
  private final AtomicReference<Map<K, V>> delegate;
  private final Supplier<Map<K, V>> reloader;

  /**
   * Creates a new ReloadableMap with an initial map and a reloader function.
   *
   * @param initial the initial map contents
   * @param reloader a supplier that provides a new map when reload is called
   */
  public ReloadableMap(final Map<K, V> initial, final Supplier<Map<K, V>> reloader) {
    this.delegate = new AtomicReference<>(initial);
    this.reloader = reloader;
  }

  /** Reload the map from the configured source by invoking the reloader supplier. */
  public void reload() {
    delegate.set(reloader.get());
  }

  /**
   * Get the underlying AtomicReference. This is useful for components that expect an
   * AtomicReference&lt;Map&lt;K,V&gt;&gt;.
   *
   * @return the underlying AtomicReference
   */
  public AtomicReference<Map<K, V>> getReference() {
    return delegate;
  }

  // Delegate all Map read methods to the underlying map

  @Override
  public int size() {
    return delegate.get().size();
  }

  @Override
  public boolean isEmpty() {
    return delegate.get().isEmpty();
  }

  @Override
  public boolean containsKey(final Object key) {
    return delegate.get().containsKey(key);
  }

  @Override
  public boolean containsValue(final Object value) {
    return delegate.get().containsValue(value);
  }

  @Override
  public V get(final Object key) {
    return delegate.get().get(key);
  }

  @NotNull
  @Override
  public Set<K> keySet() {
    return delegate.get().keySet();
  }

  @NotNull
  @Override
  public Collection<V> values() {
    return delegate.get().values();
  }

  @NotNull
  @Override
  public Set<Entry<K, V>> entrySet() {
    return delegate.get().entrySet();
  }

  // Mutation methods throw UnsupportedOperationException (immutable view)

  @Override
  public V put(final K key, final V value) {
    throw new UnsupportedOperationException(
        "ReloadableMap is immutable; use reload() to change contents");
  }

  @Override
  public V remove(final Object key) {
    throw new UnsupportedOperationException(
        "ReloadableMap is immutable; use reload() to change contents");
  }

  @Override
  public void putAll(@NotNull final Map<? extends K, ? extends V> m) {
    throw new UnsupportedOperationException(
        "ReloadableMap is immutable; use reload() to change contents");
  }

  @Override
  public void clear() {
    throw new UnsupportedOperationException(
        "ReloadableMap is immutable; use reload() to change contents");
  }
}
