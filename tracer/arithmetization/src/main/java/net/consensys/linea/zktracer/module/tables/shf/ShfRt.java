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

package net.consensys.linea.zktracer.module.tables.shf;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import org.apache.tuweni.bytes.Bytes;

public record ShfRt() implements Module {
  @Override
  public String jsonKey() {
    return "shfRT";
  }

  @Override
  public void enterTransaction() {}

  @Override
  public void popTransaction() {}

  @Override
  public int lineCount() {
    return 256 * 9 + 1;
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    for (int a = 0; a <= 255; a++) {
      for (int uShp = 0; uShp <= 8; uShp++) {
        trace
            ._byte(BigInteger.valueOf(a))
            .las(BigInteger.valueOf(Bytes.of((byte) a).shiftLeft(8 - uShp).toInt())) // a<<(8-µShp)
            .mshp(BigInteger.valueOf(uShp))
            .rap(BigInteger.valueOf(Bytes.ofUnsignedShort(a).shiftRight(uShp).toInt()))
            .ones(BigInteger.valueOf((Bytes.fromHexString("0xFF").shiftRight(uShp)).not().toInt()))
            .isInRt(BigInteger.ONE)
            .validateRow();
      }
    }

    trace
        ._byte(BigInteger.ZERO)
        .isInRt(BigInteger.ZERO)
        .las(BigInteger.ZERO)
        .mshp(BigInteger.ZERO)
        .ones(BigInteger.ZERO)
        .rap(BigInteger.ZERO)
        .validateRow();
  }
}
