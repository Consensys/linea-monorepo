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

package net.consensys.linea.zktracer.container.stacked.list;

import java.util.ArrayList;
import java.util.Collection;
import java.util.Iterator;
import java.util.List;
import java.util.ListIterator;
import java.util.NoSuchElementException;

import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.container.StackedContainer;
import org.jetbrains.annotations.NotNull;

/**
 * Implements a system of nested lists behaving as a single one, where the current context
 * modification can transparently be dropped.
 *
 * @param <E> the type of elements stored in the list
 */
public class StackedList<E extends ModuleOperation> implements List<E>, StackedContainer {

  private final List<CountedList<E>> lists = new ArrayList<>();
  /** The cached number of elements in this container */
  private int totalSize;

  @Override
  public String toString() {
    StringBuilder r = new StringBuilder();
    r.append("[[");
    for (var l : this.lists) {
      r.append(l.toString());
    }
    r.append("]]");
    return r.toString();
  }

  @Override
  public void enter() {
    this.lists.add(new CountedList<>());
  }

  public void enter(int initialCapacity) {
    this.lists.add(new CountedList<>(initialCapacity));
  }

  public E getLast() {
    return this.lists.get(this.lists.size() - 1).getLast();
  }

  @Override
  public void pop() {
    if (this.lists.isEmpty()) {
      throw new RuntimeException("asymmetric pop");
    }
    this.totalSize -= this.lists.remove(this.lists.size() - 1).size();
  }

  @Override
  public int size() {
    return this.totalSize;
  }

  public int lineCount() {
    int sum = 0;
    // Deliberate use of old-style for loops out of performances concerns
    for (int i = 0; i < this.lists.size(); i++) {
      sum += this.lists.get(i).lineCount();
    }
    return sum;
  }

  @Override
  public boolean isEmpty() {
    return this.totalSize == 0;
  }

  @Override
  public boolean contains(Object o) {
    for (CountedList<E> l : this.lists) {
      if (l.contains(o)) {
        return true;
      }
    }

    return false;
  }

  @NotNull
  @Override
  public Iterator<E> iterator() {
    return new StackedListIterator<>(this);
  }

  @NotNull
  @Override
  public Object[] toArray() {
    throw new UnsupportedOperationException("toArray is not supported");
  }

  @NotNull
  @Override
  public <T> T[] toArray(@NotNull T[] a) {
    throw new UnsupportedOperationException("toArray is not supported");
  }

