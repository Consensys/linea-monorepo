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

import static net.consensys.linea.zktracer.Trace.LLARGE;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.Bytes16;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

@Accessors(fluent = true)
@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
public final class AddOperation extends ModuleOperation {
  private static final UInt256 TWO_TO_THE_128 = UInt256.ONE.shiftLeft(128);

  @EqualsAndHashCode.Include @Getter private final OpCode opCode;
  @EqualsAndHashCode.Include @Getter private final Bytes32 arg1;
  @EqualsAndHashCode.Include @Getter private final Bytes32 arg2;
  private BaseBytes res;
  private int ctMax;

  /**
   * Returns the appropriate state of the overflow bit depending on the position within the cycle.
   *
   * @param counter current position within the tracing cycle
   * @param overflowHi putative overflow bit of the high part
   * @param overflowLo putative overflow bit of the high part
   * @return the overflow bit to trace
   */
  private boolean overflowBit(
      final int counter, final boolean overflowHi, final boolean overflowLo) {
    if (counter == this.ctMax - 1) {
      return overflowHi;
    }

    if (counter == this.ctMax) {
      return overflowLo;
    }

    return false;
  }

  public AddOperation(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {
    this.opCode = opCode;
    this.arg1 = arg1;
    this.arg2 = arg2;
  }

  private int computeCtMax() {
    res = Adder.addSub(opCode, arg1, arg2);
    return Math.max(
        1,
        Math.max(res.getHigh().trimLeadingZeros().size(), res.getLow().trimLeadingZeros().size())
            - 1);
  }

  void trace(int stamp, Trace.Add trace) {
    final Bytes16 arg1Hi = Bytes16.wrap(arg1.slice(0, 16));
    final Bytes16 arg1Lo = Bytes16.wrap(arg1.slice(16));
    final Bytes16 arg2Hi = Bytes16.wrap(arg2.slice(0, 16));
    final Bytes16 arg2Lo = Bytes16.wrap(arg2.slice(16));

    final int length = this.ctMax + 1;
    final int offset = LLARGE - length;
    final Bytes resHi = res.getHigh().slice(offset, length);
    final Bytes resLo = res.getLow().slice(offset, length);

    final UInt256 arg1Int = UInt256.fromBytes(arg1);
    final UInt256 arg2Int = UInt256.fromBytes(arg2);

    // set OverflowHi
    boolean overflowHi = false;

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

    // Set OverFlowLo
    Bytes32 addRes;
    if (opCode == OpCode.ADD) {
      addRes = Bytes32.wrap((UInt256.fromBytes(arg1Lo)).add(UInt256.fromBytes(arg2Lo)));
    } else {
      addRes = Bytes32.wrap((UInt256.fromBytes(resLo)).add(UInt256.fromBytes(arg2Lo)));
    }

    final boolean overflowLo = (addRes.compareTo(TWO_TO_THE_128) >= 0);
    for (int ct = 0; ct <= this.ctMax; ct++) {
      trace
          .acc1(resHi.slice(0, 1 + ct))
          .acc2(resLo.slice(0, 1 + ct))
          .arg1Hi(arg1Hi)
          .arg1Lo(arg1Lo)
          .arg2Hi(arg2Hi)
          .arg2Lo(arg2Lo)
          .byte1(UnsignedByte.of(resHi.get(ct)))
          .byte2(UnsignedByte.of(resLo.get(ct)))
          .ct(UnsignedByte.of(ct))
          .ctMax(UnsignedByte.of(this.ctMax))
          .inst(UnsignedByte.of(opCode.byteValue() & 0xff))
          .overflow(overflowBit(ct, overflowHi, overflowLo))
          .resHi(resHi)
          .resLo(resLo)
          .stamp(stamp)
          .validateRow();
    }
  }

  @Override
  protected int computeLineCount() {
    ctMax = computeCtMax();
    return ctMax + 1;
  }
}
