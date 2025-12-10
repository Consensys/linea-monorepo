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
import java.util.Iterator;
import java.util.Set;
import java.util.concurrent.atomic.AtomicReference;
import java.util.function.Supplier;
import org.jetbrains.annotations.NotNull;

/**
 * A thread-safe Set wrapper that supports atomic swapping of the underlying set for configuration
 * reload purposes. Implements Set&lt;T&gt; for drop-in compatibility.
 *
 * <p>This class is immutable from the perspective of Set operations - all mutation methods throw
 * {@link UnsupportedOperationException}. The underlying set can only be changed via {@link
 * #reload().
 *
 * <p>Example usage:
 *
 * <pre>{@code
 * // Create with initial set and reloader that reads from a file
 * ReloadableSet<Address> deniedAddresses = new ReloadableSet<>(
 *     initialSet,
 *     () -> parseDeniedAddresses(config.denyListPath())
 * );
 * <p>
 * // Use as a Set (implements Set<T>)
 * if (deniedAddresses.contains(address)) { ... }
 * <p>
 * // Reload from source when configuration changes
 * deniedAddresses.reload();
 * <p>
 * // Pass to validators that expect AtomicReference
 * new AllowedAddressValidator(deniedAddresses.getReference());
 * }</pre>
 *
 * @param <T> the type of elements in this set
 */
public class ReloadableSet<T> implements Set<T> {
  private final AtomicReference<Set<T>> delegate;
  private final Supplier<Set<T>> reloader;

  /**
   * Creates a new ReloadableSet with an initial set and a reloader function.
   *
   * @param initial the initial set contents
   * @param reloader a supplier that provides a new set when reload is called
   */
  public ReloadableSet(final Set<T> initial, final Supplier<Set<T>> reloader) {
    this.delegate = new AtomicReference<>(initial);
    this.reloader = reloader;
  }

  /** Reload the set from the configured source by invoking the reloader supplier. */
  public void reload() {
    delegate.set(reloader.get());
  }

  /**
   * Get the underlying AtomicReference. This is useful for validators that expect an
   * AtomicReference&lt;Set&lt;T&gt;&gt;.
   *
   * @return the underlying AtomicReference
   */
  public AtomicReference<Set<T>> getReference() {
    return delegate;
  }

  // Delegate all Set read methods to the underlying set

  @Override
  public int size() {
    return delegate.get().size();
  }

  @Override
  public boolean isEmpty() {
    return delegate.get().isEmpty();
  }

  @Override
  public boolean contains(final Object o) {
    return delegate.get().contains(o);
  }

  @NotNull
  @Override
  public Iterator<T> iterator() {
    return delegate.get().iterator();
  }

  @NotNull
  @Override
  public Object[] toArray() {
    return delegate.get().toArray();
  }

  @NotNull
  @Override
  public <T1> T1[] toArray(@NotNull final T1[] a) {
    return delegate.get().toArray(a);
  }

  @Override
  public boolean containsAll(@NotNull final Collection<?> c) {
    return delegate.get().containsAll(c);
  }

  // Mutation methods throw UnsupportedOperationException (immutable view)

  @Override
  public boolean add(final T t) {
    throw new UnsupportedOperationException(
        "ReloadableSet is immutable; use reload() to change contents");
  }

  @Override
  public boolean remove(final Object o) {
    throw new UnsupportedOperationException(
        "ReloadableSet is immutable; use reload() to change contents");
  }

  @Override
  public boolean addAll(@NotNull final Collection<? extends T> c) {
    throw new UnsupportedOperationException(
        "ReloadableSet is immutable; use reload() to change contents");
  }

  @Override
  public boolean retainAll(@NotNull final Collection<?> c) {
    throw new UnsupportedOperationException(
        "ReloadableSet is immutable; use reload() to change contents");
  }

  @Override
  public boolean removeAll(@NotNull final Collection<?> c) {
    throw new UnsupportedOperationException(
        "ReloadableSet is immutable; use reload() to change contents");
  }

  @Override
  public void clear() {
    throw new UnsupportedOperationException(
        "ReloadableSet is immutable; use reload() to change contents");
  }
}
