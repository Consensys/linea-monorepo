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

package net.consensys.linea.zktracer.module.add;

import com.google.common.base.Objects;
import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
import net.consensys.linea.zktracer.types.Bytes16;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

public final class AddOperation {
  private static final UInt256 TWO_TO_THE_128 = UInt256.ONE.shiftLeft(128);

  private final OpCode opCode;
  private final Bytes32 arg1;
  private final Bytes32 arg2;

  /**
   * Returns the appropriate state of the overflow bit depending on the position within the cycle.
   *
   * @param counter current position within the tracing cycle
   * @param overflowHi putative overflow bit of the high part
   * @param overflowLo putative overflow bit of the high part
   * @return the overflow bit to trace
   */
  private static boolean overflowBit(
      final int counter, final boolean overflowHi, final boolean overflowLo) {
    if (counter == 14) {
      return overflowHi;
    }

    if (counter == 15) {
      return overflowLo;
    }

    return false;
  }

  public AddOperation(OpCode opCode, Bytes arg1, Bytes arg2) {
    this.opCode = opCode;
    this.arg1 = Bytes32.leftPad(arg1);
    this.arg2 = Bytes32.leftPad(arg2);
  }

  void trace(int stamp, Trace trace) {
    final Bytes16 arg1Hi = Bytes16.wrap(arg1.slice(0, 16));
    final Bytes32 arg1Lo = Bytes32.leftPad(arg1.slice(16));
    final Bytes16 arg2Hi = Bytes16.wrap(arg2.slice(0, 16));
    final Bytes16 arg2Lo = Bytes16.wrap(arg2.slice(16));

    boolean overflowHi = false;

    final OpCodeData opCodeData = OpCodes.of(opCode);

    final BaseBytes res = Adder.addSub(opCode, arg1, arg2);

    final Bytes16 resHi = res.getHigh();
    final Bytes16 resLo = res.getLow();

    final UInt256 arg1Int = UInt256.fromBytes(arg1);
    final UInt256 arg2Int = UInt256.fromBytes(arg2);

    if (opCode == OpCode.ADD) {
      final UInt256 resultBytes = arg1Int.add(arg2Int);
      if (resultBytes.compareTo(arg1Int) < 0) {
        overflowHi = true;
      }
    } else if (opCode == OpCode.SUB) {
      if (arg1Int.compareTo(arg2Int) < 0) {
        overflowHi = true;
      }
    }

    for (int i = 0; i < 16; i++) {
      Bytes32 addRes;
      if (opCode == OpCode.ADD) {
        addRes = Bytes32.wrap((UInt256.fromBytes(arg1Lo)).add(UInt256.fromBytes(arg2Lo)));
      } else {
        addRes = Bytes32.wrap((UInt256.fromBytes(resLo)).add(UInt256.fromBytes(arg2Lo)));
      }
      final boolean overflowLo = (addRes.compareTo(TWO_TO_THE_128) >= 0);

      trace
          .acc1(resHi.slice(0, 1 + i))
          .acc2(resLo.slice(0, 1 + i))
          .arg1Hi(arg1Hi)
          .arg1Lo(arg1Lo)
          .arg2Hi(arg2Hi)
          .arg2Lo(arg2Lo)
          .byte1(UnsignedByte.of(resHi.get(i)))
          .byte2(UnsignedByte.of(resLo.get(i)))
          .ct(Bytes.of(i))
          .inst(Bytes.of(opCodeData.value()))
          .overflow(overflowBit(i, overflowHi, overflowLo))
          .resHi(resHi)
          .resLo(resLo)
          .stamp(Bytes.ofUnsignedLong(stamp))
          .validateRow();
    }
  }

  @Override
  public int hashCode() {
    return Objects.hashCode(this.opCode, this.arg1, this.arg2);
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) return true;
    if (o == null || getClass() != o.getClass()) return false;
    final AddOperation that = (AddOperation) o;
    return java.util.Objects.equals(this.opCode, that.opCode)
        && java.util.Objects.equals(this.arg1, that.arg1)
        && java.util.Objects.equals(this.arg2, that.arg2);
  }
}
