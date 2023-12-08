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

import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Deque;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.Set;

import net.consensys.linea.zktracer.container.StackedContainer;
import org.jetbrains.annotations.NotNull;

/**
 * Implements a system of nested sets behaving as a single one, where the current context
 * modification can transparently be dropped.
 *
 * @param <E> the type of elements stored in the set
 */
public class StackedSet<E> implements StackedContainer, java.util.Set<E> {
  private final Deque<Set<E>> sets = new ArrayDeque<>();
  private final Map<E, Integer> occurences = new HashMap<>();

  @Override
  public void enter() {
    this.sets.push(new HashSet<>());
  }

  @Override
  public void pop() {
    Set<E> set = this.sets.pop();
    for (E e : set) {
      Integer count = occurences.get(e);
      if (count > 0) occurences.put(e, count - 1);
      else throw new IllegalStateException("asymetric element removal !");
    }
  }

  @Override
  public int size() {
    int size = 0;
    for (Integer count : occurences.values()) {
      if (count != 0) {
        size++;
      }
    }
    return size;
  }

  @Override
  public boolean isEmpty() {
    throw new UnsupportedOperationException("empty not supported");
  }

  @Override
  public boolean contains(Object o) {
    return occurences.containsKey(o) && occurences.get(o) > 0;
  }

  @NotNull
  @Override
  public Iterator<E> iterator() {
    List<E> list = new ArrayList<>();
    for (Map.Entry<E, Integer> entry : occurences.entrySet()) {
      if (entry.getValue() > 0) {
        list.add(entry.getKey());
      }
    }
    return list.iterator();
  }

  @NotNull
  @Override
  @SuppressWarnings("unchecked")
  public E[] toArray() {
    return occurences.entrySet().stream()
        .filter(entry -> entry.getValue() > 0)
        .map(Map.Entry::getKey)
        .toArray(size -> (E[]) new Object[size]);
  }

  @NotNull
  @Override
  public <T> T[] toArray(@NotNull T[] a) {
    throw new UnsupportedOperationException("toArray not supported");
  }

  @Override
  public boolean add(E e) {
    final boolean added = this.sets.peek().add(e);
    occurences.put(e, occurences.getOrDefault(e, 0) + 1);
    return added;
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
    this.occurences.clear();
  }
}
