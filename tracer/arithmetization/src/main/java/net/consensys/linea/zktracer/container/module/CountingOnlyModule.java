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
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.stacked.CountOnlyOperation;
import net.consensys.linea.zktracer.module.limits.CountingModuleName;

/** A {@link CountingOnlyModule} is a {@link Module} that only counts certain outcomes. */
@RequiredArgsConstructor
public class CountingOnlyModule implements Module {
  private final CountingModuleName moduleKey;

  protected final CountOnlyOperation counts = new CountOnlyOperation();

  @Override
  public void commitTransactionBundle() {
    counts.commitTransactionBundle();
  }

  @Override
  public String moduleKey() {
    return moduleKey.toString();
  }

  @Override
  public void popTransactionBundle() {
    counts.popTransactionBundle();
  }

  @Override
  public int lineCount() {
    return counts.lineCount();
  }

  @Override
  public int spillage(Trace trace) {
    return 0;
  }

  public void updateTally(final int count) {
    Preconditions.checkArgument(count >= 0, "Must be non-negative");
    counts.add(count);
  }

  @Override
  public List<Trace.ColumnHeader> columnHeaders(Trace trace) {
    throw new IllegalStateException("should never be called");
  }

  @Override
  public void commit(Trace trace) {
    throw new IllegalStateException("should never be called");
  }
}
