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

package net.consensys.linea.zktracer.module.ext;

import java.math.BigInteger;
import java.util.List;

import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class Ext implements Module {
  private static final int MMEDIUM = 8;

  final Trace.TraceBuilder builder = Trace.builder();
  private int stamp = 0;

  @Override
  public String jsonKey() {
    return "ext";
  }

  @Override
  public final List<OpCode> supportedOpCodes() {
    return List.of(OpCode.MULMOD, OpCode.ADDMOD);
  }

  @Override
  public void trace(final MessageFrame frame) {
    final OpCodeData opCode = OpCodes.of(frame.getCurrentOperation().getOpcode());
    final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));
    final Bytes32 arg3 = Bytes32.wrap(frame.getStackItem(2));

    final ExtData data = new ExtData(opCode, arg1, arg2, arg3);

    stamp++;

    for (int i = 0; i < maxCounter(data); i++) {
      final int accLength = i + 1;
      builder
          // Byte A and Acc A
          .byteA0(UnsignedByte.of(data.getABytes().get(0).get(i)))
          .byteA1(UnsignedByte.of(data.getABytes().get(1).get(i)))
          .byteA2(UnsignedByte.of(data.getABytes().get(2).get(i)))
          .byteA3(UnsignedByte.of(data.getABytes().get(3).get(i)))
          .accA0(data.getABytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accA1(data.getABytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accA2(data.getABytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accA3(data.getABytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // Byte B and Acc B
          .byteB0(UnsignedByte.of(data.getBBytes().get(0).get(i)))
          .byteB1(UnsignedByte.of(data.getBBytes().get(1).get(i)))
          .byteB2(UnsignedByte.of(data.getBBytes().get(2).get(i)))
          .byteB3(UnsignedByte.of(data.getBBytes().get(3).get(i)))
          .accB0(data.getBBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accB1(data.getBBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accB2(data.getBBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accB3(data.getBBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // Byte C and Acc C
          .byteC0(UnsignedByte.of(data.getCBytes().get(0).get(i)))
          .byteC1(UnsignedByte.of(data.getCBytes().get(1).get(i)))
          .byteC2(UnsignedByte.of(data.getCBytes().get(2).get(i)))
          .byteC3(UnsignedByte.of(data.getCBytes().get(3).get(i)))
          .accC0(data.getCBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accC1(data.getCBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accC2(data.getCBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accC3(data.getCBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // Byte Delta and Acc Delta
          .byteDelta0(UnsignedByte.of(data.getDeltaBytes().get(0).get(i)))
          .byteDelta1(UnsignedByte.of(data.getDeltaBytes().get(1).get(i)))
          .byteDelta2(UnsignedByte.of(data.getDeltaBytes().get(2).get(i)))
          .byteDelta3(UnsignedByte.of(data.getDeltaBytes().get(3).get(i)))
          .accDelta0(data.getDeltaBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accDelta1(data.getDeltaBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accDelta2(data.getDeltaBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accDelta3(data.getDeltaBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // Byte H and Acc H
          .byteH0(UnsignedByte.of(data.getHBytes().get(0).get(i)))
          .byteH1(UnsignedByte.of(data.getHBytes().get(1).get(i)))
          .byteH2(UnsignedByte.of(data.getHBytes().get(2).get(i)))
          .byteH3(UnsignedByte.of(data.getHBytes().get(3).get(i)))
          .byteH4(UnsignedByte.of(data.getHBytes().get(4).get(i)))
          .byteH5(UnsignedByte.of(data.getHBytes().get(5).get(i)))
          .accH0(data.getHBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accH1(data.getHBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accH2(data.getHBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accH3(data.getHBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accH4(data.getHBytes().get(4).slice(0, accLength).toUnsignedBigInteger())
          .accH5(data.getHBytes().get(5).slice(0, accLength).toUnsignedBigInteger())
          // Byte I and Acc I
          .byteI0(UnsignedByte.of(data.getIBytes().get(0).get(i)))
          .byteI1(UnsignedByte.of(data.getIBytes().get(1).get(i)))
          .byteI2(UnsignedByte.of(data.getIBytes().get(2).get(i)))
          .byteI3(UnsignedByte.of(data.getIBytes().get(3).get(i)))
          .byteI4(UnsignedByte.of(data.getIBytes().get(4).get(i)))
          .byteI5(UnsignedByte.of(data.getIBytes().get(5).get(i)))
          .byteI6(UnsignedByte.of(data.getIBytes().get(6).get(i)))
          .accI0(data.getIBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accI1(data.getIBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accI2(data.getIBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accI3(data.getIBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accI4(data.getIBytes().get(4).slice(0, accLength).toUnsignedBigInteger())
          .accI5(data.getIBytes().get(5).slice(0, accLength).toUnsignedBigInteger())
          .accI6(data.getIBytes().get(6).slice(0, accLength).toUnsignedBigInteger())
          // Byte J and Acc J
          .byteJ0(UnsignedByte.of(data.getJBytes().get(0).get(i)))
          .byteJ1(UnsignedByte.of(data.getJBytes().get(1).get(i)))
          .byteJ2(UnsignedByte.of(data.getJBytes().get(2).get(i)))
          .byteJ3(UnsignedByte.of(data.getJBytes().get(3).get(i)))
          .byteJ4(UnsignedByte.of(data.getJBytes().get(4).get(i)))
          .byteJ5(UnsignedByte.of(data.getJBytes().get(5).get(i)))
          .byteJ6(UnsignedByte.of(data.getJBytes().get(6).get(i)))
          .byteJ7(UnsignedByte.of(data.getJBytes().get(7).get(i)))
          .accJ0(data.getJBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accJ1(data.getJBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accJ2(data.getJBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accJ3(data.getJBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accJ4(data.getJBytes().get(4).slice(0, accLength).toUnsignedBigInteger())
          .accJ5(data.getJBytes().get(5).slice(0, accLength).toUnsignedBigInteger())
          .accJ6(data.getJBytes().get(6).slice(0, accLength).toUnsignedBigInteger())
          .accJ7(data.getJBytes().get(7).slice(0, accLength).toUnsignedBigInteger())
          // Byte Q and Acc Q
          .byteQ0(UnsignedByte.of(data.getQBytes().get(0).get(i)))
          .byteQ1(UnsignedByte.of(data.getQBytes().get(1).get(i)))
          .byteQ2(UnsignedByte.of(data.getQBytes().get(2).get(i)))
          .byteQ3(UnsignedByte.of(data.getQBytes().get(3).get(i)))
          .byteQ4(UnsignedByte.of(data.getQBytes().get(4).get(i)))
          .byteQ5(UnsignedByte.of(data.getQBytes().get(5).get(i)))
          .byteQ6(UnsignedByte.of(data.getQBytes().get(6).get(i)))
          .byteQ7(UnsignedByte.of(data.getQBytes().get(7).get(i)))
          .accQ0(data.getQBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accQ1(data.getQBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accQ2(data.getQBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accQ3(data.getQBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accQ4(data.getQBytes().get(4).slice(0, accLength).toUnsignedBigInteger())
          .accQ5(data.getQBytes().get(5).slice(0, accLength).toUnsignedBigInteger())
          .accQ6(data.getQBytes().get(6).slice(0, accLength).toUnsignedBigInteger())
          .accQ7(data.getQBytes().get(7).slice(0, accLength).toUnsignedBigInteger())
          // Byte R and Acc R
          .byteR0(UnsignedByte.of(data.getRBytes().get(0).get(i)))
          .byteR1(UnsignedByte.of(data.getRBytes().get(1).get(i)))
          .byteR2(UnsignedByte.of(data.getRBytes().get(2).get(i)))
          .byteR3(UnsignedByte.of(data.getRBytes().get(3).get(i)))
          .accR0(data.getRBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accR1(data.getRBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accR2(data.getRBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accR3(data.getRBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // other
          .arg1Hi(data.getArg1().getHigh().toUnsignedBigInteger())
          .arg1Lo(data.getArg1().getLow().toUnsignedBigInteger())
          .arg2Hi(data.getArg2().getHigh().toUnsignedBigInteger())
          .arg2Lo(data.getArg2().getLow().toUnsignedBigInteger())
          .arg3Hi(data.getArg3().getHigh().toUnsignedBigInteger())
          .arg3Lo(data.getArg3().getLow().toUnsignedBigInteger())
          .resHi(data.getResult().getHigh().toUnsignedBigInteger())
          .resLo(data.getResult().getLow().toUnsignedBigInteger())
          .cmp(data.getCmp()[i])
          .ofH(data.getOverflowH()[i])
          .ofJ(data.getOverflowJ()[i])
          .ofI(data.getOverflowI()[i])
          .ofRes(data.getOverflowRes()[i])
          .ct(BigInteger.valueOf(i))
          .inst(BigInteger.valueOf(opCode.value()))
          .oli(data.isOli())
          .bit1(data.getBit1())
          .bit2(data.getBit2())
          .bit3(data.getBit3())
          .stamp(BigInteger.valueOf(stamp))
          .validateRow();
    }
  }

  @Override
  public Object commit() {
    return new ExtTrace(builder.build());
  }

  private int maxCounter(ExtData data) {
    if (data.isOli()) {
      return 1;
    }

    return MMEDIUM;
  }
}