  @Override
  public boolean add(E e) {
    this.totalSize++;
    return this.lists.get(this.lists.size() - 1).add(e);
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
  public boolean addAll(int index, @NotNull Collection<? extends E> c) {
    throw new UnsupportedOperationException("addAll(int) not supported");
  }

  @Override
  public boolean removeAll(@NotNull Collection<?> c) {
    boolean r = false;
    for (var x : c) {
      r |= this.remove(x);
    }
    return r;
  }

  @Override
  public boolean retainAll(@NotNull Collection<?> c) {
    throw new UnsupportedOperationException("retainAll not supported");
  }

  @Override
  public void clear() {
    this.lists.clear();
    this.totalSize = 0;
  }

  @Override
  public E get(int i) {
    for (CountedList<E> list : this.lists) {
      if (i >= list.size()) {
        i -= list.size();
      } else {
        return list.get(i);
      }
    }
    return null;
  }

  @Override
  public E set(int index, E element) {
    throw new UnsupportedOperationException("set not supported");
  }

  @Override
  public void add(int index, E element) {
    throw new UnsupportedOperationException("add(int) not supported");
  }

  @Override
  public E remove(int index) {
    throw new UnsupportedOperationException("remove not supported");
  }

  @Override
  public int indexOf(Object o) {
    int i = 0;
    for (List<E> l : this.lists) {
      final int ii = l.indexOf(o);
      if (ii != -1) {
        return i + ii;
      }
      i += l.size();
    }
    return -1;
  }

  @Override
  public int lastIndexOf(Object o) {
    throw new UnsupportedOperationException("lastIndexOf not supported");
  }

  @NotNull
  @Override
  public ListIterator<E> listIterator() {
    return new StackedListIterator<>(this);
  }

  @NotNull
  @Override
  public ListIterator<E> listIterator(int index) {
    throw new UnsupportedOperationException("listIterator(int) not supported");
  }

  @NotNull
  @Override
  public List<E> subList(int fromIndex, int toIndex) {
    throw new UnsupportedOperationException("subList not supported");
  }

  private static class StackedListIterator<F extends ModuleOperation>
      implements Iterator<F>, ListIterator<F> {
    private final StackedList<F> sl;
    /** Position of the iterator in the list of lists */
    private int head = 0;
    /** Position of the iterator within the current list, i.e. this.sl.lists[head] */
    private int offset = -1;

    StackedListIterator(StackedList<F> lists) {
      this.sl = lists;
    }

    private List<F> list() {
      if (sl.lists.isEmpty()) {
        return List.of();
      }
      return this.sl.lists.get(this.head);
    }

    private List<F> list(int i) {
      return this.sl.lists.get(i);
    }

    F read() {
      return this.sl.lists.get(head).get(this.offset);
    }

    int toIndex() {
      int idx = 0;
      for (int i = 0; i < this.head; i++) {
        idx += this.list(i).size();
      }
      return idx + this.offset;
    }

    boolean done() {
      return this.head >= this.sl.lists.size() - 1 && this.offset >= this.list().size();
    }

    @Override
    public boolean hasNext() {
      // First check if the current list is done for
      if (offset < this.list().size() - 1) {
        return true;
      }
      // Then check for a next non-empty list
      for (int i = this.head + 1; i < this.sl.lists.size(); i++) {
        if (!this.list(i).isEmpty()) {
          return true;
        }
      }
      // We're empty
      return false;
    }

    @Override
    public F next() {
      // If we are not yet at the end of the current list, return the natural next element
      if (offset < this.list().size() - 1) {
        this.offset++;
        return this.read();
      }

      this.head++;
      // Otherwise, jump to the next non-empty list if we can
      for (int i = head; i < this.sl.lists.size(); i++) {
        if (!this.list(i).isEmpty()) {
          this.head = i;
          this.offset = 0;
          return this.read();
        }
      }

      this.head = this.sl.lists.size() - 1;
      this.offset = this.list().size() - 1;
      throw new NoSuchElementException();
    }

    @Override
    public boolean hasPrevious() {
      if (offset > 0) {
        return true;
      }

      if (head > 0) {
        for (int i = head - 1; i >= 0; i--) {
          if (!this.list(i).isEmpty()) {
            return true;
          }
        }
      }

      return false;
    }

    @Override
    public F previous() {
      // If we are not yet at the very beginning of the current list, return the natural previous
      // element
      if (offset > 0) {
        this.offset--;
        return this.read();
      }

      // Otherwise, jump to the previous non-empty list if we can
      for (int i = head - 1; i >= 0; i--) {
        if (!this.list(i).isEmpty()) {
          this.head = i;
          this.offset = this.list().size() - 1;
          return this.read();
        }
      }

      this.head = 0;
      this.offset = 0;
      return null;
    }

    @Override
    public int nextIndex() {
      final int idx = this.toIndex();
      if (this.done()) {
        return idx;
      }
      return idx + 1;
    }

    @Override
    public int previousIndex() {
      final int idx = this.toIndex();
      if (idx > 0) {
        return idx - 1;
      }
      return 0;
    }

    @Override
    public void remove() {
      throw new UnsupportedOperationException("remove not supported");
    }

    @Override
    public void set(F e) {
      this.sl.lists.get(head).set(this.offset, e);
    }

    @Override
    public void add(F f) {
      throw new UnsupportedOperationException("add not supported");
    }
  }
}
