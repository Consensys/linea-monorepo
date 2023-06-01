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
package net.consensys.linea.zktracer.module.alu.ext;

import org.hyperledger.besu.evm.frame.MessageFrame;

import java.util.List;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes32;

public class ExtTracer implements ModuleTracer {
  private int stamp = 0;

  private final int MMEDIUM = 8;

  @Override
  public String jsonKey() {
    return "ext";
  }

  @Override
  public List<OpCode> supportedOpCodes() {
    return List.of(OpCode.MULMOD, OpCode.ADDMOD);
  }

  @Override
  public Object trace(final MessageFrame frame) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));
    final Bytes32 arg3 = Bytes32.wrap(frame.getStackItem(2));

    final ExtData data = new ExtData(opCode, arg1, arg2, arg3);
    final ExtTrace.Trace.Builder builder = ExtTrace.Trace.Builder.newInstance();
    stamp++;
    for (int ct = 0; ct < maxCounter(data); ct++) {
      final int accLength = ct + 1;
      builder
          .appendStamp(stamp)
          // Byte A and Acc A
          .appendByteA0(UnsignedByte.of(data.getABytes().get(0).get(ct)))
          .appendByteA1(UnsignedByte.of(data.getABytes().get(1).get(ct)))
          .appendByteA2(UnsignedByte.of(data.getABytes().get(2).get(ct)))
          .appendByteA3(UnsignedByte.of(data.getABytes().get(3).get(ct)))
          .appendAccA0(data.getABytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccA1(data.getABytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccA2(data.getABytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccA3(data.getABytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // Byte B and Acc B
          .appendByteB0(UnsignedByte.of(data.getBBytes().get(0).get(ct)))
          .appendByteB1(UnsignedByte.of(data.getBBytes().get(1).get(ct)))
          .appendByteB2(UnsignedByte.of(data.getBBytes().get(2).get(ct)))
          .appendByteB3(UnsignedByte.of(data.getBBytes().get(3).get(ct)))
          .appendAccB0(data.getBBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccB1(data.getBBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccB2(data.getBBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccB3(data.getBBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // Byte C and Acc C
          .appendByteC0(UnsignedByte.of(data.getCBytes().get(0).get(ct)))
          .appendByteC1(UnsignedByte.of(data.getCBytes().get(1).get(ct)))
          .appendByteC2(UnsignedByte.of(data.getCBytes().get(2).get(ct)))
          .appendByteC3(UnsignedByte.of(data.getCBytes().get(3).get(ct)))
          .appendAccC0(data.getCBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccC1(data.getCBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccC2(data.getCBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccC3(data.getCBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // Byte Delta and Acc Delta
          .appendByteDelta0(UnsignedByte.of(data.getDeltaBytes().get(0).get(ct)))
          .appendByteDelta1(UnsignedByte.of(data.getDeltaBytes().get(1).get(ct)))
          .appendByteDelta2(UnsignedByte.of(data.getDeltaBytes().get(2).get(ct)))
          .appendByteDelta3(UnsignedByte.of(data.getDeltaBytes().get(3).get(ct)))
          .appendAccDelta0(data.getDeltaBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccDelta1(data.getDeltaBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccDelta2(data.getDeltaBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccDelta3(data.getDeltaBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // Byte H and Acc H
          .appendByteH0(UnsignedByte.of(data.getHBytes().get(0).get(ct)))
          .appendByteH1(UnsignedByte.of(data.getHBytes().get(1).get(ct)))
          .appendByteH2(UnsignedByte.of(data.getHBytes().get(2).get(ct)))
          .appendByteH3(UnsignedByte.of(data.getHBytes().get(3).get(ct)))
          .appendByteH4(UnsignedByte.of(data.getHBytes().get(4).get(ct)))
          .appendByteH5(UnsignedByte.of(data.getHBytes().get(5).get(ct)))
          .appendAccH0(data.getHBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccH1(data.getHBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccH2(data.getHBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccH3(data.getHBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .appendAccH4(data.getHBytes().get(4).slice(0, accLength).toUnsignedBigInteger())
          .appendAccH5(data.getHBytes().get(5).slice(0, accLength).toUnsignedBigInteger())
          // Byte I and Acc I
          .appendByteI0(UnsignedByte.of(data.getIBytes().get(0).get(ct)))
          .appendByteI1(UnsignedByte.of(data.getIBytes().get(1).get(ct)))
          .appendByteI2(UnsignedByte.of(data.getIBytes().get(2).get(ct)))
          .appendByteI3(UnsignedByte.of(data.getIBytes().get(3).get(ct)))
          .appendByteI4(UnsignedByte.of(data.getIBytes().get(4).get(ct)))
          .appendByteI5(UnsignedByte.of(data.getIBytes().get(5).get(ct)))
          .appendByteI6(UnsignedByte.of(data.getIBytes().get(6).get(ct)))
          .appendAccI0(data.getIBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccI1(data.getIBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccI2(data.getIBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccI3(data.getIBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .appendAccI4(data.getIBytes().get(4).slice(0, accLength).toUnsignedBigInteger())
          .appendAccI5(data.getIBytes().get(5).slice(0, accLength).toUnsignedBigInteger())
          .appendAccI6(data.getIBytes().get(6).slice(0, accLength).toUnsignedBigInteger())
          // Byte J and Acc J
          .appendByteJ0(UnsignedByte.of(data.getJBytes().get(0).get(ct)))
          .appendByteJ1(UnsignedByte.of(data.getJBytes().get(1).get(ct)))
          .appendByteJ2(UnsignedByte.of(data.getJBytes().get(2).get(ct)))
          .appendByteJ3(UnsignedByte.of(data.getJBytes().get(3).get(ct)))
          .appendByteJ4(UnsignedByte.of(data.getJBytes().get(4).get(ct)))
          .appendByteJ5(UnsignedByte.of(data.getJBytes().get(5).get(ct)))
          .appendByteJ6(UnsignedByte.of(data.getJBytes().get(6).get(ct)))
          .appendByteJ7(UnsignedByte.of(data.getJBytes().get(7).get(ct)))
          .appendAccJ0(data.getJBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccJ1(data.getJBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccJ2(data.getJBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccJ3(data.getJBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .appendAccJ4(data.getJBytes().get(4).slice(0, accLength).toUnsignedBigInteger())
          .appendAccJ5(data.getJBytes().get(5).slice(0, accLength).toUnsignedBigInteger())
          .appendAccJ6(data.getJBytes().get(6).slice(0, accLength).toUnsignedBigInteger())
          .appendAccJ7(data.getJBytes().get(7).slice(0, accLength).toUnsignedBigInteger())
          // Byte Q and Acc Q
          .appendByteQ0(UnsignedByte.of(data.getQBytes().get(0).get(ct)))
          .appendByteQ1(UnsignedByte.of(data.getQBytes().get(1).get(ct)))
          .appendByteQ2(UnsignedByte.of(data.getQBytes().get(2).get(ct)))
          .appendByteQ3(UnsignedByte.of(data.getQBytes().get(3).get(ct)))
          .appendByteQ4(UnsignedByte.of(data.getQBytes().get(4).get(ct)))
          .appendByteQ5(UnsignedByte.of(data.getQBytes().get(5).get(ct)))
          .appendByteQ6(UnsignedByte.of(data.getQBytes().get(6).get(ct)))
          .appendByteQ7(UnsignedByte.of(data.getQBytes().get(7).get(ct)))
          .appendAccQ0(data.getQBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccQ1(data.getQBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccQ2(data.getQBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccQ3(data.getQBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .appendAccQ4(data.getQBytes().get(4).slice(0, accLength).toUnsignedBigInteger())
          .appendAccQ5(data.getQBytes().get(5).slice(0, accLength).toUnsignedBigInteger())
          .appendAccQ6(data.getQBytes().get(6).slice(0, accLength).toUnsignedBigInteger())
          .appendAccQ7(data.getQBytes().get(7).slice(0, accLength).toUnsignedBigInteger())
          // Byte R and Acc R
          .appendByteR0(UnsignedByte.of(data.getRBytes().get(0).get(ct)))
          .appendByteR1(UnsignedByte.of(data.getRBytes().get(1).get(ct)))
          .appendByteR2(UnsignedByte.of(data.getRBytes().get(2).get(ct)))
          .appendByteR3(UnsignedByte.of(data.getRBytes().get(3).get(ct)))
          .appendAccR0(data.getRBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccR1(data.getRBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccR2(data.getRBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccR3(data.getRBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // other
          .appendArg1Hi(data.getArg1().getHigh().toUnsignedBigInteger())
          .appendArg1Lo(data.getArg1().getLow().toUnsignedBigInteger())
          .appendArg2Hi(data.getArg2().getHigh().toUnsignedBigInteger())
          .appendArg2Lo(data.getArg2().getLow().toUnsignedBigInteger())
          .appendArg3Hi(data.getArg3().getHigh().toUnsignedBigInteger())
          .appendArg3Lo(data.getArg3().getLow().toUnsignedBigInteger())
          .appendResHi(data.getResult().getHigh().toUnsignedBigInteger())
          .appendResLo(data.getResult().getLow().toUnsignedBigInteger())
          .appendCmp(data.getCmp()[ct])
          .appendOfH(data.getOverflowH()[ct])
          .appendOfJ(data.getOverflowJ()[ct])
          .appendOfI(data.getOverflowI()[ct])
          .appendOfRes(data.getOverflowRes()[ct])
          .appendCt(ct)
          .appendInst(UnsignedByte.of(opCode.value))
          .appendOli(data.isOli())
          .appendBit1(data.getBit1())
          .appendBit2(data.getBit2())
          .appendBit3(data.getBit3());
    }
    builder.setStamp(stamp);
    return builder.build();
  }

  private int maxCounter(ExtData data) {
    if (data.isOli()) {
      return 1;
    } else {
      return MMEDIUM;
    }
  }
}
