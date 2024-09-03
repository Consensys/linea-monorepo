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
import java.util.List;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import org.jetbrains.annotations.NotNull;

/**
 * Implements a system of pseudo-stacked squashed List where {@link
 * operationsCommitedToTheConflation} represents the List of all operations since the beginning of
 * the conflation and {@link operationsInTransaction} represents the operations added by the last
 * transaction. We can pop only the operations added by last transaction. The line counting is done
 * by a separate {@link CountOnlyOperation}.
 *
 * @param <E> the type of elements stored in the set
 */
@Accessors(fluent = true)
public class StackedList<E extends ModuleOperation> {
  private final List<E> operationsCommitedToTheConflation;
  @Getter private final List<E> operationsInTransaction;
  private final CountOnlyOperation lineCounter = new CountOnlyOperation();
  private boolean conflationFinished = false;

  public StackedList() {
    operationsCommitedToTheConflation = new ArrayList<>();
    operationsInTransaction = new ArrayList<>();
  }

  /** Prefer this constructor as we preallocate more needed memory */
  public StackedList(
      final int expectedConflationNumberOperations, final int expectedTransactionNumberOperations) {
    operationsCommitedToTheConflation = new ArrayList<>(expectedConflationNumberOperations);
    operationsInTransaction = new ArrayList<>(expectedTransactionNumberOperations);
  }

  /**
   * when we enter a transaction, the previous transaction is definitely added to the block and
   * can't be pop
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

  public E getFirst() {
    return operationsCommitedToTheConflation.isEmpty()
        ? operationsInTransaction.getFirst()
        : operationsCommitedToTheConflation.getFirst();
  }

  public E getLast() {
    return operationsInTransaction.isEmpty()
        ? operationsCommitedToTheConflation.getLast()
        : operationsInTransaction.getLast();
  }

  public int size() {
    return operationsInTransaction.size() + operationsCommitedToTheConflation.size();
  }

  public int lineCount() {
    return lineCounter.lineCount();
  }

  public E get(int index) {
    if (index < operationsCommitedToTheConflation.size()) {
      return operationsCommitedToTheConflation.get(index);
    } else {
      return operationsInTransaction.get(index - operationsCommitedToTheConflation.size());
    }
  }

  public List<E> getAll() {
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
    lineCounter.add(e.lineCount());
    return operationsInTransaction.add(e);
  }

  public boolean addAll(@NotNull Collection<? extends E> c) {
    boolean r = false;
    for (var x : c) {
      r |= this.add(x);
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
