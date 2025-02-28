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

import static net.consensys.linea.zktracer.Trace.EVM_INST_LT;
import static net.consensys.linea.zktracer.Trace.WCP_INST_LEQ;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Utils.initArray;

import java.math.BigInteger;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.UnsignedByte;

@Accessors(fluent = true)
@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
public class GasOperation extends ModuleOperation {
  @EqualsAndHashCode.Include @Getter GasParameters gasParameters;
  BigInteger[] wcpArg1Lo;
  BigInteger[] wcpArg2Lo;
  UnsignedByte[] wcpInst;
  boolean[] wcpRes;
  int ctMax;

  public GasOperation(GasParameters gasParameters, Wcp wcp) {
    this.gasParameters = gasParameters;
    ctMax = compareGasActualAndGasCost() ? 2 : 1;

    // init arrays
    wcpArg1Lo = initArray(BigInteger.ZERO, ctMax + 1);
    wcpArg2Lo = initArray(BigInteger.ZERO, ctMax + 1);
    wcpInst = initArray(UnsignedByte.of(0), ctMax + 1);
    wcpRes = new boolean[ctMax + 1];

    // row 0
    wcpArg1Lo[0] = BigInteger.ZERO;
    wcpArg2Lo[0] = gasParameters.gasActual();
    wcpInst[0] = UnsignedByte.of(WCP_INST_LEQ);
    final boolean gasActualIsNonNegative = wcp.callLEQ(0, gasParameters.gasActual().longValue());
    wcpRes[0] = gasActualIsNonNegative; // supposed to be true

    // row 1
    wcpArg1Lo[1] = BigInteger.ZERO;
    wcpArg2Lo[1] = gasParameters.gasCost();
    wcpInst[1] = UnsignedByte.of(WCP_INST_LEQ);
    final boolean gasCostIsNonNegative = wcp.callLEQ(0, gasParameters.gasCost().longValue());
    wcpRes[1] = gasCostIsNonNegative; // supposed to be true

    // row 2
    if (compareGasActualAndGasCost()) {
      wcpArg1Lo[2] = gasParameters.gasActual();
      wcpArg2Lo[2] = gasParameters.gasCost();
      wcpInst[2] = UnsignedByte.of(EVM_INST_LT);
      final boolean gasActualLTGasCost =
          wcp.callLT(gasParameters.gasActual().longValue(), gasParameters.gasCost().longValue());
      wcpRes[2] = gasActualLTGasCost; // supposed to be equal to gasParameters.isOogx()
    }
  }

  private boolean compareGasActualAndGasCost() {
    return !gasParameters.xahoy() || gasParameters.oogx();
  }

  @Override
  protected int computeLineCount() {
    return ctMax + 1;
  }

  public void trace(Trace.Gas trace) {
    for (short i = 0; i < ctMax + 1; i++) {
      trace
          .inputsAndOutputsAreMeaningful(true)
          .first(i == 0)
          .ct(i)
          .ctMax(ctMax)
          .gasActual(bigIntegerToBytes(gasParameters.gasActual()))
          .gasCost(bigIntegerToBytes(gasParameters.gasCost()))
          .exceptionsAhoy(gasParameters.xahoy())
          .outOfGasException(gasParameters.oogx())
          .wcpArg1Lo(bigIntegerToBytes(wcpArg1Lo[i]))
          .wcpArg2Lo(bigIntegerToBytes(wcpArg2Lo[i]))
          .wcpInst(wcpInst[i])
          .wcpRes(wcpRes[i])
          .validateRow();
    }
  }
}
