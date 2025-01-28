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

package net.consensys.linea.zktracer.module.euc;

import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.OperationSetModule;
import net.consensys.linea.zktracer.container.stacked.CountOnlyOperation;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedSet;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.apache.tuweni.bytes.Bytes;

@RequiredArgsConstructor
@Accessors(fluent = true)
public class Euc implements OperationSetModule<EucOperation> {
  private final Wcp wcp;

  @Getter
  private final ModuleOperationStackedSet<EucOperation> operations =
      new ModuleOperationStackedSet<>();

  /** count the number of rows that could be added after the sequencer counts the number of line */
  public final CountOnlyOperation additionalRows = new CountOnlyOperation();

  @Override
  public String moduleKey() {
    return "EUC";
  }

  @Override
  public void enterTransaction() {
    OperationSetModule.super.enterTransaction();
    additionalRows.lineCount();
  }

  @Override
  public void popTransaction() {
    OperationSetModule.super.popTransaction();
    additionalRows.pop();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    for (EucOperation eucOperation : operations.sortOperations(new EucOperationComparator())) {
      eucOperation.trace(trace);
    }
  }

  @Override
  public int lineCount() {
    return operations.conflationFinished()
        ? operations.lineCount()
        : operations().lineCount() + additionalRows.lineCount();
  }

  public EucOperation callEUC(final Bytes dividend, final Bytes divisor) {
    final BigInteger dividendBI = dividend.toUnsignedBigInteger();
    final BigInteger divisorBI = divisor.toUnsignedBigInteger();
    final Bytes quotient = bigIntegerToBytes(dividendBI.divide(divisorBI));
    final Bytes remainder = bigIntegerToBytes(dividendBI.remainder(divisorBI));

    final EucOperation operation = new EucOperation(dividend, divisor, quotient, remainder);

    final boolean isNew = operations.add(operation);
    if (isNew) {
      wcp.callLT(operation.remainder(), operation.divisor());
    }

    return operation;
  }
}
