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

import static net.consensys.linea.zktracer.module.exp.Trace.EXP_EXPLOG;
import static net.consensys.linea.zktracer.module.exp.Trace.ISZERO;
import static net.consensys.linea.zktracer.module.exp.Trace.MAX_CT_CMPTN_EXP_LOG;
import static net.consensys.linea.zktracer.module.exp.Trace.MAX_CT_PRPRC_EXP_LOG;
import static net.consensys.linea.zktracer.opcode.gas.GasConstants.G_EXP_BYTE;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.ExpLogCall;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;

@RequiredArgsConstructor
public class ExpLogChunk extends ExpChunk {
  private final EWord exponent;
  private final long dynCost;

  @Override
  protected boolean isExpLog() {
    return true;
  }

  public static ExpLogChunk fromMessageFrame(final Wcp wcp, final MessageFrame frame) {
    EWord exponent = EWord.of(frame.getStackItem(1));
    final ExpLogChunk r =
        new ExpLogChunk(exponent, (long) G_EXP_BYTE.cost() * exponent.byteLength());
    r.wcp = wcp;
    r.preCompute();

    return r;
  }

  public static ExpLogChunk fromExpLogCall(final Wcp wcp, final ExpLogCall c) {
    final ExpLogChunk r =
        new ExpLogChunk(c.exponent(), (long) G_EXP_BYTE.cost() * c.exponent().byteLength());
    r.wcp = wcp;
    r.preCompute();

    return r;
  }

  @Override
  public void preCompute() {
    pMacroExpInst = EXP_EXPLOG;
    pMacroData1 = this.exponent.hi();
    pMacroData2 = this.exponent.lo();
    pMacroData5 = Bytes.ofUnsignedLong(this.dynCost);
    initArrays(MAX_CT_PRPRC_EXP_LOG + 1);

    // Preprocessing
    // First row
    pPreprocessingWcpFlag[0] = true;
    pPreprocessingWcpArg1Hi[0] = Bytes.EMPTY;
    pPreprocessingWcpArg1Lo[0] = this.exponent.hi();
    pPreprocessingWcpArg2Hi[0] = Bytes.EMPTY;
    pPreprocessingWcpArg2Lo[0] = Bytes.EMPTY;
    pPreprocessingWcpInst[0] = UnsignedByte.of(ISZERO);
    pPreprocessingWcpRes[0] = this.exponent.hi().isZero();
    final boolean expnHiIsZero = pPreprocessingWcpRes[0];

    // Lookup
    wcp.callISZERO(this.exponent.hi());

    // Linking constraints and fill rawAcc
    pComputationPltJmp = 16;
    pComputationRawAcc = this.exponent.hi();
    if (expnHiIsZero) {
      pComputationRawAcc = this.exponent.lo();
    }

    // Fill trimAcc
    short maxCt = (short) MAX_CT_CMPTN_EXP_LOG;
    for (short i = 0; i < maxCt + 1; i++) {
      boolean pltBit = i >= pComputationPltJmp;
      byte rawByte = pComputationRawAcc.get(i);
      byte trimByte = pltBit ? 0 : rawByte;
      pComputationTrimAcc = Bytes.concatenate(pComputationTrimAcc, Bytes.of(trimByte));
    }
  }
}
