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

package net.consensys.linea.zktracer.module.mod;

import java.math.BigInteger;
import java.util.List;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.Module;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class Mod implements Module {
  private int stamp = 0;
  private static final int MMEDIUM = 8;

  final Trace.TraceBuilder builder = Trace.builder();

  @Override
  public String jsonKey() {
    return "mod";
  }

  @Override
  public final List<OpCode> supportedOpCodes() {
    return List.of(OpCode.DIV, OpCode.SDIV, OpCode.MOD, OpCode.SMOD);
  }

  @Override
  public void trace(final MessageFrame frame) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));

    final ModData data = new ModData(opCode, arg1, arg2);

    stamp++;

    for (int i = 0; i < maxCounter(data); i++) {
      final int accLength = i + 1;
      builder
          .stamp(BigInteger.valueOf(stamp))
          .oli(data.isOli())
          .ct(BigInteger.valueOf(i))
          .inst(BigInteger.valueOf(opCode.value))
          .decSigned(data.isSigned())
          .decOutput(data.isDiv())
          .arg1Hi(data.getArg1().getHigh().toUnsignedBigInteger())
          .arg1Lo(data.getArg1().getLow().toUnsignedBigInteger())
          .arg2Hi(data.getArg2().getHigh().toUnsignedBigInteger())
          .arg2Lo(data.getArg2().getLow().toUnsignedBigInteger())
          .resHi(data.getResult().getHigh().toUnsignedBigInteger())
          .resLo(data.getResult().getLow().toUnsignedBigInteger())
          .acc12(data.getArg1().getBytes32().slice(8, i + 1).toUnsignedBigInteger())
          .acc13(data.getArg1().getBytes32().slice(0, i + 1).toUnsignedBigInteger())
          .acc22(data.getArg2().getBytes32().slice(8, i + 1).toUnsignedBigInteger())
          .acc23(data.getArg2().getBytes32().slice(0, i + 1).toUnsignedBigInteger())
          .accB0(data.getBBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accB1(data.getBBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accB2(data.getBBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accB3(data.getBBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accR0(data.getRBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accR1(data.getRBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accR2(data.getRBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accR3(data.getRBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accQ0(data.getQBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accQ1(data.getQBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accQ2(data.getQBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accQ3(data.getQBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accDelta0(data.getDBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accDelta1(data.getDBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accDelta2(data.getDBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accDelta3(data.getDBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .byte22(UnsignedByte.of(data.getArg2().getByte(i + 8)))
          .byte23(UnsignedByte.of(data.getArg2().getByte(i)))
          .byte12(UnsignedByte.of(data.getArg1().getByte(i + 8)))
          .byte13(UnsignedByte.of(data.getArg1().getByte(i)))
          .byteB0(UnsignedByte.of(data.getBBytes().get(0).get(i)))
          .byteB1(UnsignedByte.of(data.getBBytes().get(1).get(i)))
          .byteB2(UnsignedByte.of(data.getBBytes().get(2).get(i)))
          .byteB3(UnsignedByte.of(data.getBBytes().get(3).get(i)))
          .byteR0(UnsignedByte.of(data.getRBytes().get(0).get(i)))
          .byteR1(UnsignedByte.of(data.getRBytes().get(1).get(i)))
          .byteR2(UnsignedByte.of(data.getRBytes().get(2).get(i)))
          .byteR3(UnsignedByte.of(data.getRBytes().get(3).get(i)))
          .byteQ0(UnsignedByte.of(data.getQBytes().get(0).get(i)))
          .byteQ1(UnsignedByte.of(data.getQBytes().get(1).get(i)))
          .byteQ2(UnsignedByte.of(data.getQBytes().get(2).get(i)))
          .byteQ3(UnsignedByte.of(data.getQBytes().get(3).get(i)))
          .byteDelta0(UnsignedByte.of(data.getDBytes().get(0).get(i)))
          .byteDelta1(UnsignedByte.of(data.getDBytes().get(1).get(i)))
          .byteDelta2(UnsignedByte.of(data.getDBytes().get(2).get(i)))
          .byteDelta3(UnsignedByte.of(data.getDBytes().get(3).get(i)))
          .byteH0(UnsignedByte.of(data.getHBytes().get(0).get(i)))
          .byteH1(UnsignedByte.of(data.getHBytes().get(1).get(i)))
          .byteH2(UnsignedByte.of(data.getHBytes().get(2).get(i)))
          .accH0(Bytes.wrap(data.getHBytes().get(0)).slice(0, i + 1).toUnsignedBigInteger())
          .accH1(Bytes.wrap(data.getHBytes().get(1)).slice(0, i + 1).toUnsignedBigInteger())
          .accH2(Bytes.wrap(data.getHBytes().get(2)).slice(0, i + 1).toUnsignedBigInteger())
          .cmp1(data.getCmp1()[i])
          .cmp2(data.getCmp2()[i])
          .msb1(data.getMsb1()[i])
          .msb2(data.getMsb2()[i])
          .validateRow();
    }
  }

  @Override
  public Object commit() {
    return new ModTrace(builder.build());
  }

  private int maxCounter(ModData data) {
    if (data.isOli()) {
      return 1;
    }

    return MMEDIUM;
  }
}
