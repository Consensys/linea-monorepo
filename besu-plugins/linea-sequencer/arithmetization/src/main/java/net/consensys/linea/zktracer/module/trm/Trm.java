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

package net.consensys.linea.zktracer.module.trm;

import static net.consensys.linea.zktracer.module.rlputils.Pattern.bitDecomposition;
import static net.consensys.linea.zktracer.module.rlputils.Pattern.padToGivenSizeWithLeftZero;
import static net.consensys.linea.zktracer.types.Address.isPrecompile;

import java.math.BigInteger;
import java.util.List;

import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.ModuleTrace;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class Trm implements Module {
  @Override
  public String jsonKey() {
    return "trm";
  }

  private int stamp = 0;
  private final int maxCT = 16;
  static final int LLARGE = 16;
  static final int PIVOT_BIT_FLIPS_TO_TRUE = 12;

  private final StackedSet<EWord> trimmings = new StackedSet<>();

  @Override
  public void enterTransaction() {
    this.trimmings.enter();
  }

  @Override
  public void popTransaction() {
    this.trimmings.pop();
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {

    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    switch (opCode) {
      case BALANCE, EXTCODESIZE, EXTCODECOPY, EXTCODEHASH, SELFDESTRUCT -> {
        this.trimmings.add(EWord.of(frame.getStackItem(0)));
      }
      case CALL, CALLCODE, DELEGATECALL, STATICCALL -> {
        this.trimmings.add(EWord.of(frame.getStackItem(1)));
      }
    }
  }

  public static boolean isPrec(EWord data) {
    BigInteger trmAddrParamAsBigInt = data.slice(12, 20).toUnsignedBigInteger();
    return (!trmAddrParamAsBigInt.equals(BigInteger.ZERO)
        && (trmAddrParamAsBigInt.compareTo(BigInteger.TEN) < 0));
  }

  private void traceTrimming(EWord data, Trace.TraceBuilder trace) {
    this.stamp++;
    Bytes trmHi = padToGivenSizeWithLeftZero(data.hi().slice(PIVOT_BIT_FLIPS_TO_TRUE, 4), LLARGE);
    Boolean isPrec = isPrecompile(Address.extract(data));
    final int accLastByte = isPrec ? 9 - (0xff & data.get(31)) : (0xff & data.get(31)) - 10;
    List<Boolean> ones = bitDecomposition(accLastByte, maxCT).bitDecList();

    for (int ct = 0; ct < this.maxCT; ct++) {
      trace
          .ct(BigInteger.valueOf(ct))
          .stamp(BigInteger.valueOf(this.stamp))
          .isPrec(isPrec)
          .pbit(ct >= PIVOT_BIT_FLIPS_TO_TRUE)
          .addrHi(data.hi().toUnsignedBigInteger())
          .addrLo(data.lo().toUnsignedBigInteger())
          .trmAddrHi(trmHi.toUnsignedBigInteger())
          .accHi(data.hi().slice(0, ct + 1).toUnsignedBigInteger())
          .accLo(data.lo().slice(0, ct + 1).toUnsignedBigInteger())
          .accT(trmHi.slice(0, ct + 1).toUnsignedBigInteger())
          .byteHi(UnsignedByte.of(data.hi().get(ct)))
          .byteLo(UnsignedByte.of(data.lo().get(ct)))
          .one(ones.get(ct))
          .validateRow();
    }
  }

  @Override
  public ModuleTrace commit() {
    final Trace.TraceBuilder trace = Trace.builder(this.lineCount());

    for (EWord data : this.trimmings) {
      traceTrimming(data, trace);
    }
    return new TrmTrace(trace.build());
  }

  @Override
  public int lineCount() {
    return this.trimmings.size() * maxCT;
  }
}
