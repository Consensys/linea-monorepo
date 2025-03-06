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

import java.util.List;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.container.stacked.CountOnlyOperation;
import net.consensys.linea.zktracer.module.hub.Hub;

@Getter
@Accessors(fluent = true)
@RequiredArgsConstructor
public class BlockTransactions implements Module {
  private final Hub hub;
  private final CountOnlyOperation counts = new CountOnlyOperation();

  @Override
  public String moduleKey() {
    return "BLOCK_TRANSACTIONS";
  }

  @Override
  public void popTransactionBundle() {}

  @Override
  public void commitTransactionBundle() {}

  @Override
  public int lineCount() {
    final int hubNumberOfTx = hub.state.txCount();
    final int txnDataNumberOfTx = hub.txnData().operations().size();
    assert (hubNumberOfTx == txnDataNumberOfTx);

    return txnDataNumberOfTx;
  }

  @Override
  public List<Trace.ColumnHeader> columnHeaders() {
    throw new UnsupportedOperationException("Not implemented");
  }
}
