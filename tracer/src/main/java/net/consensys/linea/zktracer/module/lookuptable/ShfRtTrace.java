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

package net.consensys.linea.zktracer.module.lookuptable;

import java.util.ArrayList;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import org.apache.tuweni.bytes.Bytes;

public record ShfRtTrace(@JsonProperty("Trace") Trace trace) {
  public static ShfRtTrace generate() {
    final List<Integer> bytes = new ArrayList<>(2305);
    final List<Integer> isInRt = new ArrayList<>(2305);
    final List<Integer> las = new ArrayList<>(2305);
    final List<Integer> mshp = new ArrayList<>(2305);
    final List<Integer> ones = new ArrayList<>(2305);
    final List<Integer> rap = new ArrayList<>(2305);

    for (int a = 0; a <= 255; a++) {
      for (int uShp = 0; uShp <= 8; uShp++) {
        bytes.add(a);
        // rt.Trace.PushByte(LAS.Name(), a<<(8-µShp))
        las.add((Bytes.of((byte) a).shiftLeft(8 - uShp)).toInt());
        mshp.add(uShp);
        rap.add(Bytes.ofUnsignedShort(a).shiftRight(uShp).toInt());
        ones.add((Bytes.fromHexString("0xFF").shiftRight(uShp)).not().toInt());
        isInRt.add(1);
      }
    }

    bytes.add(0);
    isInRt.add(0);
    las.add(0);
    mshp.add(0);
    ones.add(0);
    rap.add(0);

    return new ShfRtTrace(new Trace(bytes, isInRt, las, mshp, ones, rap));
  }
}
