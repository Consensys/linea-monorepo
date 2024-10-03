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

package net.consensys.linea.zktracer.module.gas;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.OperationSetModule;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedSet;
import org.hyperledger.besu.evm.frame.MessageFrame;

@Accessors(fluent = true)
public class Gas implements OperationSetModule<GasOperation> {
  /** A list of the operations to trace */
  @Getter
  private final ModuleOperationStackedSet<GasOperation> operations =
      new ModuleOperationStackedSet<>();

  @Override
  public String moduleKey() {
    return "GAS";
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return null;
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    GasParameters gasParameters = extractGasParameters(frame);
    this.operations.add(new GasOperation(gasParameters));
  }

  private GasParameters extractGasParameters(MessageFrame frame) {
    // TODO: fill it with the actual values
    return new GasParameters(0, BigInteger.ZERO, BigInteger.ZERO, false, false);
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    int stamp = 0;
    for (GasOperation gasOperation : operations.sortOperations(new GasOperationComparator())) {
      // TODO: I thought we don't have stamp for gas anymore ?
      stamp++;
      gasOperation.trace(stamp, trace);
    }
  }
}
