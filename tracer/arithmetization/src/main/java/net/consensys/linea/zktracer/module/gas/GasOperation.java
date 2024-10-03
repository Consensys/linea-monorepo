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

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_LT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.WCP_INST_LEQ;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.math.BigInteger;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.types.UnsignedByte;

@Accessors(fluent = true)
@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
public class GasOperation extends ModuleOperation {
  @EqualsAndHashCode.Include @Getter GasParameters gasParameters;
  BigInteger[] wcpArg1Lo;
  BigInteger[] wcpArg2Lo;
  UnsignedByte[] wcpInst;
  boolean[] wcpRes;
  int CT_MAX;

  public GasOperation(GasParameters gasParameters) {
    this.gasParameters = gasParameters;
    CT_MAX = gasParameters.ctMax();

    // init arrays
    wcpArg1Lo = new BigInteger[CT_MAX + 1];
    wcpArg2Lo = new BigInteger[CT_MAX + 1];
    wcpInst = new UnsignedByte[CT_MAX + 1];
    wcpRes = new boolean[CT_MAX + 1];

    // row 0
    wcpArg1Lo[0] = BigInteger.ZERO;
    wcpArg2Lo[0] = gasParameters.gasActual();
    wcpInst[0] = UnsignedByte.of(WCP_INST_LEQ);
    wcpRes[0] = true;

    // row 1
    wcpArg1Lo[1] = BigInteger.ZERO;
    wcpArg2Lo[1] = gasParameters.gasCost();
    wcpInst[1] = UnsignedByte.of(WCP_INST_LEQ);
    wcpRes[1] = true;

    // row 2
    if (gasParameters.oogx()) {
      wcpArg1Lo[2] = gasParameters.gasActual();
      wcpArg2Lo[2] = gasParameters.gasCost();
      wcpInst[2] = UnsignedByte.of(EVM_INST_LT);
      wcpRes[2] = gasParameters.oogx();
    } else {
      // TODO: init the lists with zeros (or something equivalent) instead
      wcpArg1Lo[2] = BigInteger.ZERO;
      wcpArg2Lo[2] = BigInteger.ZERO;
      wcpInst[2] = UnsignedByte.of(0);
      wcpRes[2] = false;
    }
  }

  @Override
  protected int computeLineCount() {
    return gasParameters.ctMax() + 1;
  }

  public void trace(int stamp, Trace trace) {
    for (short i = 0; i < CT_MAX + 1; i++) {
      // TODO: review traced values
      trace
          .inputsAndOutputsAreMeaningful(stamp != 0)
          .first(i == 0)
          .ct(i)
          .ctMax(CT_MAX)
          .gasActual(gasParameters.gasActual().longValue())
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
