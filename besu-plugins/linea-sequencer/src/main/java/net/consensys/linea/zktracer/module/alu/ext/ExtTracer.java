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

import java.util.List;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class ExtTracer implements ModuleTracer {
  private int stamp = 0;

  private static final int MMEDIUM = 8;

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
    final Trace.TraceBuilder builder = Trace.builder();
    stamp++;

    for (int ct = 0; ct < maxCounter(data); ct++) {
      final int accLength = ct + 1;
      builder
          .extStampArg(stamp)
          // Byte A and Acc A
          .byteA0Arg(UnsignedByte.of(data.getABytes().get(0).get(ct)))
          .byteA1Arg(UnsignedByte.of(data.getABytes().get(1).get(ct)))
          .byteA2Arg(UnsignedByte.of(data.getABytes().get(2).get(ct)))
          .byteA3Arg(UnsignedByte.of(data.getABytes().get(3).get(ct)))
          .accA0Arg(data.getABytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accA1Arg(data.getABytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accA2Arg(data.getABytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accA3Arg(data.getABytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // Byte B and Acc B
          .byteB0Arg(UnsignedByte.of(data.getBBytes().get(0).get(ct)))
          .byteB1Arg(UnsignedByte.of(data.getBBytes().get(1).get(ct)))
          .byteB2Arg(UnsignedByte.of(data.getBBytes().get(2).get(ct)))
          .byteB3Arg(UnsignedByte.of(data.getBBytes().get(3).get(ct)))
          .accB0Arg(data.getBBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accB1Arg(data.getBBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accB2Arg(data.getBBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accB3Arg(data.getBBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // Byte C and Acc C
          .byteC0Arg(UnsignedByte.of(data.getCBytes().get(0).get(ct)))
          .byteC1Arg(UnsignedByte.of(data.getCBytes().get(1).get(ct)))
          .byteC2Arg(UnsignedByte.of(data.getCBytes().get(2).get(ct)))
          .byteC3Arg(UnsignedByte.of(data.getCBytes().get(3).get(ct)))
          .accC0Arg(data.getCBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accC1Arg(data.getCBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accC2Arg(data.getCBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accC3Arg(data.getCBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // Byte Delta and Acc Delta
          .byteDelta0Arg(UnsignedByte.of(data.getDeltaBytes().get(0).get(ct)))
          .byteDelta1Arg(UnsignedByte.of(data.getDeltaBytes().get(1).get(ct)))
          .byteDelta2Arg(UnsignedByte.of(data.getDeltaBytes().get(2).get(ct)))
          .byteDelta3Arg(UnsignedByte.of(data.getDeltaBytes().get(3).get(ct)))
          .accDelta0Arg(data.getDeltaBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accDelta1Arg(data.getDeltaBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accDelta2Arg(data.getDeltaBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accDelta3Arg(data.getDeltaBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // Byte H and Acc H
          .byteH0Arg(UnsignedByte.of(data.getHBytes().get(0).get(ct)))
          .byteH1Arg(UnsignedByte.of(data.getHBytes().get(1).get(ct)))
          .byteH2Arg(UnsignedByte.of(data.getHBytes().get(2).get(ct)))
          .byteH3Arg(UnsignedByte.of(data.getHBytes().get(3).get(ct)))
          .byteH4Arg(UnsignedByte.of(data.getHBytes().get(4).get(ct)))
          .byteH5Arg(UnsignedByte.of(data.getHBytes().get(5).get(ct)))
          .accH0Arg(data.getHBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accH1Arg(data.getHBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accH2Arg(data.getHBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accH3Arg(data.getHBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accH4Arg(data.getHBytes().get(4).slice(0, accLength).toUnsignedBigInteger())
          .accH5Arg(data.getHBytes().get(5).slice(0, accLength).toUnsignedBigInteger())
          // Byte I and Acc I
          .byteI0Arg(UnsignedByte.of(data.getIBytes().get(0).get(ct)))
          .byteI1Arg(UnsignedByte.of(data.getIBytes().get(1).get(ct)))
          .byteI2Arg(UnsignedByte.of(data.getIBytes().get(2).get(ct)))
          .byteI3Arg(UnsignedByte.of(data.getIBytes().get(3).get(ct)))
          .byteI4Arg(UnsignedByte.of(data.getIBytes().get(4).get(ct)))
          .byteI5Arg(UnsignedByte.of(data.getIBytes().get(5).get(ct)))
          .byteI6Arg(UnsignedByte.of(data.getIBytes().get(6).get(ct)))
          .accI0Arg(data.getIBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accI1Arg(data.getIBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accI2Arg(data.getIBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accI3Arg(data.getIBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accI4Arg(data.getIBytes().get(4).slice(0, accLength).toUnsignedBigInteger())
          .accI5Arg(data.getIBytes().get(5).slice(0, accLength).toUnsignedBigInteger())
          .accI6Arg(data.getIBytes().get(6).slice(0, accLength).toUnsignedBigInteger())
          // Byte J and Acc J
          .byteJ0Arg(UnsignedByte.of(data.getJBytes().get(0).get(ct)))
          .byteJ1Arg(UnsignedByte.of(data.getJBytes().get(1).get(ct)))
          .byteJ2Arg(UnsignedByte.of(data.getJBytes().get(2).get(ct)))
          .byteJ3Arg(UnsignedByte.of(data.getJBytes().get(3).get(ct)))
          .byteJ4Arg(UnsignedByte.of(data.getJBytes().get(4).get(ct)))
          .byteJ5Arg(UnsignedByte.of(data.getJBytes().get(5).get(ct)))
          .byteJ6Arg(UnsignedByte.of(data.getJBytes().get(6).get(ct)))
          .byteJ7Arg(UnsignedByte.of(data.getJBytes().get(7).get(ct)))
          .accJ0Arg(data.getJBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accJ1Arg(data.getJBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accJ2Arg(data.getJBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accJ3Arg(data.getJBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accJ4Arg(data.getJBytes().get(4).slice(0, accLength).toUnsignedBigInteger())
          .accJ5Arg(data.getJBytes().get(5).slice(0, accLength).toUnsignedBigInteger())
          .accJ6Arg(data.getJBytes().get(6).slice(0, accLength).toUnsignedBigInteger())
          .accJ7Arg(data.getJBytes().get(7).slice(0, accLength).toUnsignedBigInteger())
          // Byte Q and Acc Q
          .byteQ0Arg(UnsignedByte.of(data.getQBytes().get(0).get(ct)))
          .byteQ1Arg(UnsignedByte.of(data.getQBytes().get(1).get(ct)))
          .byteQ2Arg(UnsignedByte.of(data.getQBytes().get(2).get(ct)))
          .byteQ3Arg(UnsignedByte.of(data.getQBytes().get(3).get(ct)))
          .byteQ4Arg(UnsignedByte.of(data.getQBytes().get(4).get(ct)))
          .byteQ5Arg(UnsignedByte.of(data.getQBytes().get(5).get(ct)))
          .byteQ6Arg(UnsignedByte.of(data.getQBytes().get(6).get(ct)))
          .byteQ7Arg(UnsignedByte.of(data.getQBytes().get(7).get(ct)))
          .accQ0Arg(data.getQBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accQ1Arg(data.getQBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accQ2Arg(data.getQBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accQ3Arg(data.getQBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accQ4Arg(data.getQBytes().get(4).slice(0, accLength).toUnsignedBigInteger())
          .accQ5Arg(data.getQBytes().get(5).slice(0, accLength).toUnsignedBigInteger())
          .accQ6Arg(data.getQBytes().get(6).slice(0, accLength).toUnsignedBigInteger())
          .accQ7Arg(data.getQBytes().get(7).slice(0, accLength).toUnsignedBigInteger())
          // Byte R and Acc R
          .byteR0Arg(UnsignedByte.of(data.getRBytes().get(0).get(ct)))
          .byteR1Arg(UnsignedByte.of(data.getRBytes().get(1).get(ct)))
          .byteR2Arg(UnsignedByte.of(data.getRBytes().get(2).get(ct)))
          .byteR3Arg(UnsignedByte.of(data.getRBytes().get(3).get(ct)))
          .accR0Arg(data.getRBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accR1Arg(data.getRBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accR2Arg(data.getRBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accR3Arg(data.getRBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          // other
          .arg1HiArg(data.getArg1().getHigh().toUnsignedBigInteger())
          .arg1LoArg(data.getArg1().getLow().toUnsignedBigInteger())
          .arg2HiArg(data.getArg2().getHigh().toUnsignedBigInteger())
          .arg2LoArg(data.getArg2().getLow().toUnsignedBigInteger())
          .arg3HiArg(data.getArg3().getHigh().toUnsignedBigInteger())
          .arg3LoArg(data.getArg3().getLow().toUnsignedBigInteger())
          .resHiArg(data.getResult().getHigh().toUnsignedBigInteger())
          .resLoArg(data.getResult().getLow().toUnsignedBigInteger())
          .cmpArg(data.getCmp()[ct])
          .ofHArg(data.getOverflowH()[ct])
          .ofJArg(data.getOverflowJ()[ct])
          .ofIArg(data.getOverflowI()[ct])
          .ofResArg(data.getOverflowRes()[ct])
          .ctArg(ct)
          .instArg(UnsignedByte.of(opCode.value))
          .oliArg(data.isOli())
          .bit1Arg(data.getBit1())
          .bit2Arg(data.getBit2())
          .bit3Arg(data.getBit3());
    }

    return new ExtTrace(builder.build(), stamp);
  }

  private int maxCounter(ExtData data) {
    if (data.isOli()) {
      return 1;
    }

    return MMEDIUM;
  }
}
