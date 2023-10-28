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
import java.util.ArrayList;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import org.apache.tuweni.bytes.Bytes;

public record ShfRtTrace(@JsonProperty("Trace") Trace trace) {
  public static ShfRtTrace generate() {
    final List<BigInteger> bytes = new ArrayList<>(2305);
    final List<BigInteger> isInRt = new ArrayList<>(2305);
    final List<BigInteger> las = new ArrayList<>(2305);
    final List<BigInteger> mshp = new ArrayList<>(2305);
    final List<BigInteger> ones = new ArrayList<>(2305);
    final List<BigInteger> rap = new ArrayList<>(2305);

    for (int a = 0; a <= 255; a++) {
      for (int uShp = 0; uShp <= 8; uShp++) {
        bytes.add(BigInteger.valueOf(a));
        // rt.Trace.PushByte(LAS.Name(), a<<(8-µShp))
        las.add(BigInteger.valueOf(Bytes.of((byte) a).shiftLeft(8 - uShp).toInt()));
        mshp.add(BigInteger.valueOf(uShp));
        rap.add(BigInteger.valueOf(Bytes.ofUnsignedShort(a).shiftRight(uShp).toInt()));
        ones.add(BigInteger.valueOf((Bytes.fromHexString("0xFF").shiftRight(uShp)).not().toInt()));
        isInRt.add(BigInteger.ONE);
      }
    }

    bytes.add(BigInteger.ZERO);
    isInRt.add(BigInteger.ZERO);
    las.add(BigInteger.ZERO);
    mshp.add(BigInteger.ZERO);
    ones.add(BigInteger.ZERO);
    rap.add(BigInteger.ZERO);

    return new ShfRtTrace(new Trace(bytes, isInRt, las, mshp, ones, rap));
  }
}
