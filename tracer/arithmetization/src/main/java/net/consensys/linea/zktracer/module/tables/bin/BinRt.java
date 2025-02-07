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

package net.consensys.linea.zktracer.module.tables.bin;

import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

public class BinRt implements Module {
  @Override
  public String moduleKey() {
    return "BIN_REFERENCE_TABLE";
  }

  @Override
  public void commitTransactionBundle() {}

  @Override
  public void popTransactionBundle() {}

  @Override
  public int lineCount() {
    return 3 * 256 * 256 + 256; // 256*256 lines for AND, OR and XOR, and 256 lines for NOT
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    // AND
    UnsignedByte opCode = UnsignedByte.of(OpCode.AND.byteValue());

    for (short input1 = 0; input1 <= 255; input1++) {
      final Bytes input1Bytes = Bytes.of(input1);
      final UnsignedByte input1UByte = UnsignedByte.of(input1);

      for (short input2 = 0; input2 <= 255; input2++) {
        final Bytes input2Bytes = Bytes.of(input2);
        final UnsignedByte input2UByte = UnsignedByte.of(input2);

        final UnsignedByte result = UnsignedByte.of(input1Bytes.and(input2Bytes).get(0));
        trace
            .inst(opCode)
            .resultByte(result)
            .inputByte1(input1UByte)
            .inputByte2(input2UByte)
            .validateRow();
      }
    }

    // OR
    opCode = UnsignedByte.of(OpCode.OR.byteValue());

    for (short input1 = 0; input1 <= 255; input1++) {
      final Bytes input1Bytes = Bytes.of(input1);
      final UnsignedByte input1UByte = UnsignedByte.of(input1);

      for (short input2 = 0; input2 <= 255; input2++) {
        final Bytes input2Bytes = Bytes.of(input2);
        final UnsignedByte input2UByte = UnsignedByte.of(input2);

        final UnsignedByte result = UnsignedByte.of(input1Bytes.or(input2Bytes).get(0));
        trace
            .inst(opCode)
            .resultByte(result)
            .inputByte1(input1UByte)
            .inputByte2(input2UByte)
            .validateRow();
      }
    }

    // XOR
    opCode = UnsignedByte.of(OpCode.XOR.byteValue());

    for (short input1 = 0; input1 <= 255; input1++) {
      final Bytes input1Bytes = Bytes.of(input1);
      final UnsignedByte input1UByte = UnsignedByte.of(input1);

      for (short input2 = 0; input2 <= 255; input2++) {
        final Bytes input2Bytes = Bytes.of(input2);
        final UnsignedByte input2UByte = UnsignedByte.of(input2);

        final UnsignedByte result = UnsignedByte.of(input1Bytes.xor(input2Bytes).get(0));
        trace
            .inst(opCode)
            .resultByte(result)
            .inputByte1(input1UByte)
            .inputByte2(input2UByte)
            .validateRow();
      }
    }

    // NOT
    opCode = UnsignedByte.of(OpCode.NOT.byteValue());

    for (short input1 = 0; input1 <= 255; input1++) {
      final Bytes input1Bytes = Bytes.of(input1);
      final UnsignedByte input1UByte = UnsignedByte.of(input1);

      final UnsignedByte result = UnsignedByte.of(input1Bytes.not().get(0));
      trace
          .inst(opCode)
          .resultByte(result)
          .inputByte1(input1UByte)
          .inputByte2(UnsignedByte.ZERO)
          .validateRow();
    }
  }
}
