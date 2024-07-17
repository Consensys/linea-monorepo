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

package net.consensys.linea.zktracer.module.exp;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_ISZERO;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EXP_INST_EXPLOG;
import static net.consensys.linea.zktracer.module.exp.Trace.CT_MAX_CMPTN_EXP_LOG;
import static net.consensys.linea.zktracer.module.exp.Trace.CT_MAX_PRPRC_EXP_LOG;
import static net.consensys.linea.zktracer.opcode.gas.GasConstants.G_EXP_BYTE;

import lombok.EqualsAndHashCode;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.exp.ExpCallForExpPricing;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

@RequiredArgsConstructor
@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
public class ExpLogOperation extends ExpOperation {
  @EqualsAndHashCode.Include private final EWord exponent;
  private final long dynCost;

  @Override
  protected boolean isExpLog() {
    return true;
  }

  public static ExpLogOperation fromExpLogCall(final Wcp wcp, final ExpCallForExpPricing c) {
    final ExpLogOperation r =
        new ExpLogOperation(c.exponent(), (long) G_EXP_BYTE.cost() * c.exponent().byteLength());
    r.wcp = wcp;
    r.preCompute();

    return r;
  }

  @Override
  public void preCompute() {
    pMacroExpInst = EXP_INST_EXPLOG;
    pMacroData1 = this.exponent.hi();
    pMacroData2 = this.exponent.lo();
    pMacroData5 = Bytes.ofUnsignedLong(this.dynCost);
    initArrays(CT_MAX_PRPRC_EXP_LOG + 1);

    // Preprocessing
    // First row
    pPreprocessingWcpFlag[0] = true;
    pPreprocessingWcpArg1Hi[0] = Bytes.EMPTY;
    pPreprocessingWcpArg1Lo[0] = this.exponent.hi();
    pPreprocessingWcpArg2Hi[0] = Bytes.EMPTY;
    pPreprocessingWcpArg2Lo[0] = Bytes.EMPTY;
    pPreprocessingWcpInst[0] = UnsignedByte.of(EVM_INST_ISZERO);
    final boolean expnHiIsZero = wcp.callISZERO(this.exponent.hi());
    ;
    pPreprocessingWcpRes[0] = expnHiIsZero;

    // Linking constraints and fill rawAcc
    pComputationPltJmp = 16;
    pComputationRawAcc = this.exponent.hi();
    if (expnHiIsZero) {
      pComputationRawAcc = this.exponent.lo();
    }

    // Fill trimAcc
    short maxCt = (short) CT_MAX_CMPTN_EXP_LOG;
    for (short i = 0; i < maxCt + 1; i++) {
      boolean pltBit = i >= pComputationPltJmp;
      byte rawByte = pComputationRawAcc.get(i);
      byte trimByte = pltBit ? 0 : rawByte;
      pComputationTrimAcc = Bytes.concatenate(pComputationTrimAcc, Bytes.of(trimByte));
    }
  }
}
