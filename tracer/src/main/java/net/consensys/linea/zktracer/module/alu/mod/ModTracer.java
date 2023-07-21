/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.module.alu.mod;

import java.util.List;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class ModTracer implements ModuleTracer {
  private int stamp = 0;
  private static final int MMEDIUM = 8;

  @Override
  public String jsonKey() {
    return "mod";
  }

  @Override
  public List<OpCode> supportedOpCodes() {
    return List.of(OpCode.DIV, OpCode.SDIV, OpCode.MOD, OpCode.SMOD);
  }

  @Override
  public Object trace(final MessageFrame frame) {

    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));

    final ModData data = new ModData(opCode, arg1, arg2);
    final Trace.TraceBuilder builder = Trace.builder();

    stamp++;

    for (int ct = 0; ct < maxCounter(data); ct++) {
      final int accLength = ct + 1;
      builder
          .modStampArg(stamp)
          .oliArg(data.isOli())
          .ctArg(ct)
          .instArg(UnsignedByte.of(opCode.value))
          .decSignedArg(data.isSigned())
          .decOutputArg(data.isDiv())
          .arg1HiArg(data.getArg1().getHigh().toUnsignedBigInteger())
          .arg1LoArg(data.getArg1().getLow().toUnsignedBigInteger())
          .arg2HiArg(data.getArg2().getHigh().toUnsignedBigInteger())
          .arg2LoArg(data.getArg2().getLow().toUnsignedBigInteger())
          .resHiArg(data.getResult().getHigh().toUnsignedBigInteger())
          .resLoArg(data.getResult().getLow().toUnsignedBigInteger())
          .acc12Arg(data.getArg1().getBytes32().slice(8, ct + 1).toUnsignedBigInteger())
          .acc13Arg(data.getArg1().getBytes32().slice(0, ct + 1).toUnsignedBigInteger())
          .acc22Arg(data.getArg2().getBytes32().slice(8, ct + 1).toUnsignedBigInteger())
          .acc23Arg(data.getArg2().getBytes32().slice(0, ct + 1).toUnsignedBigInteger())
          .accB0Arg(data.getBBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accB1Arg(data.getBBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accB2Arg(data.getBBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accB3Arg(data.getBBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accR0Arg(data.getRBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accR1Arg(data.getRBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accR2Arg(data.getRBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accR3Arg(data.getRBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accQ0Arg(data.getQBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accQ1Arg(data.getQBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accQ2Arg(data.getQBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accQ3Arg(data.getQBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accDelta0Arg(data.getDBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accDelta1Arg(data.getDBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accDelta2Arg(data.getDBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accDelta3Arg(data.getDBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .byte22Arg(UnsignedByte.of(data.getArg2().getByte(ct + 8)))
          .byte23Arg(UnsignedByte.of(data.getArg2().getByte(ct)))
          .byte12Arg(UnsignedByte.of(data.getArg1().getByte(ct + 8)))
          .byte13Arg(UnsignedByte.of(data.getArg1().getByte(ct)))
          .byteB0Arg(UnsignedByte.of(data.getBBytes().get(0).get(ct)))
          .byteB1Arg(UnsignedByte.of(data.getBBytes().get(1).get(ct)))
          .byteB2Arg(UnsignedByte.of(data.getBBytes().get(2).get(ct)))
          .byteB3Arg(UnsignedByte.of(data.getBBytes().get(3).get(ct)))
          .byteR0Arg(UnsignedByte.of(data.getRBytes().get(0).get(ct)))
          .byteR1Arg(UnsignedByte.of(data.getRBytes().get(1).get(ct)))
          .byteR2Arg(UnsignedByte.of(data.getRBytes().get(2).get(ct)))
          .byteR3Arg(UnsignedByte.of(data.getRBytes().get(3).get(ct)))
          .byteQ0Arg(UnsignedByte.of(data.getQBytes().get(0).get(ct)))
          .byteQ1Arg(UnsignedByte.of(data.getQBytes().get(1).get(ct)))
          .byteQ2Arg(UnsignedByte.of(data.getQBytes().get(2).get(ct)))
          .byteQ3Arg(UnsignedByte.of(data.getQBytes().get(3).get(ct)))
          .byteDelta0Arg(UnsignedByte.of(data.getDBytes().get(0).get(ct)))
          .byteDelta1Arg(UnsignedByte.of(data.getDBytes().get(1).get(ct)))
          .byteDelta2Arg(UnsignedByte.of(data.getDBytes().get(2).get(ct)))
          .byteDelta3Arg(UnsignedByte.of(data.getDBytes().get(3).get(ct)))
          .byteH0Arg(UnsignedByte.of(data.getHBytes().get(0).get(ct)))
          .byteH1Arg(UnsignedByte.of(data.getHBytes().get(1).get(ct)))
          .byteH2Arg(UnsignedByte.of(data.getHBytes().get(2).get(ct)))
          .accH0Arg(Bytes.wrap(data.getHBytes().get(0)).slice(0, ct + 1).toUnsignedBigInteger())
          .accH1Arg(Bytes.wrap(data.getHBytes().get(1)).slice(0, ct + 1).toUnsignedBigInteger())
          .accH2Arg(Bytes.wrap(data.getHBytes().get(2)).slice(0, ct + 1).toUnsignedBigInteger())
          .cmp1Arg(data.getCmp1()[ct])
          .cmp2Arg(data.getCmp2()[ct])
          .msb1Arg(data.getMsb1()[ct])
          .msb2Arg(data.getMsb2()[ct]);
    }

    Trace trace = builder.build();

    return new ModTrace(trace, stamp);
  }

  private int maxCounter(ModData data) {
    if (data.isOli()) {
      return 1;
    }

    return MMEDIUM;
  }
}
