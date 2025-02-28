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

package net.consensys.linea.zktracer.container.module;

import java.util.List;

import com.google.common.base.Preconditions;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.stacked.CountOnlyOperation;

/** A {@link CountingOnlyModule} is a {@link Module} that only counts certain outcomes. */
public interface CountingOnlyModule extends Module {
  CountOnlyOperation counts();

  @Override
  default void commitTransactionBundle() {
    counts().commitTransactionBundle();
  }

  @Override
  default void popTransactionBundle() {
    counts().popTransactionBundle();
  }

  @Override
  default int lineCount() {
    return counts().lineCount();
  }

  default void addPrecompileLimit(final int count) {
    Preconditions.checkArgument(count >= 0, "Must be positive");
    counts().add(count);
  }

  @Override
  default List<Trace.ColumnHeader> columnHeaders() {
    throw new IllegalStateException("should never be called");
  }

  @Override
  default void commit(Trace trace) {
    throw new IllegalStateException("should never be called");
  }
}
