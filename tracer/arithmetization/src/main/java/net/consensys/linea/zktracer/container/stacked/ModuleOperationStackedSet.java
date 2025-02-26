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

package net.consensys.linea.zktracer.container.stacked;

import java.util.ArrayList;
import java.util.Collection;
import java.util.Comparator;
import java.util.List;
import java.util.Set;
import java.util.stream.Stream;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import org.jetbrains.annotations.NotNull;

/**
 * Implements a system of pseudo-stacked squashed sets where {@link
 * ModuleOperationStackedSet#operationsCommitedToTheConflation()} represents the set of all
 * operations since the beginning of the conflation and {@link
 * ModuleOperationStackedSet#operationsInTransactionBundle()} represents the operations added by the
 * last transaction. We can pop only the operations added by last transaction. The line counting is
 * done by a separate {@link CountOnlyOperation}.
 *
 * @param <E> the type of elements stored in the set
 */
@Accessors(fluent = true)
public class ModuleOperationStackedSet<E extends ModuleOperation> extends StackedSet<E> {
  private final CountOnlyOperation lineCounter = new CountOnlyOperation();
  @Getter private boolean conflationFinished = false;

  public ModuleOperationStackedSet() {
    super();
  }

  /** Prefer this constructor as we preallocate more needed memory */
  public ModuleOperationStackedSet(
      final int expectedConflationNumberOperations, final int expectedTransactionNumberOperations) {
    super(expectedConflationNumberOperations, expectedTransactionNumberOperations);
  }

  public void commitTransactionBundle() {
    super.commitTransactionBundle();
    lineCounter.commitTransactionBundle();
  }

  public void popTransactionBundle() {
    super.popTransactionBundle();
    lineCounter.popTransactionBundle();
  }

  public int lineCount() {
    return lineCounter.lineCount();
  }

  public Set<E> getAll() {
    Preconditions.checkState(conflationFinished, "Conflation not finished");
    return operationsCommitedToTheConflation();
  }

  public boolean isEmpty() {
    return size() == 0;
  }

  public boolean contains(Object o) {
    return operationsInTransactionBundle().contains(o)
        || operationsCommitedToTheConflation().contains(o);
  }

  public boolean add(E e) {
    if (!operationsCommitedToTheConflation().contains(e)) {
      final boolean isNew = operationsInTransactionBundle().add(e);
      if (isNew) {
        lineCounter.add(e.lineCount());
      }
      return isNew;
    }
    return false;
  }

  public boolean containsAll(@NotNull Collection<?> c) {
    for (var x : c) {
      if (!contains(x)) {
        return false;
      }
    }
    return true;
  }

  public boolean addAll(@NotNull Collection<? extends E> c) {
    boolean r = false;
    for (var x : c) {
      r |= add(x);
    }
    return r;
  }

  public void clear() {
    operationsCommitedToTheConflation().clear();
    operationsInTransactionBundle().clear();
    lineCounter.clear();
  }

  public void finishConflation() {
    conflationFinished = true;
    operationsCommitedToTheConflation().addAll(operationsInTransactionBundle());
    operationsInTransactionBundle().clear();
    lineCounter.commitTransactionBundle(); // this is not mandatory but it is more consistent
  }

  public List<E> sortOperations(Comparator<E> comparator) {
    final List<E> sortedOperations = new ArrayList<>(getAll());
    sortedOperations.sort(comparator);
    return sortedOperations;
  }

  public Stream<E> stream() {
    return this.getAll().stream();
  }
}
