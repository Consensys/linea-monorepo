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

package net.consensys.linea.zktracer.module.mxp.moduleCall;

import static net.consensys.linea.zktracer.Trace.GAS_CONST_G_MEMORY;
import static net.consensys.linea.zktracer.module.mxp.MxpUtils.isDoubleOffsetOpcode;
import static net.consensys.linea.zktracer.types.Conversions.*;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.math.BigInteger;

import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.mxp.MxpExoCall;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;

public abstract class CancunStateUpdateMxpCall extends CancunNotMSizeNorTrivialMxpCall {

  public CancunStateUpdateMxpCall(Hub hub) {
    super(hub);
    computeStateUpdt(hub.wcp(), hub.euc());
  }

  public void computeStateUpdt(Wcp wcp, Euc euc) {
    final OpCode opCode = this.opCodeData.mnemonic();

    // We compute and assign the computation's result for each row

    // Row i + 7
    // Compute useParams1 and useParams2
    boolean useParams2 = false; // default value if opcode is single offset
    boolean useParams1 = true;
    // we filter the row i + 7 wcp call by double_offset to prevent unnecessary comparisons
    if (isDoubleOffsetOpcode(opCode)) {
      final BigInteger max1 =
          this.offset1.toUnsignedBigInteger().add(this.size1.toUnsignedBigInteger());
      final BigInteger max2 =
          this.offset2.toUnsignedBigInteger().add(this.size2.toUnsignedBigInteger());
      exoCalls[6] = MxpExoCall.callToLT(wcp, bigIntegerToBytes(max1), bigIntegerToBytes(max2));
      useParams2 = bytesToBoolean(exoCalls[6].resultA()); // result of row i + 7
      useParams1 = !useParams2;
    }

    // Row i + 8
    // Compute floor and EYPa
    final BigInteger maxOffset1 =
        this.offset1
            .lo()
            .toUnsignedBigInteger()
            .add(this.size1.lo().toUnsignedBigInteger())
            .subtract(BigInteger.ONE);
    final BigInteger maxOffset2 =
        this.offset2
            .lo()
            .toUnsignedBigInteger()
            .add(this.size2.lo().toUnsignedBigInteger())
            .subtract(BigInteger.ONE);
    ;
    final BigInteger maxOffset =
        booleanToBigInteger(useParams1)
            .multiply(maxOffset1)
            .add(booleanToBigInteger(useParams2).multiply(maxOffset2));
    exoCalls[7] = MxpExoCall.callToEUC(euc, bigIntegerToBytes(maxOffset), Bytes.of(32));
    final Bytes floor = exoCalls[7].resultA();
    final BigInteger EYPa = floor.toUnsignedBigInteger().add(BigInteger.ONE);

    // row i + 9
    // Compute cMemQuadPart
    exoCalls[8] = MxpExoCall.callToEUC(euc, bigIntegerToBytes(EYPa.multiply(EYPa)), Bytes.of(512));
    final Bytes cMemQuadPart = exoCalls[8].resultA();

    // row i + 10
    // Compute updateInternalState
    exoCalls[9] = MxpExoCall.callToLT(wcp, longToBytes(words), bigIntegerToBytes(EYPa));
    final boolean updateInternalState = bytesToBoolean(exoCalls[9].resultA());

    // Determine state update
    final Bytes cMemLinearPart =
        bigIntegerToBytes(EYPa.multiply(BigInteger.valueOf(GAS_CONST_G_MEMORY)));
    this.isStateUpdate = updateInternalState;
    this.wordsNew = updateInternalState ? EYPa.longValue() : this.words;
    this.cMemNew =
        updateInternalState
            ? bigIntegerToBytes(
                    cMemQuadPart.toUnsignedBigInteger().add(cMemLinearPart.toUnsignedBigInteger()))
                .toLong()
            : this.cMem;
  }
}
