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

package net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.modexp;

import static java.lang.Math.max;
import static net.consensys.linea.zktracer.Trace.OOB_INST_MODEXP_PRICING;
import static net.consensys.linea.zktracer.Trace.Oob.CT_MAX_MODEXP_PRICING;
import static net.consensys.linea.zktracer.Trace.Oob.G_QUADDIVISOR;
import static net.consensys.linea.zktracer.module.oob.OobExoCall.*;
import static net.consensys.linea.zktracer.module.oob.OobOperation.computeExponentLog;
import static net.consensys.linea.zktracer.runtime.callstack.CallFrame.getOpCode;
import static net.consensys.linea.zktracer.types.Conversions.*;
import static org.hyperledger.besu.evm.internal.Words.clampedToInt;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.add.Add;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;
import net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.oob.OobExoCall;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;

@Getter
@Setter
public class ModexpPricingOobCall extends OobCall {

  final ModexpMetadata metadata;
  final Bytes callGas;
  EWord returnAtCapacity;
  boolean ramSuccess;
  BigInteger exponentLog;
  int maxMbsBbs;

  BigInteger returnGas;
  boolean returnAtCapacityNonZero;

  public ModexpPricingOobCall(ModexpMetadata metadata, long calleeGas) {
    super();
    this.metadata = metadata;
    this.callGas = Bytes.ofUnsignedLong(calleeGas);
  }

  @Override
  public void setInputData(MessageFrame frame, Hub hub) {
    final OpCode opCode = getOpCode(frame);
    final int returnAtCapacityIndex = opCode.callHasValueArgument() ? 6 : 5;
    setReturnAtCapacity(EWord.of(frame.getStackItem(returnAtCapacityIndex)));

    final int cdsIndex = opCode.callHasValueArgument() ? 4 : 3;
    final int cds = clampedToInt(frame.getStackItem(cdsIndex));

    setExponentLog(BigInteger.valueOf(computeExponentLog(metadata, cds)));

    final int maxMbsBbs = max(metadata.mbsInt(), metadata.bbsInt());
    setMaxMbsBbs(maxMbsBbs);
  }

  @Override
  public void callExoModules(Add add, Mod mod, Wcp wcp) {
    // row i
    final OobExoCall returnAtCapacityIsZeroCall = callToIsZero(wcp, returnAtCapacity);
    exoCalls.add(returnAtCapacityIsZeroCall);
    final boolean returnAtCapacityIsZero = bytesToBoolean(returnAtCapacityIsZeroCall.result());

    // row i + 1
    final OobExoCall exponentLogIsZeroCall = callToIsZero(wcp, bigIntegerToBytes(exponentLog));
    exoCalls.add(exponentLogIsZeroCall);
    final boolean exponentLogIsZero = bytesToBoolean(exponentLogIsZeroCall.result());

    // row i + 2
    final OobExoCall ceilngOfMaxDividedBy8Call =
        callToDIV(mod, Bytes.of(maxMbsBbs + 7), Bytes.ofUnsignedInt(8));
    exoCalls.add(ceilngOfMaxDividedBy8Call);
    final BigInteger ceilingOfMaxDividedBy8 =
        ceilngOfMaxDividedBy8Call.result().toUnsignedBigInteger();
    final BigInteger fOfMax = ceilingOfMaxDividedBy8.multiply(ceilingOfMaxDividedBy8);

    // row i + 3
    final BigInteger bigNumerator = exponentLogIsZero ? fOfMax : fOfMax.multiply(exponentLog);
    final OobExoCall bigQuotientCall =
        callToDIV(mod, bigIntegerToBytes(bigNumerator), Bytes.ofUnsignedInt(G_QUADDIVISOR));
    exoCalls.add(bigQuotientCall);
    final Bytes bigQuotient = bigQuotientCall.result();

    // row i + 4
    final OobExoCall bigQuotientLT200Call = callToLT(wcp, bigQuotient, Bytes.ofUnsignedInt(200));
    exoCalls.add(bigQuotientLT200Call);
    final boolean bigQuotientLT200 = bytesToBoolean(bigQuotientLT200Call.result());

    // row i + 5
    final Bytes precompileCost = bigQuotientLT200 ? Bytes.ofUnsignedInt(200) : bigQuotient;
    final OobExoCall ramSuccessCall = callToLT(wcp, callGas, precompileCost);
    exoCalls.add(ramSuccessCall);
    setRamSuccess(!bytesToBoolean(ramSuccessCall.result()));

    final BigInteger returnGas =
        ramSuccess
            ? callGas.toUnsignedBigInteger().subtract(precompileCost.toUnsignedBigInteger())
            : BigInteger.ZERO;
    setReturnGas(returnGas);

    setReturnAtCapacityNonZero(!returnAtCapacityIsZero);
  }

  @Override
  public int ctMax() {
    return CT_MAX_MODEXP_PRICING;
  }

  @Override
  public Trace.Oob trace(Trace.Oob trace) {
    return trace
        .isModexpPricing(true)
        .oobInst(OOB_INST_MODEXP_PRICING)
        .data1(callGas)
        .data3(returnAtCapacity.trimLeadingZeros())
        .data4(booleanToBytes(ramSuccess))
        .data5(bigIntegerToBytes(returnGas))
        .data6(bigIntegerToBytes(exponentLog))
        .data7(Bytes.ofUnsignedInt(maxMbsBbs))
        .data8(booleanToBytes(returnAtCapacityNonZero));
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(OOB_INST_MODEXP_PRICING)
        .pMiscOobData1(callGas)
        .pMiscOobData3(returnAtCapacity.trimLeadingZeros())
        .pMiscOobData4(booleanToBytes(ramSuccess))
        .pMiscOobData5(bigIntegerToBytes(returnGas))
        .pMiscOobData6(bigIntegerToBytes(exponentLog))
        .pMiscOobData7(Bytes.ofUnsignedInt(maxMbsBbs))
        .pMiscOobData8(booleanToBytes(returnAtCapacityNonZero));
  }
}
