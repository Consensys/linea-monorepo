/*
 * Copyright ConsenSys Inc.
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

import java.util.HashSet;
import java.util.Set;

import lombok.Getter;
import lombok.experimental.Accessors;

@Accessors(fluent = true)
@Getter
public class StackedSet<E> {
  private final Set<E> operationsCommitedToTheConflation;
  private final Set<E> operationsInTransactionBundle;

  public StackedSet() {
    operationsCommitedToTheConflation = new HashSet<>();
    operationsInTransactionBundle = new HashSet<>();
  }

  /** Prefer this constructor as we preallocate more needed memory */
  public StackedSet(
      final int expectedConflationNumberOperations, final int expectedTransactionNumberOperations) {
    operationsCommitedToTheConflation = new HashSet<>(expectedConflationNumberOperations);
    operationsInTransactionBundle = new HashSet<>(expectedTransactionNumberOperations);
  }

  public void commitTransactionBundle() {
    operationsCommitedToTheConflation().addAll(operationsInTransactionBundle());
    operationsInTransactionBundle().clear();
  }

  public void popTransactionBundle() {
    operationsInTransactionBundle().clear();
  }

  public boolean add(E e) {
    if (!operationsCommitedToTheConflation().contains(e)) {
      return operationsInTransactionBundle().add(e);
    }
    return false;
  }

  public int size() {
    return operationsInTransactionBundle().size() + operationsCommitedToTheConflation().size();
  }
}
