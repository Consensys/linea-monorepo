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

package net.consensys.linea.zktracer.module.limits;

import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.CountOnlyOperation;
import net.consensys.linea.zktracer.module.Module;

public class CountingOnlyModule implements Module {
  public final CountOnlyOperation counts = new CountOnlyOperation();

  @Override
  public String moduleKey() {
    throw new IllegalStateException("must have been implemented by the extended class");
  }

  @Override
  public void enterTransaction() {
    counts.enterTransaction();
  }

  @Override
  public void popTransaction() {
    counts.popTransaction();
  }

  @Override
  public int lineCount() {
    return counts.lineCount();
  }

  public void addPrecompileLimit(final int input) {
    throw new IllegalStateException("must have been implemented by the extended class");
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    throw new IllegalStateException("should never be called");
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    throw new IllegalStateException("should never be called");
  }
}
