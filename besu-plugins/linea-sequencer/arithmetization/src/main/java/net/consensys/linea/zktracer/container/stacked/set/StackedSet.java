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

package net.consensys.linea.zktracer.container.stacked.set;

import java.util.Collection;
import java.util.HashSet;
import java.util.Iterator;
import java.util.Set;
import java.util.Stack;

import net.consensys.linea.zktracer.container.StackedContainer;
import org.jetbrains.annotations.NotNull;

/**
 * Implements a system of nested sets behaving as a single one, where the current context
 * modification can transparently be dropped.
 *
 * @param <E> the type of elements stored in the set
 */
public class StackedSet<E> implements StackedContainer, java.util.Set<E> {
  private final Stack<Set<E>> sets = new Stack<>();
  private Set<E> collapsed;
  private boolean dirty = true;

  @Override
  public void enter() {
    this.sets.push(new HashSet<>());
  }

  @Override
  public void pop() {
    this.sets.pop();
    this.dirty = true;
  }

  @Override
  public int size() {
    return this.collapse().size();
  }

  @Override
  public boolean isEmpty() {
    throw new UnsupportedOperationException("empty not supported");
  }

  @Override
  public boolean contains(Object o) {
    for (Set<E> set : this.sets) {
      if (set.contains(o)) {
        return true;
      }
    }
    return false;
  }

  private Set<E> collapse() {
    if (this.dirty) {
      this.collapsed = new HashSet<>();
      for (Set<E> set : this.sets) {
        this.collapsed.addAll(set);
      }
      this.dirty = false;
    }

    return this.collapsed;
  }

  @NotNull
  @Override
  public Iterator<E> iterator() {
    return this.collapse().iterator();
  }

  @NotNull
  @Override
  public Object[] toArray() {
    return this.collapse().toArray();
  }

  @NotNull
  @Override
  public <T> T[] toArray(@NotNull T[] a) {
    throw new UnsupportedOperationException("toArray not supported");
  }

  @Override
  public boolean add(E e) {
    final boolean r = this.sets.peek().add(e);
    this.dirty = true;
    return r;
  }

  @Override
  public boolean remove(Object o) {
    throw new UnsupportedOperationException("remove not supported");
  }

  @Override
  public boolean containsAll(@NotNull Collection<?> c) {
    for (var x : c) {
      if (!this.contains(x)) {
        return false;
      }
    }
    return true;
  }

  @Override
  public boolean addAll(@NotNull Collection<? extends E> c) {
    boolean r = false;
    for (var x : c) {
      r |= this.add(x);
    }
    return r;
  }

  @Override
  public boolean retainAll(@NotNull Collection<?> c) {
    throw new UnsupportedOperationException("retainAll not supported");
  }

  @Override
  public boolean removeAll(@NotNull Collection<?> c) {
    throw new UnsupportedOperationException("removeAll not supported");
  }

  @Override
  public void clear() {
    this.sets.clear();
    this.dirty = true;
  }
}
