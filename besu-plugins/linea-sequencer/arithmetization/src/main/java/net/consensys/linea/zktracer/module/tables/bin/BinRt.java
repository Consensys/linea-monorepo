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
import net.consensys.linea.zktracer.module.Module;
import org.apache.tuweni.bytes.Bytes;

public record BinRt() implements Module {
  @Override
  public String moduleKey() {
    return "binRT";
  }

  @Override
  public void enterTransaction() {}

  @Override
  public void popTransaction() {}

  @Override
  public int lineCount() {
    return 256 * 256 + 1;
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    for (short a = 0; a <= 255; a++) {
      final Bytes aByte = Bytes.of((byte) a);
      for (short b = 0; b <= 255; b++) {
        final Bytes bByte = Bytes.of((byte) b);

        trace
            .byteArg1(aByte)
            .byteArg2(bByte)
            .andByte(aByte.and(bByte))
            .orByte(aByte.or(bByte))
            .xorByte(aByte.xor(bByte))
            .notByte(aByte.not())
            .isInRt(Bytes.of(1))
            .validateRow();
      }
    }

    // zero row
    trace
        .byteArg1(Bytes.EMPTY)
        .byteArg2(Bytes.EMPTY)
        .andByte(Bytes.EMPTY)
        .orByte(Bytes.EMPTY)
        .xorByte(Bytes.EMPTY)
        .notByte(Bytes.EMPTY)
        .isInRt(Bytes.EMPTY)
        .validateRow();
  }
}
