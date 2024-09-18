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

import java.util.Collection;
import java.util.HashSet;
import java.util.Set;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import org.jetbrains.annotations.NotNull;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Implements a system of pseudo-stacked squashed sets where {@link
 * operationsCommitedToTheConflation} represents the set of all operations since the beginning of
 * the conflation and {@link operationsInTransaction} represents the operations added by the last
 * transaction. We can pop only the operations added by last transaction. The line counting is done
 * by a separate {@link CountOnlyOperation}.
 *
 * @param <E> the type of elements stored in the set
 */
@Accessors(fluent = true)
public class StackedSet<E extends ModuleOperation> {
  private static final Logger log = LoggerFactory.getLogger(StackedSet.class);
  @Getter private final Set<E> operationsCommitedToTheConflation;
  @Getter private final Set<E> operationsInTransaction;
  private final CountOnlyOperation lineCounter = new CountOnlyOperation();
  @Getter private boolean conflationFinished = false;

  public StackedSet() {
    operationsCommitedToTheConflation = new HashSet<>();
    operationsInTransaction = new HashSet<>();
  }

  /** Prefer this constructor as we preallocate more needed memory */
  public StackedSet(
      final int expectedConflationNumberOperations, final int expectedTransactionNumberOperations) {
    operationsCommitedToTheConflation = new HashSet<>(expectedConflationNumberOperations);
    operationsInTransaction = new HashSet<>(expectedTransactionNumberOperations);
  }

  /**
   * When we enter transaction n, the previous transaction n-1 {@link operationsInTransaction} (or
   * an empty set if n == 0) is definitely added to the set of all transactions {@link
   * operationsCommitedToTheConflation}. We reset {@link operationsInTransaction} to an empty set as
   * it now represents the operations added at transaction n.
   */
  public void enter() {
    operationsCommitedToTheConflation.addAll(operationsInTransaction);
    operationsInTransaction.clear();
    lineCounter.enter();
  }

  public void pop() {
    operationsInTransaction.clear();
    lineCounter.pop();
  }

  public int size() {
    return operationsInTransaction.size() + operationsCommitedToTheConflation.size();
  }

  public int lineCount() {
    return lineCounter.lineCount();
  }

  public Set<E> getAll() {
    Preconditions.checkState(conflationFinished, "Conflation not finished");
    return operationsCommitedToTheConflation;
  }

  public boolean isEmpty() {
    return size() == 0;
  }

  public boolean contains(Object o) {
    return operationsInTransaction.contains(o) || operationsCommitedToTheConflation.contains(o);
  }

  public boolean add(E e) {
    if (!operationsCommitedToTheConflation.contains(e)) {
      final boolean isNew = operationsInTransaction.add(e);
      if (isNew) {
        lineCounter.add(e.lineCount());
      }
      return isNew;
    } else {
      log.trace(
          "Operation of type {} was already in operationsCommitedToTheConflation hashset, reference is ",
          e.getClass().getName(),
          e);
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
    operationsCommitedToTheConflation.clear();
    operationsInTransaction.clear();
    lineCounter.clear();
  }

  public void finishConflation() {
    conflationFinished = true;
    operationsCommitedToTheConflation.addAll(operationsInTransaction);
    operationsInTransaction.clear();
    lineCounter.enter(); // this is not mandatory but it is more consistent
  }
}
